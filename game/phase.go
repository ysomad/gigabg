package game

type Phase uint8

const (
	PhaseRecruit Phase = iota + 1
	PhaseCombat
)

func (p Phase) String() string {
	switch p {
	case PhaseRecruit:
		return "Recruit"
	case PhaseCombat:
		return "Combat"
	default:
		return "Unknown"
	}
}
