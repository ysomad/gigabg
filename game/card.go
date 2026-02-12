package game

// Card is the runtime interface for minions and spells.
type Card interface {
	Template() CardTemplate
	IsSpell() bool
	IsMinion() bool
	IsGolden() bool
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
	Attack() int
	Health() int
	Abilities() Abilities
	GoldenAbilities() Abilities
}
