package catalog

import "github.com/ysomad/gigabg/game"

// demons returns all demons tribe card templates.
func demons() map[string]*template {
	return map[string]*template{
		// Tier 1
		"imp": {
			name:   "Imp",
			tier:   game.Tier1,
			attack: 1,
			health: 1,
		},
		// Tier 2
		"imprisoner": {
			name:        "Imprisoner",
			description: "Taunt. Deathrattle: Summon a 1/1 Imp.",
			tier:        game.Tier2,
			attack:      3,
			health:      3,
			keywords:    game.NewKeywords(game.KeywordTaunt),
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.SummonMinion{TemplateID: "imp"}},
			},
		},
		// Tier 3
		"doom_lord": {
			name:   "Doom Lord",
			tier:   game.Tier3,
			attack: 5,
			health: 4,
		},

		// Tier 4
		"pit_lord": {
			name:   "Pit Lord",
			tier:   game.Tier4,
			attack: 5,
			health: 6,
		},

		// Tier 5
		"voidlord": {
			name:        "Voidlord",
			description: "Taunt. Deathrattle: Summon a 1/3 Voidwalker.",
			tier:        game.Tier5,
			attack:      3,
			health:      9,
			keywords:    game.NewKeywords(game.KeywordTaunt),
			effects: []game.TriggeredEffect{
				{Trigger: game.TriggerDeathrattle, Effect: &game.SummonMinion{TemplateID: "voidwalker"}},
			},
		},

		// Tier 6
		"supreme_abyssal": {
			name:     "Supreme Abyssal",
			tier:     game.Tier6,
			attack:   10,
			health:   10,
			keywords: game.NewKeywords(game.KeywordWindfury),
		},
	}
}
