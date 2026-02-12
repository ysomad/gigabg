package card

import "github.com/ysomad/gigabg/game"

// mechs returns all mechs tribe card templates.
func mechs() map[string]*template {
	return map[string]*template{
		// Tier 1
		"annoy_o_tron": {
			name:   "Annoy-o-Tron",
			tier:   game.Tier1,
			attack: 1,
			health: 2,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},
		// Tier 2
		"shield_bot": {
			name:   "Shield Bot",
			tier:   game.Tier2,
			attack: 2,
			health: 2,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},

		// Tier 3
		"mech_tank": {
			name:   "Mech Tank",
			tier:   game.Tier3,
			attack: 4,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},

		// Tier 4
		"mechano_tank": {
			name:   "Mechano-Tank",
			tier:   game.Tier4,
			attack: 5,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},

		// Tier 5
		"foe_reaper_4000": {
			name:   "Foe Reaper 4000",
			tier:   game.Tier5,
			attack: 6,
			health: 9,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordCleave},
			),
		},

		// Tier 6
		"mekgineer_thermaplugg": {
			name:   "Mekgineer Thermaplugg",
			tier:   game.Tier6,
			attack: 9,
			health: 7,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDivineShield},
				game.Ability{Keyword: game.KeywordWindfury},
			),
		},
	}
}
