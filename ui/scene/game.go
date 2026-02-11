package scene

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

type tierFade struct {
	tier  game.Tier
	timer float64 // seconds remaining
}

// Game orchestrates phase transitions and delegates to phase-specific handlers.
type Game struct {
	client *client.GameClient
	cr     *widget.CardRenderer
	font   *text.GoTextFace

	recruit      *recruitPhase
	combat       *combatPanel
	toast        *widget.Toast
	lastPhase    game.Phase
	sidebarHover int
	tierFades    map[string]tierFade
	sidebarSnap  []client.PlayerEntry // frozen sidebar during combat animation
	backBtn      *widget.Button
	onBackToMenu func()
}

func NewGame(c *client.GameClient, cs *cards.Cards, font *text.GoTextFace, onBackToMenu func()) *Game {
	cr := &widget.CardRenderer{Cards: cs, Font: font}
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	btnW := w * 0.15
	btnH := h * 0.06

	g := &Game{
		client: c,
		cr:     cr,
		font:   font,
		recruit: &recruitPhase{
			client: c,
			cr:     cr,
			shop:   &shopPanel{client: c, cr: cr},
		},
		toast:        widget.NewToast(font),
		sidebarHover: -1,
		tierFades:    make(map[string]tierFade),
		onBackToMenu: onBackToMenu,
	}

	g.backBtn = &widget.Button{
		Rect:      ui.Rect{X: w/2 - btnW/2, Y: h * 0.75, W: btnW, H: btnH},
		Text:      "Back to Menu",
		Color:     color.RGBA{60, 60, 90, 255},
		BorderClr: color.RGBA{100, 100, 140, 255},
		TextClr:   color.RGBA{200, 200, 255, 255},
		OnClick: func() {
			if g.onBackToMenu != nil {
				g.onBackToMenu()
			}
		},
	}

	return g
}

func (g *Game) OnEnter() {}
func (g *Game) OnExit()  {}

func (g *Game) updateSidebarHover() {
	g.sidebarHover = -1
	sb := ui.CalcGameLayout().Sidebar
	mx, my := ebiten.CursorPosition()
	if !sb.Contains(mx, my) {
		return
	}
	rowH := sb.H / float64(game.MaxPlayers)
	players := g.sidebarPlayers()
	for i := range players {
		row := ui.Rect{X: sb.X, Y: sb.Y + float64(i)*rowH, W: sb.W, H: rowH}
		if row.Contains(mx, my) {
			g.sidebarHover = i
			return
		}
	}
}

const tierFadeDuration = 2.0

func (g *Game) Update() error {
	g.updateSidebarHover()

	// Drain opponent tier updates.
	for _, u := range g.client.DrainOpponentUpdates() {
		g.tierFades[u.PlayerID] = tierFade{tier: u.ShopTier, timer: tierFadeDuration}
	}

	// Tick fade timers.
	for id, f := range g.tierFades {
		f.timer -= 1.0 / 60.0
		if f.timer <= 0 {
			delete(g.tierFades, id)
		} else {
			g.tierFades[id] = f
		}
	}

	phase := g.client.Phase()

	// Keep sidebar snapshot fresh during recruit, but freeze it during combat animation and toasts.
	if phase == game.PhaseRecruit && !g.toast.Active() && g.combat == nil {
		g.sidebarSnap = g.client.PlayerList()
	}

	// Phase transition toasts.
	if g.lastPhase == game.PhaseRecruit && phase == game.PhaseCombat {
		g.toast.Show("COMBAT")
		g.recruit.SyncOrders()
	}
	if g.lastPhase == game.PhaseCombat && phase == game.PhaseRecruit && g.combat == nil {
		g.toast.Show("RECRUIT")
	}
	if phase == game.PhaseFinished && g.lastPhase != game.PhaseFinished {
		g.toast.Show("GAME OVER")
	}
	g.lastPhase = phase

	g.toast.Update()

	if phase == game.PhaseFinished {
		g.backBtn.Update()
		return nil
	}

	// Start combat animation if events arrived and toast is done.
	if combatEvents := g.client.CombatEvents(); combatEvents != nil && g.combat == nil && !g.toast.Active() {
		state := g.client.State()
		g.combat = newCombatPanel(
			g.client.Turn(),
			state.CombatBoard,
			state.OpponentBoard,
			combatEvents,
			g.cr.Cards,
			g.font,
		)
		g.client.ClearCombatAnimation()
	}

	// Combat animation blocks all input.
	if g.combat != nil {
		const dt = 1.0 / 60.0
		if g.combat.Update(dt) {
			g.combat = nil
			g.toast.Show("RECRUIT")
		}
		return nil
	}

	if phase == game.PhaseRecruit {
		return g.recruit.Update()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ui.ColorBackground)

	if !g.client.Connected() {
		g.drawConnecting(screen)
		return
	}

	switch g.client.Phase() {
	case game.PhaseWaiting:
		g.drawWaiting(screen)
	case game.PhaseRecruit:
		if g.combat != nil {
			g.drawCombat(screen)
		} else {
			g.recruit.Draw(screen, g.font, g.client.Turn(), g.timeRemaining())
		}
	case game.PhaseCombat:
		g.drawCombat(screen)
	case game.PhaseFinished:
		g.drawGameResult(screen)
		g.toast.Draw(screen)
		return
	}

	g.drawSidebar(screen)
	g.drawSidebarTooltip(screen)
	g.toast.Draw(screen)
}

