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
	attackMoveDuration    = 0.55 // attacker moves to target
	attackBackDuration    = 0.45 // attacker returns to slot
	damageFlashTime       = 0.20 // initial white flash on hit
	damageShakeTime       = 0.50 // shake duration after hit
	shakeFreq             = 35.0 // shake frequency multiplier
	hitFlashTime          = 0.80 // hit damage indicator on minion body
	poisonEffectTime      = 0.60 // green drip/splash on poisoned target before death
	deathFadeDuration     = 0.50 // opacity fade on death
	deathParticleTime     = 0.80 // particle burst duration
	deathParticleCount    = 12   // number of particles on death
	divineShieldBreakTime = 0.50 // divine shield shard burst
	venomBreakTime        = 0.50 // venomous vial shatter
	rebornDelayTime       = 0.30 // pause after death before reborn glow starts
	spawnGlowTime         = 0.70 // blue glow pillar on reborn
	spawnFadeStart        = 0.4  // glow progress fraction when opacity fade-in begins
	eventPause            = 0.30 // pause between attacks
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
// Opacity and spawning state are derived from effect progress — effects
// never mutate minion fields directly.
type animMinion struct {
	card    api.Card
	offsetX float64 // position offset from base slot (attack animation)
	offsetY float64

	dying        bool
	deathReason  game.DeathReason
	pendingDeath bool // will die after poison effect completes
	spawning     bool

	effects effect.List
}

