package lobby

import (
	"github.com/ysomad/gigabg/errorsx"
	"github.com/ysomad/gigabg/game"
)

const MaxPlayers = 8

const (
	ErrGameStarted      errorsx.Error = "game already started"
	ErrGameNotStarted   errorsx.Error = "game not started"
	ErrLobbyFull        errorsx.Error = "lobby is full"
	ErrNotEnoughPlayers errorsx.Error = "not enough players"
)

type State int

const (
	StateWaiting State = iota + 1
	StatePlaying
	StateFinished
)

type Lobby struct {
	state   State
	players []*game.Player
	pool    *game.CardPool
	turn    uint32
	phase   game.Phase
}

func New(cards game.CardStore) *Lobby {
	return &Lobby{
		state:   StateWaiting,
		players: make([]*game.Player, 0, MaxPlayers),
		pool:    game.NewCardPool(cards),
	}
}

// AddPlayer adds a player to the lobby. Auto-starts when 8 players join.
func (l *Lobby) AddPlayer(id string) error {
	if l.state != StateWaiting {
		return ErrGameStarted
	}
	if len(l.players) >= MaxPlayers {
		return ErrLobbyFull
	}

	l.players = append(l.players, game.NewPlayer(id))

	if len(l.players) == MaxPlayers {
		l.start()
	}
	return nil
}

func (l *Lobby) start() {
	l.state = StatePlaying
	l.turn = 1
	l.phase = game.PhaseRecruit

	for _, p := range l.players {
		p.StartTurn(l.pool)
	}
}

// State returns the current lobby state.
func (l *Lobby) State() State {
	return l.state
}

// PlayerCount returns the number of players in the lobby.
func (l *Lobby) PlayerCount() int {
	return len(l.players)
}

// Players returns all players in the lobby.
func (l *Lobby) Players() []*game.Player {
	return l.players
}

// Player returns the player with the given ID.
func (l *Lobby) Player(id string) *game.Player {
	for _, p := range l.players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Turn returns the current turn number.
func (l *Lobby) Turn() uint32 {
	return l.turn
}

// Phase returns the current game phase.
func (l *Lobby) Phase() game.Phase {
	return l.phase
}

// Pool returns the card pool.
func (l *Lobby) Pool() *game.CardPool {
	return l.pool
}

// nextTurn advances to the next turn.
func (l *Lobby) nextTurn() {
	l.turn++

	for _, p := range l.players {
		p.StartTurn(l.pool)
	}

	l.phase = game.PhaseRecruit
}

// StartCombat transitions from recruit phase to combat phase.
func (l *Lobby) StartCombat() {
	l.phase = game.PhaseCombat
	// TODO: Pair players and run combat
}

// EndCombat transitions back to recruit phase.
func (l *Lobby) EndCombat() {
	l.nextTurn()
}
