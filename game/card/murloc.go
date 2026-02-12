package card

import "github.com/ysomad/gigabg/game"

// murlocs returns all murlocs tribe card templates.
func murlocs() map[string]*template {
	return map[string]*template{
		// Tier 1
		"murloc_tidehunter": {
			name:   "Murloc Tidehunter",
			tier:   game.Tier1,
			attack: 2,
			health: 1,
		},

		// Tier 2
		"toxfin": {
			name:     "Toxfin",
			tier:     game.Tier2,
			attack:   1,
			health:   2,
			keywords: game.NewKeywords(game.KeywordPoisonous),
		},

		// Tier 3
		"primalfin_champion": {
			name:     "Primalfin Champion",
			tier:     game.Tier3,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordDivineShield),
		},

		// Tier 4
		"primalfin_prime": {
			name:     "Primalfin Prime",
			tier:     game.Tier4,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordPoisonous, game.KeywordDivineShield),
		},

		// Tier 5
		"king_bagurgle": {
			name:        "King Bagurgle",
			description: "Deathrattle: Give your Murlocs +2/+2.",
			tier:        game.Tier5,
			attack:      6,
			health:      3,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.BuffStats{
					Target: game.Target{
						Type:   game.TargetAllFriendly,
						Filter: game.TargetFilter{Tribe: game.TribeMurloc, ExcludeSource: true},
					},
					Attack:     2,
					Health:     2,
					Persistent: true,
				}},
			},
		},
		"brann_bronzebeard": {
			name:   "Brann Bronzebeard",
			tier:   game.Tier5,
			attack: 2,
			health: 4,
		},

		// Tier 6
		"megasaur": {
			name:     "Megasaur",
			tier:     game.Tier6,
			attack:   8,
			health:   8,
			keywords: game.NewKeywords(game.KeywordPoisonous, game.KeywordWindfury),
		},
	}
}
