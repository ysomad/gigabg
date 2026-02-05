package server

import (
	"sync"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/lobby"
)

// LobbyStore manages lobby persistence and access control.
type LobbyStore interface {
	// Create creates a new lobby with allowed player IDs.
	// If playerIDs is empty, any player can join.
	Create(lobbyID string, playerIDs []string) error

	// Get returns the lobby by ID.
	Get(lobbyID string) (*lobby.Lobby, error)

	// CanJoin checks if a player is allowed to join the lobby.
	CanJoin(lobbyID string, playerID string) bool

	// Delete removes a lobby.
	Delete(lobbyID string) error
}

// LobbyEntry holds lobby data and access control.
type LobbyEntry struct {
	Lobby     *lobby.Lobby
	PlayerIDs map[string]struct{} // allowed players, empty = accept all
}

// MemoryStore is an in-memory implementation of LobbyStore.
type MemoryStore struct {
	cards   game.CardStore
	lobbies map[string]*LobbyEntry
	mu      sync.RWMutex
}

func NewMemoryStore(cards game.CardStore) *MemoryStore {
	return &MemoryStore{
		cards:   cards,
		lobbies: make(map[string]*LobbyEntry),
	}
}

func (s *MemoryStore) Create(lobbyID string, playerIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[lobbyID]; exists {
		return ErrLobbyExists
	}

	allowed := make(map[string]struct{}, len(playerIDs))
	for _, id := range playerIDs {
		allowed[id] = struct{}{}
	}

	s.lobbies[lobbyID] = &LobbyEntry{
		Lobby:     lobby.New(s.cards),
		PlayerIDs: allowed,
	}

	return nil
}

func (s *MemoryStore) Get(lobbyID string) (*lobby.Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.lobbies[lobbyID]
	if !exists {
		return nil, ErrLobbyNotFound
	}

	return entry.Lobby, nil
}

func (s *MemoryStore) CanJoin(lobbyID string, playerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.lobbies[lobbyID]
	if !exists {
		return false
	}

	// Empty playerIDs = accept all (for testing)
	if len(entry.PlayerIDs) == 0 {
		return true
	}

	_, allowed := entry.PlayerIDs[playerID]
	return allowed
}

func (s *MemoryStore) Delete(lobbyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[lobbyID]; !exists {
		return ErrLobbyNotFound
	}

	delete(s.lobbies, lobbyID)
	return nil
}
