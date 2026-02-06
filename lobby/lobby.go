package lobby

import (
	"math/rand/v2"
	"time"

	"github.com/ysomad/gigabg/errors"
	"github.com/ysomad/gigabg/game"
)

const (
	ErrGameStarted      errors.Error = "game already started"
	ErrGameNotStarted   errors.Error = "game not started"
	ErrLobbyFull        errors.Error = "lobby is full"
	ErrNotEnoughPlayers errors.Error = "not enough players"
)

type State uint8

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

const maxCombatLogs = 3

type combatPairing struct {
	opponentID    string
	playerBoard   game.Board // cloned with combat IDs
	opponentBoard game.Board // cloned with combat IDs
}

func newCombatPairing(opponentID string, playerBoard, opponentBoard game.Board) combatPairing {
	return combatPairing{
		opponentID:    opponentID,
		playerBoard:   playerBoard,
		opponentBoard: opponentBoard,
	}
}

type Lobby struct {
	state             State
	players           []*game.Player
	cards             game.CardStore
	pool              *game.CardPool
	turn              int
	phase             game.Phase
	phaseEndTimestamp int64                          // unix seconds
	combatLogs        map[string][]game.CombatResult // playerID -> last N results
	combatAnimations  []game.CombatAnimation         // ephemeral, cleared after send
	combatPairings    map[string]combatPairing       // playerID -> pairing, combat phase only
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
	l.phaseEndTimestamp = time.Now().Add(game.RecruitDuration).Unix()

	for _, p := range l.players {
		p.StartTurn(l.pool, l.turn)
	}
}

func (l *Lobby) startCombat() {
	l.phase = game.PhaseCombat
	l.phaseEndTimestamp = time.Now().Add(game.CombatDuration).Unix()
	l.resolveDiscovers()
	l.runCombat()
}

func (l *Lobby) runCombat() {
	if len(l.players) < 2 {
		return
	}

	if l.combatLogs == nil {
		l.combatLogs = make(map[string][]game.CombatResult, len(l.players))
	}

	// Shuffle for random pairing.
	perm := rand.Perm(len(l.players))

	l.combatAnimations = l.combatAnimations[:0]
	l.combatPairings = make(map[string]combatPairing, len(l.players))

	// Pair consecutive players; odd player out gets a bye.
	for i := 0; i+1 < len(perm); i += 2 {
		l.resolvePairing(l.players[perm[i]], l.players[perm[i+1]])
	}

	l.checkFinished()
}

func (l *Lobby) checkFinished() {
	alive := 0
	for _, p := range l.players {
		if p.Alive() {
			alive++
		}
	}
	if alive <= 1 {
		l.state = StateFinished
	}
}

func (l *Lobby) resolvePairing(p1, p2 *game.Player) {
	combat := game.NewCombat(p1, p2)

	p1Board, p2Board := combat.Boards()
	l.combatPairings[p1.ID()] = newCombatPairing(p2.ID(), p1Board, p2Board)
	l.combatPairings[p2.ID()] = newCombatPairing(p1.ID(), p2Board, p1Board)

	r1, r2 := combat.Run()

	if r1.WinnerID != "" && r1.Damage > 0 {
		loser := p2
		if r1.WinnerID == p2.ID() {
			loser = p1
		}
		if loser.Alive() {
			loser.TakeDamage(r1.Damage)
		}
	}

	l.appendCombatResult(p1.ID(), r1)
	l.appendCombatResult(p2.ID(), r2)
	l.combatAnimations = append(l.combatAnimations, combat.Animation())
}

func (l *Lobby) appendCombatResult(playerID string, result game.CombatResult) {
	logs := append(l.combatLogs[playerID], result)
	if len(logs) > maxCombatLogs {
		logs = logs[len(logs)-maxCombatLogs:]
	}
	l.combatLogs[playerID] = logs
}

// resolveDiscovers auto-picks a random discover option for players who
// didn't pick before combat started, so the spell isn't wasted.
// Unpicked options are returned to the pool.
func (l *Lobby) resolveDiscovers() {
	for _, p := range l.players {
		p.ResolveDiscover(l.pool)
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
		if p.ID() == id {
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

// CombatResults returns combat results for the given player (last 3).
func (l *Lobby) CombatResults(playerID string) []game.CombatResult {
	return l.combatLogs[playerID]
}

// CombatAnimations returns pending combat animations and clears them.
func (l *Lobby) CombatAnimations() []game.CombatAnimation {
	anims := l.combatAnimations
	l.combatAnimations = nil
	return anims
}

// CombatPairing returns the combat pairing for the given player (combat phase only).
func (l *Lobby) CombatPairing(playerID string) (opponentID string, playerBoard, opponentBoard game.Board, ok bool) {
	p, ok := l.combatPairings[playerID]
	if !ok {
		return "", game.Board{}, game.Board{}, false
	}
	return p.opponentID, p.playerBoard, p.opponentBoard, true
}

// Pool returns the card pool.
func (l *Lobby) Pool() *game.CardPool {
	return l.pool
}
