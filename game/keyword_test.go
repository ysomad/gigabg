package game

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKeywords(t *testing.T) {
	t.Parallel()

	type args struct {
		kws []Keyword
	}

	tests := []struct {
		name string
		args args
		want []Keyword
	}{
		{name: "empty", want: nil},
		{name: "single", args: args{kws: []Keyword{KeywordTaunt}}, want: []Keyword{KeywordTaunt}},
		{
			name: "multiple",
			args: args{kws: []Keyword{KeywordTaunt, KeywordDivineShield}},
			want: []Keyword{KeywordTaunt, KeywordDivineShield},
		},
		{
			name: "dedup",
			args: args{kws: []Keyword{KeywordReborn, KeywordReborn, KeywordReborn}},
			want: []Keyword{KeywordReborn},
		},
		{
			name: "all",
			args: args{kws: []Keyword{
				KeywordTaunt, KeywordDivineShield, KeywordWindfury, KeywordPoisonous,
				KeywordVenomous, KeywordCleave, KeywordStealth, KeywordReborn,
				KeywordMagnetic,
			}},
			want: []Keyword{
				KeywordTaunt, KeywordDivineShield, KeywordWindfury, KeywordPoisonous,
				KeywordVenomous, KeywordCleave, KeywordStealth, KeywordReborn,
				KeywordMagnetic,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKeywords(tt.args.kws...)
			assert.Equal(t, tt.want, k.All())
		})
	}
}

func TestKeywords_Has(t *testing.T) {
	t.Parallel()

	type args struct {
		kw Keyword
	}

	tests := []struct {
		name string
		kws  Keywords
		args args
		want bool
	}{
		{name: "empty", kws: NewKeywords(), args: args{kw: KeywordTaunt}, want: false},
		{name: "present", kws: NewKeywords(KeywordTaunt, KeywordReborn), args: args{kw: KeywordTaunt}, want: true},
		{name: "missing", kws: NewKeywords(KeywordTaunt, KeywordReborn), args: args{kw: KeywordCleave}, want: false},
		{name: "first", kws: NewKeywords(KeywordTaunt), args: args{kw: KeywordTaunt}, want: true},
		{name: "last", kws: NewKeywords(KeywordMagnetic), args: args{kw: KeywordMagnetic}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kws.Has(tt.args.kw))
		})
	}
}

func TestKeywords_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kws  Keywords
		want int
	}{
		{name: "empty", kws: NewKeywords(), want: 0},
		{name: "one", kws: NewKeywords(KeywordTaunt), want: 1},
		{name: "three", kws: NewKeywords(KeywordTaunt, KeywordDivineShield, KeywordWindfury), want: 3},
		{
			name: "all",
			kws: NewKeywords(
				KeywordTaunt, KeywordDivineShield, KeywordWindfury, KeywordPoisonous,
				KeywordVenomous, KeywordCleave, KeywordStealth, KeywordReborn,
				KeywordMagnetic,
			),
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kws.Len())
		})
	}
}

func TestKeywords_All(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kws  Keywords
		want []Keyword
	}{
		{name: "empty", kws: NewKeywords(), want: nil},
		{name: "single", kws: NewKeywords(KeywordCleave), want: []Keyword{KeywordCleave}},
		{
			name: "enum order",
			kws:  NewKeywords(KeywordMagnetic, KeywordReborn, KeywordTaunt),
			want: []Keyword{KeywordTaunt, KeywordReborn, KeywordMagnetic},
		},
		{
			name: "all",
			kws: NewKeywords(
				KeywordTaunt, KeywordDivineShield, KeywordWindfury, KeywordPoisonous,
				KeywordVenomous, KeywordCleave, KeywordStealth, KeywordReborn,
				KeywordMagnetic,
			),
			want: []Keyword{
				KeywordTaunt, KeywordDivineShield, KeywordWindfury, KeywordPoisonous,
				KeywordVenomous, KeywordCleave, KeywordStealth, KeywordReborn,
				KeywordMagnetic,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kws.All())
		})
	}
}

func TestKeywords_Iter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kws  Keywords
		want []Keyword
	}{
		{name: "empty", kws: NewKeywords(), want: nil},
		{name: "single", kws: NewKeywords(KeywordCleave), want: []Keyword{KeywordCleave}},
		{
			name: "enum order",
			kws:  NewKeywords(KeywordMagnetic, KeywordReborn, KeywordTaunt),
			want: []Keyword{KeywordTaunt, KeywordReborn, KeywordMagnetic},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, slices.Collect(tt.kws.Iter()))
		})
	}
}

func TestKeywords_Iter_Break(t *testing.T) {
	t.Parallel()

	k := NewKeywords(KeywordTaunt, KeywordDivineShield, KeywordReborn)
	var first Keyword
	for kw := range k.Iter() {
		first = kw
		break
	}

	assert.Equal(t, KeywordTaunt, first)
}

func TestKeywords_Add(t *testing.T) {
	t.Parallel()

	type args struct {
		kw Keyword
	}

	tests := []struct {
		name    string
		initial Keywords
		args    args
		want    []Keyword
	}{
		{name: "to empty", initial: NewKeywords(), args: args{kw: KeywordTaunt}, want: []Keyword{KeywordTaunt}},
		{
			name:    "new",
			initial: NewKeywords(KeywordTaunt),
			args:    args{kw: KeywordReborn},
			want:    []Keyword{KeywordTaunt, KeywordReborn},
		},
		{
			name:    "dup noop",
			initial: NewKeywords(KeywordTaunt),
			args:    args{kw: KeywordTaunt},
			want:    []Keyword{KeywordTaunt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Add(tt.args.kw)
			assert.Equal(t, tt.want, tt.initial.All())
		})
	}
}

func TestKeywords_Remove(t *testing.T) {
	t.Parallel()

	type args struct {
		kw Keyword
	}

	tests := []struct {
		name    string
		initial Keywords
		args    args
		want    []Keyword
	}{
		{name: "from empty", initial: NewKeywords(), args: args{kw: KeywordTaunt}, want: nil},
		{
			name:    "present",
			initial: NewKeywords(KeywordTaunt, KeywordReborn),
			args:    args{kw: KeywordTaunt},
			want:    []Keyword{KeywordReborn},
		},
		{
			name:    "absent noop",
			initial: NewKeywords(KeywordTaunt),
			args:    args{kw: KeywordReborn},
			want:    []Keyword{KeywordTaunt},
		},
		{name: "last remaining", initial: NewKeywords(KeywordTaunt), args: args{kw: KeywordTaunt}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.Remove(tt.args.kw)
			assert.Equal(t, tt.want, tt.initial.All())
		})
	}
}
