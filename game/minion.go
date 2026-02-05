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

// Minion is a runtime card instance with mutable state.
type Minion struct {
	template *CardTemplate
	attack   int
	health   int
	golden   bool
	keywords Keywords
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
func (m *Minion) Battlecry() *Effect        { return m.template.Battlecry }
func (m *Minion) Deathrattle() *Effect      { return m.template.Deathrattle }
func (m *Minion) Avenge() *AvengeEffect     { return m.template.Avenge }
func (m *Minion) StartOfCombat() *Effect    { return m.template.StartOfCombat }
func (m *Minion) StartOfTurn() *Effect      { return m.template.StartOfTurn }
func (m *Minion) EndOfTurn() *Effect        { return m.template.EndOfTurn }

func (m *Minion) Attack() int        { return m.attack }
func (m *Minion) Health() int        { return m.health }
func (m *Minion) Golden() bool       { return m.golden }
func (m *Minion) Keywords() Keywords { return m.keywords }

func (m *Minion) HasKeyword(k Keyword) bool { return m.keywords.Has(k) }

func (m *Minion) SetAttack(v int)  { m.attack = v }
func (m *Minion) SetHealth(v int)  { m.health = v }
func (m *Minion) SetGolden(v bool) { m.golden = v }

func (m *Minion) AddKeyword(k Keyword)    { m.keywords = m.keywords.Add(k) }
func (m *Minion) RemoveKeyword(k Keyword) { m.keywords = m.keywords.Remove(k) }

func (m *Minion) Clone() *Minion {
	if m == nil {
		return nil
	}
	clone := *m
	return &clone
}
