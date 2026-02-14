package game

import "time"

const (
	MaxPlayers = 8 // max players in lobby
	MinPlayers = 2 // min players in lobby
)

const (
	RecruitDuration = 120 * time.Second
	CombatDuration  = 5 * time.Second
)

const (
	maxBoardSize = 7  // cards on board
	maxHandSize  = 10 // cards in hand
)

const (
	MinionCost = 3

	initialHP       = 30
	initialGold     = 99 // each player has at game start
	maxGold         = 99 // each player may have during the game by upgrading MaxGold each turn without spells
	minionSellValue = 1
	ShopRefreshCost = 1
)

const discoverCount = 3

const TripleRewardID = "triple_reward"
