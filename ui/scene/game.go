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

// Game orchestrates phase transitions and delegates to phase-specific handlers.
type Game struct {
	client *client.GameClient
	cr     *widget.CardRenderer
	font   *text.GoTextFace

	recruit        *recruitPhase
	combatAnimator *CombatAnimator
	toast          *widget.Toast
	lastPhase      game.Phase
}

func NewGame(c *client.GameClient, cs *cards.Cards, font *text.GoTextFace) *Game {
	cr := &widget.CardRenderer{Cards: cs, Font: font}
	return &Game{
		client:  c,
		cr:      cr,
		font:    font,
		recruit: &recruitPhase{client: c, cr: cr},
		toast:   widget.NewToast(font),
	}
}

func (g *Game) OnEnter() {}
func (g *Game) OnExit()  {}

func (g *Game) Update() error {
	phase := g.client.Phase()

	// Phase transition toasts.
	if g.lastPhase == game.PhaseRecruit && phase == game.PhaseCombat {
		g.toast.Show("COMBAT")
		g.recruit.SyncOrders()
	}
	if g.lastPhase == game.PhaseCombat && phase == game.PhaseRecruit && g.combatAnimator == nil {
		g.toast.Show("RECRUIT")
	}
	g.lastPhase = phase

	g.toast.Update()

	// Start combat animation if events arrived and toast is done.
	if combatEvents := g.client.CombatEvents(); combatEvents != nil && g.combatAnimator == nil && !g.toast.Active() {
		state := g.client.State()
		g.combatAnimator = NewCombatAnimator(
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
	if g.combatAnimator != nil {
		const dt = 1.0 / 60.0
		if g.combatAnimator.Update(dt) {
			g.combatAnimator = nil
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

	// Combat animation takes over.
	if g.combatAnimator != nil {
		g.combatAnimator.Draw(screen)
		g.drawPlayers(screen)
		g.toast.Draw(screen)
		return
	}

	switch g.client.Phase() {
	case game.PhaseWaiting:
		g.drawWaiting(screen)
	case game.PhaseRecruit:
		g.recruit.Draw(screen, g.font, g.client.Turn(), g.timeRemaining())
	case game.PhaseCombat:
		g.drawCombat(screen)
	}

	g.drawPlayers(screen)
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
	ui.DrawText(screen, g.font, header, lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{200, 200, 200, 255})

	lineRect := lay.Header.Screen()
	lineY := float32(lineRect.Bottom())
	vector.StrokeLine(
		screen,
		float32(lineRect.W*0.03),
		lineY,
		float32(lineRect.W*0.97),
		lineY,
		float32(ui.ActiveRes.Scale()),
		color.RGBA{60, 60, 80, 255},
		false,
	)
}

func (g *Game) drawCombat(screen *ebiten.Image) {
	lay := ui.CalcCombatLayout()

	header := fmt.Sprintf("Turn %d | COMBAT", g.client.Turn())
	ui.DrawText(screen, g.font, header, lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{255, 100, 100, 255})

	remaining := g.timeRemaining()
	timer := fmt.Sprintf("%d:%02d", remaining/60, remaining%60)
	ui.DrawText(screen, g.font, timer, lay.Header.W*0.9, lay.Header.H*0.5, color.RGBA{255, 255, 255, 255})

	lineRect := lay.Header.Screen()
	lineY := float32(lineRect.Bottom())
	vector.StrokeLine(
		screen,
		float32(lineRect.W*0.03),
		lineY,
		float32(lineRect.W*0.97),
		lineY,
		float32(ui.ActiveRes.Scale()),
		color.RGBA{60, 60, 80, 255},
		false,
	)

	state := g.client.State()
	if state == nil {
		return
	}

	// Opponent board.
	ui.DrawText(screen, g.font, "OPPONENT", lay.Opponent.W*0.04, lay.Opponent.Y+lay.Opponent.H*0.02, color.RGBA{255, 120, 120, 255})
	for i, c := range state.OpponentBoard {
		r := ui.CardRect(lay.Opponent, i, len(state.OpponentBoard), lay.CardW, lay.CardH, lay.Gap)
		g.cr.DrawMinion(screen, c, r, 255, 0)
	}

	// Player board.
	ui.DrawText(screen, g.font, "YOUR BOARD", lay.Player.W*0.04, lay.Player.Y+lay.Player.H*0.02, color.RGBA{120, 255, 120, 255})
	for i, c := range state.CombatBoard {
		r := ui.CardRect(lay.Player, i, len(state.CombatBoard), lay.CardW, lay.CardH, lay.Gap)
		g.cr.DrawMinion(screen, c, r, 255, 0)
	}
}

func (g *Game) drawPlayers(screen *ebiten.Image) {
	lay := ui.CalcGameLayout()
	x := lay.PlayerBar.X + lay.PlayerBar.W*0.04
	y := lay.PlayerBar.Y + lay.PlayerBar.H*0.2

	playerID := g.client.PlayerID()
	for _, e := range g.client.PlayerList() {
		label := fmt.Sprintf("%s (%d)", e.ID, e.HP)
		clr := color.RGBA{200, 200, 200, 255}
		if e.ID == playerID {
			clr = color.RGBA{100, 255, 100, 255}
		}
		ui.DrawText(screen, g.font, label, x, y, clr)
		x += lay.PlayerBar.W * 0.08
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
