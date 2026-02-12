package card

import "github.com/ysomad/gigabg/game"

// pirates returns all pirates tribe card templates.
func pirates() map[string]*template {
	return map[string]*template{
		// Tier 1
		"pirate_recruit": {
			name:   "Pirate Recruit",
			tier:   game.Tier1,
			attack: 1,
			health: 2,
		},

		// Tier 2
		"yo_ho_ogre": {
			name:   "Yo-Ho-Ogre",
			tier:   game.Tier2,
			attack: 2,
			health: 6,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},

		// Tier 3
		"salty_veteran": {
			name:   "Salty Veteran",
			tier:   game.Tier3,
			attack: 3,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},

		// Tier 4
		"goldgrubber": {
			name:   "Goldgrubber",
			tier:   game.Tier4,
			attack: 4,
			health: 4,
		},

		// Tier 5
		"cap_n_hoggarr": {
			name:   "Cap'n Hoggarr",
			tier:   game.Tier5,
			attack: 6,
			health: 6,
		},

		// Tier 6
		"tony_two_tusk": {
			name:   "Tony Two-Tusk",
			tier:   game.Tier6,
			attack: 8,
			health: 8,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordWindfury},
				game.Ability{Keyword: game.KeywordCleave},
			),
		},
	}
}
