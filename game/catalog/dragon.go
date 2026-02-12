package catalog

import "github.com/ysomad/gigabg/game"

// dragons returns all dragons tribe card templates.
func dragons() map[string]*template {
	return map[string]*template{
		// Tier 1
		"whelp": {
			name:   "Whelp",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},

		// Tier 2
		"twilight_emissary": {
			name:     "Twilight Emissary",
			tier:     game.Tier2,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordTaunt),
		},
		"prized_promo_drake": {
			name:     "Prized Promo-Drake",
			tier:     game.Tier3,
			attack:   5,
			health:   4,
			keywords: game.NewKeywords(game.KeywordDivineShield),
		},

		// Tier 4
		"prestor_prince": {
			name:     "Prestor's Prince",
			tier:     game.Tier4,
			attack:   6,
			health:   6,
			keywords: game.NewKeywords(game.KeywordReborn),
		},

		// Tier 5
		"nadina": {
			name:        "Nadina the Red",
			description: "Deathrattle: Give your Dragons Divine Shield.",
			tier:        game.Tier5,
			attack:      7,
			health:      4,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.GiveKeyword{
					Target: game.Target{
						Type:   game.TargetAllFriendly,
						Filter: game.TargetFilter{Tribe: game.TribeDragon, ExcludeSource: true},
					},
					Keyword: game.KeywordDivineShield,
				}},
			},
		},
	}
}
