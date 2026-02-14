package game

import "strings"

// Tribe represents a single minion tribe.
type Tribe uint8

const (
	TribeNeutral Tribe = iota // no tribe affiliation
	TribeBeast
	TribeDemon
	TribeDragon
	TribeElemental
	TribeMech
	TribeMurloc
	TribeNaga
	TribePirate
	TribeQuilboar
	TribeUndead
	tribeMax

	TribeMixed Tribe = 0xFF // sentinel: tied tribes in CalcTopTribe
)

// String returns the name for a single tribe value. Multi-tribe values return "".
func (t Tribe) String() string {
	switch t {
	case TribeNeutral:
		return "Neutral"
	case TribeBeast:
		return "Beast"
	case TribeDemon:
		return "Demon"
	case TribeDragon:
		return "Dragon"
	case TribeElemental:
		return "Elemental"
	case TribeMech:
		return "Mech"
	case TribeMurloc:
		return "Murloc"
	case TribeNaga:
		return "Naga"
	case TribePirate:
		return "Pirate"
	case TribeQuilboar:
		return "Quilboar"
	case TribeUndead:
		return "Undead"
	case TribeMixed:
		return "Mixed"
	default:
		return ""
	}
}

// Tribes is a bitmask of tribes.
type Tribes uint16

// NewTribes creates a Tribes bitmask from the given tribes.
func NewTribes(tribes ...Tribe) Tribes {
	var t Tribes
	for _, tr := range tribes {
		t |= 1 << (tr - 1)
	}
	return t
}

// TribeAll is the bitmask containing all tribes.
const TribeAll Tribes = (1 << (tribeMax - 1)) - 1

// Has returns true if the tribe is set.
func (t Tribes) Has(tr Tribe) bool { return t&(1<<(tr-1)) != 0 }

// HasAny returns true if any tribe in other is also set in t.
func (t Tribes) HasAny(other Tribes) bool { return t&other != 0 }

// Len returns the number of tribes set.
func (t Tribes) Len() int {
	n := 0
	for tr := TribeBeast; tr < tribeMax; tr++ {
		if t.Has(tr) {
			n++
		}
	}
	return n
}

// First returns the lowest-numbered tribe in the set, or TribeNeutral if empty.
func (t Tribes) First() Tribe {
	for tr := TribeBeast; tr < tribeMax; tr++ {
		if t.Has(tr) {
			return tr
		}
	}
	return TribeNeutral
}

// String returns tribe names separated by newlines, or "All" for TribeAll.
func (t Tribes) String() string {
	if t == TribeAll {
		return "All"
	}
	var b strings.Builder
	for tr := TribeBeast; tr < tribeMax; tr++ {
		if !t.Has(tr) {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(tr.String())
	}
	return b.String()
}

// TopTribe holds a dominant tribe and its count.
type TopTribe struct {
	Tribe Tribe
	Count int
}

// CalcTopTribe returns the dominant tribe and its count.
// Multi-tribe minions increment each component tribe's count.
// All-tribe minions are added to the majority count after it is determined.
// Returns TribeMixed when multiple tribes exist but none dominates.
// Returns TribeNeutral when no countable tribes are present.
func CalcTopTribe(tribes []Tribes) (Tribe, int) {
	var allCount int

	counts := make(map[Tribe]int)
	for _, t := range tribes {
		if t == TribeAll {
			allCount++
			continue
		}
		for _, s := range t.List() {
			counts[s]++
		}
	}

	switch len(counts) {
	case 0:
		return TribeNeutral, 0
	case 1:
		for t, n := range counts {
			return t, n + allCount
		}
	}

	// 2+ tribes: find single dominant.
	var (
		best  Tribe
		bestN int
		tied  bool
	)

	for t, n := range counts {
		switch {
		case n > bestN:
			best, bestN, tied = t, n, false
		case n == bestN:
			tied = true
			if t > best {
				best = t
			}
		}
	}

	if !tied {
		return best, bestN + allCount
	}

	// Tied: All minions break the tie by boosting the highest tribe.
	if allCount > 0 {
		return best, bestN + allCount
	}

	return TribeMixed, len(counts)
}

// List returns all set tribes in enum order.
func (t Tribes) List() []Tribe {
	var result []Tribe
	for tr := TribeBeast; tr < tribeMax; tr++ {
		if t.Has(tr) {
			result = append(result, tr)
		}
	}
	return result
}
