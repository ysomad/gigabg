package lobby

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"time"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/pkg/errors"
)

const (
	ErrGameStarted        errors.Error = "game already started"
	ErrGameNotStarted     errors.Error = "game not started"
	ErrLobbyFull          errors.Error = "lobby is full"
	ErrLobbyNotFound      errors.Error = "lobby not found"
	ErrLobbyExists        errors.Error = "lobby already exists"
	ErrNotAllowed         errors.Error = "player not allowed in lobby"
	ErrNotEnoughPlayers   errors.Error = "not enough players"
	ErrInvalidPlayerCount errors.Error = "max players must be even, between 2 and 8"
	ErrAlreadyConnected   errors.Error = "player already connected"
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

type CombatPairing struct {
	OpponentID    string
	PlayerBoard   game.Board // cloned with combat IDs
	OpponentBoard game.Board // cloned with combat IDs
}

func newCombatPairing(opponentID string, pb, ob game.Board) CombatPairing {
	return CombatPairing{
		OpponentID:    opponentID,
		PlayerBoard:   pb,
		OpponentBoard: ob,
	}
}

type Lobby struct {
	id         string
	state      State
	maxPlayers int
	players    []*game.Player
	pool       *game.CardPool
	turn       int

	phase       game.Phase
	phaseEndsAt time.Time // when current phase ends

	combatResults  map[string][]game.CombatResult // playerID -> last N results
	combatLogs     []game.CombatLog               // ephemeral, cleared after send
	combatPairings map[string]CombatPairing       // playerID -> pairing, combat phase only
	nextPairings   map[string]string              // playerID -> next opponentID, recruit phase only

	majorityTribes map[string]game.TribeSnapshot // playerID -> snapshot from last combat

	startedAt  time.Time
	eliminated int              // number of eliminated players
	gameResult *game.GameResult // set when game finishes
}

func New(cards game.CardCatalog, maxPlayers int) (*Lobby, error) {
	if maxPlayers < game.MinPlayers || maxPlayers > game.MaxPlayers || maxPlayers%2 != 0 {
		return nil, ErrInvalidPlayerCount
	}
	return &Lobby{
		id:         strconv.Itoa(rand.IntN(100_000_000)),
		state:      StateWaiting,
		maxPlayers: maxPlayers,
		players:    make([]*game.Player, 0, maxPlayers),
		pool:       game.NewCardPool(cards, maxPlayers),
	}, nil
}

func (l *Lobby) ID() string      { return l.id }
func (l *Lobby) SetID(id string) { l.id = id }

// MaxPlayers returns the lobby's max player count.
func (l *Lobby) MaxPlayers() int { return l.maxPlayers }

// AddPlayer adds a player to the lobby. Auto-starts when 8 players join.
func (l *Lobby) AddPlayer(id string) error {
	if l.state != StateWaiting {
		return ErrGameStarted
	}
	if len(l.players) >= l.maxPlayers {
		return ErrLobbyFull
	}
	for _, p := range l.players {
		if p.ID() == id {
			return ErrAlreadyConnected
		}
	}

	l.players = append(l.players, game.NewPlayer(id))

	if len(l.players) == l.maxPlayers {
		l.start()
	}

	return nil
}

func (l *Lobby) start() {
	l.state = StatePlaying
	l.turn = 1
	l.startedAt = time.Now()
	l.startRecruit()
}

func (l *Lobby) startRecruit() {
	l.phase = game.PhaseRecruit
	l.phaseEndsAt = time.Now().Add(game.RecruitDuration)
	l.computeNextPairings()

	for _, p := range l.players {
		p.StartTurn(l.pool, l.turn)
	}
}

// prevOpponents returns a map of playerID -> previous opponentID from the
// current combat pairings.
func (l *Lobby) prevOpponents() map[string]string {
	prev := make(map[string]string, len(l.combatPairings))
	for id, cp := range l.combatPairings {
		prev[id] = cp.OpponentID
	}
	return prev
}

// computeNextPairings pre-determines the next combat opponents so clients
// can display them during recruit phase. Avoids repeating the previous
// opponent when other players are available.
func (l *Lobby) computeNextPairings() {
	alive := make([]*game.Player, 0, len(l.players))
	for _, p := range l.players {
		if p.IsAlive() {
			alive = append(alive, p)
		}
	}

	prev := l.prevOpponents()

	rand.Shuffle(len(alive), func(i, j int) { alive[i], alive[j] = alive[j], alive[i] })

	l.nextPairings = make(map[string]string, len(alive))
	paired := make(map[string]struct{}, len(alive))

	for _, p := range alive {
		if _, ok := paired[p.ID()]; ok {
			continue
		}

		var fallback *game.Player
		var found bool
		for _, q := range alive {
			if _, ok := paired[q.ID()]; q.ID() == p.ID() || ok {
				continue
			}
			if fallback == nil {
				fallback = q
			}
			if prev[p.ID()] != q.ID() && prev[q.ID()] != p.ID() {
				l.nextPairings[p.ID()] = q.ID()
				l.nextPairings[q.ID()] = p.ID()
				paired[p.ID()] = struct{}{}
				paired[q.ID()] = struct{}{}
				found = true
				break
			}
		}

		if !found && fallback != nil {
			l.nextPairings[p.ID()] = fallback.ID()
			l.nextPairings[fallback.ID()] = p.ID()
			paired[p.ID()] = struct{}{}
			paired[fallback.ID()] = struct{}{}
		}
	}
}

// NextOpponentID returns the pre-determined next opponent for the given player.
func (l *Lobby) NextOpponentID(playerID string) string {
	return l.nextPairings[playerID]
}

func (l *Lobby) startCombat() {
	l.phase = game.PhaseCombat
	l.phaseEndsAt = time.Now().Add(game.CombatDuration)
	l.resolveDiscovers()
	l.runCombat()

	// Skip combat timer if no pairing had minions on both sides.
	if l.isCombatTrivial() {
		l.phaseEndsAt = time.Now()
	}
}

// isCombatTrivial returns true if no combat animation had any events (no real fights).
func (l *Lobby) isCombatTrivial() bool {
	for _, anim := range l.combatLogs {
		if len(anim.Events) > 0 {
			return false
		}
	}
	return true
}

func (l *Lobby) runCombat() {
	if len(l.players) < 2 {
		return
	}

	if l.combatResults == nil {
		l.combatResults = make(map[string][]game.CombatResult, len(l.players))
	}

	l.combatLogs = l.combatLogs[:0]
	l.combatPairings = make(map[string]CombatPairing, len(l.players))

	// Use pre-computed pairings from recruit phase.
	resolved := make(map[string]struct{}, len(l.nextPairings))
	for pid, oid := range l.nextPairings {
		if _, ok := resolved[pid]; ok {
			continue
		}
		p1 := l.Player(pid)
		p2 := l.Player(oid)
		if p1 != nil && p2 != nil && p1.IsAlive() && p2.IsAlive() {
			l.resolvePairing(p1, p2)
			resolved[pid] = struct{}{}
			resolved[oid] = struct{}{}
		}
	}

	l.checkFinished()
}

func (l *Lobby) checkFinished() {
	var alive int
	var winner *game.Player
	for _, p := range l.players {
		if p.IsAlive() {
			alive++
			winner = p
		}
	}
	if alive > 1 {
		return
	}

	l.state = StateFinished
	l.phase = game.PhaseFinished

	if winner != nil {
		winner.SetPlacement(1)
	}

	now := time.Now()
	placements := make([]game.PlayerPlacement, len(l.players))
	for i, p := range l.players {
		snap := l.majorityTribes[p.ID()]
		placements[i] = game.PlayerPlacement{
			PlayerID:      p.ID(),
			Placement:     p.Placement(),
			MajorityTribe: snap.Tribe,
			MajorityCount: snap.Count,
		}
	}
	slices.SortFunc(placements, func(a, b game.PlayerPlacement) int {
		return a.Placement - b.Placement
	})

	winnerID := ""
	if winner != nil {
		winnerID = winner.ID()
	}

	l.gameResult = &game.GameResult{
		WinnerID:   winnerID,
		Placements: placements,
		Duration:   now.Sub(l.startedAt),
		StartedAt:  l.startedAt,
		FinishedAt: now,
	}
}

func (l *Lobby) resolvePairing(p1, p2 *game.Player) {
	l.snapshotTribe(p1)
	l.snapshotTribe(p2)

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
		if loser.IsAlive() {
			if loser.TakeDamage(r1.Damage) {
				l.eliminated++
				loser.SetPlacement(len(l.players) - l.eliminated + 1)
			}
		}
	}

	l.appendCombatResult(p1.ID(), r1)
	l.appendCombatResult(p2.ID(), r2)
	l.combatLogs = append(l.combatLogs, combat.Log())
}

