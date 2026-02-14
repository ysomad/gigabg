package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ysomad/gigabg/game"
)

func Test_template_validate(t *testing.T) {
	t.Parallel()
	type fields struct {
		name     string
		kind     game.CardKind
		tier     game.Tier
		cost     int
		attack   int
		health   int
		tribes   game.Tribes
		keywords game.Keywords
	}
	tests := []struct {
		name      string
		fields    fields
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "valid_minion",
			fields: fields{
				name:   "Imp",
				kind:   game.CardKindMinion,
				tier:   game.Tier1,
				cost:   game.MinionCost,
				attack: 1,
				health: 1,
			},
			assertion: assert.NoError,
		},
		{
			name:      "minion_no_cost",
			fields:    fields{name: "Imp", kind: game.CardKindMinion, tier: game.Tier1, health: 1},
			assertion: assert.Error,
		},
		{
			name:      "valid_spell",
			fields:    fields{name: "Bolt", kind: game.CardKindSpell, tier: game.Tier1, cost: 1},
			assertion: assert.NoError,
		},
		{
			name:      "spell_no_tier", // spell without tier for example is triple reward
			fields:    fields{name: "Bolt", kind: game.CardKindSpell},
			assertion: assert.NoError,
		},
		{
			name:      "empty_name",
			fields:    fields{kind: game.CardKindMinion, tier: game.Tier1, health: 1},
			assertion: assert.Error,
		},
		{
			name:      "minion_invalid_tier",
			fields:    fields{name: "Imp", kind: game.CardKindMinion, health: 1},
			assertion: assert.Error,
		},
		{
			name: "minion_negative_attack",
			fields: fields{
				name:   "Imp",
				kind:   game.CardKindMinion,
				tier:   game.Tier1,
				cost:   game.MinionCost,
				attack: -1,
				health: 1,
			},
			assertion: assert.Error,
		},
		{
			name:      "minion_zero_health",
			fields:    fields{name: "Imp", kind: game.CardKindMinion, tier: game.Tier1, cost: game.MinionCost},
			assertion: assert.Error,
		},
		{
			name:      "spell_no_cost",
			fields:    fields{name: "Bolt", kind: game.CardKindSpell, tier: game.Tier2},
			assertion: assert.Error,
		},
		{
			name: "magnetic_mech",
			fields: fields{
				name: "Bot", kind: game.CardKindMinion, tier: game.Tier1, cost: game.MinionCost, health: 1,
				tribes: game.NewTribes(game.TribeMech), keywords: game.NewKeywords(game.KeywordMagnetic),
			},
			assertion: assert.NoError,
		},
		{
			name: "magnetic_mech_undead",
			fields: fields{
				name: "Bot", kind: game.CardKindMinion, tier: game.Tier1, cost: game.MinionCost, health: 1,
				tribes: game.NewTribes(game.TribeMech, game.TribeUndead), keywords: game.NewKeywords(game.KeywordMagnetic),
			},
			assertion: assert.NoError,
		},
		{
			name: "magnetic_not_mech",
			fields: fields{
				name: "Bot", kind: game.CardKindMinion, tier: game.Tier1, cost: game.MinionCost, health: 1,
				tribes: game.NewTribes(game.TribeBeast), keywords: game.NewKeywords(game.KeywordMagnetic),
			},
			assertion: assert.Error,
		},
		{
			name: "magnetic_neutral",
			fields: fields{
				name: "Bot", kind: game.CardKindMinion, tier: game.Tier1, cost: game.MinionCost, health: 1,
				keywords: game.NewKeywords(game.KeywordMagnetic),
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tr := &template{
				name:     tt.fields.name,
				kind:     tt.fields.kind,
				tier:     tt.fields.tier,
				cost:     tt.fields.cost,
				attack:   tt.fields.attack,
				health:   tt.fields.health,
				_tribes:  tt.fields.tribes,
				keywords: tt.fields.keywords,
			}
			tt.assertion(t, tr.validate())
		})
	}
}
