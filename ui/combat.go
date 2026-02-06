package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
)

// Animation timing (seconds).
const (
	attackMoveDuration = 0.8
	attackBackDuration = 0.8
	damageFlashTime    = 0.6
	floatTextDuration  = 2.0
	deathFadeDuration  = 1.2
	eventPause         = 0.5
)

// Combat layout (base coords, scaled at runtime).
const (
	baseOpponentBoardY = 200
	basePlayerBoardY   = 420
)

type animPhase uint8

const (
	animPhaseForward animPhase = iota + 1
	animPhaseBack
)

type animMinion struct {
	card    api.Card
	alive   bool
	offsetX float64
	offsetY float64
	flash   float64 // remaining flash time
	opacity float64 // 1.0 = fully visible
}

type floatingText struct {
	text  string
	x, y  float64
	timer float64
	clr   color.RGBA
}

type attackAnimation struct {
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

// CombatAnimator replays combat events as animations.
type CombatAnimator struct {
	cards *cards.Cards
	font  *text.GoTextFace
	turn  int

	playerBoard   []animMinion
	opponentBoard []animMinion

	events     []game.CombatEvent
	eventIndex int
	pauseTimer float64

	attackAnim  *attackAnimation
	floatingDmg []floatingText
	deathAnims  []int // indices being faded (tracked via opacity on animMinion)

	done bool
}

func NewCombatAnimator(
	turn int,
	playerBoard, opponentBoard []api.Card,
	events []game.CombatEvent,
	c *cards.Cards,
	font *text.GoTextFace,
) *CombatAnimator {
	return &CombatAnimator{
		cards:         c,
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
		board[i] = animMinion{card: c, alive: true, opacity: 1.0}
	}
	return board
}

// Update advances animation state. Returns true when all animations are done.
func (ca *CombatAnimator) Update(dt float64) bool {
	if ca.done {
		return true
	}

	// Update floating damage text.
	ca.updateFloatingText(dt)

	// Update death fades.
	ca.updateDeathFades(dt)

	// Update flash timers.
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

	// Active attack animation.
	if ca.attackAnim != nil {
		ca.updateAttackAnim(dt)
		return false
	}

	// Pause between events.
	if ca.pauseTimer > 0 {
		ca.pauseTimer -= dt
		return false
	}

	// Process next event.
	if ca.eventIndex >= len(ca.events) {
		// Wait for remaining animations to finish.
		if len(ca.floatingDmg) == 0 && !ca.hasDeathFading() {
			ca.done = true
			return true
		}
		return false
	}

	ev := ca.events[ca.eventIndex]
	ca.eventIndex++

	switch ev.Type {
	case game.CombatEventAttack:
		ca.startAttack(ev)
	case game.CombatEventDamage:
		ca.applyDamage(ev)
		ca.pauseTimer = eventPause
	case game.CombatEventDeath:
		ca.startDeath(ev)
		ca.pauseTimer = deathFadeDuration
	}

	return false
}

func (ca *CombatAnimator) startAttack(ev game.CombatEvent) {
	srcIdx, srcIsPlayer := ca.findMinion(ev.SourceID)
	dstIdx, dstIsPlayer := ca.findMinion(ev.TargetID)
	if srcIdx < 0 || dstIdx < 0 {
		return
	}

	srcX, srcY := ca.minionPos(srcIdx, srcIsPlayer)
	dstX, dstY := ca.minionPos(dstIdx, dstIsPlayer)

	ca.attackAnim = &attackAnimation{
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
			a.progress = 0
			a.phase = animPhaseBack
		}
	case animPhaseBack:
		a.progress += dt / attackBackDuration
		if a.progress >= 1.0 {
			// Reset offset.
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

	// Interpolate position.
	board := ca.boardFor(a.srcIsPlayer)
	if a.srcIdx >= len(board) {
		ca.attackAnim = nil
		return
	}

	var t float64
	switch a.phase {
	case animPhaseForward:
		t = easeOut(a.progress)
		board[a.srcIdx].offsetX = (a.targetX - a.startX) * t
		board[a.srcIdx].offsetY = (a.targetY - a.startY) * t
	case animPhaseBack:
		t = 1.0 - easeOut(a.progress)
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
	board[idx].flash = damageFlashTime

	x, y := ca.minionPos(idx, isPlayer)
	ca.floatingDmg = append(ca.floatingDmg, floatingText{
		text:  fmt.Sprintf("-%d", ev.Amount),
		x:     x + sc(baseCardWidth)/2,
		y:     y,
		timer: floatTextDuration,
		clr:   color.RGBA{255, 60, 60, 255},
	})
}

func (ca *CombatAnimator) startDeath(ev game.CombatEvent) {
	idx, isPlayer := ca.findMinion(ev.TargetID)
	if idx < 0 {
		return
	}
	board := ca.boardFor(isPlayer)
	board[idx].alive = false
	// Opacity will be faded in updateDeathFades.
}

func (ca *CombatAnimator) updateFloatingText(dt float64) {
	n := 0
	for i := range ca.floatingDmg {
		ca.floatingDmg[i].timer -= dt
		ca.floatingDmg[i].y -= dt * scf(60) // float upward
		if ca.floatingDmg[i].timer > 0 {
			ca.floatingDmg[n] = ca.floatingDmg[i]
			n++
		}
	}
	ca.floatingDmg = ca.floatingDmg[:n]
}

func (ca *CombatAnimator) updateDeathFades(dt float64) {
	fade := dt / deathFadeDuration
	for i := range ca.playerBoard {
		if !ca.playerBoard[i].alive && ca.playerBoard[i].opacity > 0 {
			ca.playerBoard[i].opacity -= fade
		}
	}
	for i := range ca.opponentBoard {
		if !ca.opponentBoard[i].alive && ca.opponentBoard[i].opacity > 0 {
			ca.opponentBoard[i].opacity -= fade
		}
	}
}

func (ca *CombatAnimator) hasDeathFading() bool {
	for i := range ca.playerBoard {
		if !ca.playerBoard[i].alive && ca.playerBoard[i].opacity > 0 {
			return true
		}
	}
	for i := range ca.opponentBoard {
		if !ca.opponentBoard[i].alive && ca.opponentBoard[i].opacity > 0 {
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

func (ca *CombatAnimator) minionPos(idx int, isPlayer bool) (float64, float64) {
	board := ca.boardFor(isPlayer)
	totalW := float64(len(board))*sc(baseCardWidth+baseCardGap) - sc(baseCardGap)
	startX := (float64(ActiveRes.Width) - totalW) / 2
	x := startX + float64(idx)*sc(baseCardWidth+baseCardGap)

	var y float64
	if isPlayer {
		y = sc(basePlayerBoardY)
	} else {
		y = sc(baseOpponentBoardY)
	}
	return x, y
}

// Draw renders the combat animation.
func (ca *CombatAnimator) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBackground)

	w := float64(ActiveRes.Width)

	// Header.
	header := fmt.Sprintf("Turn %d | COMBAT", ca.turn)
	drawText(screen, ca.font, header, sc(50), sc(50), color.RGBA{255, 100, 100, 255})
	vector.StrokeLine(
		screen,
		float32(sc(40)),
		float32(sc(80)),
		float32(w-sc(40)),
		float32(sc(80)),
		1,
		color.RGBA{60, 60, 80, 255},
		false,
	)

	// Labels.
	drawText(screen, ca.font, "OPPONENT", sc(50), sc(baseOpponentBoardY)-scf(30), color.RGBA{255, 120, 120, 255})
	drawText(screen, ca.font, "YOUR BOARD", sc(50), sc(basePlayerBoardY)-scf(30), color.RGBA{120, 255, 120, 255})

	// Draw boards.
	ca.drawBoard(screen, ca.opponentBoard, false)
	ca.drawBoard(screen, ca.playerBoard, true)

	// Draw floating damage text.
	for _, ft := range ca.floatingDmg {
		alpha := uint8(255 * (ft.timer / floatTextDuration))
		clr := color.RGBA{ft.clr.R, ft.clr.G, ft.clr.B, alpha}
		drawText(screen, ca.font, ft.text, ft.x, ft.y, clr)
	}
}

func (ca *CombatAnimator) drawBoard(screen *ebiten.Image, board []animMinion, isPlayer bool) {
	for i, m := range board {
		if m.opacity <= 0 {
			continue
		}

		x, y := ca.minionPos(i, isPlayer)
		x += m.offsetX
		y += m.offsetY

		cw := float32(sc(baseCardWidth))
		ch := float32(sc(baseCardHeight))

		// Card background with opacity.
		alpha := uint8(255 * m.opacity)
		bg := color.RGBA{40, 40, 60, alpha}
		if m.flash > 0 {
			bg = color.RGBA{180, 40, 40, alpha}
		}
		vector.FillRect(screen, float32(x), float32(y), cw, ch, bg, false)

		// Border.
		border := color.RGBA{80, 80, 100, alpha}
		borderWidth := float32(2)
		if m.card.IsGolden {
			border = color.RGBA{255, 215, 0, alpha}
			borderWidth = 3
		}
		vector.StrokeRect(screen, float32(x), float32(y), cw, ch, borderWidth, border, false)

		// Name.
		name := m.card.TemplateID
		if t := ca.cards.ByTemplateID(m.card.TemplateID); t != nil {
			name = t.Name
		}
		drawText(screen, ca.font, name, x+scf(5), y+scf(5), color.RGBA{255, 255, 255, alpha})

		// Attack (bottom-left).
		drawText(screen, ca.font, fmt.Sprintf("%d", m.card.Attack),
			x+scf(5), y+sc(baseCardHeight)-scf(18),
			color.RGBA{255, 215, 0, alpha})

		// Health (bottom-right).
		hpClr := color.RGBA{255, 80, 80, alpha}
		drawText(screen, ca.font, fmt.Sprintf("%d", m.card.Health),
			x+sc(baseCardWidth)-scf(20), y+sc(baseCardHeight)-scf(18),
			hpClr)
	}
}

// easeOut gives a smooth deceleration curve.
func easeOut(t float64) float64 {
	return 1 - math.Pow(1-t, 2)
}
