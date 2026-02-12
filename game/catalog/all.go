package catalog

import "github.com/ysomad/gigabg/game"

// all returns all all-tribe card templates.
func all() map[string]*template {
	return map[string]*template{
		// Tier 1
		"amalgam": {
			name:   "Amalgam",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},

		// Tier 3
		"stasis_dragon": {
			name:     "Stasis Dragon",
			tier:     game.Tier3,
			attack:   4,
			health:   4,
			keywords: game.NewKeywords(game.KeywordReborn),
		},

		// Tier 5
		"lightfang_enforcer": {
			name:   "Lightfang Enforcer",
			tier:   game.Tier5,
			attack: 2,
			health: 2,
		},

		// Tier 6
		"amalgadon": {
			name:     "Amalgadon",
			tier:     game.Tier6,
			attack:   6,
			health:   6,
			keywords: game.NewKeywords(game.KeywordTaunt, game.KeywordDivineShield, game.KeywordPoisonous),
		},
	}
}