func (g *Game) drawConnecting(screen *ebiten.Image) {
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	ui.DrawText(screen, g.font, "Connecting...", w*0.45, h*0.5, color.RGBA{255, 255, 255, 255})
}

func (g *Game) drawWaiting(screen *ebiten.Image) {
	lay := ui.CalcGameLayout()

	playerCount := len(g.client.Opponents()) + 1
	header := fmt.Sprintf(
		"You are Player %s | Waiting for players... %d/%d",
		g.client.PlayerID(),
		playerCount,
		game.MaxPlayers,
	)
	ui.DrawText(screen, g.font, header, lay.Header.X+lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{200, 200, 200, 255})

	lineRect := lay.Header.Screen()
	lineY := float32(lineRect.Bottom())
	vector.StrokeLine(
		screen,
		float32(lineRect.X+lineRect.W*0.03),
		lineY,
		float32(lineRect.X+lineRect.W*0.97),
		lineY,
		float32(ui.ActiveRes.Scale()),
		color.RGBA{60, 60, 80, 255},
		false,
	)
}

// drawCombat renders the combat phase using GameLayout zones.
func (g *Game) drawCombat(screen *ebiten.Image) {
	lay := ui.CalcGameLayout()

	// Header.
	header := fmt.Sprintf("Turn %d", g.client.Turn())
	ui.DrawText(screen, g.font, header,
		lay.Header.X+lay.Header.W*0.04, lay.Header.H*0.5,
		color.RGBA{200, 200, 200, 255})

	ui.DrawText(screen, g.font, "Combat",
		lay.Header.X+lay.Header.W*0.9, lay.Header.H*0.5,
		color.RGBA{200, 200, 200, 255})

	lineRect := lay.Header.Screen()
	lineY := float32(lineRect.Bottom())
	vector.StrokeLine(
		screen,
		float32(lineRect.X+lineRect.W*0.03), lineY,
		float32(lineRect.X+lineRect.W*0.97), lineY,
		float32(ui.ActiveRes.Scale()),
		color.RGBA{60, 60, 80, 255},
		false,
	)

	// Player stats (gold + tier, no buttons).
	if p := g.client.Player(); p != nil {
		stats := fmt.Sprintf("Gold: %d/%d | Tier: %d", p.Gold, p.MaxGold, p.ShopTier)
		ui.DrawText(screen, g.font, stats,
			lay.BtnRow.X+lay.BtnRow.W*0.04, lay.BtnRow.Y+lay.BtnRow.H*0.15,
			color.RGBA{255, 215, 0, 255})
	}

	// Opponent board in Shop zone.
	ui.DrawText(screen, g.font, "OPPONENT",
		lay.Shop.X+lay.Shop.W*0.04, lay.Shop.Y+lay.Shop.H*0.02,
		color.RGBA{150, 150, 150, 255})

	// Player board in Board zone.
	ui.DrawText(screen, g.font, "BOARD",
		lay.Board.X+lay.Board.W*0.04, lay.Board.Y+lay.Board.H*0.02,
		color.RGBA{150, 150, 150, 255})

	if g.combat != nil {
		g.combat.drawOpponentBoard(screen, lay)
		g.combat.drawPlayerBoard(screen, lay)
	} else {
		// Static boards before animation starts.
		state := g.client.State()
		if state != nil {
			for i, c := range state.OpponentBoard {
				r := ui.CardRect(lay.Shop, i, len(state.OpponentBoard), lay.CardW, lay.CardH, lay.Gap)
				g.cr.DrawMinion(screen, c, r, 255, 0)
			}
			for i, c := range state.CombatBoard {
				r := ui.CardRect(lay.Board, i, len(state.CombatBoard), lay.CardW, lay.CardH, lay.Gap)
				g.cr.DrawMinion(screen, c, r, 255, 0)
			}
		}
	}

	// Hand (read-only).
	ui.DrawText(screen, g.font, "HAND",
		lay.Hand.X+lay.Hand.W*0.04, lay.Hand.Y+lay.Hand.H*0.02,
		color.RGBA{100, 100, 100, 255})
	hand := g.client.Hand()
	for i, c := range hand {
		rect := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
		if g.recruit.isSpell(c) {
			g.cr.DrawSpellCard(screen, c, rect)
		} else {
			g.cr.DrawMinionCard(screen, c, rect)
		}
	}
}

