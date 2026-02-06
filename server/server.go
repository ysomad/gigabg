package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/lobby"
)

type Server struct {
	store   LobbyStore
	clients map[string][]*ClientConn // lobbyID -> clients
	mu      sync.RWMutex
}

type ClientConn struct {
	playerID string
	lobbyID  string
	conn     *websocket.Conn
	send     chan []byte
}

func New(store LobbyStore) *Server {
	s := &Server{
		store:   store,
		clients: make(map[string][]*ClientConn),
	}
	go s.gameLoop()
	return s
}

// gameLoop runs periodically to advance phases in all lobbies.
func (s *Server) gameLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		lobbyIDs := make([]string, 0, len(s.clients))
		for id := range s.clients {
			lobbyIDs = append(lobbyIDs, id)
		}
		s.mu.RUnlock()

		for _, lobbyID := range lobbyIDs {
			l, err := s.store.Get(lobbyID)
			if err != nil {
				continue
			}

			if l.AdvancePhase() {
				slog.Info("phase changed",
					"lobby", lobbyID,
					"turn", l.Turn(),
					"phase", l.Phase().String(),
				)
				s.broadcastState(lobbyID, l)
				s.sendCombatAnimations(lobbyID, l)
			}
		}
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}

	client := &ClientConn{
		conn: conn,
		send: make(chan []byte, 256),
	}

	go s.writePump(r.Context(), client)
	s.readPump(r.Context(), client)
}

func (s *Server) writePump(ctx context.Context, client *ClientConn) {
	defer client.conn.CloseNow()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-client.send:
			if !ok {
				return
			}
			if err := client.conn.Write(ctx, websocket.MessageBinary, data); err != nil {
				slog.Error("write failed", "error", err)
				return
			}
		}
	}
}

func (s *Server) readPump(ctx context.Context, client *ClientConn) {
	defer func() {
		s.removeClient(client)
		close(client.send)
	}()

	for {
		_, data, err := client.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || errors.Is(err, io.EOF) {
				return
			}
			slog.Error("read failed", "error", err)
			return
		}

		var msg api.ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("decode failed", "error", err)
			continue
		}

		s.handleMessage(client, &msg)
	}
}

func decodePayload[T any](msg *api.ClientMessage) (T, error) {
	var v T
	if err := json.Unmarshal(msg.Payload, &v); err != nil {
		return v, err
	}
	return v, nil
}

func (s *Server) handleMessage(client *ClientConn, msg *api.ClientMessage) {
	switch msg.Action {
	case api.ActionJoinLobby:
		payload, err := decodePayload[api.JoinLobby](msg)
		if err != nil {
			s.sendError(client, "invalid payload")
			return
		}
		s.handleJoin(client, &payload)

	case api.ActionBuyCard:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.BuyCard](msg)
			if err != nil {
				return err
			}
			if err := p.BuyCard(payload.ShopIndex); err != nil {
				return err
			}
			p.CheckTriples()
			return nil
		})

	case api.ActionSellMinion:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.SellMinion](msg)
			if err != nil {
				return err
			}
			return p.SellMinion(payload.BoardIndex, l.Pool())
		})

	case api.ActionPlaceMinion:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.PlaceMinion](msg)
			if err != nil {
				return err
			}
			return p.PlaceMinion(payload.HandIndex, payload.BoardPosition, l.Cards())
		})

	case api.ActionRemoveMinion:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.RemoveMinion](msg)
			if err != nil {
				return err
			}
			return p.RemoveMinion(payload.BoardIndex)
		})

	case api.ActionUpgradeShop:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			return p.UpgradeShop()
		})

	case api.ActionRefreshShop:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			return p.RefreshShop(l.Pool())
		})

	case api.ActionPlaySpell:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.PlaySpell](msg)
			if err != nil {
				return err
			}
			return p.PlaySpell(payload.HandIndex, l.Pool())
		})

	case api.ActionDiscoverPick:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.DiscoverPick](msg)
			if err != nil {
				return err
			}
			return p.DiscoverPick(payload.Index, l.Pool())
		})

	case api.ActionFreezeShop:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			p.FreezeShop()
			return nil
		})

	case api.ActionSyncBoard:
		s.handleAction(client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.SyncBoard](msg)
			if err != nil {
				return err
			}
			return p.ReorderBoard(payload.Order)
		})
	}
}

func (s *Server) handleAction(
	client *ClientConn,
	action api.Action,
	fn func(l *lobby.Lobby, p *game.Player) error,
) {
	if client.lobbyID == "" {
		s.sendError(client, "not in lobby")
		return
	}

	l, err := s.store.Get(client.lobbyID)
	if err != nil {
		s.sendError(client, err.Error())
		return
	}

	if l.State() != lobby.StatePlaying || l.Phase() != game.PhaseRecruit {
		return
	}

	p := l.Player(client.playerID)
	if p == nil {
		s.sendError(client, "player not found")
		return
	}

	beforeGold := p.Gold()
	beforeHP := p.HP()
	beforeShopTier := p.Shop().Tier()
	beforeBoard := p.BoardSize()
	beforeHand := p.HandSize()
	beforeShop := len(p.Shop().Cards())

	if err := fn(l, p); err != nil {
		s.sendError(client, err.Error())
		return
	}

	attrs := []slog.Attr{
		slog.String("player", client.playerID),
		slog.String("lobby", client.lobbyID),
	}
	if d := p.Gold() - beforeGold; d != 0 {
		attrs = append(attrs, slog.Int("gold", d))
	}
	if d := p.HP() - beforeHP; d != 0 {
		attrs = append(attrs, slog.Int("hp", d))
	}
	if d := int(p.Shop().Tier()) - int(beforeShopTier); d != 0 {
		attrs = append(attrs, slog.Int("shop_tier", d))
	}
	if d := p.BoardSize() - beforeBoard; d != 0 {
		attrs = append(attrs, slog.Int("board", d))
	}
	if d := p.HandSize() - beforeHand; d != 0 {
		attrs = append(attrs, slog.Int("hand", d))
	}
	if d := len(p.Shop().Cards()) - beforeShop; d != 0 {
		attrs = append(attrs, slog.Int("shop", d))
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo, action.String(), attrs...)

	s.sendPlayerState(client, l, p)
}

