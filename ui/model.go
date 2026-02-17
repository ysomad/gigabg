package ui

import (
	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
)

// PlayerEntry is a player summary for the sidebar.
type PlayerEntry struct {
	ID            game.PlayerID
	HP            int
	ShopTier      game.Tier
	CombatResults []api.CombatResult
	TopTribe      game.Tribe
	TopTribeCount int
}
