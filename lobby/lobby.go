package lobby

import (
	"math/rand/v2"
	"time"

	"github.com/ysomad/gigabg/errorsx"
	"github.com/ysomad/gigabg/game"
)

const (
	RecruitDuration = 10 * time.Second
	CombatDuration  = time.Second
)

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

func (s State) String() string {
	switch s {
	case StateWaiting:
		return "Waiting"
	case StatePlaying:
		return "Playing"
	case StateFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type Lobby struct {
	state             State
	players           []*game.Player
	cards             game.CardStore
	pool              *game.CardPool
	turn              int
	phase             game.Phase
	phaseEndTimestamp int64 // unix seconds
}

func New(cards game.CardStore) *Lobby {
	return &Lobby{
		state:   StateWaiting,
		players: make([]*game.Player, 0, game.MaxPlayers),
		cards:   cards,
		pool:    game.NewCardPool(cards),
	}
}

// AddPlayer adds a player to the lobby. Auto-starts when 8 players join.
func (l *Lobby) AddPlayer(id string) error {
	if l.state != StateWaiting {
		return ErrGameStarted
	}
	if len(l.players) >= game.MaxPlayers {
		return ErrLobbyFull
	}

	l.players = append(l.players, game.NewPlayer(id))

	if len(l.players) == game.MaxPlayers {
		l.start()
	}
	return nil
}

func (l *Lobby) start() {
	l.state = StatePlaying
	l.turn = 1
	l.startRecruit()
}

func (l *Lobby) startRecruit() {
	l.phase = game.PhaseRecruit
	l.phaseEndTimestamp = time.Now().Add(RecruitDuration).Unix()

	for _, p := range l.players {
		p.StartTurn(l.pool, l.turn)
	}
}

func (l *Lobby) startCombat() {
	l.phase = game.PhaseCombat
	l.phaseEndTimestamp = time.Now().Add(CombatDuration).Unix()
	l.resolveDiscovers()
	// TODO: run combat logic
}

// resolveDiscovers auto-picks a random discover option for players who
// didn't pick before combat started, so the spell isn't wasted.
// Unpicked options are returned to the pool.
func (l *Lobby) resolveDiscovers() {
	for _, p := range l.players {
		l.resolvePlayerDiscover(p)
	}
}

func (l *Lobby) resolvePlayerDiscover(p *game.Player) {
	defer func() { p.DiscoverOptions = nil }()

	if len(p.DiscoverOptions) == 0 {
		return
	}

	if len(p.Hand) >= game.MaxHandSize {
		l.pool.ReturnCards(p.DiscoverOptions)
		return
	}

	idx := rand.IntN(len(p.DiscoverOptions))
	p.Hand = append(p.Hand, p.DiscoverOptions[idx])

	for i, c := range p.DiscoverOptions {
		if i != idx {
			l.pool.ReturnCard(c)
		}
	}
}

// AdvancePhase checks if phase should advance and does so.
// Returns true if phase changed.
func (l *Lobby) AdvancePhase() bool {
	if l.state != StatePlaying {
		return false
	}

	now := time.Now().Unix()
	if now < l.phaseEndTimestamp {
		return false
	}

	switch l.phase {
	case game.PhaseRecruit:
		l.startCombat()
	case game.PhaseCombat:
		l.turn++
		l.startRecruit()
	}

	return true
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
func (l *Lobby) Turn() int {
	return l.turn
}

// Phase returns the current game phase.
func (l *Lobby) Phase() game.Phase {
	return l.phase
}

// PhaseEndTimestamp returns when the current phase ends (unix seconds).
func (l *Lobby) PhaseEndTimestamp() int64 {
	return l.phaseEndTimestamp
}

// Cards returns the card store.
func (l *Lobby) Cards() game.CardStore {
	return l.cards
}

// Pool returns the card pool.
func (l *Lobby) Pool() *game.CardPool {
	return l.pool
}
