package card

import "github.com/ysomad/gigabg/game"

// quilboars returns all quilboars tribe card templates.
func quilboars() map[string]*template {
	return map[string]*template{
		// Tier 1
		"bristleback_knight": {
			name:   "Bristleback Knight",
			tier:   game.Tier1,
			attack: 1,
			health: 2,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDivineShield},
				game.Ability{Keyword: game.KeywordWindfury},
			),
		},
		"razorfen_beastmaster": {
			name:   "Razorfen Beastmaster",
			tier:   game.Tier1,
			attack: 2,
			health: 2,
		},

		// Tier 2
		"necrolyte_boar": {
			name:   "Necrolyte Boar",
			tier:   game.Tier2,
			attack: 3,
			health: 4,
		},

		// Tier 3
		"charlga": {
			name:   "Charlga",
			tier:   game.Tier3,
			attack: 4,
			health: 4,
		},
		"aggem_thorncurse": {
			name:   "Aggem Thorncurse",
			tier:   game.Tier3,
			attack: 3,
			health: 6,
		},
		"bristleback_bloodbrother": {
			name:   "Bristleback Bloodbrother",
			tier:   game.Tier3,
			attack: 5,
			health: 4,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"thornweaver": {
			name:   "Thornweaver",
			tier:   game.Tier3,
			attack: 4,
			health: 4,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDivineShield},
			),
		},

		// Tier 4
		"dynamic_duo": {
			name:   "Dynamic Duo",
			tier:   game.Tier4,
			attack: 5,
			health: 6,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"bonker": {
			name:   "Bonker",
			tier:   game.Tier4,
			attack: 5,
			health: 5,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordWindfury},
			),
		},
		"bristleback_reaver": {
			name:   "Bristleback Reaver",
			tier:   game.Tier4,
			attack: 6,
			health: 4,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordCleave},
			),
		},

		// Tier 5
		"necrolyte_overlord": {
			name:   "Necrolyte Overlord",
			tier:   game.Tier5,
			attack: 6,
			health: 6,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordReborn},
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
		"agamaggan": {
			name:        "Agamaggan",
			description: "Deathrattle: Give your Quilboars +3/+3.",
			tier:        game.Tier5,
			attack:      6,
			health:      6,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordDeathrattle, Effect: &game.BuffStats{
					Target: game.Target{
						Type:   game.TargetAllFriendly,
						Filter: game.TargetFilter{Tribe: game.TribeQuilboar, ExcludeSource: true},
					},
					Attack:     3,
					Health:     3,
					Persistent: true,
				}},
			),
		},

		// Tier 6
		"charlga_razorflank": {
			name:   "Charlga Razorflank",
			tier:   game.Tier6,
			attack: 7,
			health: 7,
			abilities: game.NewAbilities(
				game.Ability{Keyword: game.KeywordPoisonous},
				game.Ability{Keyword: game.KeywordDivineShield},
				game.Ability{Keyword: game.KeywordTaunt},
			),
		},
	}
}
