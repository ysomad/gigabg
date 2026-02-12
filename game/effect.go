package game

import (
	"maps"
	"slices"
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

// Effect is a triggered ability's behavior.
// Concrete types: BuffStats, GiveKeyword, Summon, DiscoverCard.
type Effect interface {
	Apply(ctx EffectContext)
	double() Effect
}

// EffectContext provides game state for effect execution.
type EffectContext struct {
	Source        *Minion   // minion that triggered the effect
	Board         *Board    // player's board
	Hand          *[]Card   // player's hand (pointer so effects can append)
	Shop          *Shop     // player's shop (nil during combat)
	Pool          *CardPool // shared card pool (nil during combat)
	OpponentBoard *Board    // opponent's board (nil outside combat)
}

// Ability pairs a keyword with an optional effect payload.
// Passive keywords (Taunt, Divine Shield) have nil Effect.
// Triggered keywords (Deathrattle, Battlecry) carry an Effect.
type Ability struct {
	Keyword   Keyword
	Effect    Effect // nil for passive keywords
	Threshold int    // only for KeywordAvenge
}

// Abilities is an ordered collection of abilities with map-based lookup.
type Abilities struct {
	list []Ability
	has  map[Keyword]struct{}
}

// NewAbilities creates an Abilities collection from the given abilities.
func NewAbilities(abilities ...Ability) Abilities {
	has := make(map[Keyword]struct{}, len(abilities))
	for _, ab := range abilities {
		has[ab.Keyword] = struct{}{}
	}
	return Abilities{list: abilities, has: has}
}

// Has returns true if any ability with the given keyword exists.
func (a *Abilities) Has(kw Keyword) bool {
	_, ok := a.has[kw]
	return ok
}

// Len returns the total number of abilities.
func (a *Abilities) Len() int { return len(a.list) }

// ByKeyword returns all effect payloads matching the keyword, in order.
// Skips abilities with nil Effect.
func (a Abilities) ByKeyword(kw Keyword) []Effect {
	var result []Effect
	for _, ab := range a.list {
		if ab.Keyword == kw && ab.Effect != nil {
			result = append(result, ab.Effect)
		}
	}
	return result
}

// Keywords returns all unique keywords in order of first appearance.
func (a Abilities) Keywords() []Keyword {
	seen := make(map[Keyword]struct{})
	var kws []Keyword
	for _, ab := range a.list {
		if _, ok := seen[ab.Keyword]; !ok {
			kws = append(kws, ab.Keyword)
			seen[ab.Keyword] = struct{}{}
		}
	}
	return kws
}

// All returns a copy of all abilities.
func (a *Abilities) All() []Ability { return slices.Clone(a.list) }

// Add appends an ability.
func (a *Abilities) Add(ab Ability) {
	a.list = append(a.list, ab)
	if a.has == nil {
		a.has = make(map[Keyword]struct{})
	}
	a.has[ab.Keyword] = struct{}{}
}

// Remove removes the first ability matching the keyword.
func (a *Abilities) Remove(kw Keyword) {
	for i, ab := range a.list {
		if ab.Keyword != kw {
			continue
		}

		a.list = append(a.list[:i], a.list[i+1:]...)
		var found bool

		for _, ab2 := range a.list {
			if ab2.Keyword == kw {
				found = true
				break
			}
		}

		if !found {
			delete(a.has, kw)
		}

		return
	}
}

// Clone returns a deep copy.
func (a Abilities) Clone() Abilities {
	if len(a.list) == 0 {
		return Abilities{}
	}
	return Abilities{
		list: slices.Clone(a.list),
		has:  maps.Clone(a.has),
	}
}

// Double returns a copy with all effect payloads doubled (for golden cards).
// Passive keywords (nil Effect) are copied as-is.
func (a *Abilities) Double() Abilities {
	if len(a.list) == 0 {
		return Abilities{}
	}
	doubled := make([]Ability, len(a.list))
	for i, ab := range a.list {
		doubled[i] = Ability{
			Keyword:   ab.Keyword,
			Threshold: ab.Threshold,
		}
		if ab.Effect != nil {
			doubled[i].Effect = ab.Effect.double()
		}
	}
	return NewAbilities(doubled...)
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

func (e *BuffStats) double() Effect {
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

func (e *GiveKeyword) double() Effect {
	return &GiveKeyword{
		Target:  Target{Type: e.Target.Type, Filter: e.Target.Filter, Count: e.Target.Count * 2},
		Keyword: e.Keyword,
	}
}

// SummonTemplateMinion spawns a minion on the board.
type SummonTemplateMinion struct {
	TemplateID string
}

func (e *SummonTemplateMinion) Apply(ctx EffectContext) {
	panic("SummonMinion.Apply not implemented")
}

func (e *SummonTemplateMinion) double() Effect {
	return &SummonTemplateMinion{TemplateID: e.TemplateID}
}

// Discover lets the player discover cards.
// TODO: specify how much cards, and target (tribe, kind, tier etc)
type Discover struct{}

func (e *Discover) Apply(ctx EffectContext) {
	panic("DiscoverCard.Apply not implemented")
}

func (e *Discover) double() Effect {
	return &Discover{}
}
