package card

import (
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/pkg/errors"
)

const (
	errEmptyName         errors.Error = "name is empty"
	errNoSpellEffect     errors.Error = "spell has no effect"
	errInvalidTier       errors.Error = "invalid tier"
	errNegativeAttack    errors.Error = "negative attack"
	errHealthNotPositive errors.Error = "health must be positive"
	errAvengeThreshold   errors.Error = "avenge threshold must be positive"
)

type template struct {
	// predefined in runtime
	_id    string
	_tribe game.Tribe

	name        string
	description string
	kind        game.CardKind
	tier        game.Tier
	attack      int
	health      int
	abilities   game.Abilities
	goldenAbs   game.Abilities
}

func (t *template) ID() string                      { return t._id }
func (t *template) Name() string                    { return t.name }
func (t *template) Description() string             { return t.description }
func (t *template) Kind() game.CardKind             { return t.kind }
func (t *template) Tribe() game.Tribe               { return t._tribe }
func (t *template) Tier() game.Tier                 { return t.tier }
func (t *template) Attack() int                     { return t.attack }
func (t *template) Health() int                     { return t.health }
func (t *template) Abilities() game.Abilities       { return t.abilities }
func (t *template) GoldenAbilities() game.Abilities { return t.goldenAbs }

func (t *template) setID(id string)           { t._id = id }
func (t *template) setTribe(tribe game.Tribe) { t._tribe = tribe }

func (t *template) validate() error {
	if t.name == "" {
		return errEmptyName
	}

	if t.kind == game.CardKindSpell {
		if !t.abilities.Has(game.KeywordSpell) {
			return errNoSpellEffect
		}
		return nil
	}

	// Minion validation
	if !t.tier.IsValid() {
		return errInvalidTier
	}
	if t.attack < 0 {
		return errNegativeAttack
	}
	if t.health <= 0 {
		return errHealthNotPositive
	}

	for _, ab := range t.abilities.All() {
		if ab.Keyword == game.KeywordAvenge && ab.Threshold <= 0 {
			return errAvengeThreshold
		}
	}

	return nil
}

func (t *template) initGoldenDefaults() {
	if t.kind != game.CardKindMinion {
		return
	}
	if t.abilities.Len() > 0 {
		t.goldenAbs = t.abilities.Double()
	}
}
