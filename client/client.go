package client

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/api"
)

var (
	ErrNotConnected = errors.New("not connected to server")
	ErrNotJoined    = errors.New("not joined to lobby")
)

// Client connects to a game server via WebSocket.
type Client struct {
	conn *websocket.Conn

	state *api.GameState
	mu    sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	errCh chan error
}

// New connects to a game server and joins the specified lobby.
func New(ctx context.Context, serverAddr, lobbyID string) (*Client, error) {
	conn, _, err := websocket.Dial(ctx, serverAddr, nil)
	if err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithCancel(context.Background())
	c := &Client{
		conn:   conn,
		ctx:    connCtx,
		cancel: cancel,
		errCh:  make(chan error, 1),
	}

	go c.readPump()

	// Send join message
	if err := c.send(api.ActionJoinLobby, api.JoinLobby{LobbyID: lobbyID}); err != nil {
		conn.CloseNow()
		return nil, err
	}

	return c, nil
}

func (c *Client) readPump() {
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

		var msg api.ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *Client) handleMessage(msg *api.ServerMessage) {
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

func (c *Client) send(action api.Action, payload any) error {
	msg := api.ClientMessage{Action: action}

	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		msg.Payload = raw
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.Write(c.ctx, websocket.MessageBinary, data)
}

// State returns the current game state.
func (c *Client) State() *api.GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// PlayerID returns this client's player ID.
func (c *Client) PlayerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return ""
	}
	return c.state.PlayerID
}

// Players returns all players in the lobby.
func (c *Client) Players() []api.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Players
}

// Turn returns the current turn number.
func (c *Client) Turn() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.Turn
}

// Phase returns the current phase.
func (c *Client) Phase() game.Phase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return game.PhaseWaiting
	}
	return c.state.Phase
}

// PhaseEndTimestamp returns when the current phase ends (unix seconds).
func (c *Client) PhaseEndTimestamp() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.PhaseEndTimestamp
}

// Shop returns the current shop cards.
func (c *Client) Shop() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Shop
}

// Hand returns the current hand cards.
func (c *Client) Hand() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Hand
}

// Board returns the current board cards.
func (c *Client) Board() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Board
}

// Player returns the current player's info.
func (c *Client) Player() *api.Player {
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
func (c *Client) BuyCard(shopIndex int) error {
	return c.send(api.ActionBuyCard, api.BuyCard{ShopIndex: shopIndex})
}

// SellCard sends a sell card action.
func (c *Client) SellCard(handIndex int) error {
	return c.send(api.ActionSellCard, api.SellCard{HandIndex: handIndex})
}

// PlaceMinion sends a place minion action.
func (c *Client) PlaceMinion(handIndex, boardPosition int) error {
	return c.send(api.ActionPlaceMinion, api.PlaceMinion{
		HandIndex:     handIndex,
		BoardPosition: boardPosition,
	})
}

// RemoveMinion sends a remove minion action.
func (c *Client) RemoveMinion(boardIndex int) error {
	return c.send(api.ActionRemoveMinion, api.RemoveMinion{BoardIndex: boardIndex})
}

// UpgradeShop sends an upgrade shop action.
func (c *Client) UpgradeShop() error {
	return c.send(api.ActionUpgradeShop, nil)
}

// RefreshShop sends a refresh shop action.
func (c *Client) RefreshShop() error {
	return c.send(api.ActionRefreshShop, nil)
}

// SyncBoard sends the board order to server.
func (c *Client) SyncBoard(order []int) error {
	return c.send(api.ActionSyncBoard, api.SyncBoard{Order: order})
}

// Discover returns the current discover offer, or nil if none pending.
func (c *Client) Discover() *api.DiscoverOffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Discover
}

// PlaySpell sends a play spell action.
func (c *Client) PlaySpell(handIndex int) error {
	return c.send(api.ActionPlaySpell, api.PlaySpell{HandIndex: handIndex})
}

// DiscoverPick sends a discover pick action.
func (c *Client) DiscoverPick(index int) error {
	return c.send(api.ActionDiscoverPick, api.DiscoverPick{Index: index})
}

// Close closes the connection.
func (c *Client) Close() error {
	c.cancel()
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

// Connected returns true if the client has received state.
func (c *Client) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state != nil
}

// WaitForState blocks until state is received or context is cancelled.
func (c *Client) WaitForState(ctx context.Context) error {
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
