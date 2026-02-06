package api

import (
	"encoding/json"

	"github.com/ysomad/gigabg/game"
)

type Action uint8

const (
	ActionJoinLobby Action = iota + 1
	ActionBuyCard
	ActionSellCard
	ActionPlaceMinion
	ActionRemoveMinion
	ActionSyncBoard
	ActionUpgradeShop
	ActionRefreshShop
	ActionPlaySpell
	ActionDiscoverPick
)

func (a Action) String() string {
	switch a {
	case ActionJoinLobby:
		return "join_lobby"
	case ActionBuyCard:
		return "buy_card"
	case ActionSellCard:
		return "sell_card"
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
	case ActionPlaySpell:
		return "play_spell"
	case ActionDiscoverPick:
		return "discover_pick"
	default:
		return "unknown"
	}
}

// Client -> Server
type ClientMessage struct {
	Action  Action
	Payload json.RawMessage
}

type JoinLobby struct {
	LobbyID string
}

type BuyCard struct {
	ShopIndex int
}

type SellCard struct {
	HandIndex int
}

type PlaceMinion struct {
	HandIndex     int
	BoardPosition int
}

type RemoveMinion struct {
	BoardIndex int
}

type SyncBoard struct {
	Order []int
}

type PlaySpell struct {
	HandIndex int
}

type DiscoverPick struct {
	Index int
}

// Server -> Client
type ServerMessage struct {
	State *GameState
	Error *Error
}

type GameState struct {
	PlayerID          string
	Turn              int
	Phase             game.Phase
	PhaseEndTimestamp int64
	Players           []Player
	Shop              []Card
	Hand              []Card
	Board             []Card
	Discover          *DiscoverOffer
}

type DiscoverOffer struct {
	Cards []Card
}

type Player struct {
	ID          string
	HP          int
	Gold        int
	MaxGold     int
	ShopTier    int
	UpgradeCost int
}

type Card struct {
	TemplateID string
	Attack     int
	Health     int
	Golden     bool
}

type Error struct {
	Message string
}

func NewPlayer(p *game.Player) Player {
	return Player{
		ID:          p.ID,
		HP:          p.HP,
		Gold:        p.Gold,
		MaxGold:     p.MaxGold,
		ShopTier:    int(p.ShopTier),
		UpgradeCost: p.UpgradeCost(),
	}
}

func NewCard(c game.Card) Card {
	if m, ok := c.(*game.Minion); ok {
		return Card{
			TemplateID: m.TemplateID(),
			Attack:     m.Attack(),
			Health:     m.Health(),
			Golden:     m.Golden(),
		}
	}
	return Card{TemplateID: c.TemplateID()}
}

func NewCardFromMinion(m *game.Minion) Card {
	return Card{
		TemplateID: m.TemplateID(),
		Attack:     m.Attack(),
		Health:     m.Health(),
		Golden:     m.Golden(),
	}
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

func NewPlayers(players []*game.Player) []Player {
	res := make([]Player, 0, len(players))
	for _, p := range players {
		res = append(res, NewPlayer(p))
	}
	return res
}
