package catalog

import "github.com/ysomad/gigabg/game"

// neutrals returns all neutrals tribe card templates.
func neutrals() map[string]*template {
	return map[string]*template{
		// Tier 2
		"spawn_of_nzoth": {
			name:        "Spawn of N'Zoth",
			description: "Deathrattle: Give your minions +1/+1.",
			tier:        game.Tier2,
			attack:      2,
			health:      2,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.BuffStats{
					Target: game.Target{
						Type:   game.TargetAllFriendly,
						Filter: game.TargetFilter{ExcludeSource: true},
					},
					Attack:     1,
					Health:     1,
					Persistent: true,
				}},
			},
		},
		"selfless_hero": {
			name:        "Selfless Hero",
			description: "Deathrattle: Give a random friendly minion Divine Shield.",
			tier:        game.Tier2,
			attack:      2,
			health:      1,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.GiveKeyword{
					Target: game.Target{
						Type:   game.TargetRandomFriendly,
						Filter: game.TargetFilter{ExcludeSource: true},
					},
					Keyword: game.KeywordDivineShield,
				}},
			},
		},
	}
}
