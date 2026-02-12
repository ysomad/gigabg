package api

import (
	"encoding/json"

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
	ActionSyncBoards
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
	case ActionSyncBoards:
		return "sync_boards"
	default:
		return "unknown"
	}
}

// ClientMessage represents message which client must send to a server.
type ClientMessage struct {
	Action  Action          `json:"action"`
	Payload json.RawMessage `json:"payload"`
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

type SyncBoards struct {
	BoardOrder []int `json:"board_order"`
	ShopOrder  []int `json:"shop_order"`
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
	State          *GameState         `json:"state"`
	Error          *Error             `json:"error,omitempty"`
	CombatEvents   []game.CombatEvent `json:"combat_events,omitempty"`
	OpponentUpdate *OpponentUpdate    `json:"opponent_update,omitempty"`
}

// OpponentUpdate is a lightweight notification about an opponent's tier change.
type OpponentUpdate struct {
	PlayerID string    `json:"player_id"`
	ShopTier game.Tier `json:"shop_tier"`
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
	Discovers         []Card              `json:"discovers"`
	CombatResults     []game.CombatResult `json:"combat_results"`
	OpponentID        string              `json:"opponent_id"`    // combat phase only
	CombatBoard       []Card              `json:"combat_board"`   // combat phase only
	OpponentBoard     []Card              `json:"opponent_board"` // combat phase only
	GameResult        *GameResult         `json:"game_result,omitempty"`
}

type Player struct {
	ID             string    `json:"id"`
	HP             int       `json:"hp"`
	Gold           int       `json:"gold"`
	MaxGold        int       `json:"max_gold"`
	ShopTier       game.Tier `json:"shop_tier"`
	UpgradeCost    int       `json:"upgrade_cost"`
}

type Opponent struct {
	ID            string              `json:"id"`
	HP            int                 `json:"hp"`
	ShopTier      game.Tier           `json:"shop_tier"`
	CombatResults []game.CombatResult `json:"combat_results"`
	MajorityTribe game.Tribe          `json:"majority_tribe"`
	MajorityCount int                 `json:"majority_count"`
}

type Card struct {
	TemplateID string     `json:"template_id"`
	Attack     int        `json:"attack"`
	Health     int        `json:"health"`
	IsGolden   bool       `json:"is_golden"`
	Tribe      game.Tribe `json:"tribe"`
	CombatID   int        `json:"combat_id"` // set only in combat context
}

type Error struct {
	Message string
}

func NewPlayer(p *game.Player) Player {
	return Player{
		ID:             p.ID(),
		HP:             p.HP(),
		Gold:           p.Gold(),
		MaxGold:        p.MaxGold(),
		ShopTier:       p.Shop().Tier(),
		UpgradeCost:    p.Shop().UpgradeCost(),
	}
}

func NewOpponents(
	players []*game.Player,
	excludeID string,
	combatResults map[string][]game.CombatResult,
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
			MajorityTribe: snap.Tribe,
			MajorityCount: snap.Count,
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
			Tribe:      m.Tribe(),
		}
	}
	return Card{TemplateID: c.Template().ID()}
}

func NewCardFromMinion(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		IsGolden:   m.IsGolden(),
		Tribe:      m.Tribe(),
	}
}

func NewCombatCard(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		IsGolden:   m.IsGolden(),
		Tribe:      m.Tribe(),
		CombatID:   m.CombatID(),
	}
}

// CombatCards converts a board (with combat IDs) to API cards.
func CombatCards(b game.Board) []Card {
	cards := make([]Card, b.Len())
	for i := range b.Len() {
		cards[i] = NewCombatCard(b.GetMinion(i))
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
	MajorityTribe game.Tribe `json:"majority_tribe"`
	MajorityCount int        `json:"majority_count"`
}

type GameResult struct {
	WinnerID   string            `json:"winner_id"`
	Placements []PlayerPlacement `json:"placements"`
	Duration   int64             `json:"duration"`
	StartedAt  int64             `json:"started_at"`
	FinishedAt int64             `json:"finished_at"`
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
			MajorityTribe: p.MajorityTribe,
			MajorityCount: p.MajorityCount,
		}
	}
	return &GameResult{
		WinnerID:   r.WinnerID,
		Placements: placements,
		Duration:   int64(r.Duration.Seconds()),
		StartedAt:  r.StartedAt.Unix(),
		FinishedAt: r.FinishedAt.Unix(),
	}
}
