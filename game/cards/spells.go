package cards

import "github.com/ysomad/gigabg/game"

// Spells returns all spell card templates.
func Spells() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		"triple_reward": {
			Kind:        game.CardKindSpell,
			Name:        "Triple Reward",
			Description: "Discover a minion from a higher tavern tier.",
			SpellEffect: &game.Effect{Type: game.EffectDiscover},
			Keywords:    game.Keywords(0).Add(game.KeywordDiscover),
		},
	}
}
