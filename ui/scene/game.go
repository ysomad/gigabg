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
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// Game orchestrates phase transitions and delegates to phase-specific handlers.
type Game struct {
	client   *client.GameClient
	cr       *widget.CardRenderer
	font     *text.GoTextFace
	boldFont *text.GoTextFace

	recruit      *recruitPhase
	combat       *combatPanel
	sidebar      *widget.Sidebar
	phaseToast   *widget.Toast
	lastPhase    game.Phase
	backBtn      *widget.Button
	onBackToMenu func()
	lay          ui.GameLayout
}

func NewGame(c *client.GameClient, cs *catalog.Catalog, font, boldFont *text.GoTextFace, onBackToMenu func()) *Game {
	cr := &widget.CardRenderer{Cards: cs, Font: font, BoldFont: boldFont}
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	btnW := w * 0.15
	btnH := h * 0.06

	g := &Game{
		client:   c,
		cr:       cr,
		font:     font,
		boldFont: boldFont,
		recruit: &recruitPhase{
			client: c,
			cr:     cr,
			shop:   &shopPanel{client: c, cr: cr},
		},
		sidebar:      widget.NewSidebar(font),
		phaseToast:   widget.NewToast(font),
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

func (g *Game) Update() error {
	g.cr.Tick++
	g.lay = ui.CalcGameLayout()
	dt := 1.0 / float64(ebiten.TPS())
	phase := g.client.Phase()

	// Keep sidebar snapshot fresh during recruit, but freeze it during combat animation and toasts.
	if phase == game.PhaseRecruit && !g.phaseToast.Active() && g.combat == nil {
		g.sidebar.Update(g.lay.Sidebar, g.client.PlayerList(), g.client.DrainOpponentUpdates(), dt)
	} else {
		g.sidebar.Update(g.lay.Sidebar, nil, g.client.DrainOpponentUpdates(), dt)
	}

	if phase != g.lastPhase {
		g.onPhaseTransition(g.lastPhase, phase)
	}
	g.lastPhase = phase

	g.phaseToast.Update(dt)

	// Start combat animation if events arrived and toast is done.
	if combatEvents := g.client.CombatEvents(); combatEvents != nil && g.combat == nil && !g.phaseToast.Active() {
		state := g.client.State()
		g.combat = newCombatPanel(
			g.client.Turn(),
			g.client.PlayerID(),
			state.CombatBoard,
			state.OpponentBoard,
			combatEvents,
			g.cr.Cards,
			g.font,
			g.boldFont,
		)
		g.client.ClearCombatLog()
	}

	// Combat animation blocks all input.
	if g.combat != nil {
		done, err := g.combat.Update(dt, g.lay)
		if err != nil {
			return fmt.Errorf("combat update: %w", err)
		}
		if done {
			g.combat = nil
			if phase == game.PhaseFinished {
				g.phaseToast.Show("GAME OVER")
			} else {
				g.phaseToast.Show("RECRUIT")
			}
		}
		return nil
	}

	if phase == game.PhaseFinished {
		g.backBtn.Update()
		return nil
	}

	if phase == game.PhaseRecruit {
		return g.recruit.Update(g.lay)
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
			g.recruit.Draw(screen, g.font, g.lay, g.client.Turn(), g.timeRemaining())
		}
	case game.PhaseCombat:
		g.drawCombat(screen)
	case game.PhaseFinished:
		if g.hasPendingCombat() {
			g.drawCombat(screen)
		} else {
			g.drawGameResult(screen)
			g.phaseToast.Draw(screen, g.toastRect())
			return
		}
	}

	playerID := g.client.PlayerID()
	state := g.client.State()
	var opponentID string
	if state != nil {
		opponentID = state.OpponentID
	}
	g.sidebar.Draw(screen, g.lay.Sidebar, playerID, opponentID)
	g.phaseToast.Draw(screen, g.toastRect())
}

func (g *Game) drawConnecting(screen *ebiten.Image) {
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	ui.DrawText(screen, g.font, "Connecting...", w*0.45, h*0.5, color.RGBA{255, 255, 255, 255})
}

func (g *Game) drawWaiting(screen *ebiten.Image) {
	playerCount := len(g.client.Opponents()) + 1
	header := fmt.Sprintf(
		"You are Player %s | Waiting for players... %d/%d",
		g.client.PlayerID(),
		playerCount,
		game.MaxPlayers,
	)
	ui.DrawText(
		screen,
		g.font,
		header,
		g.lay.Header.X+g.lay.Header.W*0.04,
		g.lay.Header.H*0.5,
		color.RGBA{200, 200, 200, 255},
	)

	lineRect := g.lay.Header.Screen()
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
	lay := g.lay

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
		g.cr.DrawHandCard(screen, c, rect)
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

	winnerText := "Winner: " + result.WinnerID
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
		minutes := int(result.Duration.Minutes())
		seconds := int(result.Duration.Seconds()) % 60
		durationText := fmt.Sprintf("Duration: %dm %ds", minutes, seconds)
		ui.DrawText(screen, g.font, durationText,
			w*0.38, h*0.85, color.RGBA{150, 150, 170, 255})
	}

	g.backBtn.Draw(screen, g.font)
}

func (g *Game) toastRect() ui.Rect {
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	boxW := w * 0.3
	boxH := h * 0.1
	return ui.Rect{X: w/2 - boxW/2, Y: h/2 - boxH/2, W: boxW, H: boxH}
}

func (g *Game) timeRemaining() time.Duration {
	remaining := time.Until(g.client.PhaseEndsAt())
	if remaining < 0 {
		return 0
	}
	return remaining
}

// hasPendingCombat returns true if a combat animation is playing or combat events are waiting.
func (g *Game) hasPendingCombat() bool {
	return g.combat != nil || g.client.CombatEvents() != nil
}

func (g *Game) onPhaseTransition(from, to game.Phase) {
	switch {
	case from == game.PhaseRecruit:
		// Recruit → Combat or Recruit → Finished (game ended during combat).
		g.phaseToast.Show("COMBAT")
		g.recruit.SyncOrders()
	case to == game.PhaseRecruit && g.combat == nil:
		g.phaseToast.Show("RECRUIT")
	case from == game.PhaseCombat && to == game.PhaseFinished:
		if !g.hasPendingCombat() {
			g.phaseToast.Show("GAME OVER")
		}
	}
}
