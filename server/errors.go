package server

import "github.com/ysomad/gigabg/errorsx"

const (
	ErrLobbyNotFound errorsx.Error = "lobby not found"
	ErrLobbyExists   errorsx.Error = "lobby already exists"
	ErrLobbyFull     errorsx.Error = "lobby is full"
	ErrNotAllowed    errorsx.Error = "player not allowed in lobby"
)
