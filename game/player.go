package game

import "github.com/ysomad/gigabg/errorsx"

const (
	MaxBoardSize = 7
	MaxHandSize  = 10
	MaxGold      = 10
	BuyCost      = 3
	SellValue    = 1
	RefreshCost  = 1
)

const (
	ErrNotEnoughGold errorsx.Error = "not enough gold"
	ErrBoardFull     errorsx.Error = "board is full"
	ErrHandFull      errorsx.Error = "hand is full"
	ErrInvalidIndex  errorsx.Error = "invalid index"
	ErrMaxTier       errorsx.Error = "already at max tier"
)

// Shop tier upgrade costs (index = current tier)
var _upgradeCosts = [6]int{0, 5, 7, 8, 9, 10} // tier 1->2 costs 5, etc.

type Player struct {
	ID       string
	HP       int
	Gold     int
	MaxGold  int
	ShopTier Tier

	Board []*Minion // only minions on board
	Hand  []Card    // can hold minions and spells
	Shop  []Card    // shop offerings
}

func NewPlayer(id string) *Player {
	return &Player{
		ID:       id,
		HP:       40,
		Gold:     3,
		MaxGold:  3,
		ShopTier: 1,
		Board:    make([]*Minion, 0, MaxBoardSize),
		Hand:     make([]Card, 0, MaxHandSize),
		Shop:     make([]Card, 0),
	}
}

// UpgradeCost returns the cost to upgrade to the next tavern tier.
func (p *Player) UpgradeCost() int {
	if !p.ShopTier.IsValid() {
		return 0
	}
	return _upgradeCosts[p.ShopTier]
}

// StartTurn prepares the player for a new turn.
func (p *Player) StartTurn(pool *CardPool) {
	// Increase max gold (cap at 10)
	if p.MaxGold < MaxGold {
		p.MaxGold++
	}
	p.Gold = p.MaxGold

	// Refresh shop for free
	p.Shop = pool.Roll(p.ShopTier, nil, p.ShopSize())
}

// TakeDamage reduces player HP and returns true if player is dead.
func (p *Player) TakeDamage(damage int) bool {
	p.HP -= damage
	return p.HP <= 0
}

// ShopSize returns how many minions to show based on shop tier.
func (p *Player) ShopSize() int {
	switch p.ShopTier {
	case Tier1:
		return 3
	case Tier2, Tier3:
		return 4
	case Tier4:
		return 5
	case Tier5:
		return 6
	case Tier6:
		return 7
	default:
		return 0
	}
}
