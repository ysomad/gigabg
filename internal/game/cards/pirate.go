package cards

import "github.com/ysomad/gigabg/internal/game"

// Pirate returns all Pirate tribe card templates.
func Pirate() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"pirate_recruit": {
			ID:     "pirate_recruit",
			Name:   "Pirate Recruit",
			Tribe:  game.TribePirate,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 2,
		},

		// Tier 2
		"yo_ho_ogre": {
			ID:       "yo_ho_ogre",
			Name:     "Yo-Ho-Ogre",
			Tribe:    game.TribePirate,
			Tier:     game.Tier2,
			Attack:   2,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 3
		"salty_veteran": {
			ID:       "salty_veteran",
			Name:     "Salty Veteran",
			Tribe:    game.TribePirate,
			Tier:     game.Tier3,
			Attack:   3,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 4
		"goldgrubber": {
			ID:     "goldgrubber",
			Name:   "Goldgrubber",
			Tribe:  game.TribePirate,
			Tier:   game.Tier4,
			Attack: 4,
			Health: 4,
		},

		// Tier 5
		"cap_n_hoggarr": {
			ID:     "cap_n_hoggarr",
			Name:   "Cap'n Hoggarr",
			Tribe:  game.TribePirate,
			Tier:   game.Tier5,
			Attack: 6,
			Health: 6,
		},

		// Tier 6
		"tony_two_tusk": {
			ID:       "tony_two_tusk",
			Name:     "Tony Two-Tusk",
			Tribe:    game.TribePirate,
			Tier:     game.Tier6,
			Attack:   8,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury).Add(game.KeywordCleave),
		},
	}
}
