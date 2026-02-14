package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTribe_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   Tribe
		want string
	}{
		{"neutral", TribeNeutral, "Neutral"},
		{"beast", TribeBeast, "Beast"},
		{"demon", TribeDemon, "Demon"},
		{"dragon", TribeDragon, "Dragon"},
		{"elemental", TribeElemental, "Elemental"},
		{"mech", TribeMech, "Mech"},
		{"murloc", TribeMurloc, "Murloc"},
		{"naga", TribeNaga, "Naga"},
		{"pirate", TribePirate, "Pirate"},
		{"quilboar", TribeQuilboar, "Quilboar"},
		{"undead", TribeUndead, "Undead"},
		{"mixed", TribeMixed, "Mixed"},
		{"unknown", Tribe(200), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tr.String())
		})
	}
}

func TestTribes_Has(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		tribes Tribes
		tribe  Tribe
		want   bool
	}{
		{"beast has beast", NewTribes(TribeBeast), TribeBeast, true},
		{"beast has demon", NewTribes(TribeBeast), TribeDemon, false},
		{"all has beast", TribeAll, TribeBeast, true},
		{"all has mech", TribeAll, TribeMech, true},
		{"all has neutral", TribeAll, TribeNeutral, false},
		{"empty has beast", 0, TribeBeast, false},
		{"empty has neutral", 0, TribeNeutral, false},
		{"multi has beast", NewTribes(TribeBeast, TribeDemon), TribeBeast, true},
		{"multi has demon", NewTribes(TribeBeast, TribeDemon), TribeDemon, true},
		{"multi has dragon", NewTribes(TribeBeast, TribeDemon), TribeDragon, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tribes.Has(tt.tribe))
		})
	}
}

func TestTribes_HasAny(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b Tribes
		want bool
	}{
		{"both empty", 0, 0, false},
		{"empty and beast", 0, NewTribes(TribeBeast), false},
		{"same single", NewTribes(TribeMech), NewTribes(TribeMech), true},
		{"disjoint singles", NewTribes(TribeMech), NewTribes(TribeBeast), false},
		{"overlap in multi", NewTribes(TribeMech, TribeUndead), NewTribes(TribeUndead, TribeDemon), true},
		{"no overlap in multi", NewTribes(TribeMech, TribeBeast), NewTribes(TribeUndead, TribeDemon), false},
		{"all overlaps single", TribeAll, NewTribes(TribePirate), true},
		{"multi overlaps all", NewTribes(TribeMech, TribeUndead), TribeAll, true},
		{"all overlaps all", TribeAll, TribeAll, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.a.HasAny(tt.b))
		})
	}
}

func TestTribes_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   Tribes
		want string
	}{
		{"empty", 0, ""},
		{"single", NewTribes(TribeBeast), "Beast"},
		{"two tribes", NewTribes(TribeMech, TribeUndead), "Mech\nUndead"},
		{"three tribes", NewTribes(TribeBeast, TribeDemon, TribeDragon), "Beast\nDemon\nDragon"},
		{"all", TribeAll, "All"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tr.String())
		})
	}
}

func TestTribes_List(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   Tribes
		want []Tribe
	}{
		{"empty returns nil", 0, nil},
		{"single tribe", NewTribes(TribeBeast), []Tribe{TribeBeast}},
		{"two tribes", NewTribes(TribeBeast, TribeDemon), []Tribe{TribeBeast, TribeDemon}},
		{"three tribes", NewTribes(TribeBeast, TribeDragon, TribeMech), []Tribe{TribeBeast, TribeDragon, TribeMech}},
		{"all tribes", TribeAll, []Tribe{
			TribeBeast, TribeDemon, TribeDragon, TribeElemental, TribeMech,
			TribeMurloc, TribeNaga, TribePirate, TribeQuilboar, TribeUndead,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tr.List())
		})
	}
}

func TestTribes_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   Tribes
		want int
	}{
		{"empty", 0, 0},
		{"single", NewTribes(TribeBeast), 1},
		{"two", NewTribes(TribeBeast, TribeDemon), 2},
		{"all", TribeAll, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tr.Len())
		})
	}
}

