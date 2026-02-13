package game

// Trigger defines when a triggered effect fires.
// Triggers are timing conditions paired with Effect payloads in TriggeredEffect.
type Trigger uint8

const (
	TriggerBattlecry Trigger = iota + 1
	TriggerDeathrattle
	TriggerAvenge
	TriggerStartOfCombat
	TriggerStartOfTurn
	TriggerEndOfTurn
	TriggerSpell
	TriggerGolden
)

func (t Trigger) String() string {
	switch t {
	case TriggerBattlecry:
		return "Battlecry"
	case TriggerDeathrattle:
		return "Deathrattle"
	case TriggerAvenge:
		return "Avenge"
	case TriggerStartOfCombat:
		return "Start of Combat"
	case TriggerStartOfTurn:
		return "Start of Turn"
	case TriggerEndOfTurn:
		return "End of Turn"
	case TriggerSpell:
		return "Spell"
	case TriggerGolden:
		return "Golden"
	default:
		return ""
	}
}
