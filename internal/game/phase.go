package game

type Phase uint8

const (
	PhaseWaiting Phase = iota
	PhaseRecruit
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
