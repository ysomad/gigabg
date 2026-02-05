package game

import (
	"math/rand/v2"
)

// CardStore provides access to card templates.
type CardStore interface {
	ByTemplateID(id string) *CardTemplate
	Query(maxTier Tier, tribes []Tribe) []*CardTemplate
}

// CardPool manages the card pool for rolling shop cards.
type CardPool struct {
	cards CardStore
}

// NewCardPool creates a new card pool.
func NewCardPool(cards CardStore) *CardPool {
	return &CardPool{cards: cards}
}

// Roll returns a random selection of cards for the shop.
func (p *CardPool) Roll(maxTier Tier, tribes []Tribe, count int) []Card {
	available := p.cards.Query(maxTier, tribes)
	if len(available) == 0 {
		return nil
	}

	res := make([]Card, 0, count)

	for range count {
		idx := rand.IntN(len(available))
		res = append(res, NewMinion(available[idx]))
	}

	return res
}
