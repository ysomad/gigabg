package game

import (
	"iter"
	"slices"
)

var _ Card = (*Minion)(nil)

// CombatID uniquely identifies a minion within a single combat. Zero means not assigned.
type CombatID int

// Minion is a runtime card instance with mutable state.
type Minion struct {
	template CardTemplate
	attack   int
	health   int
	cost     int
	golden   bool
	keywords Keywords
	effects  []TriggeredEffect
	combatID CombatID
}

// NewMinion creates a minion from a card template.
func NewMinion(t CardTemplate) *Minion {
	return &Minion{
		template: t,
		attack:   t.Attack(),
		health:   t.Health(),
		cost:     t.Cost(),
		keywords: t.Keywords(),
		effects:  t.Effects(),
	}
}

func (m *Minion) Template() CardTemplate { return m.template }
func (m *Minion) TemplateID() string     { return m.template.ID() }
func (m *Minion) Name() string           { return m.template.Name() }
func (m *Minion) Description() string    { return m.template.Description() }
func (m *Minion) Tribes() Tribes          { return m.template.Tribes() }
func (m *Minion) Tier() Tier             { return m.template.Tier() }

func (m *Minion) CombatID() CombatID { return m.combatID }
func (m *Minion) Attack() int        { return m.attack }
func (m *Minion) Health() int        { return m.health }
func (m *Minion) Cost() int          { return m.cost }

func (m *Minion) IsAlive() bool   { return m.health > 0 }
func (m *Minion) CanAttack() bool { return m.health > 0 && m.attack > 0 }

func (m *Minion) IsSpell() bool  { return false }
func (m *Minion) IsMinion() bool { return true }
func (m *Minion) IsGolden() bool { return m.golden }

// CanMagnetizeTo reports whether this minion can magnetize onto target.
func (m *Minion) CanMagnetizeTo(target *Minion) bool {
	return m.HasKeyword(KeywordMagnetic) && target != nil && target.Tribes().HasAny(m.Tribes())
}

func (m *Minion) TakeDamage(amount int) { m.health -= amount }

// Keywords returns the minion's current keywords bitmask.
func (m *Minion) Keywords() Keywords { return m.keywords }

// HasKeyword returns true if the minion has the given static keyword.
func (m *Minion) HasKeyword(kw Keyword) bool { return m.keywords.Has(kw) }

// AddKeyword adds a static keyword to the minion.
func (m *Minion) AddKeyword(kw Keyword) { m.keywords.Add(kw) }

// RemoveKeyword removes a static keyword from the minion.
func (m *Minion) RemoveKeyword(kw Keyword) { m.keywords.Remove(kw) }

// AddEffect appends a triggered effect to the minion.
func (m *Minion) AddEffect(te TriggeredEffect) { m.effects = append(m.effects, te) }

// RemoveEffect removes the first triggered effect matching the trigger.
func (m *Minion) RemoveEffect(t Trigger) {
	for i, te := range m.effects {
		if te.Trigger == t {
			m.effects = append(m.effects[:i], m.effects[i+1:]...)
			return
		}
	}
}

// EffectsByTrigger yields effect payloads matching any of the given triggers.
func (m *Minion) EffectsByTrigger(triggers ...Trigger) iter.Seq[Effect] {
	return EffectsByTrigger(m.effects, triggers...)
}

// MergeGolden combines two copies into a golden minion.
// Stats = sum of both copies (retains buffs). Effects = doubled from template + Battlecry triple reward.
func (m *Minion) MergeGolden(other *Minion) *Minion {
	effects := m.template.GoldenEffects()
	effects = append(effects, TriggeredEffect{
		Trigger: TriggerGolden,
		Effect:  &AddCard{TemplateID: TripleRewardID},
	})
	return &Minion{
		template: m.template,
		attack:   m.attack + other.attack,
		health:   m.health + other.health,
		cost:     m.template.Cost(),
		golden:   true,
		keywords: m.template.Keywords(),
		effects:  effects,
	}
}

func (m *Minion) Clone() *Minion {
	if m == nil {
		return nil
	}
	return &Minion{
		template: m.template,
		attack:   m.attack,
		health:   m.health,
		cost:     m.cost,
		golden:   m.golden,
		keywords: m.keywords,
		effects:  slices.Clone(m.effects),
		combatID: m.combatID,
	}
}
