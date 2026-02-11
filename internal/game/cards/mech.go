package cards

import "github.com/ysomad/gigabg/internal/game"

// Mech returns all Mech tribe card templates.
func Mech() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"annoy_o_tron": {
			ID:       "annoy_o_tron",
			Name:     "Annoy-o-Tron",
			Tribe:    game.TribeMech,
			Tier:     game.Tier1,
			Attack:   1,
			Health:   2,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDivineShield),
		},
		// Tier 2
		"shield_bot": {
			ID:       "shield_bot",
			Name:     "Shield Bot",
			Tribe:    game.TribeMech,
			Tier:     game.Tier2,
			Attack:   2,
			Health:   2,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield),
		},

		// Tier 3
		"mech_tank": {
			ID:       "mech_tank",
			Name:     "Mech Tank",
			Tribe:    game.TribeMech,
			Tier:     game.Tier3,
			Attack:   4,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 4
		"mechano_tank": {
			ID:       "mechano_tank",
			Name:     "Mechano-Tank",
			Tribe:    game.TribeMech,
			Tier:     game.Tier4,
			Attack:   5,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield),
		},

		// Tier 5
		"foe_reaper_4000": {
			ID:       "foe_reaper_4000",
			Name:     "Foe Reaper 4000",
			Tribe:    game.TribeMech,
			Tier:     game.Tier5,
			Attack:   6,
			Health:   9,
			Keywords: game.Keywords(0).Add(game.KeywordCleave),
		},

		// Tier 6
		"mekgineer_thermaplugg": {
			ID:       "mekgineer_thermaplugg",
			Name:     "Mekgineer Thermaplugg",
			Tribe:    game.TribeMech,
			Tier:     game.Tier6,
			Attack:   9,
			Health:   7,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield).Add(game.KeywordWindfury),
		},
	}
}
