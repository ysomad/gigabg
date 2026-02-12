package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysomad/gigabg/game"
)

func testMinion(t *testing.T, id string, tier game.Tier, tribe game.Tribe) *template {
	t.Helper()
	return &template{_id: id, name: id, kind: game.CardKindMinion, tier: tier, _tribe: tribe, health: 1}
}

func testSpell(t *testing.T, id string, tier game.Tier, cost int) *template {
	t.Helper()
	return &template{_id: id, name: id, kind: game.CardKindSpell, tier: tier, cost: cost}
}

func Test_newCatalog_empty(t *testing.T) {
	t.Parallel()

	c := newCatalog(map[string]game.CardTemplate{})

	assert.Empty(t, c.templates)
	assert.Empty(t, c.byKind)
	assert.Empty(t, c.byTribe)
	assert.Empty(t, c.byTier)
	assert.Empty(t, c.byKindTierTribe)
}

func Test_newCatalog(t *testing.T) {
	t.Parallel()

	t1Beast := testMinion(t, "t1_beast", game.Tier1, game.TribeBeast)
	t1Demon := testMinion(t, "t1_demon", game.Tier1, game.TribeDemon)
	t2Beast := testMinion(t, "t2_beast", game.Tier2, game.TribeBeast)
	t1Spell := testSpell(t, "t1_spell", game.Tier1, 1)

	templates := map[string]game.CardTemplate{
		"t1_beast": t1Beast,
		"t1_demon": t1Demon,
		"t2_beast": t2Beast,
		"t1_spell": t1Spell,
	}

	c := newCatalog(templates)

	// templates indexed by ID
	assert.Equal(t, game.CardTemplate(t1Beast), c.ByTemplateID("t1_beast"))
	assert.Equal(t, game.CardTemplate(t1Demon), c.ByTemplateID("t1_demon"))
	assert.Equal(t, game.CardTemplate(t2Beast), c.ByTemplateID("t2_beast"))
	assert.Equal(t, game.CardTemplate(t1Spell), c.ByTemplateID("t1_spell"))
	assert.Nil(t, c.ByTemplateID("nonexistent"))

	// by kind
	assert.ElementsMatch(t, []game.CardTemplate{t1Beast, t1Demon, t2Beast}, c.ByKind(game.CardKindMinion))
	assert.ElementsMatch(t, []game.CardTemplate{t1Spell}, c.ByKind(game.CardKindSpell))

	// by tribe
	assert.ElementsMatch(t, []game.CardTemplate{t1Beast, t2Beast}, c.ByTribe(game.TribeBeast))
	assert.ElementsMatch(t, []game.CardTemplate{t1Demon}, c.ByTribe(game.TribeDemon))

	// by tier
	assert.ElementsMatch(t, []game.CardTemplate{t1Beast, t1Demon, t1Spell}, c.ByTier(game.Tier1))
	assert.ElementsMatch(t, []game.CardTemplate{t2Beast}, c.ByTier(game.Tier2))

	// by kind+tier+tribe (exact)
	assert.ElementsMatch(t, []game.CardTemplate{t1Beast}, c.ByKindTierTribe(game.CardKindMinion, game.Tier1, game.TribeBeast))
	assert.ElementsMatch(t, []game.CardTemplate{t1Demon}, c.ByKindTierTribe(game.CardKindMinion, game.Tier1, game.TribeDemon))
	assert.ElementsMatch(t, []game.CardTemplate{t2Beast}, c.ByKindTierTribe(game.CardKindMinion, game.Tier2, game.TribeBeast))
	assert.ElementsMatch(t, []game.CardTemplate{t1Spell}, c.ByKindTierTribe(game.CardKindSpell, game.Tier1, 0))

	// by kind+tier, tribe=0 means all tribes
	assert.ElementsMatch(t, []game.CardTemplate{t1Beast, t1Demon}, c.ByKindTierTribe(game.CardKindMinion, game.Tier1, 0))
	assert.ElementsMatch(t, []game.CardTemplate{t2Beast}, c.ByKindTierTribe(game.CardKindMinion, game.Tier2, 0))

	// no matches
	assert.Empty(t, c.ByKindTierTribe(game.CardKindMinion, game.Tier6, game.TribeBeast))
}
