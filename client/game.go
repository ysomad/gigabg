package client

import (
	"context"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
)

// PlayerEntry is a player summary for the sidebar.
type PlayerEntry struct {
	ID            game.PlayerID
	HP            int
	ShopTier      game.Tier
	CombatResults []api.CombatResult
	TopTribe      game.Tribe
	TopTribeCount int
}

// GameClient connects to a game server via WebSocket.
type GameClient struct {
	conn *websocket.Conn

	state           *api.GameState
	combatEvents    []api.CombatEvent
	opponentUpdates []api.OpponentUpdate
	mu              sync.RWMutex

	errCh chan error
}

// NewGameClient dials the game server WebSocket and returns a GameClient.
// addr is host:port (e.g. "localhost:8080").
// If proxyURL is non-empty, the WebSocket connection is routed through the given HTTP proxy.
func NewGameClient(
	ctx context.Context,
	addr string,
	player game.PlayerID,
	lobbyID, proxyURL string,
) (*GameClient, error) {
	wsURL := fmt.Sprintf("ws://%s/ws?player=%d&lobby=%s", addr, player, lobbyID)

	var opts *websocket.DialOptions
	if proxyURL != "" {
		u, _ := url.Parse(proxyURL)
		opts = &websocket.DialOptions{
			HTTPClient: &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(u)}},
		}
	}

	conn, resp, err := websocket.Dial(ctx, wsURL, opts)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("connect to lobby: %w", err)
	}

	c := &GameClient{conn: conn, errCh: make(chan error, 1)}
	go c.readPump(context.WithoutCancel(ctx))
	return c, nil
}

func (c *GameClient) readPump(ctx context.Context) {
	defer c.conn.CloseNow() //nolint:errcheck // best-effort cleanup

	for {
		_, data, err := c.conn.Read(ctx)
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

func (c *GameClient) handleMessage(msg *api.ServerMessage) {
	c.mu.Lock()
	if len(msg.CombatEvents) > 0 {
		c.combatEvents = msg.CombatEvents
	}
	if msg.State != nil {
		c.state = msg.State
	}
	if u := msg.OpponentUpdate; u != nil {
		c.opponentUpdates = append(c.opponentUpdates, *u)
		if c.state != nil {
			for i := range c.state.Opponents {
				if c.state.Opponents[i].ID == u.Player {
					c.state.Opponents[i].ShopTier = u.ShopTier
					break
				}
			}
		}
	}
	c.mu.Unlock()

	if msg.Error != nil {
		select {
		case c.errCh <- errors.New(msg.Error.Message):
		default:
		}
	}
}

func (c *GameClient) send(action api.Action, payload any) error {
	msg := api.ClientMessage{Action: action}

	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("payload marshal: %w", err)
		}
		msg.Payload = raw
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("msg marshal: %w", err)
	}

	return c.conn.Write(context.Background(), websocket.MessageBinary, data)
}

// State returns the current game state.
func (c *GameClient) State() *api.GameState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// PlayerID returns this client's player ID.
func (c *GameClient) PlayerID() game.PlayerID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.Player.ID
}

// Opponents returns all opponents in the lobby.
func (c *GameClient) Opponents() []api.Opponent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Opponents
}

// Turn returns the current turn number.
func (c *GameClient) Turn() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return 0
	}
	return c.state.Turn
}

// Phase returns the current phase.
func (c *GameClient) Phase() game.Phase {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return game.PhaseWaiting
	}
	return c.state.Phase
}

// PhaseEndsAt returns when the current phase ends.
func (c *GameClient) PhaseEndsAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return time.Time{}
	}
	return c.state.PhaseEndsAt
}

// Shop returns the current shop cards.
func (c *GameClient) Shop() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Shop
}

// IsShopFrozen returns whether the shop is frozen.
func (c *GameClient) IsShopFrozen() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return false
	}
	return c.state.IsShopFrozen
}

// Hand returns the current hand cards.
func (c *GameClient) Hand() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Hand
}

// Board returns the current board cards.
func (c *GameClient) Board() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Board
}

// Player returns the current player's info.
func (c *GameClient) Player() *api.Player {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return &c.state.Player
}

// GameResult returns the game result, or nil if the game hasn't finished.
func (c *GameClient) GameResult() *api.GameResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.GameResult
}

