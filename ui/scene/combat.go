package scene

import (
	"fmt"
	"image/color"
	"log/slog"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// Animation timing (seconds).
const (
	attackMoveDuration  = 0.8
	attackBackDuration  = 0.8
	damageIndicatorTime = 0.8
	deathFadeDuration   = 0.4
	eventPause          = 0.5
)

type animPhase uint8

const (
	animPhaseForward animPhase = iota + 1
	animPhaseBack
)

type animMinion struct {
	card    api.Card
	offsetX float64
	offsetY float64
	flash   float64 // remaining damage indicator time
	dmgText string
	opacity float64 // 1.0 = visible, fades to 0 on death
	dying   bool
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

// CombatAnimator replays combat events as visual animations.
type CombatAnimator struct {
	cr   *widget.CardRenderer
	font *text.GoTextFace
	turn int

	playerBoard   []animMinion
	opponentBoard []animMinion

	events     []game.CombatEvent
	eventIndex int
	pauseTimer float64

	attackAnim *attackAnimation
	done       bool
}

func NewCombatAnimator(
	turn int,
	playerBoard, opponentBoard []api.Card,
	events []game.CombatEvent,
	c *cards.Cards,
	font *text.GoTextFace,
) *CombatAnimator {
	slog.Info("combat animator created",
		"player_board", len(playerBoard),
		"opponent_board", len(opponentBoard),
		"events", len(events))
	return &CombatAnimator{
		cr:            &widget.CardRenderer{Cards: c, Font: font},
		font:          font,
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
func (ca *CombatAnimator) Update(dt float64) bool {
	if ca.done {
		return true
	}

	ca.updateDeathFades(dt)

	for i := range ca.playerBoard {
		if ca.playerBoard[i].flash > 0 {
			ca.playerBoard[i].flash -= dt
		}
	}
	for i := range ca.opponentBoard {
		if ca.opponentBoard[i].flash > 0 {
			ca.opponentBoard[i].flash -= dt
		}
	}

	if ca.attackAnim != nil {
		ca.updateAttackAnim(dt)
		return false
	}

	if ca.pauseTimer > 0 {
		ca.pauseTimer -= dt
		return false
	}

	for ca.eventIndex < len(ca.events) {
		ev := ca.events[ca.eventIndex]
		ca.eventIndex++
		if ev.Type == game.CombatEventAttack {
			ca.startAttack(ev)
			return false
		}
	}

	if !ca.hasActiveIndicator() && !ca.hasDying() {
		ca.done = true
		return true
	}
	return false
}

func (ca *CombatAnimator) startAttack(ev game.CombatEvent) {
	srcIdx, srcIsPlayer := ca.findMinion(ev.SourceID)
	dstIdx, dstIsPlayer := ca.findMinion(ev.TargetID)
	if srcIdx < 0 || dstIdx < 0 {
		return
	}

	lay := ui.CalcCombatLayout()
	srcX, srcY := ca.minionPos(lay, srcIdx, srcIsPlayer)
	dstX, dstY := ca.minionPos(lay, dstIdx, dstIsPlayer)

	ca.attackAnim = &attackAnimation{
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

func (ca *CombatAnimator) updateAttackAnim(dt float64) {
	a := ca.attackAnim

	switch a.phase {
	case animPhaseForward:
		a.progress += dt / attackMoveDuration
		if a.progress >= 1.0 {
			ca.consumeHitEvents()

			srcIdx, srcIsPlayer := ca.findMinion(a.srcCombatID)
			if srcIdx < 0 || ca.boardFor(srcIsPlayer)[srcIdx].dying {
				ca.attackAnim = nil
				ca.pauseTimer = eventPause
				return
			}

			lay := ui.CalcCombatLayout()
			a.srcIdx = srcIdx
			a.srcIsPlayer = srcIsPlayer
			a.startX, a.startY = ca.minionPos(lay, srcIdx, srcIsPlayer)
			a.progress = 0
			a.phase = animPhaseBack
		}
	case animPhaseBack:
		a.progress += dt / attackBackDuration
		if a.progress >= 1.0 {
			board := ca.boardFor(a.srcIsPlayer)
			if a.srcIdx < len(board) {
				board[a.srcIdx].offsetX = 0
				board[a.srcIdx].offsetY = 0
			}
			ca.attackAnim = nil
			ca.pauseTimer = eventPause
			return
		}
	}

	board := ca.boardFor(a.srcIsPlayer)
	if a.srcIdx >= len(board) {
		ca.attackAnim = nil
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

func (ca *CombatAnimator) applyDamage(ev game.CombatEvent) {
	idx, isPlayer := ca.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}
	board := ca.boardFor(isPlayer)
	board[idx].card.Health -= ev.Amount
	board[idx].flash = damageIndicatorTime
	board[idx].dmgText = fmt.Sprintf("-%d", ev.Amount)
}

func (ca *CombatAnimator) markDying(combatID int) {
	for i, m := range ca.playerBoard {
		if m.card.CombatID == combatID {
			ca.playerBoard[i].dying = true
			return
		}
	}
	for i, m := range ca.opponentBoard {
		if m.card.CombatID == combatID {
			ca.opponentBoard[i].dying = true
			return
		}
	}
}

func (ca *CombatAnimator) consumeHitEvents() {
	for ca.eventIndex < len(ca.events) {
		ev := ca.events[ca.eventIndex]
		switch ev.Type {
		case game.CombatEventDamage:
			ca.applyDamage(ev)
			ca.eventIndex++
		case game.CombatEventDeath:
			ca.markDying(ev.TargetID)
			ca.eventIndex++
		default:
			return
		}
	}
}

func (ca *CombatAnimator) updateDeathFades(dt float64) {
	fade := dt / deathFadeDuration
	ca.playerBoard = fadeAndRemove(ca.playerBoard, fade)
	ca.opponentBoard = fadeAndRemove(ca.opponentBoard, fade)
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

func (ca *CombatAnimator) hasDying() bool {
	for _, m := range ca.playerBoard {
		if m.dying {
			return true
		}
	}
	for _, m := range ca.opponentBoard {
		if m.dying {
			return true
		}
	}
	return false
}

func (ca *CombatAnimator) hasActiveIndicator() bool {
	for _, m := range ca.playerBoard {
		if m.flash > 0 {
			return true
		}
	}
	for _, m := range ca.opponentBoard {
		if m.flash > 0 {
			return true
		}
	}
	return false
}

func (ca *CombatAnimator) findMinion(combatID int) (idx int, isPlayer bool) {
	for i, m := range ca.playerBoard {
		if m.card.CombatID == combatID {
			return i, true
		}
	}
	for i, m := range ca.opponentBoard {
		if m.card.CombatID == combatID {
			return i, false
		}
	}
	return -1, false
}

func (ca *CombatAnimator) boardFor(isPlayer bool) []animMinion {
	if isPlayer {
		return ca.playerBoard
	}
	return ca.opponentBoard
}

func (ca *CombatAnimator) minionPos(lay ui.CombatLayout, idx int, isPlayer bool) (float64, float64) {
	board := ca.boardFor(isPlayer)
	zone := lay.Opponent
	if isPlayer {
		zone = lay.Player
	}
	r := ui.CardRect(zone, idx, len(board), lay.CardW, lay.CardH, lay.Gap)
	return r.X, r.Y
}

// Draw renders the combat animation.
func (ca *CombatAnimator) Draw(screen *ebiten.Image) {
	screen.Fill(ui.ColorBackground)
	lay := ui.CalcCombatLayout()

	// Header.
	header := fmt.Sprintf("Turn %d | COMBAT", ca.turn)
	ui.DrawText(screen, ca.font, header, lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{255, 100, 100, 255})

	sh := lay.Header.Screen()
	lineY := float32(sh.Bottom())
	sw := float32(ui.ActiveRes.Scale())
	vector.StrokeLine(screen, float32(sh.X+sh.W*0.03), lineY, float32(sh.X+sh.W*0.97), lineY, sw, color.RGBA{60, 60, 80, 255}, false)

	// Labels.
	ui.DrawText(screen, ca.font, "OPPONENT", lay.Opponent.W*0.04, lay.Opponent.Y+lay.Opponent.H*0.02, color.RGBA{255, 120, 120, 255})
	ui.DrawText(screen, ca.font, "YOUR BOARD", lay.Player.W*0.04, lay.Player.Y+lay.Player.H*0.02, color.RGBA{120, 255, 120, 255})

	ca.drawBoard(screen, lay, ca.opponentBoard, false)
	ca.drawBoard(screen, lay, ca.playerBoard, true)
}

func (ca *CombatAnimator) drawBoard(screen *ebiten.Image, lay ui.CombatLayout, board []animMinion, isPlayer bool) {
	zone := lay.Opponent
	if isPlayer {
		zone = lay.Player
	}

	for i, m := range board {
		r := ui.CardRect(zone, i, len(board), lay.CardW, lay.CardH, lay.Gap)
		r.X += m.offsetX
		r.Y += m.offsetY

		// Shake on damage (base-space units).
		if m.flash > 0 {
			t := m.flash / damageIndicatorTime
			intensity := t * 4.0
			r.X += math.Sin(m.flash*30) * intensity
			r.Y += math.Cos(m.flash*25) * intensity * 0.5
		}

		alpha := uint8(255 * m.opacity)
		flashPct := m.flash / damageIndicatorTime

		ca.cr.DrawMinion(screen, m.card, r, alpha, flashPct)

		// Damage splat.
		if m.flash > 0 && m.dmgText != "" {
			ca.drawDamageSplat(screen, m, r, alpha)
		}
	}
}

func (ca *CombatAnimator) drawDamageSplat(screen *ebiten.Image, m animMinion, r ui.Rect, alpha uint8) {
	sr := r.Screen()
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := ui.ActiveRes.Scale()

	t := m.flash / damageIndicatorTime
	splatScale := 0.85 + 0.15*ui.EaseOut(t)

	// Outer glow.
	outerR := float32(28 * s * splatScale)
	vector.FillCircle(screen, cx, cy, outerR, color.RGBA{200, 120, 0, alpha}, false)

	// Spiky rays.
	for j := range 8 {
		angle := float64(j) * math.Pi * 2 / 8
		rx := cx + float32(math.Cos(angle)*22*s*splatScale)
		ry := cy + float32(math.Sin(angle)*22*s*splatScale)
		vector.FillCircle(screen, rx, ry, float32(8*s*splatScale), color.RGBA{255, 180, 0, alpha}, false)
	}

	// Inner core.
	innerR := float32(20 * s * splatScale)
	vector.FillCircle(screen, cx, cy, innerR, color.RGBA{255, 220, 50, alpha}, false)

	// Damage text.
	textScale := 2.2 * splatScale
	charW := 7 * s * textScale
	textW := charW * float64(len(m.dmgText))
	textH := 10 * s * textScale
	op := &text.DrawOptions{}
	op.GeoM.Scale(textScale, textScale)
	op.GeoM.Translate(float64(cx)-textW/2, float64(cy)-textH/2)
	op.ColorScale.ScaleWithColor(color.RGBA{180, 20, 0, alpha})
	text.Draw(screen, m.dmgText, ca.font, op)
}
