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

var _ http.Handler = (*Server)(nil)

type Server struct {
	mux     *http.ServeMux
	store   *lobby.MemoryStore
	cards   game.CardCatalog
	clients map[string][]*ClientConn // lobbyID -> clients
	mu      sync.RWMutex
}

type ClientConn struct {
	playerID string
	lobbyID  string
	conn     *websocket.Conn
	send     chan []byte
}

func New(store *lobby.MemoryStore, cards game.CardCatalog) *Server {
	s := &Server{
		store:   store,
		cards:   cards,
		clients: make(map[string][]*ClientConn),
		mux:     http.NewServeMux(),
	}

	s.mux.HandleFunc("POST /lobbies", s.createLobby)
	s.mux.HandleFunc("/ws", s.handleWS)

	go s.gameLoop()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) createLobby(w http.ResponseWriter, r *http.Request) {
	var req api.CreateLobbyReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	l, err := lobby.New(s.cards, req.MaxPlayers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.CreateLobby(l); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("lobby created", "lobby", l.ID(), "max_players", req.MaxPlayers)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(api.CreateLobbyResp{LobbyID: l.ID()}); err != nil {
		slog.Error("encode failed", "error", err)
	}
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
			l, err := s.store.Lobby(lobbyID)
			if err != nil {
				continue
			}

			if l.AdvancePhase() {
				slog.Info("phase changed",
					"lobby", lobbyID,
					"turn", l.Turn(),
					"phase", l.Phase().String(),
				)
				s.sendCombatLogs(lobbyID, l)
				s.broadcastState(lobbyID, l)

				if l.State() == lobby.StateFinished {
					if err := s.store.DeleteLobby(lobbyID); err != nil {
						slog.Error("delete lobby", "error", err, "lobby", lobbyID)
					}

					s.mu.Lock()
					delete(s.clients, lobbyID)
					s.mu.Unlock()

					slog.Info("game finished, lobby removed", "lobby", lobbyID)
				}
			}
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	lobbyID := r.URL.Query().Get("lobby")
	playerID := r.URL.Query().Get("player")

	slog.Info("ws connect", "player", playerID, "lobby", lobbyID, "remote", r.RemoteAddr)

	if lobbyID == "" {
		slog.Info("ws rejected, missing lobby param", "remote", r.RemoteAddr)
		http.Error(w, "lobby query param required", http.StatusBadRequest)
		return
	}

	if playerID == "" {
		slog.Info("ws rejected, missing player param", "remote", r.RemoteAddr)
		http.Error(w, "player query param required", http.StatusBadRequest)
		return
	}

	l, err := s.store.Lobby(lobbyID)
	if err != nil {
		slog.Info("ws rejected, lobby not found", "lobby", lobbyID, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

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

	// Join lobby on connect.
	s.mu.Lock()

	if err := l.AddPlayer(playerID); err != nil {
		s.mu.Unlock()
		if cerr := conn.Close(websocket.StatusPolicyViolation, err.Error()); cerr != nil {
			slog.Error("close rejected player conn", "error", cerr, "player", playerID)
		}
		return
	}

	client.playerID = playerID
	client.lobbyID = lobbyID
	s.clients[lobbyID] = append(s.clients[lobbyID], client)
	s.mu.Unlock()

	slog.Info("player joined",
		"player", playerID,
		"lobby", lobbyID,
		"lobby_players", l.PlayerCount(),
		"lobby_max_players", l.MaxPlayers(),
	)

	if l.State() == lobby.StatePlaying {
		slog.Info("game started",
			"lobby", lobbyID,
			"turn", l.Turn(),
			"phase", l.Phase().String(),
		)
	}

	s.broadcastState(lobbyID, l)

	go s.writePump(r.Context(), client)
	s.readPump(r.Context(), client)
}

func (s *Server) writePump(ctx context.Context, client *ClientConn) {
	defer client.conn.CloseNow() //nolint:errcheck // best-effort cleanup

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

		s.handleMessage(ctx, client, &msg)
	}
}

func decodePayload[T any](msg *api.ClientMessage) (T, error) {
	var v T
	if err := json.Unmarshal(msg.Payload, &v); err != nil {
		return v, err
	}
	return v, nil
}

func (s *Server) handleMessage(ctx context.Context, client *ClientConn, msg *api.ClientMessage) {
	switch msg.Action {
	case api.ActionBuyCard:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
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
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.SellMinion](msg)
			if err != nil {
				return err
			}
			return p.SellMinion(payload.BoardIndex, l.Pool())
		})

	case api.ActionPlaceMinion:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.PlaceMinion](msg)
			if err != nil {
				return err
			}
			return p.PlaceMinion(payload.HandIndex, payload.BoardPosition, l.Pool())
		})

	case api.ActionRemoveMinion:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.RemoveMinion](msg)
			if err != nil {
				return err
			}
			return p.RemoveMinion(payload.BoardIndex)
		})

	case api.ActionUpgradeShop:
		s.handleUpgradeShop(ctx, client, msg.Action)

	case api.ActionRefreshShop:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			return p.RefreshShop(l.Pool())
		})

	case api.ActionPlaySpell:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.PlaySpell](msg)
			if err != nil {
				return err
			}
			return p.PlaySpell(payload.HandIndex, l.Pool())
		})

	case api.ActionDiscoverPick:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.DiscoverPick](msg)
			if err != nil {
				return err
			}
			return p.DiscoverPick(payload.Index, l.Pool())
		})

	case api.ActionFreezeShop:
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			p.FreezeShop()
			return nil
		})

	case api.ActionSyncBoards:
		slog.Info(msg.Action.String(), "player", client.playerID, "lobby", client.lobbyID)
		s.handleAction(ctx, client, msg.Action, func(l *lobby.Lobby, p *game.Player) error {
			payload, err := decodePayload[api.SyncBoards](msg)
			if err != nil {
				return err
			}
			if err := p.ReorderBoard(payload.BoardOrder); err != nil {
				return err
			}
			return p.ReorderShop(payload.ShopOrder)
		})
	}
}

