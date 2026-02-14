package game

// CombatEventType identifies the kind of combat event.
type CombatEventType uint8

const (
	CombatEventAttack        CombatEventType = iota + 1 // minion initiates attack
	CombatEventDamage                                   // damage dealt to a minion
	CombatEventDeath                                    // minion dies and is removed
	CombatEventBuff                                     // future: stat buff applied
	CombatEventRemoveKeyword                            // keyword removed from minion
	CombatEventReborn                                   // minion respawned via Reborn
	CombatEventTrigger                                  // future: triggered ability fires
)

// CombatEvent is implemented by all combat event types.
type CombatEvent interface {
	EventType() CombatEventType
}

var (
	_ CombatEvent = (*AttackEvent)(nil)
	_ CombatEvent = (*DamageEvent)(nil)
	_ CombatEvent = (*DeathEvent)(nil)
	_ CombatEvent = (*RemoveKeywordEvent)(nil)
	_ CombatEvent = (*RebornEvent)(nil)
)

// AttackEvent is emitted when a minion initiates an attack.
type AttackEvent struct {
	Source CombatID `json:"source"`
	Target CombatID `json:"target"`
	Owner  PlayerID `json:"owner"`
}

func (AttackEvent) EventType() CombatEventType { return CombatEventAttack }

// DamageEvent is emitted when damage is dealt to a minion.
type DamageEvent struct {
	Source CombatID `json:"source"`
	Target CombatID `json:"target"`
	Amount int      `json:"amount"`
	Owner  PlayerID `json:"owner"`
}

func (DamageEvent) EventType() CombatEventType { return CombatEventDamage }

// DeathReason indicates why a minion died.
type DeathReason uint8

const (
	DeathReasonDamage DeathReason = iota // killed by normal damage
	DeathReasonPoison                    // killed by Poisonous or Venomous
)

// DeathEvent is emitted when a minion dies and is removed.
type DeathEvent struct {
	Target      CombatID    `json:"target"`
	DeathReason DeathReason `json:"death_reason"`
	Owner       PlayerID    `json:"owner"`
}

func (DeathEvent) EventType() CombatEventType { return CombatEventDeath }

// RemoveKeywordEvent is emitted when a keyword is removed from a minion.
type RemoveKeywordEvent struct {
	Source  CombatID `json:"source"`
	Target  CombatID `json:"target"`
	Keyword Keyword  `json:"keyword"`
	Owner   PlayerID `json:"owner"`
}

func (RemoveKeywordEvent) EventType() CombatEventType { return CombatEventRemoveKeyword }

// RebornEvent is emitted when a minion respawns via Reborn (always 1 HP).
type RebornEvent struct {
	Target   CombatID `json:"target"`
	Owner    PlayerID `json:"owner"`
	Template string   `json:"template"`
}

func (RebornEvent) EventType() CombatEventType { return CombatEventReborn }

// CombatResult is the outcome of a single combat from one player's perspective.
// Stored per-player (last 3). Visible to other clients on hover.
type CombatResult struct {
	Opponent PlayerID
	Winner   PlayerID // 0 if tie
	Damage   int      // damage dealt to loser (0 if tie)
}

// CombatLog holds data for combat replay on the client.
// Sent once to the two fighting players, then discarded.
// Boards are delivered separately via GameState.
type CombatLog struct {
	Player1 PlayerID
	Player2 PlayerID
	Events  []CombatEvent
}