func (g *Game) drawGameResult(screen *ebiten.Image) {
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)

	result := g.client.GameResult()
	if result == nil {
		ui.DrawText(screen, g.font, "Game Over",
			w*0.42, h*0.4, color.RGBA{255, 215, 0, 255})
		g.backBtn.Draw(screen, g.font)
		return
	}

	playerID := g.client.PlayerID()
	if playerID == result.WinnerID {
		ui.DrawText(screen, g.font, "YOU WON",
			w*0.42, h*0.08, color.RGBA{255, 215, 0, 255})
	} else {
		ui.DrawText(screen, g.font, "GAME OVER",
			w*0.42, h*0.08, color.RGBA{200, 200, 200, 255})
	}

	winnerText := fmt.Sprintf("Winner: %s", result.WinnerID)
	ui.DrawText(screen, g.font, winnerText,
		w*0.38, h*0.18, color.RGBA{100, 255, 100, 255})

	startY := h * 0.30
	lineH := h * 0.06

	for i, p := range result.Placements {
		y := startY + float64(i)*lineH

		clr := color.RGBA{200, 200, 200, 255}
		if p.PlayerID == playerID {
			clr = color.RGBA{100, 255, 100, 255}
		}
		if p.Placement == 1 {
			clr = color.RGBA{255, 215, 0, 255}
		}

		line := fmt.Sprintf("#%d  %s", p.Placement, p.PlayerID)
		if p.MajorityTribe != game.TribeNeutral && p.MajorityTribe != game.TribeMixed {
			line += fmt.Sprintf("  (%s x%d)", p.MajorityTribe, p.MajorityCount)
		}
		ui.DrawText(screen, g.font, line, w*0.35, y, clr)
	}

	if result.Duration > 0 {
		minutes := result.Duration / 60
		seconds := result.Duration % 60
		durationText := fmt.Sprintf("Duration: %dm %ds", minutes, seconds)
		ui.DrawText(screen, g.font, durationText,
			w*0.38, h*0.85, color.RGBA{150, 150, 170, 255})
	}

	g.backBtn.Draw(screen, g.font)
}

func (g *Game) sidebarPlayers() []client.PlayerEntry {
	if g.sidebarSnap != nil {
		return g.sidebarSnap
	}
	return g.client.PlayerList()
}

func (g *Game) drawSidebar(screen *ebiten.Image) {
	lay := ui.CalcGameLayout()
	sb := lay.Sidebar
	sbScreen := sb.Screen()

	// Background.
	vector.FillRect(screen,
		float32(sbScreen.X), float32(sbScreen.Y),
		float32(sbScreen.W), float32(sbScreen.H),
		color.RGBA{15, 15, 25, 255}, false,
	)

	playerID := g.client.PlayerID()
	state := g.client.State()
	var opponentID string
	if state != nil {
		opponentID = state.OpponentID
	}

	players := g.sidebarPlayers()
	rowH := sb.H / float64(game.MaxPlayers)
	scale := float32(ui.ActiveRes.Scale())

	for i, e := range players {
		rowY := sb.Y + float64(i)*rowH
		row := ui.Rect{X: sb.X, Y: rowY, W: sb.W, H: rowH}
		rowScreen := row.Screen()
		padX := row.W * 0.06

		// Highlight self.
		if e.ID == playerID {
			vector.FillRect(screen,
				float32(rowScreen.X), float32(rowScreen.Y),
				float32(rowScreen.W), float32(rowScreen.H),
				color.RGBA{30, 50, 30, 255}, false,
			)
		}

		// Red border for current combat opponent.
		if e.ID == opponentID {
			vector.StrokeRect(screen,
				float32(rowScreen.X), float32(rowScreen.Y),
				float32(rowScreen.W), float32(rowScreen.H),
				scale*2, color.RGBA{255, 80, 80, 255}, false,
			)
		}

		// Line 1: Name + HP.
		nameClr := color.RGBA{200, 200, 200, 255}
		if e.ID == playerID {
			nameClr = color.RGBA{100, 255, 100, 255}
		}
		ui.DrawText(screen, g.font, fmt.Sprintf("%s  %d HP", e.ID, e.HP), row.X+padX, row.Y+rowH*0.2, nameClr)

		// Line 2: Tier + tribe.
		line2 := fmt.Sprintf("Tier %d", e.ShopTier)
		switch e.MajorityTribe {
		case game.TribeNeutral:
		case game.TribeMixed:
			line2 += " | Mixed"
		default:
			line2 += fmt.Sprintf(" | %s x%d", e.MajorityTribe, e.MajorityCount)
		}
		ui.DrawText(screen, g.font, line2, row.X+padX, row.Y+rowH*0.55, color.RGBA{160, 160, 180, 255})

		// Fading tier upgrade indicator.
		if f, ok := g.tierFades[e.ID]; ok {
			alpha := uint8(255 * (f.timer / tierFadeDuration))
			tierStr := fmt.Sprintf("T%d!", f.tier)
			ui.DrawText(screen, g.font, tierStr, row.X+row.W*0.70, row.Y+rowH*0.35, color.RGBA{255, 215, 0, alpha})
		}

		// Row separator.
		sepY := float32(rowScreen.Bottom())
		vector.StrokeLine(screen,
			float32(rowScreen.X+rowScreen.W*0.05), sepY,
			float32(rowScreen.X+rowScreen.W*0.95), sepY,
			scale, color.RGBA{40, 40, 60, 255}, false,
		)
	}
}

