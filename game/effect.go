package game

import "iter"

// Effect is a triggered ability's behavior.
// Concrete types: BuffStats, GiveKeyword, SummonMinion, DiscoverCard, MakeGolden.
type Effect interface {
	Apply(ctx EffectContext)
	golden() Effect
}

var (
	_ Effect = (*BuffStats)(nil)
	_ Effect = (*GiveKeyword)(nil)
	_ Effect = (*SummonMinion)(nil)
	_ Effect = (*DiscoverCard)(nil)
	_ Effect = (*MakeGolden)(nil)
	_ Effect = (*DealDamage)(nil)
	_ Effect = (*DestroyMinion)(nil)
	_ Effect = (*AddCard)(nil)
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
	TargetFriendlySelected
)

// TargetFilter restricts which minions/cards can be targeted.
type TargetFilter struct {
	Tribe         Tribe   // filter by tribe (0 = any)
	Tier          Tier    // filter by tier (0 = any)
	HasKeyword    Keyword // must have this keyword (0 = any)
	ExcludeSource bool    // exclude the source minion
}

// Target defines who an effect targets and how many.
type Target struct {
	Type   TargetType
	Filter TargetFilter
	Count  int // 0 = all matching
}

// EffectContext provides game state for effect execution.
type EffectContext struct {
	Source        *Minion   // minion that triggered the effect
	Board         *Board    // player's board
	Hand          *Hand     // player's hand
	Shop          *Shop     // player's shop (nil during combat)
	Pool          *CardPool // shared card pool (nil during combat)
	OpponentBoard *Board    // opponent's board (nil outside combat)
	Discovers     *[]Card   // discover options output (set by DiscoverCard)
}

func (ctx EffectContext) WithSource(m *Minion) EffectContext {
	if m != nil {
		ctx.Source = m
	}
	return ctx
}

func (ctx EffectContext) WithOpponentBoard(b *Board) EffectContext {
	if b != nil {
		ctx.OpponentBoard = b
	}
	return ctx
}

// TriggeredEffect pairs a trigger timing with an effect payload.
type TriggeredEffect struct {
	Trigger   Trigger
	Effect    Effect
	Threshold int // only for TriggerAvenge
}

// EffectsByTrigger yields effect payloads matching any of the given triggers, in order.
func EffectsByTrigger(effects []TriggeredEffect, triggers ...Trigger) iter.Seq[Effect] {
	return func(yield func(Effect) bool) {
		for _, te := range effects {
			if te.Effect == nil {
				continue
			}
			for _, t := range triggers {
				if te.Trigger == t {
					if !yield(te.Effect) {
						return
					}
					break
				}
			}
		}
	}
}

// MakeGoldenEffects returns a copy with all effect payloads upgraded for golden cards.
func MakeGoldenEffects(effects []TriggeredEffect) []TriggeredEffect {
	out := make([]TriggeredEffect, len(effects))
	for i, te := range effects {
		out[i] = TriggeredEffect{
			Trigger:   te.Trigger,
			Threshold: te.Threshold,
		}
		if te.Effect != nil {
			out[i].Effect = te.Effect.golden()
		}
	}
	return out
}

// BuffStats gives +Attack/+Health to targets.
type BuffStats struct {
	Target     Target
	Attack     int
	Health     int
	Persistent bool // true = permanent buff, false = combat only
}

func (e *BuffStats) Apply(ctx EffectContext) {
	panic("BuffStats.Apply not implemented")
}

func (e *BuffStats) golden() Effect {
	return &BuffStats{
		Target:     Target{Type: e.Target.Type, Filter: e.Target.Filter, Count: e.Target.Count * 2},
		Attack:     e.Attack * 2,
		Health:     e.Health * 2,
		Persistent: e.Persistent,
	}
}

// GiveKeyword grants a keyword to targets.
type GiveKeyword struct {
	Target  Target
	Keyword Keyword
}

