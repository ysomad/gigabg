package cards

import "github.com/ysomad/gigabg/game"

// Quilboar returns all Quilboar tribe card templates.
func Quilboar() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"bristleback_knight": {
			ID:       "bristleback_knight",
			Name:     "Bristleback Knight",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier1,
			Attack:   1,
			Health:   2,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield).Add(game.KeywordWindfury),
		},
		"razorfen_beastmaster": {
			ID:     "razorfen_beastmaster",
			Name:   "Razorfen Beastmaster",
			Tribe:  game.TribeQuilboar,
			Tier:   game.Tier1,
			Attack: 2,
			Health: 2,
		},

		// Tier 2
		"necrolyte_boar": {
			ID:     "necrolyte_boar",
			Name:   "Necrolyte Boar",
			Tribe:  game.TribeQuilboar,
			Tier:   game.Tier2,
			Attack: 3,
			Health: 4,
		},

		// Tier 3
		"charlga": {
			ID:     "charlga",
			Name:   "Charlga",
			Tribe:  game.TribeQuilboar,
			Tier:   game.Tier3,
			Attack: 4,
			Health: 4,
		},
		"aggem_thorncurse": {
			ID:     "aggem_thorncurse",
			Name:   "Aggem Thorncurse",
			Tribe:  game.TribeQuilboar,
			Tier:   game.Tier3,
			Attack: 3,
			Health: 6,
		},
		"bristleback_bloodbrother": {
			ID:       "bristleback_bloodbrother",
			Name:     "Bristleback Bloodbrother",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier3,
			Attack:   5,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"thornweaver": {
			ID:       "thornweaver",
			Name:     "Thornweaver",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier3,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield),
		},

		// Tier 4
		"dynamic_duo": {
			ID:       "dynamic_duo",
			Name:     "Dynamic Duo",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier4,
			Attack:   5,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"bonker": {
			ID:       "bonker",
			Name:     "Bonker",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier4,
			Attack:   5,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury),
		},
		"bristleback_reaver": {
			ID:       "bristleback_reaver",
			Name:     "Bristleback Reaver",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier4,
			Attack:   6,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordCleave),
		},

		// Tier 5
		"necrolyte_overlord": {
			ID:       "necrolyte_overlord",
			Name:     "Necrolyte Overlord",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier5,
			Attack:   6,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordReborn).Add(game.KeywordTaunt),
		},
		"agamaggan": {
			ID:          "agamaggan",
			Name:        "Agamaggan",
			Description: "Deathrattle: Give your Quilboars +3/+3.",
			Tribe:       game.TribeQuilboar,
			Tier:        game.Tier5,
			Attack:      6,
			Health:      6,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type: game.EffectBuffStats,
				Target: game.Target{
					Type:   game.TargetAllFriendly,
					Filter: game.TargetFilter{Tribe: game.TribeQuilboar, ExcludeSelf: true},
				},
				Attack:     3,
				Health:     3,
				Persistent: true,
			},
		},

		// Tier 6
		"charlga_razorflank": {
			ID:       "charlga_razorflank",
			Name:     "Charlga Razorflank",
			Tribe:    game.TribeQuilboar,
			Tier:     game.Tier6,
			Attack:   7,
			Health:   7,
			Keywords: game.Keywords(0).Add(game.KeywordPoisonous).Add(game.KeywordDivineShield).Add(game.KeywordTaunt),
		},
	}
}
