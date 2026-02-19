package game

import (
	"iter"
	"math/bits"
)

// Keyword represents a static card mechanic that affects combat behavior.
// Keywords have no effect payload — they are checked by game logic directly.
// Reordering or removing existing constants changes their numeric values.
// New keywords must be appended before keywordMax.
type Keyword uint8

const (
	KeywordTaunt Keyword = iota + 1
	KeywordDivineShield
	KeywordWindfury
	KeywordVenomous
	KeywordCleave
	KeywordStealth
	KeywordReborn
	KeywordMagnetic

	keywordMax // sentinel for iteration
)

func (k Keyword) String() string {
	switch k {
	case KeywordTaunt:
		return "Taunt"
	case KeywordDivineShield:
		return "Divine Shield"
	case KeywordWindfury:
		return "Windfury"
	case KeywordVenomous:
		return "Venomous"
	case KeywordCleave:
		return "Cleave"
	case KeywordStealth:
		return "Stealth"
	case KeywordReborn:
		return "Reborn"
	case KeywordMagnetic:
		return "Magnetic"
	default:
		return ""
	}
}

func (k Keyword) Description() string {
	switch k {
	case KeywordTaunt:
		return "Enemies must attack this minion."
	case KeywordDivineShield:
		return "The first time a Shielded minion takes damage, ignore it."
	case KeywordWindfury:
		return "Can attack twice each turn."
	case KeywordVenomous:
		return "Destroy the first minion this deals damage to."
	case KeywordCleave:
		return "Also damages the minions next to whomever this attacks."
	case KeywordStealth:
		return "Can't be attacked or targeted until it attacks."
	case KeywordReborn:
		return "Resurrects with 1 Health the first time it dies."
	case KeywordMagnetic:
		return "Play this to the left of a Mech to fuse them together."
	default:
		return ""
	}
}

// Keywords is a bitmask of static keywords.
type Keywords uint16

// NewKeywords creates a Keywords bitmask from the given keywords.
func NewKeywords(kws ...Keyword) Keywords {
	var k Keywords
	for _, kw := range kws {
		k |= 1 << (kw - 1)
	}
	return k
}

// Has returns true if the keyword is set.
func (k Keywords) Has(kw Keyword) bool { return k&(1<<(kw-1)) != 0 }

// Len returns the number of keywords set.
func (k Keywords) Len() int { return bits.OnesCount16(uint16(k)) }

// All returns all set keywords in enum order.
func (k Keywords) All() []Keyword {
	var result []Keyword
	for kw := KeywordTaunt; kw < keywordMax; kw++ {
		if k.Has(kw) {
			result = append(result, kw)
		}
	}
	return result
}

// Iter iterates over set keywords in enum order.
func (k Keywords) Iter() iter.Seq[Keyword] {
	return func(yield func(Keyword) bool) {
		for kw := KeywordTaunt; kw < keywordMax; kw++ {
			if k.Has(kw) && !yield(kw) {
				return
			}
		}
	}
}

// Add sets a keyword.
func (k *Keywords) Add(kw Keyword) { *k |= 1 << (kw - 1) }

// Remove clears a keyword.
func (k *Keywords) Remove(kw Keyword) { *k &^= 1 << (kw - 1) }

// Merge adds all keywords from other into k.
func (k *Keywords) Merge(other Keywords) { *k |= other }
