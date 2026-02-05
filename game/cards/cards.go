// Package cards contains all card templates for the game.
package cards

import (
	"fmt"

	"github.com/ysomad/gigabg/game"
)

// Cards provides indexed access to card templates.
type Cards struct {
	templates   map[string]*game.CardTemplate
	byTribe     map[game.Tribe][]*game.CardTemplate
	byTier      map[game.Tier][]*game.CardTemplate
	byTribeTier map[game.Tribe]map[game.Tier][]*game.CardTemplate
}

// New loads and indexes all card templates.
// Returns an error if duplicate IDs are found or validation fails.
func New() (*Cards, error) {
	templates := make(map[string]*game.CardTemplate)

	tribeMaps := []struct {
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
	}

	for _, tribe := range tribeMaps {
		for id, template := range tribe.cards {
			if existing, ok := templates[id]; ok {
				return nil, fmt.Errorf("duplicate card ID %q: found in %s and %s", id, existing.Tribe, tribe.name)
			}
			if err := template.Validate(); err != nil {
				return nil, fmt.Errorf(
					"%s: '%s' (%d) has invalid template: %w",
					template.Tribe.String(),
					id,
					template.Tier,
					err,
				)
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
		byTribe:     make(map[game.Tribe][]*game.CardTemplate),
		byTier:      make(map[game.Tier][]*game.CardTemplate),
		byTribeTier: make(map[game.Tribe]map[game.Tier][]*game.CardTemplate),
	}

	for _, t := range templates {
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

// Query returns cards matching maxTier and tribes filters.
// If tribes is empty, all tribes are included.
func (c *Cards) Query(maxTier game.Tier, tribes []game.Tribe) []*game.CardTemplate {
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
