package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysomad/gigabg/game"
)

func testMinion(t *testing.T, id string, tier game.Tier, tribes game.Tribes) *template {
	t.Helper()
	return &template{_id: id, name: id, kind: game.CardKindMinion, tier: tier, _tribes: tribes, health: 1}
}

func testSpell(t *testing.T, id string, tier game.Tier, cost int) *template {
	t.Helper()
	return &template{_id: id, name: id, kind: game.CardKindSpell, tier: tier, cost: cost}
}

func TestCatalog_ByTemplateID(t *testing.T) {
	t.Parallel()

	t1Beast := testMinion(t, "t1_beast", game.Tier1, game.NewTribes(game.TribeBeast))
	token := testMinion(t, "token", game.Tier1, game.NewTribes(game.TribeDemon))

	c := &Catalog{
		all:             map[string]game.CardTemplate{"t1_beast": t1Beast, "token": token},
		byKind:          make(map[game.CardKind][]game.CardTemplate),
		byTribe:         make(map[game.Tribe][]game.CardTemplate),
		byTier:          make(map[game.Tier][]game.CardTemplate),
		byKindTierTribe: make(map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate),
	}
	c.index(t1Beast)

	tests := []struct {
		name string
		id   string
		want game.CardTemplate
	}{
		{"shop card", "t1_beast", t1Beast},
		{"token card", "token", token},
		{"nonexistent", "missing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, c.ByTemplateID(tt.id))
		})
	}
}

func TestCatalog_ByKindTierTribe(t *testing.T) {
	t.Parallel()

	t1Beast := testMinion(t, "t1_beast", game.Tier1, game.NewTribes(game.TribeBeast))
	t1Demon := testMinion(t, "t1_demon", game.Tier1, game.NewTribes(game.TribeDemon))
	t2Beast := testMinion(t, "t2_beast", game.Tier2, game.NewTribes(game.TribeBeast))
	t1Spell := testSpell(t, "t1_spell", game.Tier1, 1)
	token := testMinion(t, "token", game.Tier1, game.NewTribes(game.TribeDemon))

	c := &Catalog{
		all:             make(map[string]game.CardTemplate),
		byKind:          make(map[game.CardKind][]game.CardTemplate),
		byTribe:         make(map[game.Tribe][]game.CardTemplate),
		byTier:          make(map[game.Tier][]game.CardTemplate),
		byKindTierTribe: make(map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate),
	}

	for _, tmpl := range []game.CardTemplate{t1Beast, t1Demon, t2Beast, t1Spell} {
		c.all[tmpl.ID()] = tmpl
		c.index(tmpl)
	}
	c.all["token"] = token // token: not indexed

	tests := []struct {
		name  string
		kind  game.CardKind
		tier  game.Tier
		tribe game.Tribe
		want  []game.CardTemplate
	}{
		{"exact match", game.CardKindMinion, game.Tier1, game.TribeBeast, []game.CardTemplate{t1Beast}},
		{"tribe zero returns all tribes", game.CardKindMinion, game.Tier1, 0, []game.CardTemplate{t1Beast, t1Demon}},
		{"tier2 beast", game.CardKindMinion, game.Tier2, game.TribeBeast, []game.CardTemplate{t2Beast}},
		{"spell tier1", game.CardKindSpell, game.Tier1, 0, []game.CardTemplate{t1Spell}},
		{"no matches", game.CardKindMinion, game.Tier6, game.TribeBeast, nil},
		{"token excluded from index", game.CardKindMinion, game.Tier1, game.TribeDemon, []game.CardTemplate{t1Demon}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ByKindTierTribe(tt.kind, tt.tier, tt.tribe)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
