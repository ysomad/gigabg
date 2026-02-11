package gameserver

import (
	"sync"

	"github.com/ysomad/gigabg/internal/game"
	"github.com/ysomad/gigabg/internal/lobby"
)

// MemoryStore is an in-memory lobby store.
type MemoryStore struct {
	cards   game.CardStore
	lobbies map[string]*lobby.Lobby
	mu      sync.RWMutex
}

func NewMemoryStore(cards game.CardStore) *MemoryStore {
	return &MemoryStore{
		cards:   cards,
		lobbies: make(map[string]*lobby.Lobby),
	}
}

func (s *MemoryStore) CreateLobby(lobbyID string, maxPlayers int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[lobbyID]; exists {
		return ErrLobbyExists
	}

	l, err := lobby.New(s.cards, maxPlayers)
	if err != nil {
		return err
	}

	s.lobbies[lobbyID] = l
	return nil
}

func (s *MemoryStore) Lobby(lobbyID string) (*lobby.Lobby, error) {
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
	return nil
}