func (s *Server) handleAction(
	ctx context.Context,
	client *ClientConn,
	action api.Action,
	fn func(l *lobby.Lobby, p *game.Player) error,
) {
	if client.lobbyID == "" {
		s.sendError(client, "not in lobby")
		return
	}

	l, err := s.store.Lobby(client.lobbyID)
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
	slog.LogAttrs(ctx, slog.LevelInfo, action.String(), attrs...)

	s.sendPlayerState(client, l, p)
}

func (s *Server) sendCombatLogs(lobbyID string, l *lobby.Lobby) {
	anims := l.CombatLogs()
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

			events, err := api.NewCombatEvents(anim.Events)
			if err != nil {
				slog.Error("failed to marshal combat events", "player", c.playerID, "err", err)
				continue
			}

			s.sendMessage(c, &api.ServerMessage{CombatEvents: events})
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
		Player: api.NewPlayer(p),
		Opponents: api.NewOpponents(
			l.Players(),
			client.playerID,
			api.NewAllCombatResults(l.AllCombatResults()),
			l.MajorityTribes(),
		),
		Turn:          l.Turn(),
		Phase:         l.Phase(),
		PhaseEndsAt:   l.PhaseEndsAt(),
		Shop:          api.NewCards(p.Shop().Cards()),
		IsShopFrozen:  p.Shop().IsFrozen(),
		Hand:          api.NewCards(p.Hand()),
		Board:         api.NewCardsFromMinions(p.Board().Minions()),
		CombatResults: api.NewCombatResults(l.CombatResults(client.playerID)),
	}

	if p.HasDiscovers() {
		state.Discovers = api.NewCards(p.Discovers())
	}

	switch l.Phase() {
	case game.PhaseRecruit:
		state.OpponentID = l.NextOpponentID(client.playerID)
	case game.PhaseCombat, game.PhaseFinished:
		if pair, ok := l.CombatPairing(client.playerID); ok {
			state.OpponentID = pair.OpponentID
			state.CombatBoard = api.CombatCards(pair.PlayerBoard)
			state.OpponentBoard = api.CombatCards(pair.OpponentBoard)
		}
	}

	if l.Phase() == game.PhaseFinished {
		state.GameResult = api.NewGameResult(l.GameResult())
	}

	s.sendMessage(client, &api.ServerMessage{State: state})
}

func (s *Server) handleUpgradeShop(ctx context.Context, client *ClientConn, action api.Action) {
	s.handleAction(ctx, client, action, func(l *lobby.Lobby, p *game.Player) error {
		if err := p.UpgradeShop(); err != nil {
			return err
		}
		s.sendOpponentUpdate(client.lobbyID, client.playerID, p.Shop().Tier())
		return nil
	})
}

func (s *Server) sendOpponentUpdate(lobbyID, playerID string, tier game.Tier) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg := &api.ServerMessage{
		OpponentUpdate: &api.OpponentUpdate{
			PlayerID: playerID,
			ShopTier: tier,
		},
	}
	for _, c := range s.clients[lobbyID] {
		if c.playerID == playerID {
			continue
		}
		s.sendMessage(c, msg)
	}
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
