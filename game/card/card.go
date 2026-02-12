// Package card contains all card templates for the game.
package card

import (
	"fmt"

	"github.com/ysomad/gigabg/game"
)

var _ game.CardCatalog = (*Catalog)(nil)

// Catalog provides indexed access to card templates.
type Catalog struct {
	templates       map[string]game.CardTemplate
	byKind          map[game.CardKind][]game.CardTemplate
	byTribe         map[game.Tribe][]game.CardTemplate
	byTier          map[game.Tier][]game.CardTemplate
	byKindTierTribe map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate
}

// NewCatalog loads and indexes all card templates.
// Returns an error if duplicate IDs are found or validation fails.
func NewCatalog() (*Catalog, error) {
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

			// set default minion cost to minions
			if t.kind == game.CardKindMinion {
				t.cost = game.MinionCost
			}

			t.initGoldenDefaults()

			if err := t.validate(); err != nil {
				return nil, fmt.Errorf("%s: '%s' has invalid template: %w", set.tribe.String(), id, err)
			}

			templates[id] = t
		}
	}

	return newCatalog(templates), nil
}

func newCatalog(templates map[string]game.CardTemplate) *Catalog {
	c := &Catalog{
		templates:       templates,
		byKind:          make(map[game.CardKind][]game.CardTemplate),
		byTribe:         make(map[game.Tribe][]game.CardTemplate),
		byTier:          make(map[game.Tier][]game.CardTemplate),
		byKindTierTribe: make(map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate),
	}

	for _, t := range templates {
		c.byKind[t.Kind()] = append(c.byKind[t.Kind()], t)
		c.byTribe[t.Tribe()] = append(c.byTribe[t.Tribe()], t)
		c.byTier[t.Tier()] = append(c.byTier[t.Tier()], t)

		if c.byKindTierTribe[t.Kind()] == nil {
			c.byKindTierTribe[t.Kind()] = make(map[game.Tier]map[game.Tribe][]game.CardTemplate)
		}

		if c.byKindTierTribe[t.Kind()][t.Tier()] == nil {
			c.byKindTierTribe[t.Kind()][t.Tier()] = make(map[game.Tribe][]game.CardTemplate)
		}

		c.byKindTierTribe[t.Kind()][t.Tier()][t.Tribe()] = append(c.byKindTierTribe[t.Kind()][t.Tier()][t.Tribe()], t)
	}

	return c
}

// ByTemplateID returns a card template by ID.
func (c *Catalog) ByTemplateID(id string) game.CardTemplate {
	return c.templates[id]
}

// ByKind returns all cards of the given kind.
func (c *Catalog) ByKind(kind game.CardKind) []game.CardTemplate {
	return c.byKind[kind]
}

// ByTribe returns all cards of the given tribe.
func (c *Catalog) ByTribe(tribe game.Tribe) []game.CardTemplate {
	return c.byTribe[tribe]
}

// ByTier returns all cards of the given tier.
func (c *Catalog) ByTier(tier game.Tier) []game.CardTemplate {
	return c.byTier[tier]
}

// ByKindTierTribe returns cards matching kind, tier, and tribe.
func (c *Catalog) ByKindTierTribe(kind game.CardKind, tier game.Tier, tribe game.Tribe) []game.CardTemplate {
	if m := c.byKindTierTribe[kind]; m != nil {
		if m2 := m[tier]; m2 != nil {
			return m2[tribe]
		}
	}
	return nil
}
