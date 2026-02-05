package cards

import "github.com/ysomad/gigabg/game"

// Dragon returns all Dragon tribe card templates.
func Dragon() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"whelp": {
			ID:     "whelp",
			Name:   "Whelp",
			Tribe:  game.TribeDragon,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 1,
		},

		// Tier 2
		"twilight_emissary": {
			ID:       "twilight_emissary",
			Name:     "Twilight Emissary",
			Tribe:    game.TribeDragon,
			Tier:     game.Tier2,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		"prized_promo_drake": {
			ID:       "prized_promo_drake",
			Name:     "Prized Promo-Drake",
			Tribe:    game.TribeDragon,
			Tier:     game.Tier3,
			Attack:   5,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield),
		},

		// Tier 4
		"prestor_prince": {
			ID:       "prestor_prince",
			Name:     "Prestor's Prince",
			Tribe:    game.TribeDragon,
			Tier:     game.Tier4,
			Attack:   6,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordReborn),
		},

		// Tier 5
		"nadina": {
			ID:          "nadina",
			Name:        "Nadina the Red",
			Description: "Deathrattle: Give your Dragons Divine Shield.",
			Tribe:       game.TribeDragon,
			Tier:        game.Tier5,
			Attack:      7,
			Health:      4,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type: game.EffectGiveKeyword,
				Target: game.Target{
					Type:   game.TargetAllFriendly,
					Filter: game.TargetFilter{Tribe: game.TribeDragon, ExcludeSelf: true},
				},
				Keyword: game.KeywordDivineShield,
			},
		},

		// Tier 6
		"amalgadon": {
			ID:       "amalgadon",
			Name:     "Amalgadon",
			Tribe:    game.TribeDragon,
			Tier:     game.Tier6,
			Attack:   6,
			Health:   6,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDivineShield).Add(game.KeywordPoisonous),
		},
	}
}
