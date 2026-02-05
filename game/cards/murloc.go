package cards

import "github.com/ysomad/gigabg/game"

// Murloc returns all Murloc tribe card templates.
func Murloc() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"murloc_tidehunter": {
			ID:     "murloc_tidehunter",
			Name:   "Murloc Tidehunter",
			Tribe:  game.TribeMurloc,
			Tier:   game.Tier1,
			Attack: 2,
			Health: 1,
		},

		// Tier 2
		"toxfin": {
			ID:       "toxfin",
			Name:     "Toxfin",
			Tribe:    game.TribeMurloc,
			Tier:     game.Tier2,
			Attack:   1,
			Health:   2,
			Keywords: game.Keywords(0).Add(game.KeywordPoisonous),
		},

		// Tier 3
		"primalfin_champion": {
			ID:       "primalfin_champion",
			Name:     "Primalfin Champion",
			Tribe:    game.TribeMurloc,
			Tier:     game.Tier3,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordDivineShield),
		},

		// Tier 4
		"primalfin_prime": {
			ID:       "primalfin_prime",
			Name:     "Primalfin Prime",
			Tribe:    game.TribeMurloc,
			Tier:     game.Tier4,
			Attack:   4,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordPoisonous).Add(game.KeywordDivineShield),
		},

		// Tier 5
		"king_bagurgle": {
			ID:          "king_bagurgle",
			Name:        "King Bagurgle",
			Description: "Deathrattle: Give your Murlocs +2/+2.",
			Tribe:       game.TribeMurloc,
			Tier:        game.Tier5,
			Attack:      6,
			Health:      3,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type: game.EffectBuffStats,
				Target: game.Target{
					Type:   game.TargetAllFriendly,
					Filter: game.TargetFilter{Tribe: game.TribeMurloc, ExcludeSelf: true},
				},
				Attack:     2,
				Health:     2,
				Persistent: true,
			},
		},
		"brann_bronzebeard": {
			ID:     "brann_bronzebeard",
			Name:   "Brann Bronzebeard",
			Tribe:  game.TribeMurloc,
			Tier:   game.Tier5,
			Attack: 2,
			Health: 4,
		},

		// Tier 6
		"megasaur": {
			ID:       "megasaur",
			Name:     "Megasaur",
			Tribe:    game.TribeMurloc,
			Tier:     game.Tier6,
			Attack:   8,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordPoisonous).Add(game.KeywordWindfury),
		},
	}
}
