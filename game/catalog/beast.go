package catalog

import "github.com/ysomad/gigabg/game"

// beasts returns all beasts tribe card templates.
func beasts() map[string]*template {
	return map[string]*template{
		"alleycat": {
			name:   "Alleycat",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},
		"big_bad_wolf": {
			name:   "Big Bad Wolf",
			tier:   game.Tier2,
			attack: 3,
			health: 2,
		},
		"savannah_highmane": {
			name:   "Savannah Highmane",
			tier:   game.Tier3,
			attack: 6,
			health: 5,
		},
		"mama_bear": {
			name:   "Mama Bear",
			tier:   game.Tier4,
			attack: 5,
			health: 5,
		},
		"king_krush": {
			name:   "King Krush",
			tier:   game.Tier5,
			attack: 8,
			health: 8,
		},
		"ghastcoiler": {
			name:     "Ghastcoiler",
			tier:     game.Tier6,
			attack:   7,
			health:   7,
			keywords: game.NewKeywords(game.KeywordReborn),
		},
	}
}
