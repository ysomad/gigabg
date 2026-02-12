package card

import "github.com/ysomad/gigabg/game"

// nagas returns all nagas tribe card templates.
func nagas() map[string]*template {
	return map[string]*template{
		// Tier 1
		"tidescale_hunter": {
			name:   "Tidescale Hunter",
			tier:   game.Tier1,
			attack: 1,
			health: 2,
		},

		// Tier 2
		"eventide_brute": {
			name:     "Eventide Brute",
			tier:     game.Tier2,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordTaunt),
		},

		// Tier 3
		"slitherspear": {
			name:   "Slitherspear",
			tier:   game.Tier3,
			attack: 4,
			health: 5,
		},

		// Tier 4
		"electric_eel": {
			name:     "Electric Eel",
			tier:     game.Tier4,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordWindfury),
		},

		// Tier 5
		"leviathan": {
			name:     "Leviathan",
			tier:     game.Tier5,
			attack:   7,
			health:   7,
			keywords: game.NewKeywords(game.KeywordTaunt, game.KeywordDivineShield),
		},

		// Tier 6
		"zola_the_gorgon": {
			name:     "Zola the Gorgon",
			tier:     game.Tier6,
			attack:   8,
			health:   8,
			keywords: game.NewKeywords(game.KeywordPoisonous, game.KeywordReborn),
		},
	}
}
