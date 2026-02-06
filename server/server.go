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

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/lobby"
	"github.com/ysomad/gigabg/message"
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

		var msg message.ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("decode failed", "error", err)
			continue
		}

		s.handleMessage(client, &msg)
	}
}

func (s *Server) handleMessage(client *ClientConn, msg *message.ClientMessage) {
	switch {
	case msg.Join != nil:
		s.handleJoin(client, msg.Join)
	case msg.Buy != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			if err := p.BuyCard(msg.Buy.ShopIndex); err != nil {
				return err
			}
			p.CheckTriples()
			return nil
		})
	case msg.SellCard != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.SellCard(msg.SellCard.HandIndex, l.Pool())
		})
	case msg.PlaceMinion != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.PlaceMinion(msg.PlaceMinion.HandIndex, msg.PlaceMinion.BoardPosition, l.Cards())
		})
	case msg.RemoveMinion != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.RemoveMinion(msg.RemoveMinion.BoardIndex)
		})
	case msg.UpgradeShop != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.UpgradeShop()
		})
	case msg.RefreshShop != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.RefreshShop(l.Pool())
		})
	case msg.PlaySpell != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.PlaySpell(msg.PlaySpell.HandIndex, l.Pool())
		})
	case msg.DiscoverPick != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.DiscoverPick(msg.DiscoverPick.Index, l.Pool())
		})
	case msg.SyncBoard != nil:
		s.handleAction(client, msg, func(l *lobby.Lobby, p *game.Player) error {
			return p.ReorderBoard(msg.SyncBoard.Order)
		})
	}
}

func (s *Server) handleAction(
	client *ClientConn,
	msg *message.ClientMessage,
	action func(l *lobby.Lobby, p *game.Player) error,
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
		s.sendError(client, "not in recruit phase")
		return
	}

	p := l.Player(client.playerID)
	if p == nil {
		s.sendError(client, "player not found")
		return
	}

	beforeGold := p.Gold
	beforeHP := p.HP
	beforeShopTier := p.ShopTier
	beforeBoard := len(p.Board)
	beforeHand := len(p.Hand)
	beforeShop := len(p.Shop)

	if err := action(l, p); err != nil {
		s.sendError(client, err.Error())
		return
	}

	attrs := []slog.Attr{
		slog.String("player", client.playerID),
		slog.String("lobby", client.lobbyID),
	}
	if d := p.Gold - beforeGold; d != 0 {
		attrs = append(attrs, slog.Int("gold", d))
	}
	if d := p.HP - beforeHP; d != 0 {
		attrs = append(attrs, slog.Int("hp", d))
	}
	if d := int(p.ShopTier) - int(beforeShopTier); d != 0 {
		attrs = append(attrs, slog.Int("shop_tier", d))
	}
	if d := len(p.Board) - beforeBoard; d != 0 {
		attrs = append(attrs, slog.Int("board", d))
	}
	if d := len(p.Hand) - beforeHand; d != 0 {
		attrs = append(attrs, slog.Int("hand", d))
	}
	if d := len(p.Shop) - beforeShop; d != 0 {
		attrs = append(attrs, slog.Int("shop", d))
	}
	slog.LogAttrs(context.Background(), slog.LevelInfo, msg.Action(), attrs...)

	s.sendPlayerState(client, l, p)
}

func (s *Server) handleJoin(client *ClientConn, join *message.JoinLobby) {
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
	phase := message.PhaseWaiting
	if l.State() == lobby.StatePlaying {
		switch l.Phase() {
		case game.PhaseRecruit:
			phase = message.PhaseRecruit
		case game.PhaseCombat:
			phase = message.PhaseCombat
		}
	}

	state := &message.GameState{
		PlayerID:          client.playerID,
		Turn:              l.Turn(),
		Phase:             phase,
		PhaseEndTimestamp: l.PhaseEndTimestamp(),
		Players:           message.NewPlayers(l.Players()),
		Shop:              message.NewCards(p.Shop),
		Hand:              message.NewCards(p.Hand),
		Board:             message.NewCardsFromMinions(p.Board),
	}

	if len(p.DiscoverOptions) > 0 {
		state.Discover = &message.DiscoverOffer{
			Cards: message.NewCards(p.DiscoverOptions),
		}
	}

	s.sendMessage(client, &message.ServerMessage{State: state})
}

func (s *Server) sendError(client *ClientConn, msg string) {
	resp := &message.ServerMessage{
		Error: &message.Error{Message: msg},
	}
	s.sendMessage(client, resp)
}

func (s *Server) sendMessage(client *ClientConn, msg *message.ServerMessage) {
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
