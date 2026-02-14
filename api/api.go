package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ysomad/gigabg/game"
)

type Action uint8

const (
	ActionBuyCard Action = iota + 1
	ActionSellMinion
	ActionPlaceMinion
	ActionRemoveMinion
	ActionUpgradeShop
	ActionRefreshShop
	ActionFreezeShop
	ActionPlaySpell
	ActionDiscoverPick
	ActionReorderCards
)

func (a Action) String() string {
	switch a {
	case ActionBuyCard:
		return "buy_card"
	case ActionSellMinion:
		return "sell_minion"
	case ActionPlaceMinion:
		return "place_minion"
	case ActionRemoveMinion:
		return "remove_minion"
	case ActionUpgradeShop:
		return "upgrade_shop"
	case ActionRefreshShop:
		return "refresh_shop"
	case ActionFreezeShop:
		return "freeze_shop"
	case ActionPlaySpell:
		return "play_spell"
	case ActionDiscoverPick:
		return "discover_pick"
	case ActionReorderCards:
		return "reorder_cards"
	default:
		return "unknown"
	}
}

// ClientMessage represents message which client must send to a server.
type ClientMessage struct {
	Action  Action          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type BuyCard struct {
	ShopIndex int `json:"shop_index"`
}

type SellMinion struct {
	BoardIndex int `json:"board_index"`
}

type PlaceMinion struct {
	HandIndex     int `json:"hand_index"`
	BoardPosition int `json:"board_position"`
}

type RemoveMinion struct {
	BoardIndex int `json:"board_index"`
}

type ReorderCards struct {
	BoardOrder []int `json:"board_order,omitempty"`
	ShopOrder  []int `json:"shop_order,omitempty"`
}

type PlaySpell struct {
	HandIndex int `json:"hand_index"`
}

type DiscoverPick struct {
	Index int `json:"index"`
}

// HTTP types for lobby creation.

type CreateLobbyReq struct {
	MaxPlayers int `json:"max_players"`
}

type CreateLobbyResp struct {
	LobbyID string `json:"lobby_id"`
}

// ServerMessage represents message which server must send to a client.
type ServerMessage struct {
	State          *GameState      `json:"state,omitempty"`
	Error          *Error          `json:"error,omitempty"`
	CombatEvents   []CombatEvent   `json:"combat_events,omitempty"`
	OpponentUpdate *OpponentUpdate `json:"opponent_update,omitempty"`
}

// OpponentUpdate is a lightweight notification about an opponent's tier change.
type OpponentUpdate struct {
	PlayerID string    `json:"player_id"`
	ShopTier game.Tier `json:"shop_tier"`
}

type GameState struct {
	Player        Player         `json:"player"`
	Opponents     []Opponent     `json:"opponents"`
	Turn          int            `json:"turn"`
	Phase         game.Phase     `json:"phase"`
	PhaseEndsAt   time.Time      `json:"phase_ends_at"`
	Shop          []Card         `json:"shop,omitempty"`
	IsShopFrozen  bool           `json:"is_shop_frozen,omitempty"`
	Hand          []Card         `json:"hand,omitempty"`
	Board         []Card         `json:"board,omitempty"`
	Discovers     []Card         `json:"discovers,omitempty"`
	CombatResults []CombatResult `json:"combat_results,omitempty"`
	OpponentID    string         `json:"opponent_id"`              // combat phase only
	CombatBoard   []Card         `json:"combat_board,omitempty"`   // combat phase only
	OpponentBoard []Card         `json:"opponent_board,omitempty"` // combat phase only
	GameResult    *GameResult    `json:"game_result,omitempty"`
}

type Player struct {
	ID          string    `json:"id"`
	HP          int       `json:"hp"`
	Gold        int       `json:"gold"`
	MaxGold     int       `json:"max_gold"`
	ShopTier    game.Tier `json:"shop_tier"`
	UpgradeCost int       `json:"upgrade_cost"`
	RefreshCost int       `json:"refresh_cost"`
}

type Opponent struct {
	ID            string         `json:"id"`
	HP            int            `json:"hp"`
	ShopTier      game.Tier      `json:"shop_tier"`
	CombatResults []CombatResult `json:"combat_results,omitempty"`
	TopTribe      game.Tribe     `json:"top_tribe,omitempty"`
	TopTribeCount int            `json:"top_tribe_count,omitempty"`
}

type Card struct {
	TemplateID string        `json:"template_id"`
	Tribe      game.Tribe    `json:"tribe"`
	Attack     int           `json:"attack"`
	Health     int           `json:"health"`
	IsGolden   bool          `json:"is_golden,omitempty"`
	Cost       int           `json:"cost,omitempty"`
	Keywords   game.Keywords `json:"keywords,omitempty"`
	CombatID   int           `json:"combat_id,omitempty"` // set only in combat context
}

// CombatEvent is the JSON envelope: type discriminator + raw payload.
type CombatEvent struct {
	Type    game.CombatEventType `json:"type"`
	Payload json.RawMessage      `json:"payload"`
}

// NewCombatEvents encodes game events to JSON envelopes.
func NewCombatEvents(events []game.CombatEvent) ([]CombatEvent, error) {
	res := make([]CombatEvent, len(events))
	for i, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			return nil, fmt.Errorf("event %d: %w", ev.EventType(), err)
		}
		res[i] = CombatEvent{Type: ev.EventType(), Payload: raw}
	}
	return res, nil
}