func (e *GiveKeyword) Apply(ctx EffectContext) {
	panic("GiveKeyword.Apply not implemented")
}

func (e *GiveKeyword) golden() Effect {
	return &GiveKeyword{
		Target:  Target{Type: e.Target.Type, Filter: e.Target.Filter, Count: e.Target.Count * 2},
		Keyword: e.Keyword,
	}
}

// SummonMinion spawns a minion on the board.
type SummonMinion struct {
	TemplateID string
}

func (e *SummonMinion) Apply(ctx EffectContext) {
	panic("SummonMinion.Apply not implemented")
}

func (e *SummonMinion) golden() Effect {
	return &SummonMinion{TemplateID: e.TemplateID}
}

// DiscoverCard lets the player discover cards from the pool.
type DiscoverCard struct {
	TierOffset int      // tiers above shop tier to discover from
	Tribe      Tribe    // filter by tribe (0 = any)
	Kind       CardKind // filter by card kind (0 = any)
}

func (e *DiscoverCard) Apply(ctx EffectContext) {
	if ctx.Shop == nil || ctx.Pool == nil || ctx.Discovers == nil {
		return
	}
	tier := min(ctx.Shop.Tier()+Tier(e.TierOffset), Tier6)
	*ctx.Discovers = ctx.Pool.RollDiscover(tier, e.Tribe, e.Kind)
}

func (e *DiscoverCard) golden() Effect { return e }

// MakeGolden makes the target minion golden.
type MakeGolden struct{}

func (e *MakeGolden) Apply(ctx EffectContext) {
	panic("MakeGolden.Apply not implemented")
}

func (e *MakeGolden) golden() Effect { return e }

// DealDamage deals damage to targets.
type DealDamage struct {
	Target Target
	Amount int
}

func (e *DealDamage) Apply(ctx EffectContext) {
	panic("DealDamage.Apply not implemented")
}

func (e *DealDamage) golden() Effect {
	return &DealDamage{
		Target: Target{Type: e.Target.Type, Filter: e.Target.Filter, Count: e.Target.Count},
		Amount: e.Amount * 2,
	}
}

// DestroyMinion destroys target minions.
type DestroyMinion struct {
	Target Target
}

func (e *DestroyMinion) Apply(ctx EffectContext) {
	panic("DestroyMinion.Apply not implemented")
}

func (e *DestroyMinion) golden() Effect {
	return &DestroyMinion{
		Target: Target{Type: e.Target.Type, Filter: e.Target.Filter, Count: e.Target.Count * 2},
	}
}

// AddCard adds cards to the player's hand.
type AddCard struct {
	TemplateID string
}

func (e *AddCard) Apply(ctx EffectContext) {
	if e.TemplateID == "" {
		return
	}

	tmpl := ctx.Pool.ByTemplateID(e.TemplateID)
	if tmpl != nil && !ctx.Hand.IsFull() {
		ctx.Hand.Add(NewCard(tmpl))
	}
}

func (e *AddCard) golden() Effect { return e }

// AuraScope defines the range of an aura effect.
type AuraScope uint8

const (
	AuraScopeAllFriendly AuraScope = iota + 1
	AuraScopeAdjacent
)

// Aura is an ongoing effect active while the source minion is on the board.
type Aura struct {
	Scope   AuraScope
	Filter  TargetFilter
	Attack  int
	Health  int
	Keyword Keyword // 0 = none
}

// MakeGoldenAuras returns a copy of auras upgraded for golden cards.
func MakeGoldenAuras(auras []Aura) []Aura {
	if len(auras) == 0 {
		return nil
	}
	out := make([]Aura, len(auras))
	for i, a := range auras {
		out[i] = Aura{
			Scope:   a.Scope,
			Filter:  a.Filter,
			Attack:  a.Attack * 2,
			Health:  a.Health * 2,
			Keyword: a.Keyword,
		}
	}
	return out
}
