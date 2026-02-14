// Package catalog contains all card templates for the game.
package catalog

import (
	"fmt"

	"github.com/ysomad/gigabg/game"
)

var _ game.CardCatalog = (*Catalog)(nil)

type cardSet struct {
	tribes game.Tribes
	cards  map[string]*template
}

// Catalog provides indexed access to card templates.
// All cards (shop + tokens) are accessible via ByTemplateID.
// Index-based methods (ByKind, ByTier, ByTribe, ByKindTierTribe) return only shop cards.
type Catalog struct {
	all             map[string]game.CardTemplate // all cards (shop + tokens)
	byKind          map[game.CardKind][]game.CardTemplate
	byTribe         map[game.Tribe][]game.CardTemplate
	byTier          map[game.Tier][]game.CardTemplate
	byKindTierTribe map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate
}

// New loads and indexes all card templates.
// Returns an error if duplicate IDs are found or validation fails.
func New() (*Catalog, error) {
	shopSets := []cardSet{
		{game.NewTribes(game.TribeBeast), beasts()},
		{game.NewTribes(game.TribeDemon), demons()},
		{game.NewTribes(game.TribeDragon), dragons()},
		{game.NewTribes(game.TribeElemental), elementals()},
		{game.NewTribes(game.TribeMech), mechs()},
		{game.NewTribes(game.TribeMurloc), murlocs()},
		{game.NewTribes(game.TribeNaga), nagas()},
		{0, neutrals()},
		{game.NewTribes(game.TribeQuilboar), quilboars()},
		{game.NewTribes(game.TribeUndead), undeads()},
		{game.NewTribes(game.TribeUndead, game.TribeMech), map[string]*template{
			"reanimated_mech": {
				name:     "Reanimated Mech",
				tier:     game.Tier4,
				attack:   4,
				health:   5,
				keywords: game.NewKeywords(game.KeywordReborn, game.KeywordMagnetic),
			},
		}},
		{game.TribeAll, all()},
		{0, spells()},
	}

	tokenSets := []cardSet{
		{game.NewTribes(game.TribeDemon), demonTokens()},
		{game.NewTribes(game.TribeUndead), undeadTokens()},
		{0, spellTokens()},
	}

	c := &Catalog{
		all:             make(map[string]game.CardTemplate),
		byKind:          make(map[game.CardKind][]game.CardTemplate),
		byTribe:         make(map[game.Tribe][]game.CardTemplate),
		byTier:          make(map[game.Tier][]game.CardTemplate),
		byKindTierTribe: make(map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate),
	}

	// Shop cards go into all + indexes.
	for _, set := range shopSets {
		for id, t := range set.cards {
			if err := c.initTemplate(id, t, set.tribes); err != nil {
				return nil, err
			}

			c.all[id] = t
			c.index(t)
		}
	}

	// Token cards go into all only (not indexed, never offered in shop).
	for _, set := range tokenSets {
		for id, t := range set.cards {
			if err := c.initTemplate(id, t, set.tribes); err != nil {
				return nil, err
			}

			c.all[id] = t
		}
	}

	return c, nil
}

func (c *Catalog) initTemplate(id string, t *template, tribes game.Tribes) error {
	if _, ok := c.all[id]; ok {
		return fmt.Errorf("duplicate card ID %q", id)
	}

	t.setID(id)
	t.setTribes(tribes)

	if t.kind == game.CardKindMinion {
		t.cost = game.MinionCost
	}

	t.initGoldenDefaults()

	if err := t.validate(); err != nil {
		return fmt.Errorf("'%s' has invalid template: %w", id, err)
	}

	return nil
}

func (c *Catalog) index(t game.CardTemplate) {
	c.byKind[t.Kind()] = append(c.byKind[t.Kind()], t)
	c.byTier[t.Tier()] = append(c.byTier[t.Tier()], t)

	if c.byKindTierTribe[t.Kind()] == nil {
		c.byKindTierTribe[t.Kind()] = make(map[game.Tier]map[game.Tribe][]game.CardTemplate)
	}
	if c.byKindTierTribe[t.Kind()][t.Tier()] == nil {
		c.byKindTierTribe[t.Kind()][t.Tier()] = make(map[game.Tribe][]game.CardTemplate)
	}

	tribes := t.Tribes().List()
	if len(tribes) == 0 {
		tribes = []game.Tribe{game.TribeNeutral}
	}
	for _, tribe := range tribes {
		c.byTribe[tribe] = append(c.byTribe[tribe], t)
		c.byKindTierTribe[t.Kind()][t.Tier()][tribe] = append(c.byKindTierTribe[t.Kind()][t.Tier()][tribe], t)
	}
}

// ByTemplateID returns a card template by ID, searching all cards (shop + tokens).
func (c *Catalog) ByTemplateID(id string) game.CardTemplate {
	return c.all[id]
}

// ByKind returns all shop cards of the given kind.
func (c *Catalog) ByKind(kind game.CardKind) []game.CardTemplate {
	return c.byKind[kind]
}

// ByTribe returns all shop cards of the given tribe.
func (c *Catalog) ByTribe(tribe game.Tribe) []game.CardTemplate {
	return c.byTribe[tribe]
}

// ByTier returns all shop cards of the given tier.
func (c *Catalog) ByTier(tier game.Tier) []game.CardTemplate {
	return c.byTier[tier]
}

// ByKindTierTribe returns shop cards matching kind, tier, and tribe.
// Zero tribe means all tribes.
func (c *Catalog) ByKindTierTribe(kind game.CardKind, tier game.Tier, tribe game.Tribe) []game.CardTemplate {
	m := c.byKindTierTribe[kind]
	if m == nil {
		return nil
	}
	m2 := m[tier]
	if m2 == nil {
		return nil
	}
	if tribe != 0 {
		return m2[tribe]
	}
	var res []game.CardTemplate
	for _, t := range m2 {
		res = append(res, t...)
	}
	return res
}
