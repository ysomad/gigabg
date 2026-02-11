package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
)

// PlayerEntry is a player summary for the player bar.
type PlayerEntry struct {
	ID       string
	HP       int
	ShopTier game.Tier
}

// GameConn connects to a game server via WebSocket.
type GameConn struct {
	conn *websocket.Conn

	state        *api.GameState
	combatEvents []game.CombatEvent
	mu           sync.RWMutex

	errCh chan error
}

func newGameConn(conn *websocket.Conn) *GameConn {
	c := &GameConn{
		conn:  conn,
		errCh: make(chan error, 1),
	}
	go c.readPump()
	return c
}

func (c *GameConn) readPump() {
	defer c.conn.CloseNow() //nolint:errcheck // best-effort cleanup

	for {
		_, data, err := c.conn.Read(context.Background())
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure || errors.Is(err, io.EOF) {
				return
			}
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

func (c *GameConn) handleMessage(msg *api.ServerMessage) {
	c.mu.Lock()
	if len(msg.CombatEvents) > 0 {
		c.combatEvents = msg.CombatEvents
	}
	if msg.State != nil {
		c.state = msg.State
	}
	c.mu.Unlock()

	if msg.Error != nil {
		select {
		case c.errCh <- errors.New(msg.Error.Message):
		default:
		}
	}
}

func (c *GameConn) send(action api.Action, payload any) error {
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
	return c.conn.Write(context.Background(), websocket.MessageBinary, data)
}

// State returns the current game state.
func (c *GameConn) State() *api.GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// PlayerID returns this client's player ID.
func (c *GameConn) PlayerID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return ""
	}
	return c.state.Player.ID
}

// Opponents returns all opponents in the lobby.
func (c *GameConn) Opponents() []api.Opponent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Opponents
}

// Turn returns the current turn number.
func (c *GameConn) Turn() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.Turn
}

// Phase returns the current phase.
func (c *GameConn) Phase() game.Phase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return game.PhaseWaiting
	}
	return c.state.Phase
}

// PhaseEndTimestamp returns when the current phase ends (unix seconds).
func (c *GameConn) PhaseEndTimestamp() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.PhaseEndTimestamp
}

// Shop returns the current shop cards.
func (c *GameConn) Shop() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Shop
}

// IsShopFrozen returns whether the shop is frozen.
func (c *GameConn) IsShopFrozen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return false
	}
	return c.state.IsShopFrozen
}

// Hand returns the current hand cards.
func (c *GameConn) Hand() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Hand
}

// Board returns the current board cards.
func (c *GameConn) Board() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Board
}

// Player returns the current player's info.
func (c *GameConn) Player() *api.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return &c.state.Player
}

// PlayerList returns all players (including self) sorted by HP descending.
func (c *GameConn) PlayerList() []PlayerEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	p := c.state.Player
	list := make([]PlayerEntry, 0, len(c.state.Opponents)+1)
	list = append(list, PlayerEntry{ID: p.ID, HP: p.HP, ShopTier: p.ShopTier})
	for _, o := range c.state.Opponents {
		list = append(list, PlayerEntry{ID: o.ID, HP: o.HP, ShopTier: o.ShopTier})
	}
	slices.SortFunc(list, func(a, b PlayerEntry) int {
		return b.HP - a.HP
	})
	return list
}

// BuyCard sends a buy card action.
func (c *GameConn) BuyCard(shopIndex int) error {
	return c.send(api.ActionBuyCard, api.BuyCard{ShopIndex: shopIndex})
}

// SellMinion sends a sell minion action.
func (c *GameConn) SellMinion(boardIndex int) error {
	return c.send(api.ActionSellMinion, api.SellMinion{BoardIndex: boardIndex})
}

// PlaceMinion sends a place minion action.
func (c *GameConn) PlaceMinion(handIndex, boardPosition int) error {
	return c.send(api.ActionPlaceMinion, api.PlaceMinion{
		HandIndex:     handIndex,
		BoardPosition: boardPosition,
	})
}

// RemoveMinion sends a remove minion action.
func (c *GameConn) RemoveMinion(boardIndex int) error {
	return c.send(api.ActionRemoveMinion, api.RemoveMinion{BoardIndex: boardIndex})
}

// UpgradeShop sends an upgrade shop action.
func (c *GameConn) UpgradeShop() error {
	return c.send(api.ActionUpgradeShop, nil)
}

// RefreshShop sends a refresh shop action.
func (c *GameConn) RefreshShop() error {
	return c.send(api.ActionRefreshShop, nil)
}

// FreezeShop sends a freeze shop action.
func (c *GameConn) FreezeShop() error {
	return c.send(api.ActionFreezeShop, nil)
}

// SyncBoard sends the board order to server.
func (c *GameConn) SyncBoard(order []int) error {
	return c.send(api.ActionSyncBoard, api.SyncBoard{Order: order})
}

// Discover returns the current discover offer, or nil if none pending.
func (c *GameConn) Discover() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Discover
}

// PlaySpell sends a play spell action.
func (c *GameConn) PlaySpell(handIndex int) error {
	return c.send(api.ActionPlaySpell, api.PlaySpell{HandIndex: handIndex})
}

// DiscoverPick sends a discover pick action.
func (c *GameConn) DiscoverPick(index int) error {
	return c.send(api.ActionDiscoverPick, api.DiscoverPick{Index: index})
}

// CombatEvents returns the pending combat events, or nil.
func (c *GameConn) CombatEvents() []game.CombatEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.combatEvents
}

// ClearCombatAnimation discards the pending combat animation.
func (c *GameConn) ClearCombatAnimation() {
	c.mu.Lock()
	c.combatEvents = nil
	c.mu.Unlock()
}

// Close closes the connection.
func (c *GameConn) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

// Connected returns true if the client has received state.
func (c *GameConn) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state != nil
}

// WaitForState blocks until state is received or context is cancelled.
func (c *GameConn) WaitForState(ctx context.Context) error {
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
