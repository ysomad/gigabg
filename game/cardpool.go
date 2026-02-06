package game

import (
	"math/rand/v2"
)

const DiscoverCount = 3

// TierPoolSize defines how many copies of each template exist per tier.
var TierPoolSize = map[Tier]int{
	Tier1: 16, Tier2: 15, Tier3: 13,
	Tier4: 11, Tier5: 9, Tier6: 7,
}

// CardStore provides access to card templates.
type CardStore interface {
	ByTemplateID(id string) *CardTemplate
	ByMaxTierTribes(maxTier Tier, tribes []Tribe) []*CardTemplate
	ByTierTribes(tier Tier, tribes []Tribe) []*CardTemplate
}

// CardPool manages a finite pool of cards shared across all players in a lobby.
type CardPool struct {
	cards      CardStore
	quantities map[string]int // template ID → available copies
}

// NewCardPool creates a new card pool with finite quantities per template.
func NewCardPool(cards CardStore) *CardPool {
	pool := &CardPool{
		cards:      cards,
		quantities: make(map[string]int),
	}

	for tier := Tier1; tier <= Tier6; tier++ {
		count := TierPoolSize[tier]
		for _, tmpl := range cards.ByTierTribes(tier, nil) {
			if !tmpl.IsSpell() {
				pool.quantities[tmpl.ID] = count
			}
		}
	}

	return pool
}

// Roll returns a random selection of cards for the shop, removing them from the pool.
func (p *CardPool) Roll(maxTier Tier, tribes []Tribe, count int) []Card {
	available := p.availableTemplates(p.cards.ByMaxTierTribes(maxTier, tribes))
	if len(available) == 0 {
		return nil
	}

	res := make([]Card, 0, count)

	for range count {
		if len(available) == 0 {
			break
		}
		idx := rand.IntN(len(available))
		tmpl := available[idx]
		res = append(res, NewMinion(tmpl))
		p.quantities[tmpl.ID]--

		if p.quantities[tmpl.ID] <= 0 {
			available = append(available[:idx], available[idx+1:]...)
		}
	}

	return res
}

// RollExactTier returns DiscoverCount random minions of the exact tier, removing them from the pool.
func (p *CardPool) RollExactTier(tier Tier, tribes []Tribe) []Card {
	available := p.availableTemplates(p.cards.ByTierTribes(tier, tribes))
	if len(available) == 0 {
		return nil
	}

	res := make([]Card, 0, DiscoverCount)

	for range DiscoverCount {
		if len(available) == 0 {
			break
		}
		idx := rand.IntN(len(available))
		tmpl := available[idx]
		res = append(res, NewMinion(tmpl))
		p.quantities[tmpl.ID]--

		if p.quantities[tmpl.ID] <= 0 {
			available = append(available[:idx], available[idx+1:]...)
		}
	}

	return res
}

// ReturnCard returns a card to the pool. Skips spells and golden minions.
func (p *CardPool) ReturnCard(c Card) {
	if c.IsMinion() && !c.IsGolden() {
		p.quantities[c.TemplateID()]++
	}
}

// ReturnCards returns multiple cards to the pool.
func (p *CardPool) ReturnCards(cards []Card) {
	for _, c := range cards {
		p.ReturnCard(c)
	}
}

// availableTemplates filters templates to only those with quantity > 0.
func (p *CardPool) availableTemplates(templates []*CardTemplate) []*CardTemplate {
	res := make([]*CardTemplate, 0, len(templates))
	for _, t := range templates {
		if p.quantities[t.ID] > 0 {
			res = append(res, t)
		}
	}
	return res
}
