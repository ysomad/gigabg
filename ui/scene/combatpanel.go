package scene

import (
	"image/color"
	"log/slog"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// Animation timing (seconds).
const (
	attackMoveDuration  = 0.8
	attackBackDuration  = 0.8
	damageIndicatorTime = 1.5
	poisonIndicatorTime = 0.8
	deathFadeDuration   = 0.4
	eventPause          = 0.5
)

type animPhase uint8

const (
	animPhaseForward animPhase = iota + 1
	animPhaseBack
)

type animMinion struct {
	card        api.Card
	offsetX     float64
	offsetY     float64
	flash       float64 // remaining damage indicator time
	dmgText     string
	opacity     float64 // 1.0 = visible, fades to 0 on death
	dying       bool
	poisonDeath bool // died from poison/venom; show indicator after damage fades
}

type attackAnimation struct {
	srcCombatID int
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
}

// combatPanel replays combat events as visual animations using GameLayout zones.
// Opponent board renders in the Shop zone, player board in the Board zone.
type combatPanel struct {
	cr       *widget.CardRenderer
	font     *text.GoTextFace
	boldFont *text.GoTextFace
	turn     int

	playerBoard   []animMinion
	opponentBoard []animMinion

	events     []api.CombatEvent
	eventIndex int
	pauseTimer float64

	attackAnim *attackAnimation
	done       bool
}

func newCombatPanel(
	turn int,
	playerBoard, opponentBoard []api.Card,
	events []api.CombatEvent,
	c *catalog.Catalog,
	font, boldFont *text.GoTextFace,
) *combatPanel {
	slog.Info("combat panel created",
		"player_board", len(playerBoard),
		"opponent_board", len(opponentBoard),
		"events", len(events))
	return &combatPanel{
		cr:            &widget.CardRenderer{Cards: c, Font: font, BoldFont: boldFont},
		font:          font,
		boldFont:      boldFont,
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
func (cp *combatPanel) Update(dt float64, lay ui.GameLayout) bool {
	cp.cr.Tick++

	if cp.done {
		return true
	}

	cp.updateDeathFades(dt)

	for i := range cp.playerBoard {
		if cp.playerBoard[i].flash > 0 {
			cp.playerBoard[i].flash -= dt
		}
	}
	for i := range cp.opponentBoard {
		if cp.opponentBoard[i].flash > 0 {
			cp.opponentBoard[i].flash -= dt
		}
	}

	cp.startPoisonIndicators()

	if cp.attackAnim != nil {
		cp.updateAttackAnim(dt, lay)
		return false
	}

	if cp.pauseTimer > 0 {
		cp.pauseTimer -= dt
		return false
	}

	for cp.eventIndex < len(cp.events) {
		ev := cp.events[cp.eventIndex]
		cp.eventIndex++
		if ev.Type == game.CombatEventAttack {
			cp.startAttack(ev, lay)
			return false
		}
	}

	if !cp.hasActiveIndicator() && !cp.hasDying() {
		cp.done = true
		return true
	}
	return false
}

func (cp *combatPanel) startAttack(ev api.CombatEvent, lay ui.GameLayout) {
	srcIdx, srcIsPlayer := cp.findMinion(ev.SourceID)
	dstIdx, dstIsPlayer := cp.findMinion(ev.TargetID)
	if srcIdx < 0 || dstIdx < 0 {
		return
	}

	srcX, srcY := cp.minionPos(lay, srcIdx, srcIsPlayer)
	dstX, dstY := cp.minionPos(lay, dstIdx, dstIsPlayer)

	cp.attackAnim = &attackAnimation{
		srcCombatID: ev.SourceID,
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

func (cp *combatPanel) updateAttackAnim(dt float64, lay ui.GameLayout) {
	a := cp.attackAnim

	switch a.phase {
	case animPhaseForward:
		a.progress += dt / attackMoveDuration
		if a.progress >= 1.0 {
			cp.consumeHitEvents()

			srcIdx, srcIsPlayer := cp.findMinion(a.srcCombatID)
			if srcIdx < 0 || cp.boardFor(srcIsPlayer)[srcIdx].dying {
				cp.attackAnim = nil
				cp.pauseTimer = eventPause
				return
			}

			a.srcIdx = srcIdx
			a.srcIsPlayer = srcIsPlayer
			a.startX, a.startY = cp.minionPos(lay, srcIdx, srcIsPlayer)
			a.progress = 0
			a.phase = animPhaseBack
		}
	case animPhaseBack:
		a.progress += dt / attackBackDuration
		if a.progress >= 1.0 {
			board := cp.boardFor(a.srcIsPlayer)
			if a.srcIdx < len(board) {
				board[a.srcIdx].offsetX = 0
				board[a.srcIdx].offsetY = 0
			}
			cp.attackAnim = nil
			cp.pauseTimer = eventPause
			return
		}
	}

	board := cp.boardFor(a.srcIsPlayer)
	if a.srcIdx >= len(board) {
		cp.attackAnim = nil
		return
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
}

func (cp *combatPanel) applyDamage(ev api.CombatEvent) {
	idx, isPlayer := cp.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}
	board := cp.boardFor(isPlayer)
	board[idx].card.Health -= ev.Amount
	board[idx].flash = damageIndicatorTime
	board[idx].dmgText = "-" + strconv.Itoa(ev.Amount)
}

func (cp *combatPanel) removeKeyword(ev api.CombatEvent) {
	idx, isPlayer := cp.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}
	board := cp.boardFor(isPlayer)
	board[idx].card.Keywords.Remove(ev.Keyword)
}

func (cp *combatPanel) markDying(ev api.CombatEvent) {
	idx, isPlayer := cp.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}

	board := cp.boardFor(isPlayer)
	if ev.DeathReason == game.DeathReasonPoison {
		board[idx].poisonDeath = true
		return
	}
	board[idx].dying = true
}

// startPoisonIndicators transitions poison-killed minions from damage display
// to the skull indicator once their damage flash expires.
func (cp *combatPanel) startPoisonIndicators() {
	start := func(board []animMinion) {
		for i := range board {
			if board[i].poisonDeath && !board[i].dying && board[i].flash <= 0 {
				board[i].flash = poisonIndicatorTime
				board[i].dmgText = ""
				board[i].dying = true
			}
		}
	}
	start(cp.playerBoard)
	start(cp.opponentBoard)
}

func (cp *combatPanel) consumeHitEvents() {
	for cp.eventIndex < len(cp.events) {
		ev := cp.events[cp.eventIndex]
		switch ev.Type {
		case game.CombatEventDamage:
			cp.applyDamage(ev)
			cp.eventIndex++
		case game.CombatEventRemoveKeyword:
			cp.removeKeyword(ev)
			cp.eventIndex++
		case game.CombatEventDeath:
			cp.markDying(ev)
			cp.eventIndex++
		default:
			return
		}
	}
}

func (cp *combatPanel) updateDeathFades(dt float64) {
	fade := dt / deathFadeDuration
	cp.playerBoard = fadeAndRemove(cp.playerBoard, fade)
	cp.opponentBoard = fadeAndRemove(cp.opponentBoard, fade)
}

func fadeAndRemove(board []animMinion, fade float64) []animMinion {
	n := 0
	for i := range board {
		if board[i].dying {
			board[i].opacity -= fade
			if board[i].opacity <= 0 {
				continue
			}
		}
		board[n] = board[i]
		n++
	}
	return board[:n]
}

func (cp *combatPanel) hasDying() bool {
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

func (cp *combatPanel) hasActiveIndicator() bool {
	for _, m := range cp.playerBoard {
		if m.flash > 0 {
			return true
		}
	}
	for _, m := range cp.opponentBoard {
		if m.flash > 0 {
			return true
		}
	}
	return false
}

func (cp *combatPanel) findMinion(combatID int) (idx int, isPlayer bool) {
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

func (cp *combatPanel) boardFor(isPlayer bool) []animMinion {
	if isPlayer {
		return cp.playerBoard
	}
	return cp.opponentBoard
}

// minionPos returns the base-space position (top-left) for the given minion.
// Opponent board uses the Shop zone, player board uses the Board zone.
func (cp *combatPanel) minionPos(lay ui.GameLayout, idx int, isPlayer bool) (float64, float64) {
	board := cp.boardFor(isPlayer)
	zone := lay.Shop // opponent board in shop zone
	if isPlayer {
		zone = lay.Board // player board in board zone
	}
	r := ui.CardRect(zone, idx, len(board), lay.CardW, lay.CardH, lay.Gap)
	return r.X, r.Y
}

// startCleaveSlash begins a blade trail effect if the attacker has Cleave.

// drawOpponentBoard renders opponent minions with animations into the Shop zone.
func (cp *combatPanel) drawOpponentBoard(screen *ebiten.Image, lay ui.GameLayout) {
	cp.drawBoard(screen, lay.Shop, cp.opponentBoard, lay.CardW, lay.CardH, lay.Gap)
}

// drawPlayerBoard renders player minions with animations into the Board zone.
func (cp *combatPanel) drawPlayerBoard(screen *ebiten.Image, lay ui.GameLayout) {
	cp.drawBoard(screen, lay.Board, cp.playerBoard, lay.CardW, lay.CardH, lay.Gap)
}

func (cp *combatPanel) drawBoard(screen *ebiten.Image, zone ui.Rect, board []animMinion, cardW, cardH, gap float64) {
	for i, m := range board {
		r := ui.CardRect(zone, i, len(board), cardW, cardH, gap)
		r.X += m.offsetX
		r.Y += m.offsetY

		// Shake on damage (base-space units).
		if m.flash > 0 && !m.poisonDeath {
			t := m.flash / damageIndicatorTime
			intensity := t * 4.0
			r.X += math.Sin(m.flash*30) * intensity
			r.Y += math.Cos(m.flash*25) * intensity * 0.5
		}

		alpha := uint8(255 * m.opacity)
		flashPct := m.flash / damageIndicatorTime

		cp.cr.DrawMinion(screen, m.card, r, alpha, flashPct)

		if m.flash > 0 {
			if m.poisonDeath && m.dying {
				cp.drawPoisonSplat(screen, m, r, alpha)
			} else if m.dmgText != "" {
				cp.drawDamageSplat(screen, m, r, alpha)
			}
		}
	}
}

func (cp *combatPanel) drawDamageSplat(screen *ebiten.Image, m animMinion, r ui.Rect, alpha uint8) {
	sr := r.Screen()
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := ui.ActiveRes.Scale()
	sf := float32(s)

	duration := damageIndicatorTime
	splatBg := color.RGBA{140, 100, 0, 0} // gold
	if m.poisonDeath && m.dying {
		duration = poisonIndicatorTime
		splatBg = color.RGBA{30, 120, 40, 0} // green
	}
	t := m.flash / duration // 1.0 → 0.0

	// Pop-in: starts large, settles to 1.0.
	splatScale := float32(1.0 + 0.15*(1.0-ui.EaseOut(1.0-t)))

	// Float up over lifetime.
	cy -= float32((1.0 - t) * 12.0 * s)

	// Fade out in last 30%.
	a := alpha
	if t < 0.3 {
		a = uint8(float64(alpha) * (t / 0.3))
	}

	// Damage background circle.
	splatBg.A = a
	vector.FillCircle(screen, cx, cy, 22*sf*splatScale, splatBg, true)

	// Bold white damage text.
	op := &text.DrawOptions{}
	op.GeoM.Scale(1.8, 1.8)
	op.GeoM.Translate(float64(cx), float64(cy))
	op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, a})
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(screen, m.dmgText, cp.boldFont, op)
}

// drawPoisonSplat draws a green circle with a vector skull for poison/venom kills.
func (cp *combatPanel) drawPoisonSplat(screen *ebiten.Image, m animMinion, r ui.Rect, alpha uint8) {
	sr := r.Screen()
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := ui.ActiveRes.Scale()
	sf := float32(s)

	t := m.flash / poisonIndicatorTime // 1.0 → 0.0

	// Pop-in: starts large, settles to 1.0.
	scale := float32(1.0 + 0.15*(1.0-ui.EaseOut(1.0-t)))

	// Float up over lifetime.
	cy -= float32((1.0 - t) * 12.0 * s)

	// Fade out in last 30%.
	a := alpha
	if t < 0.3 {
		a = uint8(float64(alpha) * (t / 0.3))
	}

	// Green background circle.
	bgR := 22 * sf * scale
	bg := color.RGBA{30, 120, 40, a}
	vector.FillCircle(screen, cx, cy, bgR, bg, true)

	// Skull: cranium + jaw in bone-white, eye/nose sockets in bg color.
	bone := color.RGBA{230, 220, 200, a}

	// Cranium (upper ellipse).
	crX := bgR * 0.48
	crY := bgR * 0.52
	crCY := cy - bgR*0.08
	ui.FillEllipse(screen, cx, crCY, crX, crY, bone)

	// Jaw (smaller ellipse below cranium).
	jrX := crX * 0.68
	jrY := crY * 0.32
	ui.FillEllipse(screen, cx, crCY+crY*0.7, jrX, jrY, bone)

	// Eye sockets.
	eyeR := crX * 0.22
	eyeOff := crX * 0.38
	eyeY := crCY - crY*0.08
	vector.FillCircle(screen, cx-eyeOff, eyeY, eyeR, bg, true)
	vector.FillCircle(screen, cx+eyeOff, eyeY, eyeR, bg, true)

	// Nose (small inverted triangle).
	noseY := crCY + crY*0.22
	noseH := crY * 0.18
	noseW := crX * 0.14
	var nose vector.Path
	nose.MoveTo(cx, noseY+noseH)
	nose.LineTo(cx-noseW, noseY)
	nose.LineTo(cx+noseW, noseY)
	nose.Close()
	noseOp := &vector.DrawPathOptions{AntiAlias: true}
	noseOp.ColorScale.ScaleWithColor(bg)
	vector.FillPath(screen, &nose, nil, noseOp)

	// Teeth (vertical dark slits in jaw area).
	teethY := crCY + crY*0.42
	teethH := crY * 0.28
	toothW := sf * 1.0
	for _, off := range []float32{-0.22, 0, 0.22} {
		tx := cx + crX*off
		vector.FillRect(screen, tx-toothW*0.5, teethY, toothW, teethH, bg, false)
	}
}