func (s *Server) handleJoin(client *ClientConn, join *api.JoinLobby) {
	lobbyID := join.LobbyID
	if lobbyID == "" {
		s.sendError(client, "lobby_id required")
		return
	}

	// Auto-create lobby if it doesn't exist
	if _, err := s.store.Get(lobbyID); errors.Is(err, ErrLobbyNotFound) {
		if err := s.store.Create(lobbyID, nil); err != nil {
			s.sendError(client, err.Error())
			return
		}

		slog.Info("lobby created", "lobby", lobbyID)
	}

	l, err := s.store.Get(lobbyID)
	if err != nil {
		s.sendError(client, err.Error())
		return
	}

	s.mu.Lock()
	clients := s.clients[lobbyID]
	playerID := itoa(len(clients) + 1)

	// Add player to lobby
	if err := l.AddPlayer(playerID); err != nil {
		s.mu.Unlock()
		s.sendError(client, err.Error())
		return
	}

	client.playerID = playerID
	client.lobbyID = lobbyID
	s.clients[lobbyID] = append(clients, client)
	s.mu.Unlock()

	slog.Info("player joined",
		"player", playerID,
		"lobby", lobbyID,
		"lobby_players", l.PlayerCount(),
		"lobby_max_players", game.MaxPlayers,
	)

	// Check if game started (all players joined)
	if l.State() == lobby.StatePlaying {
		slog.Info("game started",
			"lobby", lobbyID,
			"turn", l.Turn(),
			"phase", l.Phase().String(),
		)
	}

	// Broadcast state to all clients in lobby
	s.broadcastState(lobbyID, l)
}

func (s *Server) sendCombatAnimations(lobbyID string, l *lobby.Lobby) {
	anims := l.CombatAnimations()
	if len(anims) == 0 {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := s.clients[lobbyID]
	for _, anim := range anims {
		for _, c := range clients {
			if c.playerID != anim.Player1ID && c.playerID != anim.Player2ID {
				continue
			}
			s.sendMessage(c, &api.ServerMessage{CombatEvents: anim.Events})
		}
	}
}

func (s *Server) broadcastState(lobbyID string, l *lobby.Lobby) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.clients[lobbyID] {
		p := l.Player(c.playerID)
		if p == nil {
			continue
		}
		s.sendPlayerState(c, l, p)
	}
}

func (s *Server) sendPlayerState(client *ClientConn, l *lobby.Lobby, p *game.Player) {
	state := &api.GameState{
		Player:            api.NewPlayer(p),
		Opponents:         api.NewOpponents(l.Players(), client.playerID),
		Turn:              l.Turn(),
		Phase:             l.Phase(),
		PhaseEndTimestamp: l.PhaseEndTimestamp(),
		Shop:              api.NewCards(p.Shop().Cards()),
		IsShopFrozen:      p.Shop().IsFrozen(),
		Hand:              api.NewCards(p.Hand()),
		Board:             api.NewCardsFromMinions(p.Board().Minions()),
		CombatResults:     l.CombatResults(client.playerID),
	}

	if p.HasDiscover() {
		state.Discover = api.NewCards(p.DiscoverOptions())
	}

	if l.Phase() == game.PhaseCombat {
		if opponentID, playerBoard, opponentBoard, ok := l.CombatPairing(client.playerID); ok {
			state.OpponentID = opponentID
			state.CombatBoard = api.CombatCards(playerBoard)
			state.OpponentBoard = api.CombatCards(opponentBoard)
		}
	}

	s.sendMessage(client, &api.ServerMessage{State: state})
}

func (s *Server) sendError(client *ClientConn, msg string) {
	resp := &api.ServerMessage{
		Error: &api.Error{Message: msg},
	}
	s.sendMessage(client, resp)
}

func (s *Server) sendMessage(client *ClientConn, msg *api.ServerMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("encode failed", "error", err)
		return
	}

	select {
	case client.send <- data:
	default:
		slog.Warn("send buffer full", "player", client.playerID)
	}
}

func (s *Server) removeClient(client *ClientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client.lobbyID == "" {
		return
	}

	slog.Info("player disconnected",
		"player", client.playerID,
		"lobby", client.lobbyID,
	)

	clients := s.clients[client.lobbyID]
	for i, c := range clients {
		if c == client {
			s.clients[client.lobbyID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}

	if len(s.clients[client.lobbyID]) == 0 {
		delete(s.clients, client.lobbyID)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
