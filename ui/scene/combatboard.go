package scene

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
	"image/color"
	"log/slog"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/effect"
	"github.com/ysomad/gigabg/ui/widget"
)

// Animation timing (seconds).
const (
	attackMoveDuration = 0.55 // attacker moves to target
	attackBackDuration = 0.45 // attacker returns to slot
	damageFlashTime    = 0.20 // initial white flash on hit
	damageShakeTime    = 0.50 // shake duration after hit
	shakeFreq          = 35.0 // shake frequency multiplier
	hitFlashTime       = 0.80 // hit damage indicator on minion body
	poisonEffectTime   = 0.60 // green drip/splash on poisoned target before death
	deathFadeDuration  = 0.50 // opacity fade on death
	deathParticleTime  = 0.80 // particle burst duration
	deathParticleCount = 12   // number of particles on death
	shieldBreakTime    = 0.50 // divine shield shard burst
	rebornDelayTime    = 0.30 // pause after death before reborn glow starts
	spawnGlowTime      = 0.70 // blue glow pillar on reborn
	spawnFadeDuration  = 0.50 // opacity fade-in on reborn (starts midway through glow)
	spawnFadeStart     = 0.4  // glow progress fraction when fade begins
	eventPause         = 0.30 // pause between attacks
)

// animPhase tracks the two-phase attacker movement.
type animPhase uint8

const (
	animPhaseForward animPhase = iota + 1
	animPhaseBack
)

// pendingReborn holds data for a reborn animation that triggers after the
// dying minion is fully removed from the board.
type pendingReborn struct {
	card     api.Card
	isPlayer bool
	boardIdx int     // board index where the minion should appear
	delay    float64 // countdown before glow starts (rebornDelayTime)
	active   bool    // true once delay has elapsed and minion is inserted
}

// animMinion holds per-minion animation state during combat replay.
type animMinion struct {
	card    api.Card
	offsetX float64 // position offset from base slot (attack animation)
	offsetY float64

	opacity      float64 // 1.0 = visible, fades to 0 on death
	dying        bool
	deathReason  game.DeathReason
	pendingDeath bool // will die after poison effect completes
	spawning     bool

	effects effect.List
}

type attackAnimation struct {
	srcCombatID game.CombatID
	srcIdx      int
	srcIsPlayer bool
	dstIdx      int
	dstIsPlayer bool
	startX      float64
	startY      float64
	targetX     float64
	targetY     float64
	progress    float64
	phase       animPhase
	impactDone  bool // true after processEvents called on impact
}

// combatBoard replays combat events as visual animations using GameLayout zones.
// Opponent board renders in the Shop zone, player board in the Board zone.
//
// Animation sequence per attack:
//  1. Attacker moves forward to target (eased)
//  2. Impact: process events (damage, shield breaks, keyword removals)
//  3. Damage numbers float up, shakes play
//  4. Poison effect overlay on poisoned targets
//  5. Death animations (fade + particles)
//  6. Reborn: delay -> glow pillar -> minion fades in at same position
//  7. Attacker returns to slot (waits for reborn to finish)
type combatBoard struct {
	cr       *widget.CardRenderer
	font     *text.GoTextFace
	boldFont *text.GoTextFace
	player   game.PlayerID
	turn     int

	playerBoard   []animMinion
	opponentBoard []animMinion

	events     []api.CombatEvent
	eventIndex int
	pauseTimer float64

	attackAnim *attackAnimation
	done       bool

	// Panel-level effects (death particles, independent of minion lifecycle).
	overlays effect.List

	// Pending reborn animations waiting for death fade to complete.
	pendingReborns []pendingReborn

	// Set by markDying so rebornMinion can record the position.
	lastDeathIdx      int
	lastDeathIsPlayer bool
}

func newCombatBoard(
	turn int,
	player game.PlayerID,
	playerBoard, opponentBoard []api.Card,
	events []api.CombatEvent,
	c *catalog.Catalog,
	font, boldFont *text.GoTextFace,
) *combatBoard {
	slog.Debug("combat panel created",
		"player_board", len(playerBoard),
		"opponent_board", len(opponentBoard),
		"events", len(events))
	return &combatBoard{
		cr:            &widget.CardRenderer{Cards: c, Font: font, BoldFont: boldFont},
		font:          font,
		boldFont:      boldFont,
		player:        player,
		turn:          turn,
		playerBoard:   buildAnimBoard(playerBoard),
		opponentBoard: buildAnimBoard(opponentBoard),
		events:        events,
	}
}

