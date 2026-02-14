package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcTopTribe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		tribes    []Tribe
		want      Tribe
		wantCount int
	}{
		{"nil tribes", nil, TribeNeutral, 0},
		{"empty tribes", []Tribe{}, TribeNeutral, 0},
		{"neutral", []Tribe{TribeNeutral, TribeNeutral, TribeNeutral, TribeNeutral, TribeBeast}, TribeBeast, 1},
		{"all 1", []Tribe{TribeAll, TribeAll, TribeAll, TribeAll, TribeAll, TribeDemon}, TribeDemon, 6},
		{"all 2", []Tribe{TribeAll, TribeAll, TribeAll, TribeNeutral}, TribeNeutral, 0},
		{"all 3", []Tribe{TribeAll, TribePirate, TribePirate, TribeNeutral}, TribePirate, 3},
		{"beasts", []Tribe{TribeBeast, TribeDemon, TribeNaga, TribeBeast}, TribeBeast, 2},
		{"demons", []Tribe{TribeDemon, TribeDemon, TribeNeutral, TribeAll}, TribeDemon, 3},
		{"dragons", []Tribe{TribeAll, TribeElemental, TribeDragon, TribeDragon, TribeNaga}, TribeDragon, 3},
		{
			"elems",
			[]Tribe{TribeDemon, TribeBeast, TribeNaga, TribeElemental, TribeElemental, TribeElemental},
			TribeElemental,
			3,
		},
		{"mechs", []Tribe{TribeNeutral, TribeMurloc, TribeBeast, TribeMech, TribeMech, TribeMech}, TribeMech, 3},
		{"murloc", []Tribe{TribeMurloc, TribeMurloc, TribeMurloc, TribeMurloc, TribeMurloc}, TribeMurloc, 5},
		{"nagas", []Tribe{TribeMech, TribeNaga, TribeNaga, TribeNeutral, TribeMech, TribeNaga}, TribeNaga, 3},
		{"pirates", []Tribe{TribePirate, TribePirate, TribeUndead, TribePirate, TribeNeutral}, TribePirate, 3},
		{"quilboar", []Tribe{TribeQuilboar, TribeQuilboar, TribeBeast, TribeAll}, TribeQuilboar, 3},
		{"undead", []Tribe{TribeUndead, TribeUndead, TribeUndead, TribeDemon, TribeDragon}, TribeUndead, 3},
		{"single minion", []Tribe{TribeDragon}, TribeDragon, 1},
		{"single neutral", []Tribe{TribeNeutral}, TribeNeutral, 0},
		{"single all", []Tribe{TribeAll}, TribeNeutral, 0},
		{"mixed tied", []Tribe{TribeBeast, TribeDemon}, TribeMixed, 2},
		{"mixed tied with all", []Tribe{TribeBeast, TribeDemon, TribeAll}, TribeDemon, 2},
		{"mixed three way", []Tribe{TribeMech, TribePirate, TribeUndead}, TribeMixed, 3},
		{"all only neutrals", []Tribe{TribeNeutral, TribeNeutral, TribeNeutral}, TribeNeutral, 0},
		{"all boosts single", []Tribe{TribeAll, TribeAll, TribeMurloc}, TribeMurloc, 3},
		{"all boosts dominant", []Tribe{TribeBeast, TribeBeast, TribeDemon, TribeAll, TribeAll}, TribeBeast, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CalcTopTribe(tt.tribes)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantCount, got1)
		})
	}
}