type CombatResult struct {
	OpponentID string `json:"opponent_id"`
	WinnerID   string `json:"winner_id"`
	Damage     int    `json:"damage"`
}

func NewCombatResult(cr game.CombatResult) CombatResult {
	return CombatResult{
		OpponentID: cr.OpponentID,
		WinnerID:   cr.WinnerID,
		Damage:     cr.Damage,
	}
}

func NewCombatResults(results []game.CombatResult) []CombatResult {
	res := make([]CombatResult, len(results))
	for i, cr := range results {
		res[i] = NewCombatResult(cr)
	}
	return res
}

func NewAllCombatResults(m map[string][]game.CombatResult) map[string][]CombatResult {
	res := make(map[string][]CombatResult, len(m))
	for k, v := range m {
		res[k] = NewCombatResults(v)
	}
	return res
}

type Error struct {
	Message string `json:"message"`
}

func NewPlayer(p *game.Player) Player {
	return Player{
		ID:          p.ID(),
		HP:          p.HP(),
		Gold:        p.Gold(),
		MaxGold:     p.MaxGold(),
		ShopTier:    p.Shop().Tier(),
		UpgradeCost: p.Shop().UpgradeCost(),
		RefreshCost: p.Shop().RefreshCost(),
	}
}

func NewOpponents(
	players []*game.Player,
	excludeID string,
	combatResults map[string][]CombatResult,
	tribes map[string]game.TribeSnapshot,
) []Opponent {
	res := make([]Opponent, 0, len(players)-1)
	for _, p := range players {
		if p.ID() == excludeID {
			continue
		}
		snap := tribes[p.ID()]
		res = append(res, Opponent{
			ID:            p.ID(),
			HP:            p.HP(),
			ShopTier:      p.Shop().Tier(),
			CombatResults: combatResults[p.ID()],
			TopTribe:      snap.Tribe,
			TopTribeCount: snap.Count,
		})
	}
	return res
}

func NewCard(c game.Card) Card {
	if m, ok := c.(*game.Minion); ok {
		return Card{
			TemplateID: m.TemplateID(),
			Attack:     m.Attack(),
			Health:     m.Health(),
			Cost:       m.Cost(),
			IsGolden:   m.IsGolden(),
			Tribe:      m.Tribe(),
			Keywords:   m.Keywords(),
		}
	}
	return Card{
		TemplateID: c.Template().ID(),
		Cost:       c.Template().Cost(),
	}
}

func NewCardFromMinion(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		Cost:       m.Cost(),
		IsGolden:   m.IsGolden(),
		Tribe:      m.Tribe(),
		Keywords:   m.Keywords(),
	}
}

func NewCombatCard(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		IsGolden:   m.IsGolden(),
		Tribe:      m.Tribe(),
		Keywords:   m.Keywords(),
		CombatID:   m.CombatID(),
	}
}

// CombatCards converts a board (with combat IDs) to API cards.
func CombatCards(b game.Board) []Card {
	cards := make([]Card, b.Len())
	for i := range b.Len() {
		cards[i] = NewCombatCard(b.MinionAt(i))
	}
	return cards
}

func NewCards(cards []game.Card) []Card {
	res := make([]Card, 0, len(cards))
	for _, c := range cards {
		res = append(res, NewCard(c))
	}
	return res
}

func NewCardsFromMinions(minions []*game.Minion) []Card {
	res := make([]Card, 0, len(minions))
	for _, m := range minions {
		res = append(res, NewCardFromMinion(m))
	}
	return res
}

type PlayerPlacement struct {
	PlayerID      string     `json:"player_id"`
	Placement     int        `json:"placement"`
	TopTribe      game.Tribe `json:"majority_tribe"`
	TopTribeCount int        `json:"majority_count"`
}

type GameResult struct {
	WinnerID   string            `json:"winner_id"`
	Placements []PlayerPlacement `json:"placements"`
	Duration   time.Duration     `json:"duration"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at"`
}

func NewGameResult(r *game.GameResult) *GameResult {
	if r == nil {
		return nil
	}
	placements := make([]PlayerPlacement, len(r.Placements))
	for i, p := range r.Placements {
		placements[i] = PlayerPlacement{
			PlayerID:      p.PlayerID,
			Placement:     p.Placement,
			TopTribe:      p.TopTribe,
			TopTribeCount: p.TopTribeCount,
		}
	}
	return &GameResult{
		WinnerID:   r.WinnerID,
		Placements: placements,
		Duration:   r.Duration,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
}
