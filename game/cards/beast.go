package cards

import "github.com/ysomad/gigabg/game"

// Beast returns all Beast tribe card templates.
func Beast() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		"alleycat": {
			Name:   "Alleycat",
			Tribe:  game.TribeBeast,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 1,
		},
		"big_bad_wolf": {
			Name:   "Big Bad Wolf",
			Tribe:  game.TribeBeast,
			Tier:   game.Tier2,
			Attack: 3,
			Health: 2,
		},
		"savannah_highmane": {
			Name:   "Savannah Highmane",
			Tribe:  game.TribeBeast,
			Tier:   game.Tier3,
			Attack: 6,
			Health: 5,
		},
		"mama_bear": {
			Name:   "Mama Bear",
			Tribe:  game.TribeBeast,
			Tier:   game.Tier4,
			Attack: 5,
			Health: 5,
		},
		"king_krush": {
			Name:   "King Krush",
			Tribe:  game.TribeBeast,
			Tier:   game.Tier5,
			Attack: 8,
			Health: 8,
		},
		"ghastcoiler": {
			Name:     "Ghastcoiler",
			Tribe:    game.TribeBeast,
			Tier:     game.Tier6,
			Attack:   7,
			Health:   7,
			Keywords: game.Keywords(0).Add(game.KeywordReborn),
		},
	}
}
