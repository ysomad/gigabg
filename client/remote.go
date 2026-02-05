package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/ysomad/gigabg/proto"
)

var (
	ErrNotConnected = errors.New("not connected to server")
	ErrNotJoined    = errors.New("not joined to lobby")
)

// RemoteClient connects to a game server via WebSocket.
type RemoteClient struct {
	conn *websocket.Conn

	state *pb.GameState
	mu    sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	errCh chan error
}

// NewRemote connects to a game server and joins the specified lobby.
func NewRemote(ctx context.Context, serverAddr, lobbyID string) (*RemoteClient, error) {
	conn, _, err := websocket.Dial(ctx, serverAddr, nil)
	if err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithCancel(context.Background())
	c := &RemoteClient{
		conn:   conn,
		ctx:    connCtx,
		cancel: cancel,
		errCh:  make(chan error, 1),
	}

	go c.readPump()

	// Send join message
	if err := c.sendMessage(&pb.ClientMessage{
		Msg: &pb.ClientMessage_Join{
			Join: &pb.JoinLobby{LobbyId: lobbyID},
		},
	}); err != nil {
		conn.CloseNow()
		return nil, err
	}

	return c, nil
}

func (c *RemoteClient) readPump() {
	defer c.conn.CloseNow()

	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			select {
			case c.errCh <- err:
			default:
			}
			return
		}

		var msg pb.ServerMessage
		if err := proto.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *RemoteClient) handleMessage(msg *pb.ServerMessage) {
	switch m := msg.Msg.(type) {
	case *pb.ServerMessage_State:
		c.mu.Lock()
		c.state = m.State
		c.mu.Unlock()

	case *pb.ServerMessage_Error:
		select {
		case c.errCh <- errors.New(m.Error.Message):
		default:
		}
	}
}

func (c *RemoteClient) sendMessage(msg *pb.ClientMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.Write(c.ctx, websocket.MessageBinary, data)
}

// State returns the current game state.
func (c *RemoteClient) State() *pb.GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// PlayerID returns this client's player ID.
func (c *RemoteClient) PlayerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return ""
	}
	return c.state.PlayerId
}

// PlayerCount returns the number of players in the lobby.
func (c *RemoteClient) PlayerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return int(c.state.PlayerCount)
}

// Players returns all players in the lobby.
func (c *RemoteClient) Players() []*pb.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Players
}

// Close closes the connection.
func (c *RemoteClient) Close() error {
	c.cancel()
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

// Connected returns true if the client has received state.
func (c *RemoteClient) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state != nil
}

// WaitForState blocks until state is received or context is cancelled.
func (c *RemoteClient) WaitForState(ctx context.Context) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		c.mu.RLock()
		hasState := c.state != nil
		c.mu.RUnlock()

		if hasState {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-c.errCh:
			return err
		case <-ticker.C:
		}
	}
}