func (g *Game) drawSidebarTooltip(screen *ebiten.Image) {
	if g.sidebarHover < 0 {
		return
	}
	players := g.sidebarPlayers()
	if g.sidebarHover >= len(players) {
		return
	}

	e := players[g.sidebarHover]
	sb := ui.CalcGameLayout().Sidebar
	rowH := sb.H / float64(game.MaxPlayers)
	scale := float32(ui.ActiveRes.Scale())

	// Tooltip positioned to the right of sidebar, aligned with hovered row.
	tipW := sb.W * 1.4
	tipH := rowH * 2.5
	tipX := sb.Right()
	tipY := sb.Y + float64(g.sidebarHover)*rowH

	// Clamp to screen bottom.
	if tipY+tipH > float64(ui.BaseHeight) {
		tipY = float64(ui.BaseHeight) - tipH
	}

	tip := ui.Rect{X: tipX, Y: tipY, W: tipW, H: tipH}
	tipScreen := tip.Screen()
	padX := tip.W * 0.06

	// Background.
	vector.FillRect(screen,
		float32(tipScreen.X), float32(tipScreen.Y),
		float32(tipScreen.W), float32(tipScreen.H),
		color.RGBA{25, 25, 40, 255}, false,
	)
	vector.StrokeRect(screen,
		float32(tipScreen.X), float32(tipScreen.Y),
		float32(tipScreen.W), float32(tipScreen.H),
		scale, color.RGBA{60, 60, 90, 255}, false,
	)

	// Header: player name.
	ui.DrawText(screen, g.font, e.ID, tip.X+padX, tip.Y+tip.H*0.08, color.RGBA{220, 220, 220, 255})

	// Tribe info.
	switch e.MajorityTribe {
	case game.TribeNeutral:
	case game.TribeMixed:
		ui.DrawText(screen, g.font, "Mixed", tip.X+padX, tip.Y+tip.H*0.22, color.RGBA{180, 180, 200, 255})
	default:
		tribeStr := fmt.Sprintf("%s x%d", e.MajorityTribe, e.MajorityCount)
		ui.DrawText(screen, g.font, tribeStr, tip.X+padX, tip.Y+tip.H*0.22, color.RGBA{180, 180, 200, 255})
	}

	// Last 3 combat results.
	y := tip.Y + tip.H*0.40
	lineH := tip.H * 0.18
	for _, cr := range e.CombatResults {
		var label string
		var clr color.Color
		switch cr.WinnerID {
		case "":
			label = "Tie vs " + cr.OpponentID
			clr = color.RGBA{140, 140, 140, 255}
		case e.ID:
			label = fmt.Sprintf("Won vs %s (%d dmg)", cr.OpponentID, cr.Damage)
			clr = color.RGBA{80, 220, 80, 255}
		default:
			label = fmt.Sprintf("Lost vs %s (%d dmg)", cr.OpponentID, cr.Damage)
			clr = color.RGBA{220, 80, 80, 255}
		}
		ui.DrawText(screen, g.font, label, tip.X+padX, y, clr)
		y += lineH
	}

	if len(e.CombatResults) == 0 {
		ui.DrawText(screen, g.font, "No fights yet", tip.X+padX, tip.Y+tip.H*0.40, color.RGBA{100, 100, 100, 255})
	}
}

func (g *Game) timeRemaining() int64 {
	endTime := g.client.PhaseEndTimestamp()
	now := time.Now().Unix()
	remaining := endTime - now
	if remaining < 0 {
		return 0
	}
	return remaining
}
