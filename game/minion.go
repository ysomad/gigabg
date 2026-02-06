package game

type Tribe uint8

const (
	TribeNeutral Tribe = iota
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
)

func (t Tribe) String() string {
	switch t {
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
	default:
		return "Neutral"
	}
}

var _ Card = (*Minion)(nil)

// Minion is a runtime card instance with mutable state.
type Minion struct {
	template *CardTemplate
	attack   int
	health   int
	golden   bool
	keywords Keywords
	combatID int // unique within a single combat, 0 = not assigned
}

// NewMinion creates a minion from a card template.
func NewMinion(t *CardTemplate) *Minion {
	return &Minion{
		template: t,
		attack:   t.Attack,
		health:   t.Health,
		keywords: t.Keywords,
	}
}

func (m *Minion) Template() *CardTemplate { return m.template }
func (m *Minion) TemplateID() string      { return m.template.ID }
func (m *Minion) Name() string            { return m.template.Name }
func (m *Minion) Description() string     { return m.template.Description }
func (m *Minion) Tribe() Tribe            { return m.template.Tribe }
func (m *Minion) Tier() Tier              { return m.template.Tier }
func (m *Minion) Battlecry() *Effect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.Battlecry
	}
	return m.template.Battlecry
}

func (m *Minion) Deathrattle() *Effect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.Deathrattle
	}
	return m.template.Deathrattle
}

func (m *Minion) Avenge() *AvengeEffect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.Avenge
	}
	return m.template.Avenge
}

func (m *Minion) StartOfCombat() *Effect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.StartOfCombat
	}
	return m.template.StartOfCombat
}

func (m *Minion) StartOfTurn() *Effect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.StartOfTurn
	}
	return m.template.StartOfTurn
}

func (m *Minion) EndOfTurn() *Effect {
	if m.golden && m.template.Golden != nil {
		return m.template.Golden.EndOfTurn
	}
	return m.template.EndOfTurn
}

func (m *Minion) CombatID() int      { return m.combatID }
func (m *Minion) Attack() int        { return m.attack }
func (m *Minion) Health() int        { return m.health }
func (m *Minion) Keywords() Keywords { return m.keywords }

func (m *Minion) Alive() bool     { return m.health > 0 }
func (m *Minion) CanAttack() bool { return m.health > 0 && m.attack > 0 }

func (m *Minion) IsSpell() bool             { return false }
func (m *Minion) IsMinion() bool            { return true }
func (m *Minion) IsGolden() bool            { return m.golden }
func (m *Minion) HasKeyword(k Keyword) bool { return m.keywords.Has(k) }

func (m *Minion) TakeDamage(amount int) { m.health -= amount }

func (m *Minion) AddKeyword(k Keyword)    { m.keywords = m.keywords.Add(k) }
func (m *Minion) RemoveKeyword(k Keyword) { m.keywords = m.keywords.Remove(k) }

// ToGolden creates a golden version of the minion with 2x base stats,
// preserving keywords from the original.
func (m *Minion) ToGolden() *Minion {
	return &Minion{
		template: m.template,
		attack:   m.template.Golden.Attack,
		health:   m.template.Golden.Health,
		keywords: m.keywords,
		golden:   true,
	}
}

func (m *Minion) Clone() *Minion {
	if m == nil {
		return nil
	}
	clone := *m
	return &clone
}
