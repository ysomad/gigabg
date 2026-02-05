package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	"github.com/ysomad/gigabg/lobby"
	pb "github.com/ysomad/gigabg/proto"
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
	return &Server{
		store:   store,
		clients: make(map[string][]*ClientConn),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Printf("websocket accept error: %v", err)
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
				log.Printf("write error: %v", err)
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
			if websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Printf("read error: %v", err)
			}
			return
		}

		var msg pb.ClientMessage
		if err := proto.Unmarshal(data, &msg); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		s.handleMessage(client, &msg)
	}
}

func (s *Server) handleMessage(client *ClientConn, msg *pb.ClientMessage) {
	switch m := msg.Msg.(type) {
	case *pb.ClientMessage_Join:
		s.handleJoin(client, m.Join)
	}
}

func (s *Server) handleJoin(client *ClientConn, join *pb.JoinLobby) {
	lobbyID := join.LobbyId
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

	log.Printf("player %s joined lobby %s (%d/%d)", playerID, lobbyID, l.PlayerCount(), lobby.MaxPlayers)

	// Broadcast state to all clients in lobby
	s.broadcastState(lobbyID, l)
}

func (s *Server) broadcastState(lobbyID string, l *lobby.Lobby) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Build players list
	players := make([]*pb.Player, 0, l.PlayerCount())
	for _, p := range l.Players() {
		players = append(players, &pb.Player{
			Id: p.ID,
			Hp: int32(p.HP),
		})
	}

	// Send to each client with their player_id
	for _, c := range s.clients[lobbyID] {
		resp := &pb.ServerMessage{
			Msg: &pb.ServerMessage_State{
				State: &pb.GameState{
					PlayerId:    c.playerID,
					PlayerCount: int32(l.PlayerCount()),
					Players:     players,
				},
			},
		}
		s.sendMessage(c, resp)
	}
}

func (s *Server) sendError(client *ClientConn, message string) {
	resp := &pb.ServerMessage{
		Msg: &pb.ServerMessage_Error{
			Error: &pb.Error{
				Message: message,
			},
		},
	}
	s.sendMessage(client, resp)
}

func (s *Server) sendMessage(client *ClientConn, msg *pb.ServerMessage) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		log.Printf("client %s send buffer full", client.playerID)
	}
}

func (s *Server) removeClient(client *ClientConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client.lobbyID == "" {
		return
	}

	log.Printf("player %s disconnected from lobby %s", client.playerID, client.lobbyID)

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
