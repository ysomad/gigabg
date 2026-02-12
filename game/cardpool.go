package game

import (
	"log/slog"
	"math/rand/v2"
)

// Copies per template by tier from https://hearthstone.wiki.gg/wiki/Battlegrounds.

func minionCopies() map[Tier]int {
	return map[Tier]int{
		Tier1: 15,
		Tier2: 15,
		Tier3: 13,
		Tier4: 11,
		Tier5: 9,
		Tier6: 7,
	}
}

func spellCopies() map[Tier]int {
	return map[Tier]int{
		Tier1: 5,
		Tier2: 7,
		Tier3: 9,
		Tier4: 11,
		Tier5: 9,
		Tier6: 7,
	}
}

// scaleCopies scales base 8-player copies for the actual player count.
func scaleCopies(base, players int) int {
	return max(1, base*players/MaxPlayers)
}

// CardCatalog provides access to card templates.
type CardCatalog interface {
	ByTemplateID(id string) CardTemplate
	ByKindTierTribe(kind CardKind, tier Tier, tribe Tribe) []CardTemplate
}

// CardPool manages a finite pool of cards shared across all players in a lobby.
type CardPool struct {
	cards      CardCatalog
	quantities map[string]int          // template ID → available copies
	byTier     map[Tier][]CardTemplate // pool templates indexed by tier
}

// NewCardPool creates a new card pool with finite quantities per template.
// Copies are scaled from 8-player base values proportionally to players.
func NewCardPool(cards CardCatalog, players int) *CardPool {
	pool := &CardPool{
		cards:      cards,
		quantities: make(map[string]int),
		byTier:     make(map[Tier][]CardTemplate),
	}

	mc := minionCopies()
	sc := spellCopies()

	for tier := Tier1; tier <= Tier6; tier++ {
		mCopies := scaleCopies(mc[tier], players)
		sCopies := scaleCopies(sc[tier], players)

		minions := cards.ByKindTierTribe(CardKindMinion, tier, 0)
		for _, tmpl := range minions {
			pool.quantities[tmpl.ID()] = mCopies
			pool.byTier[tier] = append(pool.byTier[tier], tmpl)
		}

		spells := cards.ByKindTierTribe(CardKindSpell, tier, 0)
		for _, tmpl := range spells {
			pool.quantities[tmpl.ID()] = sCopies
			pool.byTier[tier] = append(pool.byTier[tier], tmpl)
		}

		slog.Debug("card pool",
			"tier", tier,
			"minions", len(minions),
			"minion_copies", mCopies,
			"spells", len(spells),
			"spell_copies", sCopies,
		)
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

// ByTemplateID returns a card template by ID from the catalog.
func (p *CardPool) ByTemplateID(id string) CardTemplate {
	return p.cards.ByTemplateID(id)
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
