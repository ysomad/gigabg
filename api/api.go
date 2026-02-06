package api

import (
	"encoding/json"

	"github.com/ysomad/gigabg/game"
)

type Action uint8

const (
	ActionJoinLobby Action = iota + 1
	ActionBuyCard
	ActionSellMinion
	ActionPlaceMinion
	ActionRemoveMinion
	ActionSyncBoard
	ActionUpgradeShop
	ActionRefreshShop
	ActionFreezeShop
	ActionPlaySpell
	ActionDiscoverPick
)

func (a Action) String() string {
	switch a {
	case ActionJoinLobby:
		return "join_lobby"
	case ActionBuyCard:
		return "buy_card"
	case ActionSellMinion:
		return "sell_minion"
	case ActionPlaceMinion:
		return "place_minion"
	case ActionRemoveMinion:
		return "remove_minion"
	case ActionSyncBoard:
		return "sync_board"
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
	default:
		return "unknown"
	}
}

// ClientMessage represents message which client must send to a server.
type ClientMessage struct {
	Action  Action          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

type JoinLobby struct {
	LobbyID string `json:"lobby_id"`
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

type SyncBoard struct {
	Order []int `json:"order"`
}

type PlaySpell struct {
	HandIndex int `json:"hand_index"`
}

type DiscoverPick struct {
	Index int `json:"index"`
}

// ServerMessage represents message which server must send to a client.
type ServerMessage struct {
	State        *GameState         `json:"state"`
	Error        *Error             `json:"error,omitempty"`
	CombatEvents []game.CombatEvent `json:"combat_events,omitempty"`
}

type GameState struct {
	Player            Player              `json:"player"`
	Opponents         []Opponent          `json:"opponents"`
	Turn              int                 `json:"turn"`
	Phase             game.Phase          `json:"phase"`
	PhaseEndTimestamp int64               `json:"phase_end_timestamp"`
	Shop              []Card              `json:"shop"`
	IsShopFrozen      bool                `json:"shop_frozen"`
	Hand              []Card              `json:"hand"`
	Board             []Card              `json:"board"`
	Discover          []Card              `json:"discover"`
	CombatResults     []game.CombatResult `json:"combat_results"`
	OpponentID        string              `json:"opponent_id"`    // combat phase only
	CombatBoard       []Card              `json:"combat_board"`   // player's board with combat IDs, combat phase only
	OpponentBoard     []Card              `json:"opponent_board"` // opponent's board with combat IDs, combat phase only
}

type Player struct {
	ID          string    `json:"id"`
	HP          int       `json:"hp"`
	Gold        int       `json:"gold"`
	MaxGold     int       `json:"max_gold"`
	ShopTier    game.Tier `json:"shop_tier"`
	UpgradeCost int       `json:"upgrade_cost"`
}

type Opponent struct {
	ID       string    `json:"id"`
	HP       int       `json:"hp"`
	ShopTier game.Tier `json:"shop_tier"`
}

type Card struct {
	TemplateID string `json:"template_id"`
	Attack     int    `json:"attack"`
	Health     int    `json:"health"`
	IsGolden   bool   `json:"is_golden"`
	CombatID   int    `json:"combat_id"` // set only in combat context
}

type Error struct {
	Message string
}

func NewPlayer(p *game.Player) Player {
	return Player{
		ID:          p.ID(),
		HP:          p.HP(),
		Gold:        p.Gold(),
		MaxGold:     p.MaxGold(),
		ShopTier:    p.Shop().Tier(),
		UpgradeCost: p.Shop().UpgradeCost(),
	}
}

func NewOpponents(players []*game.Player, excludeID string) []Opponent {
	res := make([]Opponent, 0, len(players)-1)
	for _, p := range players {
		if p.ID() == excludeID {
			continue
		}
		res = append(res, Opponent{
			ID:       p.ID(),
			HP:       p.HP(),
			ShopTier: p.Shop().Tier(),
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
			IsGolden:   m.IsGolden(),
		}
	}
	return Card{TemplateID: c.TemplateID()}
}

func NewCardFromMinion(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		IsGolden:   m.IsGolden(),
	}
}

func NewCombatCard(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		IsGolden:   m.IsGolden(),
		CombatID:   m.CombatID(),
	}
}

// CombatCards converts a board (with combat IDs) to API cards.
func CombatCards(b game.Board) []Card {
	cards := make([]Card, b.Len())
	for i := range b.Len() {
		cards[i] = NewCombatCard(b.Get(i))
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
