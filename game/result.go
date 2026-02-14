package game

import "time"

// PlayerPlacement records a player's final standing.
type PlayerPlacement struct {
	PlayerID      string
	Placement     int
	TopTribe Tribe
	TopTribeCount int
}

// GameResult holds the outcome of a completed game.
type GameResult struct {
	WinnerID   string
	Placements []PlayerPlacement // sorted 1st → last
	Duration   time.Duration
	StartedAt  time.Time
	FinishedAt time.Time
}
