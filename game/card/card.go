// Package cards contains all card templates for the game.
package card

import (
	"fmt"

	"github.com/ysomad/gigabg/game"
)

// Cards provides indexed access to card templates.
type Cards struct {
	templates   map[string]game.CardTemplate
	byKind      map[game.CardKind][]game.CardTemplate
	byKindTier  map[game.CardKind]map[game.Tier][]game.CardTemplate
	byTribe     map[game.Tribe][]game.CardTemplate
	byTier      map[game.Tier][]game.CardTemplate
	byTribeTier map[game.Tribe]map[game.Tier][]game.CardTemplate
}

// New loads and indexes all card templates.
// Returns an error if duplicate IDs are found or validation fails.
func New() (*Cards, error) {
	templates := make(map[string]game.CardTemplate)

	cardSets := []struct {
		tribe game.Tribe
		cards map[string]*template
	}{
		{game.TribeBeast, beasts()},
		{game.TribeDemon, demons()},
		{game.TribeDragon, dragons()},
		{game.TribeElemental, elementals()},
		{game.TribeMech, mechs()},
		{game.TribeMurloc, murlocs()},
		{game.TribeNaga, nagas()},
		{game.TribeNeutral, neutrals()},
		{game.TribePirate, pirates()},
		{game.TribeQuilboar, quilboars()},
		{game.TribeUndead, undeads()},
		{game.TribeAll, all()},
		{0, spells()}, // spells has no tribe
	}

	for _, set := range cardSets {
		for id, t := range set.cards {
			if _, ok := templates[id]; ok {
				return nil, fmt.Errorf("duplicate card ID %q in %s", id, set.tribe.String())
			}

			t.setID(id)
			t.setTribe(set.tribe)
			t.initGoldenDefaults()

			if err := t.validate(); err != nil {
				return nil, fmt.Errorf("%s: '%s' has invalid template: %w", set.tribe.String(), id, err)
			}

			templates[id] = t
		}
	}

	return newCards(templates), nil
}

func newCards(templates map[string]game.CardTemplate) *Cards {
	c := &Cards{
		templates:   templates,
		byKind:      make(map[game.CardKind][]game.CardTemplate),
		byKindTier:  make(map[game.CardKind]map[game.Tier][]game.CardTemplate),
		byTribe:     make(map[game.Tribe][]game.CardTemplate),
		byTier:      make(map[game.Tier][]game.CardTemplate),
		byTribeTier: make(map[game.Tribe]map[game.Tier][]game.CardTemplate),
	}

	for _, t := range templates {
		c.byKind[t.Kind()] = append(c.byKind[t.Kind()], t)

		if c.byKindTier[t.Kind()] == nil {
			c.byKindTier[t.Kind()] = make(map[game.Tier][]game.CardTemplate)
		}
		c.byKindTier[t.Kind()][t.Tier()] = append(c.byKindTier[t.Kind()][t.Tier()], t)

		c.byTribe[t.Tribe()] = append(c.byTribe[t.Tribe()], t)
		c.byTier[t.Tier()] = append(c.byTier[t.Tier()], t)

		if c.byTribeTier[t.Tribe()] == nil {
			c.byTribeTier[t.Tribe()] = make(map[game.Tier][]game.CardTemplate)
		}
		c.byTribeTier[t.Tribe()][t.Tier()] = append(c.byTribeTier[t.Tribe()][t.Tier()], t)
	}

	return c
}

// ByTemplateID returns a card template by ID.
func (c *Cards) ByTemplateID(id string) game.CardTemplate {
	return c.templates[id]
}

// ByKind returns all cards of the given kind.
func (c *Cards) ByKind(kind game.CardKind) []game.CardTemplate {
	return c.byKind[kind]
}

// ByKindTier returns all cards matching both kind and tier.
func (c *Cards) ByKindTier(kind game.CardKind, tier game.Tier) []game.CardTemplate {
	if m := c.byKindTier[kind]; m != nil {
		return m[tier]
	}
	return nil
}

// ByTribe returns all cards of the given tribe.
func (c *Cards) ByTribe(tribe game.Tribe) []game.CardTemplate {
	return c.byTribe[tribe]
}

// ByTier returns all cards of the given tier.
func (c *Cards) ByTier(tier game.Tier) []game.CardTemplate {
	return c.byTier[tier]
}

// ByTribeTier returns all cards matching both tribe and tier.
func (c *Cards) ByTribeTier(tribe game.Tribe, tier game.Tier) []game.CardTemplate {
	if m := c.byTribeTier[tribe]; m != nil {
		return m[tier]
	}
	return nil
}

// ByTierTribes returns minion cards of the exact tier.
// If tribes is empty, all tribes are included.
func (c *Cards) ByTierTribes(tier game.Tier, tribes []game.Tribe) []game.CardTemplate {
	if len(tribes) == 0 {
		return c.ByKindTier(game.CardKindMinion, tier)
	}

	var res []game.CardTemplate
	for _, tribe := range tribes {
		res = append(res, c.ByTribeTier(tribe, tier)...)
	}
	return res
}

// ByMaxTierTribes returns cards matching maxTier and tribes filters.
// If tribes is empty, all tribes are included.
func (c *Cards) ByMaxTierTribes(maxTier game.Tier, tribes []game.Tribe) []game.CardTemplate {
	if len(tribes) == 0 {
		var res []game.CardTemplate
		for tier := game.Tier1; tier <= maxTier; tier++ {
			res = append(res, c.byTier[tier]...)
		}
		return res
	}

	var res []game.CardTemplate
	for _, tribe := range tribes {
		for tier := game.Tier1; tier <= maxTier; tier++ {
			res = append(res, c.ByTribeTier(tribe, tier)...)
		}
	}
	return res
}
