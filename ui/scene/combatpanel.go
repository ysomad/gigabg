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
	"github.com/ysomad/gigabg/game/catalog"
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

// combatPanel replays combat events as visual animations using GameLayout zones.
// Opponent board renders in the Shop zone, player board in the Board zone.
type combatPanel struct {
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

func newCombatPanel(
	turn int,
	playerBoard, opponentBoard []api.Card,
	events []game.CombatEvent,
	c *catalog.Catalog,
	font *text.GoTextFace,
) *combatPanel {
	slog.Info("combat panel created",
		"player_board", len(playerBoard),
		"opponent_board", len(opponentBoard),
		"events", len(events))
	return &combatPanel{
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
func (cp *combatPanel) Update(dt float64) bool {
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

	if cp.attackAnim != nil {
		cp.updateAttackAnim(dt)
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
			cp.startAttack(ev)
			return false
		}
	}

	if !cp.hasActiveIndicator() && !cp.hasDying() {
		cp.done = true
		return true
	}
	return false
}

func (cp *combatPanel) startAttack(ev game.CombatEvent) {
	srcIdx, srcIsPlayer := cp.findMinion(ev.SourceID)
	dstIdx, dstIsPlayer := cp.findMinion(ev.TargetID)
	if srcIdx < 0 || dstIdx < 0 {
		return
	}

	lay := ui.CalcGameLayout()
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

func (cp *combatPanel) updateAttackAnim(dt float64) {
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

			lay := ui.CalcGameLayout()
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

func (cp *combatPanel) applyDamage(ev game.CombatEvent) {
	idx, isPlayer := cp.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}
	board := cp.boardFor(isPlayer)
	board[idx].card.Health -= ev.Amount
	board[idx].flash = damageIndicatorTime
	board[idx].dmgText = fmt.Sprintf("-%d", ev.Amount)
}

func (cp *combatPanel) markDying(combatID int) {
	for i, m := range cp.playerBoard {
		if m.card.CombatID == combatID {
			cp.playerBoard[i].dying = true
			return
		}
	}
	for i, m := range cp.opponentBoard {
		if m.card.CombatID == combatID {
			cp.opponentBoard[i].dying = true
			return
		}
	}
}

func (cp *combatPanel) consumeHitEvents() {
	for cp.eventIndex < len(cp.events) {
		ev := cp.events[cp.eventIndex]
		switch ev.Type {
		case game.CombatEventDamage:
			cp.applyDamage(ev)
			cp.eventIndex++
		case game.CombatEventDeath:
			cp.markDying(ev.TargetID)
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
		if m.flash > 0 {
			t := m.flash / damageIndicatorTime
			intensity := t * 4.0
			r.X += math.Sin(m.flash*30) * intensity
			r.Y += math.Cos(m.flash*25) * intensity * 0.5
		}

		alpha := uint8(255 * m.opacity)
		flashPct := m.flash / damageIndicatorTime

		cp.cr.DrawMinion(screen, m.card, r, alpha, flashPct)

		// Damage splat.
		if m.flash > 0 && m.dmgText != "" {
			cp.drawDamageSplat(screen, m, r, alpha)
		}
	}
}

func (cp *combatPanel) drawDamageSplat(screen *ebiten.Image, m animMinion, r ui.Rect, alpha uint8) {
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
	text.Draw(screen, m.dmgText, cp.font, op)
}
