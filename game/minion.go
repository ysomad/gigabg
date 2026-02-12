package game

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
	template  CardTemplate
	attack    int
	health    int
	golden    bool
	abilities Abilities
	combatID  int // unique within a single combat, 0 = not assigned
}

// NewMinion creates a minion from a card template.
func NewMinion(t CardTemplate) *Minion {
	return &Minion{
		template:  t,
		attack:    t.Attack(),
		health:    t.Health(),
		abilities: t.Abilities().Clone(),
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

func (m *Minion) Alive() bool     { return m.health > 0 }
func (m *Minion) CanAttack() bool { return m.health > 0 && m.attack > 0 }

func (m *Minion) IsSpell() bool  { return false }
func (m *Minion) IsMinion() bool { return true }
func (m *Minion) IsGolden() bool { return m.golden }

func (m *Minion) TakeDamage(amount int) { m.health -= amount }

// HasAbility returns true if the minion has an ability with the given keyword.
func (m *Minion) HasAbility(kw Keyword) bool { return m.abilities.Has(kw) }

// AddAbility appends an ability to the minion.
func (m *Minion) AddAbility(ab Ability) { m.abilities.Add(ab) }

// RemoveAbility removes the first ability matching the keyword.
func (m *Minion) RemoveAbility(kw Keyword) { m.abilities.Remove(kw) }

// EffectsByKeyword returns all effect payloads for the given keyword.
func (m *Minion) EffectsByKeyword(kw Keyword) []Effect { return m.abilities.ByKeyword(kw) }

// MergeGolden combines two copies into a golden minion.
// Stats = sum of both copies (retains buffs). Abilities = doubled from template.
func (m *Minion) MergeGolden(other *Minion) *Minion {
	return &Minion{
		template:  m.template,
		attack:    m.attack + other.attack,
		health:    m.health + other.health,
		golden:    true,
		abilities: m.template.GoldenAbilities().Clone(),
	}
}

func (m *Minion) Clone() *Minion {
	if m == nil {
		return nil
	}
	return &Minion{
		template:  m.template,
		attack:    m.attack,
		health:    m.health,
		golden:    m.golden,
		abilities: m.abilities.Clone(),
		combatID:  m.combatID,
	}
}
