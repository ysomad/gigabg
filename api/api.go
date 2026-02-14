package api

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
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
	Action  Action         `json:"action"`
	Payload jsontext.Value `json:"payload,omitzero"`
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
	Player   game.PlayerID `json:"player"`
	ShopTier game.Tier     `json:"shop_tier"`
}

type GameState struct {
	Player        Player         `json:"player"`
	Opponents     []Opponent     `json:"opponents"`
	Turn          int            `json:"turn"`
	Phase         game.Phase     `json:"phase"`
	PhaseEndsAt   time.Time      `json:"phase_ends_at,format:unix"`
	Shop          []Card         `json:"shop,omitempty"`
	IsShopFrozen  bool           `json:"is_shop_frozen,omitzero"`
	Hand          []Card         `json:"hand,omitempty"`
	Board         []Card         `json:"board,omitempty"`
	Discovers     []Card         `json:"discovers,omitempty"`
	CombatResults []CombatResult `json:"combat_results,omitempty"`
	Opponent      game.PlayerID  `json:"opponent"`                 // combat phase only
	CombatBoard   []Card         `json:"combat_board,omitempty"`   // combat phase only
	OpponentBoard []Card         `json:"opponent_board,omitempty"` // combat phase only
	GameResult    *GameResult    `json:"game_result,omitempty"`
}

type Player struct {
	ID          game.PlayerID `json:"id"`
	HP          int           `json:"hp"`
	Gold        int           `json:"gold"`
	MaxGold     int           `json:"max_gold"`
	ShopTier    game.Tier     `json:"shop_tier"`
	UpgradeCost int           `json:"upgrade_cost"`
	RefreshCost int           `json:"refresh_cost"`
}

type Opponent struct {
	ID            game.PlayerID  `json:"id"`
	HP            int            `json:"hp"`
	ShopTier      game.Tier      `json:"shop_tier"`
	CombatResults []CombatResult `json:"combat_results,omitempty"`
	TopTribe      game.Tribe     `json:"top_tribe,omitzero"`
	TopTribeCount int            `json:"top_tribe_count,omitzero"`
}

type Card struct {
	Template string        `json:"template"`
	Tribe    game.Tribe    `json:"tribe"`
	Attack   int           `json:"attack"`
	Health   int           `json:"health"`
	IsGolden bool          `json:"is_golden,omitzero"`
	Cost     int           `json:"cost,omitzero"`
	Keywords game.Keywords `json:"keywords,omitzero"`
	CombatID game.CombatID `json:"combat_id,omitzero"` // set only in combat context
}

// CombatEvent is the JSON envelope: type discriminator + raw payload.
type CombatEvent struct {
	Type    game.CombatEventType `json:"type"`
	Payload jsontext.Value       `json:"payload"`
}

// NewCombatEvents encodes game events to JSON envelopes.
func NewCombatEvents(events []game.CombatEvent) ([]CombatEvent, error) {
	res := make([]CombatEvent, len(events))
	for i, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			return nil, fmt.Errorf("event %d: %w", ev.Type(), err)
		}
		res[i] = CombatEvent{Type: ev.Type(), Payload: raw}
	}
	return res, nil
}

type CombatResult struct {
	Opponent game.PlayerID `json:"opponent"`
	Winner   game.PlayerID `json:"winner"`
	Damage   int           `json:"damage"`
}

func NewCombatResult(cr game.CombatResult) CombatResult {
	return CombatResult{
		Opponent: cr.Opponent,
		Winner:   cr.Winner,
		Damage:   cr.Damage,
	}
}

func NewCombatResults(results []game.CombatResult) []CombatResult {
	res := make([]CombatResult, len(results))
	for i, cr := range results {
		res[i] = NewCombatResult(cr)
	}
	return res
}

func NewAllCombatResults(m map[game.PlayerID][]game.CombatResult) map[game.PlayerID][]CombatResult {
	res := make(map[game.PlayerID][]CombatResult, len(m))
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
	exclude game.PlayerID,
	combatResults map[game.PlayerID][]CombatResult,
	tribes map[game.PlayerID]game.TribeSnapshot,
) []Opponent {
	res := make([]Opponent, 0, len(players)-1)
	for _, p := range players {
		if p.ID() == exclude {
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
			Template: m.TemplateID(),
			Attack:   m.Attack(),
			Health:   m.Health(),
			Cost:     m.Cost(),
			IsGolden: m.IsGolden(),
			Tribe:    m.Tribe(),
			Keywords: m.Keywords(),
		}
	}
	return Card{
		Template: c.Template().ID(),
		Cost:     c.Template().Cost(),
	}
}

func NewCardFromMinion(m *game.Minion) Card {
	return Card{
		Template: m.TemplateID(),
		Attack:   m.Attack(),
		Health:   m.Health(),
		Cost:     m.Cost(),
		IsGolden: m.IsGolden(),
		Tribe:    m.Tribe(),
		Keywords: m.Keywords(),
	}
}

func NewCombatCard(m *game.Minion) Card {
	return Card{
		Template: m.TemplateID(),
		Attack:   m.Attack(),
		Health:   m.Health(),
		IsGolden: m.IsGolden(),
		Tribe:    m.Tribe(),
		Keywords: m.Keywords(),
		CombatID: m.CombatID(),
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
	Player        game.PlayerID `json:"player"`
	Placement     int           `json:"placement"`
	TopTribe      game.Tribe    `json:"top_tribe"`
	TopTribeCount int           `json:"top_tribe_count"`
}

type GameResult struct {
	Winner     game.PlayerID     `json:"winner"`
	Placements []PlayerPlacement `json:"placements"`
	Duration   time.Duration     `json:"duration,format:nano"`
	StartedAt  time.Time         `json:"started_at,format:unix"`
	EndedAt    time.Time         `json:"ended_at,format:unix"`
}

func NewGameResult(r *game.GameResult) *GameResult {
	if r == nil {
		return nil
	}
	placements := make([]PlayerPlacement, len(r.Placements))
	for i, p := range r.Placements {
		placements[i] = PlayerPlacement{
			Player:        p.Player,
			Placement:     p.Placement,
			TopTribe:      p.TopTribe,
			TopTribeCount: p.TopTribeCount,
		}
	}
	return &GameResult{
		Winner:     r.Winner,
		Placements: placements,
		Duration:   r.Duration,
		StartedAt:  r.StartedAt,
		EndedAt:    r.EndedAt,
	}
}