func (l *Lobby) snapshotTribe(p *game.Player) {
	if l.majorityTribes == nil {
		l.majorityTribes = make(map[string]game.TribeSnapshot, len(l.players))
	}
	tribe, count := p.Board().MajorityTribe()
	l.majorityTribes[p.ID()] = game.TribeSnapshot{Tribe: tribe, Count: count}
}

// MajorityTribes returns the snapshot of majority tribes from last combat.
func (l *Lobby) MajorityTribes() map[string]game.TribeSnapshot { return l.majorityTribes }

func (l *Lobby) appendCombatResult(playerID string, result game.CombatResult) {
	logs := append(l.combatResults[playerID], result) //nolint:gocritic // intentional new slice
	if len(logs) > maxCombatLogs {
		logs = logs[len(logs)-maxCombatLogs:]
	}
	l.combatResults[playerID] = logs
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

	if time.Now().Before(l.phaseEndsAt) {
		return false
	}

	switch l.phase {
	case game.PhaseWaiting, game.PhaseFinished:
		return false
	case game.PhaseRecruit:
		l.startCombat()
	case game.PhaseCombat:
		l.turn++
		l.startRecruit()
	}

	return true
}

// State returns the current lobby state.
func (l *Lobby) State() State { return l.state }

// PlayerCount returns the number of players in the lobby.
func (l *Lobby) PlayerCount() int { return len(l.players) }

