package game

import "time"

const DiscoverCount = 3

const (
	MaxPlayers   = 2  // players in lobby
	MaxBoardSize = 7  // cards on board
	MaxHandSize  = 10 // cards in hand
)

const (
	InitialGold     = 3  // each player has at game start
	MaxGold         = 10 // each player may have during the game by upgrading MaxGold each turn without spells
	MinionPrice     = 3
	SellValue       = 1 //
	ShopRefreshCost = 1
)

const (
	RecruitDuration = 20 * time.Second
	CombatDuration  = 5 * time.Second
)
