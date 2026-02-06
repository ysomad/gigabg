package client

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/ysomad/gigabg/message"
)

var (
	ErrNotConnected = errors.New("not connected to server")
	ErrNotJoined    = errors.New("not joined to lobby")
)

// RemoteClient connects to a game server via WebSocket.
type RemoteClient struct {
	conn *websocket.Conn

	state *message.GameState
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
	if err := c.sendMessage(&message.ClientMessage{
		Join: &message.JoinLobby{LobbyID: lobbyID},
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

		var msg message.ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *RemoteClient) handleMessage(msg *message.ServerMessage) {
	if msg.State != nil {
		c.mu.Lock()
		c.state = msg.State
		c.mu.Unlock()
	}

	if msg.Error != nil {
		select {
		case c.errCh <- errors.New(msg.Error.Message):
		default:
		}
	}
}

func (c *RemoteClient) sendMessage(msg *message.ClientMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.Write(c.ctx, websocket.MessageBinary, data)
}

// State returns the current game state.
func (c *RemoteClient) State() *message.GameState {
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
	return c.state.PlayerID
}

// Players returns all players in the lobby.
func (c *RemoteClient) Players() []message.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Players
}

// Turn returns the current turn number.
func (c *RemoteClient) Turn() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.Turn
}

// Phase returns the current phase.
func (c *RemoteClient) Phase() message.Phase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return message.PhaseWaiting
	}
	return c.state.Phase
}

// PhaseEndTimestamp returns when the current phase ends (unix seconds).
func (c *RemoteClient) PhaseEndTimestamp() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.PhaseEndTimestamp
}

// Shop returns the current shop cards.
func (c *RemoteClient) Shop() []message.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Shop
}

// Hand returns the current hand cards.
func (c *RemoteClient) Hand() []message.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Hand
}

// Board returns the current board cards.
func (c *RemoteClient) Board() []message.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Board
}

// Player returns the current player's info.
func (c *RemoteClient) Player() *message.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	for i := range c.state.Players {
		if c.state.Players[i].ID == c.state.PlayerID {
			return &c.state.Players[i]
		}
	}
	return nil
}

// BuyCard sends a buy card action.
func (c *RemoteClient) BuyCard(shopIndex int) error {
	return c.sendMessage(&message.ClientMessage{
		Buy: &message.BuyCard{ShopIndex: shopIndex},
	})
}

// SellCard sends a sell card action.
func (c *RemoteClient) SellCard(handIndex int) error {
	return c.sendMessage(&message.ClientMessage{
		SellCard: &message.SellCard{HandIndex: handIndex},
	})
}

// PlaceMinion sends a place minion action.
func (c *RemoteClient) PlaceMinion(handIndex, boardPosition int) error {
	return c.sendMessage(&message.ClientMessage{
		PlaceMinion: &message.PlaceMinion{
			HandIndex:     handIndex,
			BoardPosition: boardPosition,
		},
	})
}

// RemoveMinion sends a remove minion action.
func (c *RemoteClient) RemoveMinion(boardIndex int) error {
	return c.sendMessage(&message.ClientMessage{
		RemoveMinion: &message.RemoveMinion{BoardIndex: boardIndex},
	})
}

// UpgradeShop sends an upgrade shop action.
func (c *RemoteClient) UpgradeShop() error {
	return c.sendMessage(&message.ClientMessage{
		UpgradeShop: &message.UpgradeShop{},
	})
}

// RefreshShop sends a refresh shop action.
func (c *RemoteClient) RefreshShop() error {
	return c.sendMessage(&message.ClientMessage{
		RefreshShop: &message.RefreshShop{},
	})
}

// SyncBoard sends the board order to server.
func (c *RemoteClient) SyncBoard(order []int) error {
	return c.sendMessage(&message.ClientMessage{
		SyncBoard: &message.SyncBoard{Order: order},
	})
}

// Discover returns the current discover offer, or nil if none pending.
func (c *RemoteClient) Discover() *message.DiscoverOffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Discover
}

// PlaySpell sends a play spell action.
func (c *RemoteClient) PlaySpell(handIndex int) error {
	return c.sendMessage(&message.ClientMessage{
		PlaySpell: &message.PlaySpell{HandIndex: handIndex},
	})
}

// DiscoverPick sends a discover pick action.
func (c *RemoteClient) DiscoverPick(index int) error {
	return c.sendMessage(&message.ClientMessage{
		DiscoverPick: &message.DiscoverPick{Index: index},
	})
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
