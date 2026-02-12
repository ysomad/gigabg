package catalog

import "github.com/ysomad/gigabg/game"

// spells returns all spell card templates.
func spells() map[string]*template {
	return map[string]*template{
		game.TripleRewardID: {
			kind:        game.CardKindSpell,
			name:        "Triple Reward",
			description: "Discover a minion from a higher tavern tier.",
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerSpell, Effect: &game.DiscoverCard{TierOffset: 1}},
			},
		},
		"bolster": {
			kind:        game.CardKindSpell,
			name:        "Bolster",
			description: "Give a friendly minion Taunt.",
			tier:        game.Tier1,
			cost:        1,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerSpell, Effect: &game.GiveKeyword{
					Target:  game.Target{Type: game.TargetFriendlySelected},
					Keyword: game.KeywordTaunt,
				}},
			},
		},
		"golden_touch": {
			kind:        game.CardKindSpell,
			name:        "Golden Touch",
			description: "Make a friendly minion Golden.",
			tier:        game.Tier6,
			cost:        4,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerSpell, Effect: &game.MakeGolden{}},
			},
		},
	}
}
