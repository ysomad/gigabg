package card

import (
	"slices"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/pkg/errors"
)

const (
	errEmptyName         errors.Error = "name is empty"
	errInvalidTier       errors.Error = "invalid tier"
	errNegativeAttack    errors.Error = "negative attack"
	errHealthNotPositive errors.Error = "health must be positive"
	errSpellNoCost       errors.Error = "spell with tier must have cost"
	errMinionNoCost      errors.Error = "minion must have cost"
)

var _ game.CardTemplate = (*template)(nil)

type template struct {
	// predefined in runtime
	_id    string
	_tribe game.Tribe

	name          string
	description   string
	kind          game.CardKind
	tier          game.Tier
	cost          int
	attack        int
	health        int
	keywords      game.Keywords
	effects       []game.TriggeredEffect
	goldenEffects []game.TriggeredEffect
	auras         []game.Aura
	goldenAuras   []game.Aura
}

func (t *template) ID() string                            { return t._id }
func (t *template) Name() string                          { return t.name }
func (t *template) Description() string                   { return t.description }
func (t *template) Kind() game.CardKind                   { return t.kind }
func (t *template) Tribe() game.Tribe                     { return t._tribe }
func (t *template) Tier() game.Tier                       { return t.tier }
func (t *template) Cost() int                             { return t.cost }
func (t *template) Attack() int                           { return t.attack }
func (t *template) Health() int                           { return t.health }
func (t *template) Keywords() game.Keywords               { return t.keywords }
func (t *template) Effects() []game.TriggeredEffect       { return slices.Clone(t.effects) }
func (t *template) GoldenEffects() []game.TriggeredEffect { return slices.Clone(t.goldenEffects) }
func (t *template) Auras() []game.Aura                    { return slices.Clone(t.auras) }
func (t *template) GoldenAuras() []game.Aura              { return slices.Clone(t.goldenAuras) }

func (t *template) setID(id string)           { t._id = id }
func (t *template) setTribe(tribe game.Tribe) { t._tribe = tribe }

func (t *template) validate() error {
	if t.name == "" {
		return errEmptyName
	}

	if t.kind == game.CardKindSpell {
		if t.tier.IsValid() && t.cost <= 0 {
			return errSpellNoCost
		}
		return nil
	}

	// Minion validation
	if !t.tier.IsValid() {
		return errInvalidTier
	}
	if t.cost <= 0 {
		return errMinionNoCost
	}
	if t.attack < 0 {
		return errNegativeAttack
	}
	if t.health <= 0 {
		return errHealthNotPositive
	}

	return nil
}

func (t *template) initGoldenDefaults() {
	if t.kind != game.CardKindMinion {
		return
	}
	if len(t.effects) > 0 && len(t.goldenEffects) == 0 {
		t.goldenEffects = game.MakeGoldenEffects(t.effects)
	}
	if len(t.auras) > 0 && len(t.goldenAuras) == 0 {
		t.goldenAuras = game.MakeGoldenAuras(t.auras)
	}
}
