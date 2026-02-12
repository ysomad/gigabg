package game

import "slices"

type Tribe uint8

const (
	TribeNeutral Tribe = iota + 1
	TribeBeast
	TribeDemon
	TribeDragon
	TribeElemental
	TribeMech
	TribeMurloc
	TribeNaga
	TribePirate
	TribeQuilboar
	TribeUndead
	TribeAll
	TribeMixed
)

func (t Tribe) String() string {
	switch t {
	case TribeNeutral:
		return "Neutral"
	case TribeBeast:
		return "Beast"
	case TribeDemon:
		return "Demon"
	case TribeDragon:
		return "Dragon"
	case TribeElemental:
		return "Elemental"
	case TribeMech:
		return "Mech"
	case TribeMurloc:
		return "Murloc"
	case TribeNaga:
		return "Naga"
	case TribePirate:
		return "Pirate"
	case TribeQuilboar:
		return "Quilboar"
	case TribeUndead:
		return "Undead"
	case TribeAll:
		return "All"
	case TribeMixed:
		return "Mixed"
	default:
		return ""
	}
}

var _ Card = (*Minion)(nil)

// Minion is a runtime card instance with mutable state.
type Minion struct {
	template CardTemplate
	attack   int
	health   int
	golden   bool
	keywords Keywords
	effects  []TriggeredEffect
	combatID int // unique within a single combat, 0 = not assigned
}

// NewMinion creates a minion from a card template.
func NewMinion(t CardTemplate) *Minion {
	return &Minion{
		template: t,
		attack:   t.Attack(),
		health:   t.Health(),
		keywords: t.Keywords(),
		effects:  t.Effects(),
	}
}

func (m *Minion) Template() CardTemplate { return m.template }
func (m *Minion) TemplateID() string     { return m.template.ID() }
func (m *Minion) Name() string           { return m.template.Name() }
func (m *Minion) Description() string    { return m.template.Description() }
func (m *Minion) Tribe() Tribe           { return m.template.Tribe() }
func (m *Minion) Tier() Tier             { return m.template.Tier() }

func (m *Minion) CombatID() int { return m.combatID }
func (m *Minion) Attack() int   { return m.attack }
func (m *Minion) Health() int   { return m.health }

func (m *Minion) IsAlive() bool   { return m.health > 0 }
func (m *Minion) CanAttack() bool { return m.health > 0 && m.attack > 0 }

func (m *Minion) IsSpell() bool  { return false }
func (m *Minion) IsMinion() bool { return true }
func (m *Minion) IsGolden() bool { return m.golden }

func (m *Minion) TakeDamage(amount int) { m.health -= amount }

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

// EffectsByTrigger returns all effect payloads for the given trigger.
func (m *Minion) EffectsByTrigger(t Trigger) []Effect {
	return EffectsByTrigger(m.effects, t)
}

// MergeGolden combines two copies into a golden minion.
// Stats = sum of both copies (retains buffs). Effects = doubled from template + Battlecry triple reward.
func (m *Minion) MergeGolden(other *Minion) *Minion {
	effects := m.template.GoldenEffects()
	effects = append(effects, TriggeredEffect{
		Trigger: TriggerBattlecry,
		Effect:  &AddCard{TemplateID: TripleRewardID},
	})
	return &Minion{
		template: m.template,
		attack:   m.attack + other.attack,
		health:   m.health + other.health,
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
		golden:   m.golden,
		keywords: m.keywords,
		effects:  slices.Clone(m.effects),
		combatID: m.combatID,
	}
}