// PlayerList returns all players (including self) sorted by HP descending.
func (c *GameClient) PlayerList() []PlayerEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	p := c.state.Player
	tribes := make([]game.Tribe, len(c.state.Board))
	for i, card := range c.state.Board {
		tribes[i] = card.Tribe
	}
	selfTribe, selfCount := game.CalcTopTribe(tribes)
	list := make([]PlayerEntry, 0, len(c.state.Opponents)+1)
	list = append(list, PlayerEntry{
		ID:            p.ID,
		HP:            p.HP,
		ShopTier:      p.ShopTier,
		CombatResults: c.state.CombatResults,
		TopTribe:      selfTribe,
		TopTribeCount: selfCount,
	})
	for _, o := range c.state.Opponents {
		list = append(list, PlayerEntry{
			ID:            o.ID,
			HP:            o.HP,
			ShopTier:      o.ShopTier,
			CombatResults: o.CombatResults,
			TopTribe:      o.TopTribe,
			TopTribeCount: o.TopTribeCount,
		})
	}
	slices.SortFunc(list, func(a, b PlayerEntry) int {
		return b.HP - a.HP
	})
	return list
}

// BuyCard sends a buy card action.
func (c *GameClient) BuyCard(shopIndex int) error {
	return c.send(api.ActionBuyCard, api.BuyCard{ShopIndex: shopIndex})
}

// SellMinion sends a sell minion action.
func (c *GameClient) SellMinion(boardIndex int) error {
	return c.send(api.ActionSellMinion, api.SellMinion{BoardIndex: boardIndex})
}

// PlaceMinion sends a place minion action.
func (c *GameClient) PlaceMinion(handIndex, boardPosition int) error {
	return c.send(api.ActionPlaceMinion, api.PlaceMinion{
		HandIndex:     handIndex,
		BoardPosition: boardPosition,
	})
}

// RemoveMinion sends a remove minion action.
func (c *GameClient) RemoveMinion(boardIndex int) error {
	return c.send(api.ActionRemoveMinion, api.RemoveMinion{BoardIndex: boardIndex})
}

// UpgradeShop sends an upgrade shop action.
func (c *GameClient) UpgradeShop() error {
	return c.send(api.ActionUpgradeShop, nil)
}

// RefreshShop sends a refresh shop action.
func (c *GameClient) RefreshShop() error {
	return c.send(api.ActionRefreshShop, nil)
}

// FreezeShop sends a freeze shop action.
func (c *GameClient) FreezeShop() error {
	return c.send(api.ActionFreezeShop, nil)
}

// ReorderCards sends board and shop card order to the server.
func (c *GameClient) ReorderCards(boardOrder, shopOrder []int) error {
	return c.send(api.ActionReorderCards, api.ReorderCards{
		BoardOrder: boardOrder,
		ShopOrder:  shopOrder,
	})
}

// Discovers returns the current discover offer, or nil if none pending.
func (c *GameClient) Discovers() []api.Card {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.state == nil {
		return nil
	}
	return c.state.Discovers
}

// PlaySpell sends a play spell action.
func (c *GameClient) PlaySpell(handIndex int) error {
	return c.send(api.ActionPlaySpell, api.PlaySpell{HandIndex: handIndex})
}

// DiscoverPick sends a discover pick action.
func (c *GameClient) DiscoverPick(index int) error {
	return c.send(api.ActionDiscoverPick, api.DiscoverPick{Index: index})
}

// DrainOpponentUpdates returns and clears pending opponent updates.
func (c *GameClient) DrainOpponentUpdates() []api.OpponentUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()
	updates := c.opponentUpdates
	c.opponentUpdates = nil
	return updates
}

// CombatEvents returns the pending combat events, or nil.
func (c *GameClient) CombatEvents() []api.CombatEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.combatEvents
}

// ClearCombatLog discards the pending combat log.
func (c *GameClient) ClearCombatLog() {
	c.mu.Lock()
	c.combatEvents = nil
	c.mu.Unlock()
}

// Close closes the connection.
func (c *GameClient) Close() error {
	return c.conn.Close(websocket.StatusNormalClosure, "")
}

// Connected returns true if the client has received state.
func (c *GameClient) Connected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state != nil
}

// WaitForState blocks until state is received or context is cancelled.
func (c *GameClient) WaitForState(ctx context.Context) error {
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
