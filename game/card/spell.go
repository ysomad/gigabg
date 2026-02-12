package card

import "github.com/ysomad/gigabg/game"

// spells returns all spell card templates.
func spells() map[string]*template {
	return map[string]*template{
		"triple_reward": {
			kind:        game.CardKindSpell,
			name:        "Triple Reward",
			description: "Discover a minion from a higher tavern tier.",
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordSpell, Effect: &game.Discover{}},
			),
		},
	}
}
