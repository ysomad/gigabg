package game

import (
	"math"
	"math/rand/v2"
)

// ratio of pool card copies are calculted from https://hearthstone.wiki.gg/wiki/Battlegrounds.

// minionPoolRatio returns copies-per-template ratio for each tier,
// derived from Hearthstone Battlegrounds minion pool data.
func minionPoolRatio() map[Tier]float64 {
	return map[Tier]float64{
		Tier1: 0.1136,
		Tier2: 0.0534,
		Tier3: 0.0216,
		Tier4: 0.0180,
		Tier5: 0.0189,
		Tier6: 0.0157,
	}
}

// spellPoolRatio returns copies-per-template ratio for each tier,
// derived from Hearthstone Battlegrounds spell pool data.
func spellPoolRatio() map[Tier]float64 {
	return map[Tier]float64{
		Tier1: 0.1136,
		Tier2: 0.4667,
		Tier3: 0.1216,
		Tier4: 0.1134,
		Tier5: 0.2000,
		Tier6: 0.1842,
	}
}

// poolCopies calculates copies per template.
func poolCopies(numTemplates int, ratios map[Tier]float64, tier Tier) int {
	if numTemplates <= 0 {
		return 0
	}
	return max(1, int(math.Round(float64(numTemplates)*ratios[tier])))
}

// CardCatalog provides access to card templates.
type CardCatalog interface {
	ByTemplateID(id string) CardTemplate
	ByKind(kind CardKind) []CardTemplate
	ByTribe(tribe Tribe) []CardTemplate
	ByTier(tier Tier) []CardTemplate
	ByKindTierTribe(kind CardKind, tier Tier, tribe Tribe) []CardTemplate
}

// CardPool manages a finite pool of cards shared across all players in a lobby.
type CardPool struct {
	cards      CardCatalog
	quantities map[string]int          // template ID → available copies
	byTier     map[Tier][]CardTemplate // pool templates indexed by tier
}

// NewCardPool creates a new card pool with finite quantities per template.
func NewCardPool(cards CardCatalog) *CardPool {
	pool := &CardPool{
		cards:      cards,
		quantities: make(map[string]int),
		byTier:     make(map[Tier][]CardTemplate),
	}

	minionRatio := minionPoolRatio()
	spellRatio := spellPoolRatio()

	for tier := Tier1; tier <= Tier6; tier++ {
		minions := cards.ByKindTierTribe(CardKindMinion, tier, 0)
		minionCopies := poolCopies(len(minions), minionRatio, tier)
		for _, tmpl := range minions {
			pool.quantities[tmpl.ID()] = minionCopies
			pool.byTier[tier] = append(pool.byTier[tier], tmpl)
		}

		spells := cards.ByKindTierTribe(CardKindSpell, tier, 0)
		spellCopies := poolCopies(len(spells), spellRatio, tier)
		for _, tmpl := range spells {
			pool.quantities[tmpl.ID()] = spellCopies
			pool.byTier[tier] = append(pool.byTier[tier], tmpl)
		}
	}

	return pool
}

// Roll returns a random selection of cards for the shop, removing them from the pool.
func (p *CardPool) Roll(maxTier Tier, count int) []Card {
	var templates []CardTemplate
	for tier := Tier1; tier <= maxTier; tier++ {
		templates = append(templates, p.byTier[tier]...)
	}
	return p.roll(templates, count)
}

// RollDiscover returns discoverCount random cards of the exact tier, removing them from the pool.
// If tribe is 0, cards from all tribes are included. If kind is 0, all kinds are included.
func (p *CardPool) RollDiscover(tier Tier, tribe Tribe, kind CardKind) []Card {
	templates := p.filter(p.byTier[tier], tribe, kind)
	return p.roll(templates, discoverCount)
}

// ReturnCard returns a card to the pool. Skips golden cards and cards not in the pool.
func (p *CardPool) ReturnCard(c Card) {
	if c.IsGolden() {
		return
	}
	id := c.Template().ID()
	if _, ok := p.quantities[id]; ok {
		p.quantities[id]++
	}
}

// ReturnCards returns multiple cards to the pool.
func (p *CardPool) ReturnCards(cards []Card) {
	for _, c := range cards {
		p.ReturnCard(c)
	}
}

func (p *CardPool) roll(templates []CardTemplate, count int) []Card {
	available := p.available(templates)
	if len(available) == 0 {
		return nil
	}

	res := make([]Card, 0, count)

	for range count {
		if len(available) == 0 {
			break
		}
		idx := rand.IntN(len(available)) //nolint:gosec // game logic, not crypto
		tmpl := available[idx]
		res = append(res, NewCard(tmpl))
		p.quantities[tmpl.ID()]--

		if p.quantities[tmpl.ID()] <= 0 {
			available = append(available[:idx], available[idx+1:]...)
		}
	}

	return res
}

// available filters templates to only those with quantity > 0.
func (p *CardPool) available(templates []CardTemplate) []CardTemplate {
	res := make([]CardTemplate, 0, len(templates))
	for _, t := range templates {
		if p.quantities[t.ID()] > 0 {
			res = append(res, t)
		}
	}
	return res
}

// filter returns templates matching the given tribe and kind. Zero values mean "any".
func (p *CardPool) filter(templates []CardTemplate, tribe Tribe, kind CardKind) []CardTemplate {
	if tribe == 0 && kind == 0 {
		return templates
	}

	filtered := make([]CardTemplate, 0, len(templates))
	for _, t := range templates {
		if tribe != 0 && t.Tribe() != tribe {
			continue
		}
		if kind != 0 && t.Kind() != kind {
			continue
		}
		filtered = append(filtered, t)
	}
	return filtered
}
