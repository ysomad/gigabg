package cards

import "github.com/ysomad/gigabg/internal/game"

// Demon returns all Demon tribe card templates.
func Demon() map[string]*game.CardTemplate {
	return map[string]*game.CardTemplate{
		// Tier 1
		"imp": {
			ID:     "imp",
			Name:   "Imp",
			Tribe:  game.TribeDemon,
			Tier:   game.Tier1,
			Attack: 1,
			Health: 1,
		},
		// Tier 2
		"imprisoner": {
			ID:          "imprisoner",
			Name:        "Imprisoner",
			Description: "Taunt. Deathrattle: Summon a 1/1 Imp.",
			Tribe:       game.TribeDemon,
			Tier:        game.Tier2,
			Attack:      3,
			Health:      3,
			Keywords:    game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:   game.EffectSummon,
				CardID: "imp",
			},
		},
		// Tier 3
		"doom_lord": {
			ID:     "doom_lord",
			Name:   "Doom Lord",
			Tribe:  game.TribeDemon,
			Tier:   game.Tier3,
			Attack: 5,
			Health: 4,
		},

		// Tier 4
		"pit_lord": {
			ID:     "pit_lord",
			Name:   "Pit Lord",
			Tribe:  game.TribeDemon,
			Tier:   game.Tier4,
			Attack: 5,
			Health: 6,
		},

		// Tier 5
		"voidlord": {
			ID:          "voidlord",
			Name:        "Voidlord",
			Description: "Taunt. Deathrattle: Summon a 1/3 Voidwalker.",
			Tribe:       game.TribeDemon,
			Tier:        game.Tier5,
			Attack:      3,
			Health:      9,
			Keywords:    game.Keywords(0).Add(game.KeywordTaunt).Add(game.KeywordDeathrattle),
			Deathrattle: &game.Effect{
				Type:   game.EffectSummon,
				CardID: "voidwalker",
			},
		},

		// Tier 6
		"supreme_abyssal": {
			ID:       "supreme_abyssal",
			Name:     "Supreme Abyssal",
			Tribe:    game.TribeDemon,
			Tier:     game.Tier6,
			Attack:   10,
			Health:   10,
			Keywords: game.Keywords(0).Add(game.KeywordWindfury),
		},
	}
}