// minionOpacity computes the minion's current opacity from effect state.
// Dying minions fade out based on DeathFade progress (1→0).
// Spawning minions fade in based on SpawnGlow progress (0→1).
// Normal minions are fully opaque.
func minionOpacity(m *animMinion) float64 {
	if p := m.effects.Progress(effect.KindDeathFade); p >= 0 {
		return 1.0 - p
	}
	if p := m.effects.Progress(effect.KindSpawnGlow); p >= 0 {
		sg := m.effects.Progress(effect.KindSpawnGlow)
		if sg < spawnFadeStart {
			return 0
		}
		fp := (sg - spawnFadeStart) / (1.0 - spawnFadeStart)
		if fp > 1.0 {
			return 1.0
		}
		return fp
	}
	return 1.0
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

	playerBoard   []*animMinion
	opponentBoard []*animMinion

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

func buildAnimBoard(cards []api.Card) []*animMinion {
	board := make([]*animMinion, len(cards))
	for i, c := range cards {
		board[i] = &animMinion{card: c}
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

	// Tick all minion effects and derive state transitions.
	cp.updateMinionEffects(elapsed)

	// Remove dead minions (DeathFade completed).
	cp.playerBoard = removeDeadMinions(cp.playerBoard, cp, lay, true)
	cp.opponentBoard = removeDeadMinions(cp.opponentBoard, cp, lay, false)

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

// updateMinionEffects ticks effect lists on all minions and derives state
// transitions from effect completion:
//   - PoisonDrip finished → start death (add DeathFade + DeathTint)
//   - SpawnGlow finished → clear spawning flag
func (cp *combatBoard) updateMinionEffects(elapsed float64) {
	for _, m := range cp.playerBoard {
		cp.tickMinion(m, elapsed)
	}
	for _, m := range cp.opponentBoard {
		cp.tickMinion(m, elapsed)
	}
}

func (cp *combatBoard) tickMinion(m *animMinion, elapsed float64) {
	hadPoisonDrip := m.effects.Has(effect.KindPoisonDrip)
	hadSpawnGlow := m.effects.Has(effect.KindSpawnGlow)

	m.effects.Update(elapsed)

	// PoisonDrip just finished → start death sequence.
	if hadPoisonDrip && !m.effects.Has(effect.KindPoisonDrip) {
		m.pendingDeath = false
		m.dying = true
		m.effects.Add(effect.NewDeathFade(deathFadeDuration))
		m.effects.Add(effect.NewDeathTint(deathFadeDuration, m.deathReason))
	}

	// SpawnGlow just finished → minion is fully visible.
	if hadSpawnGlow && !m.effects.Has(effect.KindSpawnGlow) {
		m.spawning = false
	}
}

// removeDeadMinions removes minions whose DeathFade has completed and spawns
// death particles at their position.
func removeDeadMinions(board []*animMinion, cp *combatBoard, lay ui.GameLayout, isPlayer bool) []*animMinion {
	n := 0
	for i, m := range board {
		if m.dying && !m.effects.Has(effect.KindDeathFade) {
			cp.spawnDeathParticles(m, lay, i, len(board), isPlayer)
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
	check := func(board []*animMinion) bool {
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
	check := func(board []*animMinion) bool {
		for _, m := range board {
			if m.effects.Has(effect.KindFlash) || m.effects.Has(effect.KindDivineShieldBreak) ||
				m.effects.Has(effect.KindVenomBreak) || m.effects.Has(effect.KindPoisonDrip) ||
				m.effects.Has(effect.KindSpawnGlow) || m.pendingDeath || m.dying || m.spawning {
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
	m := cp.boardFor(isPlayer)[idx]
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
	m := cp.boardFor(isPlayer)[idx]
	m.card.Keywords.Remove(ev.Keyword)

	switch ev.Keyword {
	case game.KeywordDivineShield:
		m.effects.Add(effect.NewDivineShieldBreak(divineShieldBreakTime))
	case game.KeywordVenomous:
		m.effects.Add(effect.NewVenomBreak(venomBreakTime))
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

	m := cp.boardFor(isPlayer)[idx]
	m.deathReason = ev.DeathReason

	// Poison deaths show a drip effect first; tickMinion starts the death
	// sequence when PoisonDrip completes.
	if ev.DeathReason == game.DeathReasonPoison {
		m.pendingDeath = true
		m.effects.Add(effect.NewPoisonDrip(poisonEffectTime, 255))
		return nil
	}

	// Immediate death: start fade + tint.
	m.dying = true
	m.effects.Add(effect.NewDeathFade(deathFadeDuration))
	m.effects.Add(effect.NewDeathTint(deathFadeDuration, m.deathReason))
	return nil
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
	for _, m := range cp.boardFor(isPlayer) {
		if m.dying {
			return true
		}
	}
	return false
}

func (cp *combatBoard) insertRebornMinion(pr pendingReborn) {
	m := &animMinion{
		card:     pr.card,
		spawning: true,
	}

	m.effects.Add(effect.NewSpawnGlow(spawnGlowTime, spawnFadeStart))

	// Clamp index to current board length.
	board := cp.boardFor(pr.isPlayer)
	idx := pr.boardIdx
	if idx > len(board) {
		idx = len(board)
	}

	if pr.isPlayer {
		cp.playerBoard = slices.Insert(cp.playerBoard, idx, m)
	} else {
		cp.opponentBoard = slices.Insert(cp.opponentBoard, idx, m)
	}
}

func (cp *combatBoard) spawnDeathParticles(m *animMinion, lay ui.GameLayout, idx, count int, isPlayer bool) {
	zone := lay.Shop
	if isPlayer {
		zone = lay.Board
	}
	r := ui.CardRect(zone, idx, count, lay.CardW, lay.CardH, lay.Gap)
	r.X += m.offsetX
	r.Y += m.offsetY
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

func (cp *combatBoard) boardFor(isPlayer bool) []*animMinion {
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

// draw renders both boards and overlays. The attacking minion is always
// drawn last so it appears on top of all other minions. Impact effects
// on the defender (shield break, venom break) are drawn on top of the
// attacker so they remain visible.
func (cp *combatBoard) draw(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout) {
	// Determine which minion (if any) is the active attacker.
	attackerIdx := -1
	attackerIsPlayer := false
	if cp.attackAnim != nil {
		attackerIdx = cp.attackAnim.srcIdx
		attackerIsPlayer = cp.attackAnim.srcIsPlayer
	}

	// Draw both boards, skipping the attacker.
	playerSkip := -1
	opponentSkip := -1
	if attackerIdx >= 0 {
		if attackerIsPlayer {
			playerSkip = attackerIdx
		} else {
			opponentSkip = attackerIdx
		}
	}

	cp.drawBoard(screen, res, lay, lay.Shop, cp.opponentBoard, opponentSkip)
	cp.drawBoard(screen, res, lay, lay.Board, cp.playerBoard, playerSkip)

	// Draw the attacker on top of both boards.
	if attackerIdx >= 0 {
		board := cp.boardFor(attackerIsPlayer)
		zone := lay.Shop
		if attackerIsPlayer {
			zone = lay.Board
		}
		if attackerIdx < len(board) {
			cp.drawMinion(screen, res, lay, zone, board, attackerIdx)
		}
	}

	// Draw the defender's front effects on top of the attacker so impact
	// visuals (divine shield break, venom break) aren't hidden.
	if cp.attackAnim != nil {
		a := cp.attackAnim
		dstBoard := cp.boardFor(a.dstIsPlayer)
		if a.dstIdx < len(dstBoard) {
			dm := dstBoard[a.dstIdx]
			dstZone := lay.Shop
			if a.dstIsPlayer {
				dstZone = lay.Board
			}
			r := ui.CardRect(dstZone, a.dstIdx, len(dstBoard), lay.CardW, lay.CardH, lay.Gap)
			r.X += dm.offsetX
			r.Y += dm.offsetY
			var a8 uint8
			var fp float64
			dm.effects.Modify(&r, &a8, &fp)
			dm.effects.DrawFront(screen, res, r)
		}
	}

	cp.overlays.DrawFront(screen, res, ui.Rect{})
}

func (cp *combatBoard) drawBoard(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout, zone ui.Rect, board []*animMinion, skipIdx int) {
	for i := range board {
		if i == skipIdx {
			continue
		}
		cp.drawMinion(screen, res, lay, zone, board, i)
	}
}

func (cp *combatBoard) drawMinion(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout, zone ui.Rect, board []*animMinion, idx int) {
	m := board[idx]
	r := ui.CardRect(zone, idx, len(board), lay.CardW, lay.CardH, lay.Gap)
	r.X += m.offsetX
	r.Y += m.offsetY

	// Compute opacity from effect state.
	opacity := minionOpacity(m)
	alpha := uint8(255 * opacity)
	flashPct := 0.0

	// Let effects modify draw params (shake, flash).
	m.effects.Modify(&r, &alpha, &flashPct)

	// Draw behind-card effects (spawn glow).
	m.effects.DrawBehind(screen, res, r)

	// Draw the minion card.
	cp.cr.DrawMinion(screen, m.card, r, alpha, flashPct)

	// Draw front-of-card effects (hit damage, poison drip, death tint, shield break).
	m.effects.DrawFront(screen, res, r)
}
