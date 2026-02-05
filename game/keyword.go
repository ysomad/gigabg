package game

// Keyword represents a minion keyword.
type Keyword uint8

const (
	// Static keywords (affect combat behavior)
	KeywordTaunt Keyword = iota + 1
	KeywordDivineShield
	KeywordWindfury
	KeywordReborn
	KeywordPoisonous
	KeywordCleave
	KeywordStealth
	KeywordImmune
	KeywordMagnetic

	// Triggered abilities (have Effect payloads)
	KeywordBattlecry
	KeywordDeathrattle
	KeywordAvenge
	KeywordStartOfCombat
	KeywordStartOfTurn
	KeywordEndOfTurn

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
	case KeywordReborn:
		return "Reborn"
	case KeywordPoisonous:
		return "Poisonous"
	case KeywordCleave:
		return "Cleave"
	case KeywordStealth:
		return "Stealth"
	case KeywordImmune:
		return "Immune"
	case KeywordMagnetic:
		return "Magnetic"
	case KeywordBattlecry:
		return "Battlecry"
	case KeywordDeathrattle:
		return "Deathrattle"
	case KeywordAvenge:
		return "Avenge"
	case KeywordStartOfCombat:
		return "Start of Combat"
	case KeywordStartOfTurn:
		return "Start of Turn"
	case KeywordEndOfTurn:
		return "End of Turn"
	default:
		return ""
	}
}

// Keywords is a set of keywords on a minion.
type Keywords uint16

// Has returns true if the keyword is present.
func (k Keywords) Has(kw Keyword) bool {
	return k&(1<<kw) != 0
}

// Add returns keywords with the given keyword added.
func (k Keywords) Add(kw Keyword) Keywords {
	return k | (1 << kw)
}

// Remove returns keywords with the given keyword removed.
func (k Keywords) Remove(kw Keyword) Keywords {
	return k &^ (1 << kw)
}

// List returns all keywords as a slice.
func (k Keywords) List() []Keyword {
	var result []Keyword
	for kw := KeywordTaunt; kw < keywordMax; kw++ {
		if k.Has(kw) {
			result = append(result, kw)
		}
	}
	return result
}
