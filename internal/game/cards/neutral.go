package cards

import "github.com/ysomad/gigabg/internal/game"

// Neutral returns all Neutral tribe card templates.
func Neutral() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 2
		"spawn_of_nzoth": {
			ID:          "spawn_of_nzoth",
			Name:        "Spawn of N'Zoth",
			Description: "Deathrattle: Give your minions +1/+1.",
			Tribe:       game.TribeNeutral,
			Tier:        game.Tier2,
			Attack:      2,
			Health:      2,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:       game.EffectBuffStats,
				Target:     game.Target{Type: game.TargetAllFriendly, Filter: game.TargetFilter{ExcludeSelf: true}},
				Attack:     1,
				Health:     1,
				Persistent: true,
			},
		},
		"selfless_hero": {
			ID:          "selfless_hero",
			Name:        "Selfless Hero",
			Description: "Deathrattle: Give a random friendly minion Divine Shield.",
			Tribe:       game.TribeNeutral,
			Tier:        game.Tier2,
			Attack:      2,
			Health:      1,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:    game.EffectGiveKeyword,
				Target:  game.Target{Type: game.TargetRandomFriendly, Filter: game.TargetFilter{ExcludeSelf: true}},
				Keyword: game.KeywordDivineShield,
			},
		},
	}
}
