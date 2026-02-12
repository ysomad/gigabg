package game

import "time"

const (
	MaxPlayers = 8 // max players in lobby
	MinPlayers = 2 // min players in lobby
)

const (
	RecruitDuration = 20 * time.Second
	CombatDuration  = 5 * time.Second
)

const (
	maxBoardSize = 7  // cards on board
	maxHandSize  = 10 // cards in hand
)

const (
	initialHP       = 10
	initialGold     = 3  // each player has at game start
	maxGold         = 10 // each player may have during the game by upgrading MaxGold each turn without spells
	MinionCost      = 3
	minionSellValue = 1
	shopRefreshCost = 1
)

const discoverCount = 3

const TripleRewardID = "triple_reward"
