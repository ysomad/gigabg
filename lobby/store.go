package lobby

import (
	"log/slog"
	"sync"
)

// MemoryStore is an in-memory lobby store.
type MemoryStore struct {
	lobbies map[string]*Lobby
	mu      sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		lobbies: make(map[string]*Lobby),
	}
}

func (s *MemoryStore) CreateLobby(l *Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[l.ID()]; exists {
		return ErrLobbyExists
	}

	s.lobbies[l.ID()] = l

	slog.Info("lobby created", "id", l.ID())

	return nil
}

func (s *MemoryStore) Lobby(lobbyID string) (*Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l, exists := s.lobbies[lobbyID]
	if !exists {
		return nil, ErrLobbyNotFound
	}

	return l, nil
}

func (s *MemoryStore) DeleteLobby(lobbyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[lobbyID]; !exists {
		return ErrLobbyNotFound
	}

	delete(s.lobbies, lobbyID)

	slog.Info("lobby deleted", "id", lobbyID)

	return nil
}
