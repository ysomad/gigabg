package game

// EffectType defines what an effect does.
type EffectType uint8

const (
	EffectBuffStats EffectType = iota + 1
	EffectGiveKeyword
	EffectSummon
	EffectDealDamage
	EffectDestroy
	EffectAddCard   // add card to hand (random, specific, or filtered by tribe/tier)
	EffectDiscover  // discover from next tier
)

// TargetType defines who the effect targets.
type TargetType uint8

const (
	TargetSelf TargetType = iota + 1
	TargetAllFriendly
	TargetAllEnemy
	TargetRandomFriendly
	TargetRandomEnemy
	TargetLeftFriendly
	TargetRightFriendly
	TargetLeftmostFriendly
	TargetRightmostFriendly
)

// TargetFilter restricts which minions/cards can be targeted.
type TargetFilter struct {
	Tribe       Tribe   // filter by tribe (0 = any)
	Tier        Tier    // filter by tier (0 = any)
	HasKeyword  Keyword // must have this keyword (0 = any)
	ExcludeSelf bool    // exclude the source minion
}

// Target defines who an effect targets and how many.
type Target struct {
	Type   TargetType
	Filter TargetFilter
	Count  int // 0 = all matching
}

// Effect defines a triggered ability's behavior.
type Effect struct {
	Type       EffectType
	Target     Target
	Attack     int     // attack buff or damage amount
	Health     int     // health buff
	Keyword    Keyword // for EffectGiveKeyword
	CardID     string  // card template ID for EffectSummon/EffectAddCard
	Persistent bool    // true = permanent buff, false = combat only
}

// Clone returns a deep copy of the effect.
func (e *Effect) Clone() *Effect {
	if e == nil {
		return nil
	}
	clone := *e
	return &clone
}

// Double returns a copy with doubled numeric values (Attack, Health, Target.Count).
func (e *Effect) Double() *Effect {
	if e == nil {
		return nil
	}
	d := e.Clone()
	d.Attack *= 2
	d.Health *= 2
	if d.Target.Count > 0 {
		d.Target.Count *= 2
	}
	return d
}

// DoubleAvenge returns an AvengeEffect with doubled effect values.
func DoubleAvenge(a *AvengeEffect) *AvengeEffect {
	if a == nil {
		return nil
	}
	doubled := a.Effect.Double()
	return &AvengeEffect{
		Effect:    *doubled,
		Threshold: a.Threshold,
	}
}
