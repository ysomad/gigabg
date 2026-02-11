package cards

import "github.com/ysomad/gigabg/game"

// Undead returns all Undead tribe card templates.
func Undead() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"skeleton": {
			ID:     "skeleton",
			Name:   "Skeleton",
			Tribe:  game.TribeUndead,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 1,
		},
		// Tier 2
		"risen_guard": {
			ID:       "risen_guard",
			Name:     "Risen Guard",
			Tribe:    game.TribeUndead,
			Tier:     game.Tier2,
			Attack:   2,
			Health:   4,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt),
		},
		// Tier 3
		"lich": {
			ID:     "lich",
			Name:   "Lich",
			Tribe:  game.TribeUndead,
			Tier:   game.Tier3,
			Attack: 4,
			Health: 6,
		},
		// Tier 4
		"bone_baron": {
			ID:          "bone_baron",
			Name:        "Bone Baron",
			Description: "Deathrattle: Summon a 1/1 Skeleton.",
			Tribe:       game.TribeUndead,
			Tier:        game.Tier4,
			Attack:      5,
			Health:      5,
			Keywords:    game.Keywords(0).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:   game.EffectSummon,
				CardID: "skeleton",
			},
		},
		// Tier 5
		"lich_king": {
			ID:       "lich_king",
			Name:     "Lich King",
			Tribe:    game.TribeUndead,
			Tier:     game.Tier5,
			Attack:   8,
			Health:   8,
			Keywords: game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordReborn),
		},
		// Tier 6
		"kelthuzad": {
			ID:          "kelthuzad",
			Name:        "Kel'Thuzad",
			Description: "Deathrattle: Give your minions +2/+2.",
			Tribe:       game.TribeUndead,
			Tier:        game.Tier6,
			Attack:      6,
			Health:      8,
			Keywords:    game.Keywords(0).Add(game.KeywordReborn).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:       game.EffectBuffStats,
				Target:     game.Target{Type: game.TargetAllFriendly, Filter: game.TargetFilter{ExcludeSelf: true}},
				Attack:     2,
				Health:     2,
				Persistent: true,
			},
		},
	}
}