func TestTribes_First(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   Tribes
		want Tribe
	}{
		{"empty", 0, TribeNeutral},
		{"beast", NewTribes(TribeBeast), TribeBeast},
		{"demon only", NewTribes(TribeDemon), TribeDemon},
		{"multi returns lowest", NewTribes(TribeDemon, TribeDragon), TribeDemon},
		{"all returns beast", TribeAll, TribeBeast},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.tr.First())
		})
	}
}

func TestCalcTopTribe(t *testing.T) {
	t.Parallel()

	tr := func(tribes ...Tribe) Tribes { return NewTribes(tribes...) }

	tests := []struct {
		name      string
		tribes    []Tribes
		want      Tribe
		wantCount int
	}{
		{"nil tribes", nil, TribeNeutral, 0},
		{"empty tribes", []Tribes{}, TribeNeutral, 0},
		{"neutral", []Tribes{0, 0, 0, 0, tr(TribeBeast)}, TribeBeast, 1},
		{"all 1", []Tribes{TribeAll, TribeAll, TribeAll, TribeAll, TribeAll, tr(TribeDemon)}, TribeDemon, 6},
		{"all 2", []Tribes{TribeAll, TribeAll, TribeAll, 0}, TribeNeutral, 0},
		{"all 3", []Tribes{TribeAll, tr(TribePirate), tr(TribePirate), 0}, TribePirate, 3},
		{"beasts", []Tribes{tr(TribeBeast), tr(TribeDemon), tr(TribeNaga), tr(TribeBeast)}, TribeBeast, 2},
		{"demons", []Tribes{tr(TribeDemon), tr(TribeDemon), 0, TribeAll}, TribeDemon, 3},
		{"dragons", []Tribes{TribeAll, tr(TribeElemental), tr(TribeDragon), tr(TribeDragon), tr(TribeNaga)}, TribeDragon, 3},
		{
			"elems",
			[]Tribes{tr(TribeDemon), tr(TribeBeast), tr(TribeNaga), tr(TribeElemental), tr(TribeElemental), tr(TribeElemental)},
			TribeElemental,
			3,
		},
		{"mechs", []Tribes{0, tr(TribeMurloc), tr(TribeBeast), tr(TribeMech), tr(TribeMech), tr(TribeMech)}, TribeMech, 3},
		{"murloc", []Tribes{tr(TribeMurloc), tr(TribeMurloc), tr(TribeMurloc), tr(TribeMurloc), tr(TribeMurloc)}, TribeMurloc, 5},
		{"nagas", []Tribes{tr(TribeMech), tr(TribeNaga), tr(TribeNaga), 0, tr(TribeMech), tr(TribeNaga)}, TribeNaga, 3},
		{"pirates", []Tribes{tr(TribePirate), tr(TribePirate), tr(TribeUndead), tr(TribePirate), 0}, TribePirate, 3},
		{"quilboar", []Tribes{tr(TribeQuilboar), tr(TribeQuilboar), tr(TribeBeast), TribeAll}, TribeQuilboar, 3},
		{"undead", []Tribes{tr(TribeUndead), tr(TribeUndead), tr(TribeUndead), tr(TribeDemon), tr(TribeDragon)}, TribeUndead, 3},
		{"single minion", []Tribes{tr(TribeDragon)}, TribeDragon, 1},
		{"single neutral", []Tribes{0}, TribeNeutral, 0},
		{"single all", []Tribes{TribeAll}, TribeNeutral, 0},
		{"mixed tied", []Tribes{tr(TribeBeast), tr(TribeDemon)}, TribeMixed, 2},
		{"mixed tied with all", []Tribes{tr(TribeBeast), tr(TribeDemon), TribeAll}, TribeDemon, 2},
		{"mixed three way", []Tribes{tr(TribeMech), tr(TribePirate), tr(TribeUndead)}, TribeMixed, 3},
		{"all only neutrals", []Tribes{0, 0, 0}, TribeNeutral, 0},
		{"all boosts single", []Tribes{TribeAll, TribeAll, tr(TribeMurloc)}, TribeMurloc, 3},
		{"all boosts dominant", []Tribes{tr(TribeBeast), tr(TribeBeast), tr(TribeDemon), TribeAll, TribeAll}, TribeBeast, 4},
		{"multi tribe counted", []Tribes{tr(TribeBeast, TribeDemon), tr(TribeBeast)}, TribeBeast, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CalcTopTribe(tt.tribes)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantCount, got1)
		})
	}
}
