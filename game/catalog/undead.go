package catalog

import "github.com/ysomad/gigabg/game"

// undeads returns pool undead tribe card templates.
func undeads() map[string]*template {
	return map[string]*template{
		// Tier 2
		"risen_guard": {
			name:     "Risen Guard",
			tier:     game.Tier2,
			attack:   2,
			health:   4,
			keywords: game.NewKeywords(game.KeywordTaunt),
		},
		// Tier 3
		"lich": {
			name:   "Lich",
			tier:   game.Tier3,
			attack: 4,
			health: 6,
		},
		// Tier 4
		"bone_baron": {
			name:        "Bone Baron",
			description: "Deathrattle: Summon a 1/1 Skeleton.",
			tier:        game.Tier4,
			attack:      5,
			health:      5,
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.SummonMinion{TemplateID: "skeleton"}},
			},
		},
		// Tier 5
		"lich_king": {
			name:     "Lich King",
			tier:     game.Tier5,
			attack:   8,
			health:   8,
			keywords: game.NewKeywords(game.KeywordTaunt, game.KeywordReborn),
		},
		// Tier 6
		"kelthuzad": {
			name:        "Kel'Thuzad",
			description: "Deathrattle: Give your minions +2/+2.",
			tier:        game.Tier6,
			attack:      6,
			health:      8,
			keywords:    game.NewKeywords(game.KeywordReborn),
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.BuffStats{
					Target: game.Target{
						Type:   game.TargetAllFriendly,
						Filter: game.TargetFilter{ExcludeSource: true},
					},
					Attack:     2,
					Health:     2,
					Persistent: true,
				}},
			},
		},
	}
}

// undeadTokens returns non-pool undead tokens created by gameplay effects.
func undeadTokens() map[string]*template {
	return map[string]*template{
		"skeleton": {
			name:   "Skeleton",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},
	}
}