func buildAnimBoard(cards []api.Card) []animMinion {
	board := make([]animMinion, len(cards))
	for i, c := range cards {
		board[i] = animMinion{card: c, opacity: 1.0}
	}
	return board
}

// Update advances animation state. Returns true when all animations are done.
func (cp *combatBoard) Update(elapsed float64, res ui.Resolution, lay ui.GameLayout) (bool, error) {
	cp.cr.Res = res
	cp.cr.Tick++

	if cp.done {
		return true, nil
	}

	// Tick all minion effects.
	cp.updateMinionEffects(elapsed)

	// Remove dead minions (opacity reached 0).
	cp.playerBoard = cp.removeDeadMinions(cp.playerBoard, lay, true)
	cp.opponentBoard = cp.removeDeadMinions(cp.opponentBoard, lay, false)

	// Tick pending reborns.
	cp.updatePendingReborns(elapsed)

	// Tick panel-level overlays (particles).
	cp.overlays.Update(elapsed)

	// Advance current attack animation.
	if cp.attackAnim != nil {
		if err := cp.updateAttackAnim(elapsed, lay); err != nil {
			return false, fmt.Errorf("update attack animation: %w", err)
		}
		return false, nil
	}

	if cp.pauseTimer > 0 {
		cp.pauseTimer -= elapsed
		return false, nil
	}

	// Wait for active effects to finish before starting next attack.
	if cp.hasActiveEffects() {
		return false, nil
	}

	// Scan for the next AttackEvent to start a new animation.
	for cp.eventIndex < len(cp.events) {
		ev := cp.events[cp.eventIndex]
		cp.eventIndex++

		if ev.Type == game.CombatEventAttack {
			var e game.AttackEvent
			if err := json.Unmarshal(ev.Payload, &e); err != nil {
				return false, fmt.Errorf("unmarshal attack event: %w", err)
			}
			cp.startAttack(e, lay)
			return false, nil
		}
	}

	// All events consumed; wait for remaining visuals to finish.
	if !cp.hasActiveEffects() && !cp.hasDying() && !cp.hasSpawning() &&
		!cp.overlays.HasAny() && len(cp.pendingReborns) == 0 {
		cp.done = true
		return true, nil
	}

	return false, nil
}

// updateMinionEffects ticks effect lists on all minions.
func (cp *combatBoard) updateMinionEffects(elapsed float64) {
	for i := range cp.playerBoard {
		cp.playerBoard[i].effects.Update(elapsed)
	}
	for i := range cp.opponentBoard {
		cp.opponentBoard[i].effects.Update(elapsed)
	}
}

// removeDeadMinions removes minions whose opacity reached 0 and spawns
// death particles at their position.
func (cp *combatBoard) removeDeadMinions(board []animMinion, lay ui.GameLayout, isPlayer bool) []animMinion {
	n := 0
	for i := range board {
		if board[i].dying && board[i].opacity <= 0 {
			cp.spawnDeathParticles(board[i], lay, i, len(board), isPlayer)
			continue
		}
		board[n] = board[i]
		n++
	}
	return board[:n]
}

// hasActiveEffects returns true if any minion has an active visual effect
// or there are pending reborns waiting.
func (cp *combatBoard) hasActiveEffects() bool {
	if len(cp.pendingReborns) > 0 {
		return true
	}
	check := func(board []animMinion) bool {
		for _, m := range board {
			if m.effects.HasAny() || m.pendingDeath || m.spawning {
				return true
			}
		}
		return false
	}
	return check(cp.playerBoard) || check(cp.opponentBoard)
}

func (cp *combatBoard) startAttack(ev game.AttackEvent, lay ui.GameLayout) {
	srcIdx, srcIsPlayer := cp.findMinion(ev.Source)
	dstIdx, dstIsPlayer := cp.findMinion(ev.Target)
	if srcIdx < 0 || dstIdx < 0 {
		return
	}

	srcX, srcY := cp.minionPos(lay, srcIdx, srcIsPlayer)
	dstX, dstY := cp.minionPos(lay, dstIdx, dstIsPlayer)

	cp.attackAnim = &attackAnimation{
		srcCombatID: ev.Source,
		srcIdx:      srcIdx,
		srcIsPlayer: srcIsPlayer,
		dstIdx:      dstIdx,
		dstIsPlayer: dstIsPlayer,
		startX:      srcX,
		startY:      srcY,
		targetX:     dstX,
		targetY:     dstY,
		phase:       animPhaseForward,
	}
}

