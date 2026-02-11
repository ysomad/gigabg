package gameserver

import "github.com/ysomad/gigabg/internal/pkg/errors"

const (
	ErrLobbyNotFound errors.Error = "lobby not found"
	ErrLobbyExists   errors.Error = "lobby already exists"
	ErrLobbyFull     errors.Error = "lobby is full"
	ErrNotAllowed    errors.Error = "player not allowed in lobby"
)
