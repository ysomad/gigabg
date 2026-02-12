package game

// CombatEventType identifies the kind of combat event.
type CombatEventType uint8

const (
	CombatEventAttack  CombatEventType = iota + 1 // minion initiates attack
	CombatEventDamage                             // damage dealt to a minion
	CombatEventDeath                              // minion dies and is removed
	CombatEventBuff                               // future: stat buff applied
	CombatEventRemoveKeyword                      // keyword removed from minion
	CombatEventSummon                             // future: minion summoned
	CombatEventTrigger                            // future: triggered ability fires
)

// CombatEvent is a single step in the combat log.
// Not all fields are used by every event type.
type CombatEvent struct {
	Type     CombatEventType
	SourceID int     // combat ID of acting minion
	TargetID int     // combat ID of target minion
	Amount   int     // damage/buff amount
	Keyword  Keyword // keyword for CombatEventRemoveKeyword
	OwnerID  string  // player who owns the affected minion
}

// CombatResult is the outcome of a single combat from one player's perspective.
// Stored per-player (last 3). Visible to other clients on hover.
type CombatResult struct {
	OpponentID string
	WinnerID   string // empty if tie
	Damage     int    // damage dealt to loser (0 if tie)
}

// CombatAnimation holds data for combat replay on the client.
// Sent once to the two fighting players, then discarded.
// Boards are delivered separately via GameState.
type CombatAnimation struct {
	Player1ID string
	Player2ID string
	Events    []CombatEvent
}
