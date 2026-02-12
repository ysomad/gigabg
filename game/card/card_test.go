// Package card contains all card templates for the game.
package card

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

func testTemplates(t *testing.T, tt ...game.CardTemplate) map[string]game.CardTemplate {
	t.Helper()
	m := make(map[string]game.CardTemplate, len(tt))
	for _, tmpl := range tt {
		m[tmpl.ID()] = tmpl
	}
	return m
}

func Test_newCatalog(t *testing.T) {
	t.Parallel()

	t1Beast := testMinion(t, "t1_beast", game.Tier1, game.TribeBeast)
	t1Demon := testMinion(t, "t1_demon", game.Tier1, game.TribeDemon)
	t2Beast := testMinion(t, "t2_beast", game.Tier2, game.TribeBeast)
	t1Spell := testSpell(t, "t1_spell", game.Tier1, 1)

	type args struct {
		templates map[string]game.CardTemplate
	}

	tests := []struct {
		name string
		args args
		want *Catalog
	}{
		{
			name: "empty",
			args: args{templates: map[string]game.CardTemplate{}},
			want: &Catalog{
				templates:       map[string]game.CardTemplate{},
				byKind:          map[game.CardKind][]game.CardTemplate{},
				byTribe:         map[game.Tribe][]game.CardTemplate{},
				byTier:          map[game.Tier][]game.CardTemplate{},
				byKindTierTribe: map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate{},
			},
		},
		{
			name: "ok",
			args: args{templates: testTemplates(t, t1Beast, t1Demon, t2Beast, t1Spell)},
			want: &Catalog{
				templates: testTemplates(t, t1Beast, t1Demon, t2Beast, t1Spell),
				byKind: map[game.CardKind][]game.CardTemplate{
					game.CardKindMinion: {t1Beast, t1Demon, t2Beast},
					game.CardKindSpell:  {t1Spell},
				},
				byTribe: map[game.Tribe][]game.CardTemplate{
					game.TribeBeast: {t1Beast, t2Beast},
					game.TribeDemon: {t1Demon},
					0:               {t1Spell},
				},
				byTier: map[game.Tier][]game.CardTemplate{
					game.Tier1: {t1Beast, t1Demon, t1Spell},
					game.Tier2: {t2Beast},
				},
				byKindTierTribe: map[game.CardKind]map[game.Tier]map[game.Tribe][]game.CardTemplate{
					game.CardKindMinion: {
						game.Tier1: {
							game.TribeBeast: {t1Beast},
							game.TribeDemon: {t1Demon},
						},
						game.Tier2: {
							game.TribeBeast: {t2Beast},
						},
					},
					game.CardKindSpell: {
						game.Tier1: {
							0: {t1Spell},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newCatalog(tt.args.templates))
		})
	}
}