func (cp *combatBoard) updateAttackAnim(elapsed float64, lay ui.GameLayout) error {
	a := cp.attackAnim

	switch a.phase {
	case animPhaseForward:
		a.progress += elapsed / attackMoveDuration
		if a.progress >= 1.0 {
			a.progress = 1.0

			if !a.impactDone {
				a.impactDone = true
				if err := cp.processEvents(lay); err != nil {
					return err
				}
			}

			// Wait for impact effects and reborns to settle before returning.
			if cp.hasImpactEffects() {
				return nil
			}

			srcIdx, srcIsPlayer := cp.findMinion(a.srcCombatID)
			if srcIdx < 0 || cp.boardFor(srcIsPlayer)[srcIdx].dying {
				cp.attackAnim = nil
				cp.pauseTimer = eventPause
				return nil
			}

			a.srcIdx = srcIdx
			a.srcIsPlayer = srcIsPlayer
			a.startX, a.startY = cp.minionPos(lay, srcIdx, srcIsPlayer)
			a.progress = 0
			a.phase = animPhaseBack
		}
	case animPhaseBack:
		a.progress += elapsed / attackBackDuration
		if a.progress >= 1.0 {
			board := cp.boardFor(a.srcIsPlayer)
			if a.srcIdx < len(board) {
				board[a.srcIdx].offsetX = 0
				board[a.srcIdx].offsetY = 0
			}
			cp.attackAnim = nil
			cp.pauseTimer = eventPause
			return nil
		}
	}

	board := cp.boardFor(a.srcIsPlayer)
	if a.srcIdx >= len(board) {
		cp.attackAnim = nil
		return nil
	}

	switch a.phase {
	case animPhaseForward:
		t := ui.EaseOut(a.progress)
		board[a.srcIdx].offsetX = (a.targetX - a.startX) * t
		board[a.srcIdx].offsetY = (a.targetY - a.startY) * t
	case animPhaseBack:
		t := 1.0 - ui.EaseOut(a.progress)
		board[a.srcIdx].offsetX = (a.targetX - a.startX) * t
		board[a.srcIdx].offsetY = (a.targetY - a.startY) * t
	}
	return nil
}

// hasImpactEffects returns true if there are still impact-phase effects playing
// that should finish before the attacker returns.
func (cp *combatBoard) hasImpactEffects() bool {
	if len(cp.pendingReborns) > 0 {
		return true
	}
	check := func(board []animMinion) bool {
		for _, m := range board {
			if m.effects.Has(effect.KindFlash) || m.effects.Has(effect.KindShieldBreak) ||
				m.effects.Has(effect.KindPoisonDrip) || m.effects.Has(effect.KindSpawnGlow) ||
				m.pendingDeath || m.dying || m.spawning {
				return true
			}
		}
		return false
	}
	return check(cp.playerBoard) || check(cp.opponentBoard)
}

func unmarshalApply[T any](payload jsontext.Value, apply func(T) error) error {
	var e T
	if err := json.Unmarshal(payload, &e); err != nil {
		return err
	}
	return apply(e)
}

func (cp *combatBoard) processEvents(lay ui.GameLayout) error {
	for cp.eventIndex < len(cp.events) {
		ev := cp.events[cp.eventIndex]
		var err error
		switch ev.Type {
		case game.CombatEventDamage:
			err = unmarshalApply(ev.Payload, cp.applyDamage)
		case game.CombatEventRemoveKeyword:
			err = unmarshalApply(ev.Payload, cp.removeKeyword)
		case game.CombatEventDeath:
			err = unmarshalApply(ev.Payload, cp.markDying)
		case game.CombatEventReborn:
			err = unmarshalApply(ev.Payload, cp.queueReborn)
		default:
			return nil
		}
		if err != nil {
			return fmt.Errorf("process combat event type %d: %w", ev.Type, err)
		}
		cp.eventIndex++
	}
	return nil
}

func (cp *combatBoard) applyDamage(ev game.DamageEvent) error {
	idx, isPlayer := cp.findMinion(ev.Target)
	if idx < 0 {
		return fmt.Errorf("minion %d not found", ev.Target)
	}
	board := cp.boardFor(isPlayer)
	m := &board[idx]
	m.card.Health -= ev.Amount

	intensity := math.Min(float64(ev.Amount)*1.5, 8.0)
	m.effects.Add(effect.NewFlash(damageFlashTime))
	m.effects.Add(effect.NewShake(damageShakeTime, intensity, shakeFreq))
	m.effects.Add(effect.NewHitDamage(hitFlashTime, ev.Amount, cp.boldFont))
	return nil
}

