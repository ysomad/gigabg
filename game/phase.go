package game

type Phase uint8

const (
	PhaseRecruit Phase = iota + 1
	PhaseCombat
)
