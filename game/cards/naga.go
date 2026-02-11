package cards

import "github.com/ysomad/gigabg/game"

// Naga returns all Naga tribe card templates.
func Naga() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"tidescale_hunter": {
			ID:     "tidescale_hunter",
			Name:   "Tidescale Hunter",
			Tribe:  game.TribeNaga,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 2,
		},

		// Tier 2
		"eventide_brute": {
			ID:       "eventide_brute",
			Name:     "Eventide Brute",
			Tribe:    game.TribeNaga,
			Tier:     game.Tier2,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 3
		"slitherspear": {
			ID:     "slitherspear",
			Name:   "Slitherspear",
			Tribe:  game.TribeNaga,
			Tier:   game.Tier3,
			Attack: 4,
			Health: 5,
		},

		// Tier 4
		"electric_eel": {
			ID:       "electric_eel",
			Name:     "Electric Eel",
			Tribe:    game.TribeNaga,
			Tier:     game.Tier4,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury),
		},

		// Tier 5
		"leviathan": {
			ID:       "leviathan",
			Name:     "Leviathan",
			Tribe:    game.TribeNaga,
			Tier:     game.Tier5,
			Attack:   7,
			Health:   7,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDivineShield),
		},

		// Tier 6
		"zola_the_gorgon": {
			ID:       "zola_the_gorgon",
			Name:     "Zola the Gorgon",
			Tribe:    game.TribeNaga,
			Tier:     game.Tier6,
			Attack:   8,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordPoisonous).Add(game.KeywordReborn),
		},
	}
}
