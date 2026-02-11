// Package cards contains all card templates for the game.
package cards

import (
	"fmt"

	"github.com/ysomad/gigabg/internal/game"
)

// Cards provides indexed access to card templates.
type Cards struct {
	templates   map[string]*game.CardTemplate
	byKind      map[game.CardKind][]*game.CardTemplate
	byKindTier  map[game.CardKind]map[game.Tier][]*game.CardTemplate
	byTribe     map[game.Tribe][]*game.CardTemplate
	byTier      map[game.Tier][]*game.CardTemplate
	byTribeTier map[game.Tribe]map[game.Tier][]*game.CardTemplate
}

// New loads and indexes all card templates.
// Returns an error if duplicate IDs are found or validation fails.
func New() (*Cards, error) {
	templates := make(map[string]*game.CardTemplate)

	cardSets := []struct {
		name  string
		cards map[string]*game.CardTemplate
	}{
		{"Beast", Beast()},
		{"Demon", Demon()},
		{"Dragon", Dragon()},
		{"Elemental", Elemental()},
		{"Mech", Mech()},
		{"Murloc", Murloc()},
		{"Naga", Naga()},
		{"Neutral", Neutral()},
		{"Pirate", Pirate()},
		{"Quilboar", Quilboar()},
		{"Undead", Undead()},
		{"Spells", Spells()},
	}

	for _, set := range cardSets {
		for id, template := range set.cards {
			if _, ok := templates[id]; ok {
				return nil, fmt.Errorf("duplicate card ID %q in %s", id, set.name)
			}
			if err := template.Validate(); err != nil {
				return nil, fmt.Errorf("%s: '%s' has invalid template: %w", set.name, id, err)
			}
			template.ID = id
			templates[id] = template
		}
	}

	return _new(templates), nil
}

func _new(templates map[string]*game.CardTemplate) *Cards {
	c := &Cards{
		templates:   templates,
		byKind:      make(map[game.CardKind][]*game.CardTemplate),
		byKindTier:  make(map[game.CardKind]map[game.Tier][]*game.CardTemplate),
		byTribe:     make(map[game.Tribe][]*game.CardTemplate),
		byTier:      make(map[game.Tier][]*game.CardTemplate),
		byTribeTier: make(map[game.Tribe]map[game.Tier][]*game.CardTemplate),
	}

	for _, t := range templates {
		// Default unset Kind to Minion so existing templates are indexed correctly
		if t.Kind == 0 {
			t.Kind = game.CardKindMinion
		}

		// Auto-populate golden defaults for minions
		if t.Kind == game.CardKindMinion {
			initGoldenDefaults(t)
		}

		c.byKind[t.Kind] = append(c.byKind[t.Kind], t)

		if c.byKindTier[t.Kind] == nil {
			c.byKindTier[t.Kind] = make(map[game.Tier][]*game.CardTemplate)
		}
		c.byKindTier[t.Kind][t.Tier] = append(c.byKindTier[t.Kind][t.Tier], t)

		c.byTribe[t.Tribe] = append(c.byTribe[t.Tribe], t)
		c.byTier[t.Tier] = append(c.byTier[t.Tier], t)

		if c.byTribeTier[t.Tribe] == nil {
			c.byTribeTier[t.Tribe] = make(map[game.Tier][]*game.CardTemplate)
		}
		c.byTribeTier[t.Tribe][t.Tier] = append(c.byTribeTier[t.Tribe][t.Tier], t)
	}

	return c
}

// ByTemplateID returns a card template by ID.
func (c *Cards) ByTemplateID(id string) *game.CardTemplate {
	return c.templates[id]
}

// ByKind returns all cards of the given kind.
func (c *Cards) ByKind(kind game.CardKind) []*game.CardTemplate {
	return c.byKind[kind]
}

// ByKindTier returns all cards matching both kind and tier.
func (c *Cards) ByKindTier(kind game.CardKind, tier game.Tier) []*game.CardTemplate {
	if m := c.byKindTier[kind]; m != nil {
		return m[tier]
	}
	return nil
}

// ByTribe returns all cards of the given tribe.
func (c *Cards) ByTribe(tribe game.Tribe) []*game.CardTemplate {
	return c.byTribe[tribe]
}

// ByTier returns all cards of the given tier.
func (c *Cards) ByTier(tier game.Tier) []*game.CardTemplate {
	return c.byTier[tier]
}

// ByTribeTier returns all cards matching both tribe and tier.
func (c *Cards) ByTribeTier(tribe game.Tribe, tier game.Tier) []*game.CardTemplate {
	if m := c.byTribeTier[tribe]; m != nil {
		return m[tier]
	}
	return nil
}

// ByTierTribes returns minion cards of the exact tier.
// If tribes is empty, all tribes are included.
func (c *Cards) ByTierTribes(tier game.Tier, tribes []game.Tribe) []*game.CardTemplate {
	if len(tribes) == 0 {
		return c.ByKindTier(game.CardKindMinion, tier)
	}

	var res []*game.CardTemplate
	for _, tribe := range tribes {
		res = append(res, c.ByTribeTier(tribe, tier)...)
	}
	return res
}

// ByMaxTierTribes returns cards matching maxTier and tribes filters.
// If tribes is empty, all tribes are included.
func (c *Cards) ByMaxTierTribes(maxTier game.Tier, tribes []game.Tribe) []*game.CardTemplate {
	if len(tribes) == 0 {
		var res []*game.CardTemplate
		for tier := game.Tier1; tier <= maxTier; tier++ {
			res = append(res, c.byTier[tier]...)
		}
		return res
	}

	var res []*game.CardTemplate
	for _, tribe := range tribes {
		for tier := game.Tier1; tier <= maxTier; tier++ {
			res = append(res, c.ByTribeTier(tribe, tier)...)
		}
	}
	return res
}

// initGoldenDefaults fills in Golden with 2x defaults for any unset fields.
// Cards that define custom Golden overrides keep their values.
func initGoldenDefaults(t *game.CardTemplate) {
	if t.Golden == nil {
		t.Golden = &game.GoldenStats{}
	}
	g := t.Golden

	if g.Attack == 0 {
		g.Attack = t.Attack * 2
	}
	if g.Health == 0 {
		g.Health = t.Health * 2
	}
	if g.Description == "" {
		g.Description = t.Description
	}
	if g.Battlecry == nil && t.Battlecry != nil {
		g.Battlecry = t.Battlecry.Double()
	}
	if g.Deathrattle == nil && t.Deathrattle != nil {
		g.Deathrattle = t.Deathrattle.Double()
	}
	if g.Avenge == nil && t.Avenge != nil {
		g.Avenge = game.DoubleAvenge(t.Avenge)
	}
	if g.StartOfCombat == nil && t.StartOfCombat != nil {
		g.StartOfCombat = t.StartOfCombat.Double()
	}
	if g.StartOfTurn == nil && t.StartOfTurn != nil {
		g.StartOfTurn = t.StartOfTurn.Double()
	}
	if g.EndOfTurn == nil && t.EndOfTurn != nil {
		g.EndOfTurn = t.EndOfTurn.Double()
	}
}
