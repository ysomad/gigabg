package game

// Keyword represents a card property or trigger type.
// Keywords are stored as entries in an Abilities collection.
// Reordering or removing existing constants changes their numeric values.
// New keywords must be appended before keywordMax.
type Keyword uint8

const (
	KeywordTaunt Keyword = iota + 1
	KeywordDivineShield
	KeywordWindfury
	KeywordPoisonous
	KeywordCleave
	KeywordStealth
	KeywordReborn
	KeywordMagnetic
	KeywordImmune
	KeywordBattlecry
	KeywordDeathrattle
	KeywordAvenge
	KeywordStartOfCombat
	KeywordStartOfTurn
	KeywordEndOfTurn
	KeywordSpell

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
	case KeywordPoisonous:
		return "Poisonous"
	case KeywordCleave:
		return "Cleave"
	case KeywordStealth:
		return "Stealth"
	case KeywordReborn:
		return "Reborn"
	case KeywordMagnetic:
		return "Magnetic"
	case KeywordImmune:
		return "Immune"
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
	case KeywordSpell:
		return "Spell"
	default:
		return ""
	}
}
