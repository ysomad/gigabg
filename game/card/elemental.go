package card

import "github.com/ysomad/gigabg/game"

// elementals returns all elementals tribe card templates.
func elementals() map[string]*template {
	return map[string]*template{
		// Tier 1
		"sellemental": {
			name:   "Sellemental",
			tier:   game.Tier1,
			attack: 2,
			health: 2,
		},
		"water_droplet": {
			name:   "Water Droplet",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},
		"refreshing_anomaly": {
			name:   "Refreshing Anomaly",
			tier:   game.Tier1,
			attack: 1,
			health: 3,
		},
		"molten_rock": {
			name:   "Molten Rock",
			tier:   game.Tier1,
			attack: 2,
			health: 3,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"party_elemental": {
			name:   "Party Elemental",
			tier:   game.Tier1,
			attack: 2,
			health: 2,
		},
		"spark": {
			name:   "Spark",
			tier:   game.Tier1,
			attack: 2,
			health: 1,
		},

		// Tier 2
		"arcane_assistant": {
			name:   "Arcane Assistant",
			tier:   game.Tier2,
			attack: 3,
			health: 3,
		},
		"crackling_cyclone": {
			name:   "Crackling Cyclone",
			tier:   game.Tier2,
			attack: 4,
			health: 1,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordWindfury},
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},
		"stasis_elemental": {
			name:   "Stasis Elemental",
			tier:   game.Tier2,
			attack: 4,
			health: 4,
		},
		"recycling_wraith": {
			name:   "Recycling Wraith",
			tier:   game.Tier2,
			attack: 5,
			health: 4,
		},
		"magmaloc": {
			name:   "Magmaloc",
			tier:   game.Tier2,
			attack: 3,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},

		// Tier 3
		"wildfire_elemental": {
			name:   "Wildfire Elemental",
			tier:   game.Tier3,
			attack: 7,
			health: 3,
		},
		"stormwatcher": {
			name:   "Stormwatcher",
			tier:   game.Tier3,
			attack: 4,
			health: 8,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordWindfury},
			),
		},
		"lieutenant_garr": {
			name:   "Lieutenant Garr",
			tier:   game.Tier3,
			attack: 5,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"dancing_barnstormer": {
			name:   "Dancing Barnstormer",
			tier:   game.Tier3,
			attack: 4,
			health: 4,
		},

		// Tier 4
		"majordomo_executus": {
			name:   "Majordomo Executus",
			tier:   game.Tier4,
			attack: 6,
			health: 3,
		},
		"gentle_djinni": {
			name:   "Gentle Djinni",
			tier:   game.Tier4,
			attack: 4,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"rock_elemental": {
			name:   "Rock Elemental",
			tier:   game.Tier4,
			attack: 6,
			health: 6,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},

		// Tier 5
		"lil_rag": {
			name:   "Lil' Rag",
			tier:   game.Tier5,
			attack: 6,
			health: 6,
		},
		"nomi": {
			name:   "Nomi, Kitchen Nightmare",
			tier:   game.Tier5,
			attack: 4,
			health: 4,
		},

		// Tier 6
		"garr": {
			name:   "Garr",
			tier:   game.Tier6,
			attack: 8,
			health: 8,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},
	}
}
