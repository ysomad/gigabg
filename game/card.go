package game

// Card is the runtime interface for minions and spells.
type Card interface {
	Template() CardTemplate
	IsSpell() bool
	IsMinion() bool
	IsGolden() bool
}

// NewCard creates a runtime card from a template.
func NewCard(t CardTemplate) Card {
	if t.Kind() == CardKindSpell {
		return NewSpell(t)
	}
	return NewMinion(t)
}

// CardKind distinguishes minions from spells.
type CardKind uint8

const (
	CardKindMinion CardKind = iota
	CardKindSpell
)

// CardTemplate is a read-only card definition.
type CardTemplate interface {
	ID() string
	Name() string
	Description() string
	Kind() CardKind
	Tribe() Tribe
	Tier() Tier
	Cost() int
	Attack() int
	Health() int
	Keywords() Keywords
	Effects() []TriggeredEffect
	GoldenEffects() []TriggeredEffect
	Auras() []Aura
	GoldenAuras() []Aura
}
