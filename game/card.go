package game

import (
	"fmt"

	"github.com/ysomad/gigabg/pkg/errors"
)

const (
	ErrEmptyName         errors.Error = "name is empty"
	ErrNoSpellEffect     errors.Error = "spell has no effect"
	ErrInvalidTier       errors.Error = "invalid tier"
	ErrNegativeAttack    errors.Error = "negative attack"
	ErrHealthNotPositive errors.Error = "health must be positive"
	ErrAvengeThreshold   errors.Error = "avenge threshold must be positive"
)

type Card interface {
	TemplateID() string
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

// GoldenStats holds predefined golden card stats and effects.
// Cards can define custom golden overrides; unset fields are auto-populated
// with 2x defaults by the card store.
type GoldenStats struct {
	Attack        int
	Health        int
	Description   string
	Battlecry     *Effect
	Deathrattle   *Effect
	Avenge        *AvengeEffect
	StartOfCombat *Effect
	StartOfTurn   *Effect
	EndOfTurn     *Effect
}

// CardTemplate is a read-only card definition.
// Created once at startup and never modified.
type CardTemplate struct {
	Kind        CardKind
	ID          string
	Name        string
	Description string

	// Minion fields
	Tribe         Tribe
	Tier          Tier
	Attack        int
	Health        int
	Keywords      Keywords
	Battlecry     *Effect
	Deathrattle   *Effect
	Avenge        *AvengeEffect
	StartOfCombat *Effect
	StartOfTurn   *Effect
	EndOfTurn     *Effect

	// Golden version (auto-populated with 2x defaults if nil)
	Golden *GoldenStats

	// Spell fields
	SpellEffect *Effect
}

// AvengeEffect is a triggered effect with a death threshold.
type AvengeEffect struct {
	Effect
	Threshold int // number of friendly deaths to trigger
}

func (t *CardTemplate) IsSpell() bool             { return t.Kind == CardKindSpell }
func (t *CardTemplate) HasKeyword(k Keyword) bool { return t.Keywords.Has(k) }

func (t *CardTemplate) Validate() error {
	if t.Name == "" {
		return ErrEmptyName
	}

	if t.Kind == CardKindSpell {
		if t.SpellEffect == nil {
			return ErrNoSpellEffect
		}
		return nil
	}

	// Minion validation
	if !t.Tier.IsValid() {
		return ErrInvalidTier
	}
	if t.Attack < 0 {
		return ErrNegativeAttack
	}
	if t.Health <= 0 {
		return ErrHealthNotPositive
	}

	// Validate keyword/effect consistency
	checks := []struct {
		kw        Keyword
		hasEffect bool
		name      string
	}{
		{KeywordBattlecry, t.Battlecry != nil, "battlecry"},
		{KeywordDeathrattle, t.Deathrattle != nil, "deathrattle"},
		{KeywordAvenge, t.Avenge != nil, "avenge"},
		{KeywordStartOfCombat, t.StartOfCombat != nil, "start of combat"},
		{KeywordStartOfTurn, t.StartOfTurn != nil, "start of turn"},
		{KeywordEndOfTurn, t.EndOfTurn != nil, "end of turn"},
	}

	for _, c := range checks {
		hasKeyword := t.Keywords.Has(c.kw)
		if hasKeyword && !c.hasEffect {
			return fmt.Errorf("has %s keyword but no effect", c.name)
		}
		if c.hasEffect && !hasKeyword {
			return fmt.Errorf("has %s effect but missing keyword", c.name)
		}
	}

	if t.Avenge != nil && t.Avenge.Threshold <= 0 {
		return ErrAvengeThreshold
	}

	return nil
}
