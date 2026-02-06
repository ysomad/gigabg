package game

// Shop tier upgrade costs (index = current tier).
var _upgradeCosts = [6]int{0, 5, 7, 8, 9, 10} // tier 1->2 costs 5, etc.

type Shop struct {
	cards    []Card
	tier     Tier
	frozen   bool
	discount int
}

func (s Shop) Cards() []Card  { return s.cards }
func (s Shop) Tier() Tier     { return s.tier }
func (s Shop) IsFrozen() bool { return s.frozen }

// Size returns how many cards to show based on shop tier.
func (s Shop) Size() int {
	switch s.tier {
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

// UpgradeCost returns the cost to upgrade to the next shop tier.
func (s Shop) UpgradeCost() int {
	if !s.tier.IsValid() {
		return 0
	}
	cost := _upgradeCosts[s.tier] - s.discount
	if cost < 0 {
		return 0
	}
	return cost
}

// StartTurn prepares the shop for a new turn.
// If frozen, keeps current cards and unfreezes. Otherwise refreshes.
func (s *Shop) StartTurn(pool *CardPool) {
	s.discount++

	if s.frozen {
		s.frozen = false
		return
	}

	pool.ReturnCards(s.cards)
	s.cards = pool.Roll(s.tier, nil, s.Size())
}

// BuyCard removes a card from the shop at the given index and returns it.
func (s *Shop) BuyCard(index int) (Card, error) {
	if index < 0 || index >= len(s.cards) {
		return nil, ErrInvalidIndex
	}

	card := s.cards[index]
	s.cards = append(s.cards[:index], s.cards[index+1:]...)
	return card, nil
}

// Refresh returns old cards to pool and rolls new ones.
func (s *Shop) Refresh(pool *CardPool) {
	s.frozen = false
	pool.ReturnCards(s.cards)
	s.cards = pool.Roll(s.tier, nil, s.Size())
}

// Upgrade upgrades the shop to the next tier.
func (s *Shop) Upgrade() {
	s.tier++
	s.discount = 0
}

// Freeze toggles the shop freeze state.
func (s *Shop) Freeze() {
	s.frozen = !s.frozen
}