// Players returns all players in the lobby.
func (l *Lobby) Players() []*game.Player { return l.players }

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
func (l *Lobby) Turn() int { return l.turn }

// Phase returns the current game phase.
func (l *Lobby) Phase() game.Phase { return l.phase }

// PhaseEndsAt returns when the current phase ends.
func (l *Lobby) PhaseEndsAt() time.Time { return l.phaseEndsAt }

// CombatResults returns combat results for the given player (last 3).
func (l *Lobby) CombatResults(playerID string) []game.CombatResult { return l.combatResults[playerID] }

// AllCombatResults returns combat results for all players.
func (l *Lobby) AllCombatResults() map[string][]game.CombatResult { return l.combatResults }

// CombatLogs returns pending combat logs and clears them.
func (l *Lobby) CombatLogs() []game.CombatLog {
	logs := l.combatLogs
	l.combatLogs = nil
	return logs
}

// CombatPairing returns the combat pairing for the given player (combat phase only).
func (l *Lobby) CombatPairing(playerID string) (CombatPairing, bool) {
	p, ok := l.combatPairings[playerID]
	return p, ok
}

// GameResult returns the game result, or nil if the game hasn't finished.
func (l *Lobby) GameResult() *game.GameResult { return l.gameResult }

// Pool returns the card pool.
func (l *Lobby) Pool() *game.CardPool { return l.pool }