func (cp *combatBoard) removeKeyword(ev game.RemoveKeywordEvent) error {
	idx, isPlayer := cp.findMinion(ev.Target)
	if idx < 0 {
		return fmt.Errorf("minion %d not found", ev.Target)
	}
	board := cp.boardFor(isPlayer)
	board[idx].card.Keywords.Remove(ev.Keyword)

	if ev.Keyword == game.KeywordDivineShield {
		board[idx].effects.Add(effect.NewShieldBreak(shieldBreakTime))
	}
	return nil
}

func (cp *combatBoard) markDying(ev game.DeathEvent) error {
	idx, isPlayer := cp.findMinion(ev.Target)
	if idx < 0 {
		return fmt.Errorf("minion %d not found", ev.Target)
	}

	cp.lastDeathIdx = idx
	cp.lastDeathIsPlayer = isPlayer

	board := cp.boardFor(isPlayer)
	m := &board[idx]
	m.deathReason = ev.DeathReason

	if ev.DeathReason == game.DeathReasonPoison {
		m.pendingDeath = true
		m.effects.Add(effect.NewPoisonDrip(poisonEffectTime, uint8(255*m.opacity), func() {
			m.pendingDeath = false
			cp.startDying(m)
		}))
		return nil
	}

	cp.startDying(m)
	return nil
}

// startDying begins the death fade and tint effects on a minion.
func (cp *combatBoard) startDying(m *animMinion) {
	m.dying = true
	m.effects.Add(effect.NewDeathFade(deathFadeDuration, &m.opacity, nil))
	m.effects.Add(effect.NewDeathTint(&m.opacity, m.deathReason))
}

// queueReborn records a pending reborn. The actual minion insertion happens
// after the dying minion has fully faded out and been removed from the board,
// so the death animation plays completely before the reborn glow appears.
func (cp *combatBoard) queueReborn(ev game.RebornEvent) error {
	t := cp.cr.Cards.ByTemplateID(ev.Template)
	if t == nil {
		return fmt.Errorf("unknown template %q", ev.Template)
	}

	kw := t.Keywords()
	kw.Remove(game.KeywordReborn)

	card := api.Card{
		Template: ev.Template,
		Attack:   t.Attack(),
		Health:   1,
		Tribes:   t.Tribes(),
		Keywords: kw,
		CombatID: ev.Target,
	}

	cp.pendingReborns = append(cp.pendingReborns, pendingReborn{
		card:     card,
		isPlayer: cp.lastDeathIsPlayer,
		boardIdx: cp.lastDeathIdx,
		delay:    rebornDelayTime,
	})
	return nil
}

// updatePendingReborns ticks pending reborn delays. Once a reborn's owning
// dying minion has been removed and the delay elapses, the new minion is
// inserted into the board with a glow + fade-in animation.
func (cp *combatBoard) updatePendingReborns(elapsed float64) {
	n := 0
	for i := range cp.pendingReborns {
		pr := &cp.pendingReborns[i]

		if pr.active {
			continue
		}

		if cp.hasDyingOnBoard(pr.isPlayer) {
			cp.pendingReborns[n] = cp.pendingReborns[i]
			n++
			continue
		}

		pr.delay -= elapsed
		if pr.delay > 0 {
			cp.pendingReborns[n] = cp.pendingReborns[i]
			n++
			continue
		}

		cp.insertRebornMinion(*pr)
	}
	cp.pendingReborns = cp.pendingReborns[:n]
}

func (cp *combatBoard) hasDyingOnBoard(isPlayer bool) bool {
	board := cp.boardFor(isPlayer)
	for _, m := range board {
		if m.dying {
			return true
		}
	}
	return false
}

