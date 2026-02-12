package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_poolCopies(t *testing.T) {
	t.Parallel()
	type args struct {
		ratios       map[Tier]float64
		numTemplates map[Tier]int
	}
	tests := []struct {
		name string
		args args
		want map[Tier]int
	}{
		{
			name: "minions",
			args: args{
				ratios: minionPoolRatio(),
				numTemplates: map[Tier]int{
					Tier1: 132,
					Tier2: 281,
					Tier3: 603,
					Tier4: 610,
					Tier5: 476,
					Tier6: 447,
				},
			},
			want: map[Tier]int{
				Tier1: 15,
				Tier2: 15,
				Tier3: 13,
				Tier4: 11,
				Tier5: 9,
				Tier6: 7,
			},
		},
		{
			name: "spells",
			args: args{
				ratios: spellPoolRatio(),
				numTemplates: map[Tier]int{
					Tier1: 44,
					Tier2: 15,
					Tier3: 74,
					Tier4: 97,
					Tier5: 45,
					Tier6: 38,
				},
			},
			want: map[Tier]int{
				Tier1: 5,
				Tier2: 7,
				Tier3: 9,
				Tier4: 11,
				Tier5: 9,
				Tier6: 7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for tier := Tier1; tier <= Tier6; tier++ {
				got := poolCopies(tt.args.numTemplates[tier], tt.args.ratios, tier)
				assert.Equal(t, tt.want[tier], got, "tier %d", tier)
			}
		})
	}
}
