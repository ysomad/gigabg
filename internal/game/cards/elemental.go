package cards

import "github.com/ysomad/gigabg/internal/game"

// Elemental returns all Elemental tribe card templates.
func Elemental() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"sellemental": {
			ID:     "sellemental",
			Name:   "Sellemental",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier1,
			Attack: 2,
			Health: 2,
		},
		"water_droplet": {
			ID:     "water_droplet",
			Name:   "Water Droplet",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 1,
		},
		"refreshing_anomaly": {
			ID:     "refreshing_anomaly",
			Name:   "Refreshing Anomaly",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 3,
		},
		"molten_rock": {
			ID:       "molten_rock",
			Name:     "Molten Rock",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier1,
			Attack:   2,
			Health:   3,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"party_elemental": {
			ID:     "party_elemental",
			Name:   "Party Elemental",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier1,
			Attack: 2,
			Health: 2,
		},
		"spark": {
			ID:     "spark",
			Name:   "Spark",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier1,
			Attack: 2,
			Health: 1,
		},

		// Tier 2
		"arcane_assistant": {
			ID:     "arcane_assistant",
			Name:   "Arcane Assistant",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier2,
			Attack: 3,
			Health: 3,
		},
		"crackling_cyclone": {
			ID:       "crackling_cyclone",
			Name:     "Crackling Cyclone",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier2,
			Attack:   4,
			Health:   1,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury).Add(game.KeywordDivineShield),
		},
		"stasis_elemental": {
			ID:     "stasis_elemental",
			Name:   "Stasis Elemental",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier2,
			Attack: 4,
			Health: 4,
		},
		"recycling_wraith": {
			ID:     "recycling_wraith",
			Name:   "Recycling Wraith",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier2,
			Attack: 5,
			Health: 4,
		},
		"magmaloc": {
			ID:       "magmaloc",
			Name:     "Magmaloc",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier2,
			Attack:   3,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 3
		"wildfire_elemental": {
			ID:     "wildfire_elemental",
			Name:   "Wildfire Elemental",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier3,
			Attack: 7,
			Health: 3,
		},
		"stormwatcher": {
			ID:       "stormwatcher",
			Name:     "Stormwatcher",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier3,
			Attack:   4,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury),
		},
		"lieutenant_garr": {
			ID:       "lieutenant_garr",
			Name:     "Lieutenant Garr",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier3,
			Attack:   5,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"dancing_barnstormer": {
			ID:     "dancing_barnstormer",
			Name:   "Dancing Barnstormer",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier3,
			Attack: 4,
			Health: 4,
		},

		// Tier 4
		"majordomo_executus": {
			ID:     "majordomo_executus",
			Name:   "Majordomo Executus",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier4,
			Attack: 6,
			Health: 3,
		},
		"gentle_djinni": {
			ID:       "gentle_djinni",
			Name:     "Gentle Djinni",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier4,
			Attack:   4,
			Health:   5,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"rock_elemental": {
			ID:       "rock_elemental",
			Name:     "Rock Elemental",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier4,
			Attack:   6,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},

		// Tier 5
		"lil_rag": {
			ID:     "lil_rag",
			Name:   "Lil' Rag",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier5,
			Attack: 6,
			Health: 6,
		},
		"nomi": {
			ID:     "nomi",
			Name:   "Nomi, Kitchen Nightmare",
			Tribe:  game.TribeElemental,
			Tier:   game.Tier5,
			Attack: 4,
			Health: 4,
		},

		// Tier 6
		"garr": {
			ID:       "garr",
			Name:     "Garr",
			Tribe:    game.TribeElemental,
			Tier:     game.Tier6,
			Attack:   8,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDivineShield),
		},
	}
}