func (cp *combatBoard) insertRebornMinion(pr pendingReborn) {
	m := animMinion{
		card:     pr.card,
		opacity:  0,
		spawning: true,
	}

	// Clamp index to current board length.
	board := cp.boardFor(pr.isPlayer)
	idx := pr.boardIdx
	if idx > len(board) {
		idx = len(board)
	}

	// Insert first, then attach effect with pointers to the actual slice
	// element. Creating the effect before insert would point at the local
	// variable which is dead after the copy into the slice.
	if pr.isPlayer {
		cp.playerBoard = slices.Insert(cp.playerBoard, idx, m)
		actual := &cp.playerBoard[idx]
		actual.effects.Add(effect.NewSpawnGlow(
			spawnGlowTime, spawnFadeDuration, spawnFadeStart,
			&actual.opacity, &actual.spawning, &cp.cr.Tick,
		))
	} else {
		cp.opponentBoard = slices.Insert(cp.opponentBoard, idx, m)
		actual := &cp.opponentBoard[idx]
		actual.effects.Add(effect.NewSpawnGlow(
			spawnGlowTime, spawnFadeDuration, spawnFadeStart,
			&actual.opacity, &actual.spawning, &cp.cr.Tick,
		))
	}
}

func (cp *combatBoard) spawnDeathParticles(m animMinion, lay ui.GameLayout, idx, count int, isPlayer bool) {
	zone := lay.Shop
	if isPlayer {
		zone = lay.Board
	}
	r := ui.CardRect(zone, idx, count, lay.CardW, lay.CardH, lay.Gap)
	cx := r.X + r.W/2
	cy := r.Y + r.H/2

	baseClr := color.RGBA{80, 80, 120, 255}
	if m.deathReason == game.DeathReasonPoison {
		baseClr = color.RGBA{30, 160, 50, 255}
	}

	cp.overlays.Add(effect.NewParticles(cx, cy, deathParticleCount, baseClr, deathParticleTime))
}

// --- Query helpers ---

func (cp *combatBoard) hasSpawning() bool {
	for _, m := range cp.playerBoard {
		if m.spawning {
			return true
		}
	}
	for _, m := range cp.opponentBoard {
		if m.spawning {
			return true
		}
	}
	return false
}

func (cp *combatBoard) hasDying() bool {
	for _, m := range cp.playerBoard {
		if m.dying {
			return true
		}
	}
	for _, m := range cp.opponentBoard {
		if m.dying {
			return true
		}
	}
	return false
}

func (cp *combatBoard) findMinion(combatID game.CombatID) (idx int, isPlayer bool) {
	for i, m := range cp.playerBoard {
		if m.card.CombatID == combatID {
			return i, true
		}
	}
	for i, m := range cp.opponentBoard {
		if m.card.CombatID == combatID {
			return i, false
		}
	}
	return -1, false
}

func (cp *combatBoard) boardFor(isPlayer bool) []animMinion {
	if isPlayer {
		return cp.playerBoard
	}
	return cp.opponentBoard
}

func (cp *combatBoard) minionPos(lay ui.GameLayout, idx int, isPlayer bool) (float64, float64) {
	board := cp.boardFor(isPlayer)
	zone := lay.Shop
	if isPlayer {
		zone = lay.Board
	}
	r := ui.CardRect(zone, idx, len(board), lay.CardW, lay.CardH, lay.Gap)
	return r.X, r.Y
}

// --- Drawing ---

func (cp *combatBoard) drawOpponentBoard(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout) {
	cp.drawBoard(screen, res, lay.Shop, cp.opponentBoard, lay.CardW, lay.CardH, lay.Gap)
}

func (cp *combatBoard) drawPlayerBoard(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout) {
	cp.drawBoard(screen, res, lay.Board, cp.playerBoard, lay.CardW, lay.CardH, lay.Gap)
}

func (cp *combatBoard) drawBoard(screen *ebiten.Image, res ui.Resolution, zone ui.Rect, board []animMinion, cardW, cardH, gap float64) {
	for i, m := range board {
		r := ui.CardRect(zone, i, len(board), cardW, cardH, gap)
		r.X += m.offsetX
		r.Y += m.offsetY

		// Let effects modify draw params.
		alpha := uint8(255 * m.opacity)
		flashPct := 0.0
		m.effects.Modify(&r, &alpha, &flashPct)

		// Draw behind-card effects (shield break, spawn glow).
		m.effects.DrawBehind(screen, res, r)

		// Draw the minion card.
		cp.cr.DrawMinion(screen, m.card, r, alpha, flashPct)

		// Draw front-of-card effects (hit damage, poison drip, death tint).
		m.effects.DrawFront(screen, res, r)
	}
}

// drawOverlays draws panel-level effects (death particles) on top of all boards.
// Must be called once per frame, after both boards are drawn.
func (cp *combatBoard) drawOverlays(screen *ebiten.Image, res ui.Resolution) {
	cp.overlays.DrawFront(screen, res, ui.Rect{})
}
