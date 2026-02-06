package message

import "github.com/ysomad/gigabg/game"

// Client -> Server
type ClientMessage struct {
	Join         *JoinLobby
	Buy          *BuyCard
	SellCard     *SellCard
	PlaceMinion  *PlaceMinion
	RemoveMinion *RemoveMinion
	SyncBoard    *SyncBoard
	UpgradeShop  *UpgradeShop
	RefreshShop  *RefreshShop
	PlaySpell    *PlaySpell
	DiscoverPick *DiscoverPick
}

// Action returns a human-readable name of the action in the message.
func (m *ClientMessage) Action() string {
	switch {
	case m.Join != nil:
		return "join_lobby"
	case m.Buy != nil:
		return "buy_card"
	case m.SellCard != nil:
		return "sell_card"
	case m.PlaceMinion != nil:
		return "place_minion"
	case m.RemoveMinion != nil:
		return "remove_minion"
	case m.SyncBoard != nil:
		return "sync_board"
	case m.UpgradeShop != nil:
		return "upgrade_shop"
	case m.RefreshShop != nil:
		return "refresh_shop"
	case m.PlaySpell != nil:
		return "play_spell"
	case m.DiscoverPick != nil:
		return "discover_pick"
	default:
		return "unknown"
	}
}

type PlaySpell struct {
	HandIndex int
}

type DiscoverPick struct {
	Index int
}

type SyncBoard struct {
	Order []int // indices representing new order
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

type UpgradeShop struct{}

type RefreshShop struct{}

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
