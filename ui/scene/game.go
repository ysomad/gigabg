package scene

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// Game renders the recruit phase, combat display, and combat animation.
type Game struct {
	client *client.GameConn
	cr     *widget.CardRenderer
	font   *text.GoTextFace

	// Drag state.
	dragging      bool
	dragIndex     int
	dragFromBoard bool
	dragCurrentX  int
	dragCurrentY  int

	// Hover tooltip.
	hoverCard *api.Card
	hoverRect ui.Rect

	// Local board order (indices into server board).
	boardOrder []int
	lastPhase  game.Phase

	// Combat animation.
	combatAnimator *CombatAnimator

	toast *widget.Toast
}

func NewGame(c *client.GameConn, cs *cards.Cards, font *text.GoTextFace) *Game {
	return &Game{
		client: c,
		cr:     &widget.CardRenderer{Cards: cs, Font: font},
		font:   font,
		toast:  widget.NewToast(font),
	}
}

func (g *Game) OnEnter() {}
func (g *Game) OnExit()  {}

func (g *Game) isSpell(c api.Card) bool {
	t := g.cr.Cards.ByTemplateID(c.TemplateID)
	return t != nil && t.IsSpell()
}

// screenToBase converts screen pixel coords to base coords.
func screenToBase(screenX, screenY int) (float64, float64) {
	s := ui.ActiveRes.Scale()
	bx := (float64(screenX) - ui.ActiveRes.OffsetX()) / s
	by := (float64(screenY) - ui.ActiveRes.OffsetY()) / s
	return bx, by
}

func (g *Game) Update() error {
	phase := g.client.Phase()

	// Phase transition toasts.
	if g.lastPhase == game.PhaseRecruit && phase == game.PhaseCombat {
		g.toast.Show("COMBAT")
	}
	if g.lastPhase == game.PhaseCombat && phase == game.PhaseRecruit && g.combatAnimator == nil {
		g.toast.Show("RECRUIT")
	}

	// Sync board order before combat starts.
	if g.lastPhase == game.PhaseRecruit && phase == game.PhaseCombat {
		if len(g.boardOrder) > 0 {
			g.client.SyncBoard(g.boardOrder)
		}
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

	// Keep board order in sync with server board size.
	board := g.client.Board()
	if len(g.boardOrder) != len(board) {
		g.boardOrder = make([]int, len(board))
		for i := range g.boardOrder {
			g.boardOrder[i] = i
		}
	}

	if phase != game.PhaseRecruit {
		return nil
	}

	lay := ui.CalcGameLayout()
	mx, my := ebiten.CursorPosition()

	// Reset hover each frame.
	g.hoverCard = nil

	// Discover overlay blocks all other input.
	if discover := g.client.Discover(); discover != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.handleDiscoverClick(lay, discover, mx, my)
		}
		return nil
	}

	// Start drag.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// From board.
		for i := range g.boardOrder {
			r := ui.CardRect(lay.Board, i, len(g.boardOrder), lay.CardW, lay.CardH, lay.Gap)
			if r.Contains(mx, my) {
				g.dragging = true
				g.dragIndex = i
				g.dragFromBoard = true
				g.dragCurrentX = mx
				g.dragCurrentY = my
				return nil
			}
		}
		// From hand.
		hand := g.client.Hand()
		for i, c := range hand {
			r := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
			if r.Contains(mx, my) {
				if t := g.cr.Cards.ByTemplateID(c.TemplateID); t != nil && t.IsSpell() {
					g.client.PlaySpell(i)
					return nil
				}
				g.dragging = true
				g.dragIndex = i
				g.dragFromBoard = false
				g.dragCurrentX = mx
				g.dragCurrentY = my
				return nil
			}
		}
	}

	// Update drag position.
	if g.dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.dragCurrentX = mx
		g.dragCurrentY = my
		return nil
	}

	// End drag.
	if g.dragging && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.endDrag(lay, mx, my)
		return nil
	}

	// Click buttons or shop.
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
		switch {
		case refresh.Contains(mx, my):
			g.client.RefreshShop()
			return nil
		case upgrade.Contains(mx, my):
			g.client.UpgradeShop()
			return nil
		case freeze.Contains(mx, my):
			g.client.FreezeShop()
			return nil
		}

		// Buy from shop.
		shop := g.client.Shop()
		for i := range shop {
			r := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
			if r.Contains(mx, my) {
				g.client.BuyCard(i)
				return nil
			}
		}
	}

	// Hover detection (skip while dragging).
	if !g.dragging {
		shop := g.client.Shop()
		for i, c := range shop {
			r := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
			if r.Contains(mx, my) {
				g.hoverCard = &c
				g.hoverRect = r
				break
			}
		}
		if g.hoverCard == nil {
			board := g.client.Board()
			for i, serverIdx := range g.boardOrder {
				if serverIdx >= 0 && serverIdx < len(board) {
					r := ui.CardRect(lay.Board, i, len(g.boardOrder), lay.CardW, lay.CardH, lay.Gap)
					if r.Contains(mx, my) {
						c := board[serverIdx]
						g.hoverCard = &c
						g.hoverRect = r
						break
					}
				}
			}
		}
	}

	return nil
}

func (g *Game) endDrag(lay ui.GameLayout, mx, my int) {
	defer func() { g.dragging = false }()

	dropPad := lay.CardH * 0.4

	if g.dragFromBoard {
		shopZone := ui.Rect{X: lay.Shop.X, Y: lay.Shop.Y - dropPad, W: lay.Shop.W, H: lay.Shop.H + 2*dropPad}
		if shopZone.Contains(mx, my) {
			// Sell by dragging board card to shop.
			g.client.SellMinion(g.boardOrder[g.dragIndex])
			return
		}
	}

	boardZone := ui.Rect{X: lay.Board.X, Y: lay.Board.Y - dropPad, W: lay.Board.W, H: lay.Board.H + 2*dropPad}
	if !boardZone.Contains(mx, my) {
		return
	}

	pos := g.getDropPosition(lay, mx)
	if g.dragFromBoard {
		// Reorder locally.
		if pos != g.dragIndex && pos >= 0 && pos <= len(g.boardOrder) {
			val := g.boardOrder[g.dragIndex]
			g.boardOrder = append(g.boardOrder[:g.dragIndex], g.boardOrder[g.dragIndex+1:]...)
			if pos > g.dragIndex {
				pos--
			}
			g.boardOrder = append(g.boardOrder[:pos], append([]int{val}, g.boardOrder[pos:]...)...)
		}
	} else {
		g.client.PlaceMinion(g.dragIndex, pos)
	}
}

func (g *Game) getDropPosition(lay ui.GameLayout, mx int) int {
	// Convert screen mx to base x.
	baseMx, _ := screenToBase(mx, 0)

	board := g.client.Board()
	if len(board) == 0 {
		return 0
	}
	for i := range board {
		r := ui.CardRect(lay.Board, i, len(board), lay.CardW, lay.CardH, lay.Gap)
		if baseMx < r.X+r.W/2 {
			return i
		}
	}
	return len(board)
}

func (g *Game) handleDiscoverClick(lay ui.GameLayout, discover []api.Card, mx, my int) {
	// Use board zone vertically centered for discover.
	discoverZone := ui.Rect{
		X: 0,
		Y: lay.Screen.H/2 - lay.CardH/2 - lay.Screen.H*0.03,
		W: lay.Screen.W,
		H: lay.CardH + lay.Screen.H*0.06,
	}
	for i := range discover {
		r := ui.CardRect(discoverZone, i, len(discover), lay.CardW, lay.CardH, lay.Gap)
		if r.Contains(mx, my) {
			g.client.DiscoverPick(i)
			return
		}
	}
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

	phase := g.client.Phase()
	switch phase {
	case game.PhaseWaiting:
		g.drawWaiting(screen)
	case game.PhaseRecruit:
		g.drawRecruit(screen)
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

func (g *Game) drawRecruit(screen *ebiten.Image) {
	lay := ui.CalcGameLayout()

	// Header.
	header := fmt.Sprintf("Turn %d | RECRUIT", g.client.Turn())
	ui.DrawText(screen, g.font, header, lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{100, 200, 100, 255})

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

	// Player stats.
	if p := g.client.Player(); p != nil {
		stats := fmt.Sprintf("HP: %d | Gold: %d/%d | Tier: %d", p.HP, p.Gold, p.MaxGold, p.ShopTier)
		ui.DrawText(
			screen,
			g.font,
			stats,
			lay.BtnRow.W*0.04,
			lay.BtnRow.Y+lay.BtnRow.H*0.15,
			color.RGBA{255, 215, 0, 255},
		)
	}

	// Buttons.
	g.drawButtons(screen, lay)

	// Shop.
	ui.DrawText(
		screen,
		g.font,
		"SHOP",
		lay.Shop.W*0.04,
		lay.Shop.Y+lay.Shop.H*0.02,
		color.RGBA{150, 150, 150, 255},
	)

	shop := g.client.Shop()
	frozen := g.client.IsShopFrozen()
	for i, c := range shop {
		r := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
		if g.isSpell(c) {
			g.cr.DrawShopSpell(screen, c, r)
		} else {
			g.cr.DrawShopMinion(screen, c, r)
		}
		if frozen {
			sr := r.Screen()
			s := ui.ActiveRes.Scale()
			ui.StrokeEllipse(screen,
				float32(sr.X+sr.W/2), float32(sr.Y+sr.H/2),
				float32(sr.W/2), float32(sr.H/2),
				float32(3*s), color.RGBA{80, 160, 255, 255})
		}
	}

	// Board.
	ui.DrawText(screen, g.font, "BOARD", lay.Board.W*0.04, lay.Board.Y+lay.Board.H*0.02, color.RGBA{150, 150, 150, 255})

	board := g.client.Board()
	for i, serverIdx := range g.boardOrder {
		if g.dragging && g.dragFromBoard && i == g.dragIndex {
			continue
		}
		if serverIdx >= 0 && serverIdx < len(board) {
			r := ui.CardRect(lay.Board, i, len(g.boardOrder), lay.CardW, lay.CardH, lay.Gap)
			g.cr.DrawMinion(screen, board[serverIdx], r, 255, 0)
		}
	}

	// Hand.
	ui.DrawText(screen, g.font, "HAND", lay.Hand.W*0.04, lay.Hand.Y+lay.Hand.H*0.02, color.RGBA{150, 150, 150, 255})

	hand := g.client.Hand()
	for i, c := range hand {
		if g.dragging && !g.dragFromBoard && i == g.dragIndex {
			continue
		}
		r := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
		if g.isSpell(c) {
			g.cr.DrawSpellCard(screen, c, r)
		} else {
			g.cr.DrawMinionCard(screen, c, r)
		}
	}

	// Dragged card at cursor — convert screen drag pos to base coords.
	if g.dragging {
		var c api.Card
		if g.dragFromBoard {
			serverIdx := g.boardOrder[g.dragIndex]
			if serverIdx >= 0 && serverIdx < len(board) {
				c = board[serverIdx]
			}
		} else {
			c = hand[g.dragIndex]
		}
		bx, by := screenToBase(g.dragCurrentX, g.dragCurrentY)
		dragRect := ui.Rect{
			X: bx - lay.CardW/2,
			Y: by - lay.CardH/2,
			W: lay.CardW,
			H: lay.CardH,
		}
		if g.dragFromBoard {
			g.cr.DrawMinion(screen, c, dragRect, 255, 0)
		} else if g.isSpell(c) {
			g.cr.DrawSpellCard(screen, c, dragRect)
		} else {
			g.cr.DrawMinionCard(screen, c, dragRect)
		}
	}

	// Hover tooltip (draw before discover so discover covers it).
	if g.hoverCard != nil {
		tooltipY := g.hoverRect.Y - lay.CardH - 8
		if tooltipY < 0 {
			tooltipY = 0
		}
		tooltipRect := ui.Rect{
			X: g.hoverRect.X + g.hoverRect.W/2 - lay.CardW/2,
			Y: tooltipY,
			W: lay.CardW,
			H: lay.CardH,
		}
		if g.isSpell(*g.hoverCard) {
			g.cr.DrawSpellCard(screen, *g.hoverCard, tooltipRect)
		} else {
			g.cr.DrawMinionCard(screen, *g.hoverCard, tooltipRect)
		}
	}

	// Discover overlay.
	if discover := g.client.Discover(); discover != nil {
		g.drawDiscoverOverlay(screen, lay, discover)
	}
}

func (g *Game) drawButtons(screen *ebiten.Image, lay ui.GameLayout) {
	refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
	s := ui.ActiveRes.Scale()
	sw := float32(s)

	// Refresh.
	sr := refresh.Screen()
	vector.FillRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		color.RGBA{60, 60, 90, 255},
		false,
	)
	vector.StrokeRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		sw,
		color.RGBA{100, 100, 140, 255},
		false,
	)
	ui.DrawText(
		screen,
		g.font,
		"Refresh (1g)",
		refresh.X+refresh.W*0.08,
		refresh.Y+refresh.H*0.25,
		color.RGBA{200, 200, 255, 255},
	)

	// Upgrade.
	sr = upgrade.Screen()
	vector.FillRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		color.RGBA{60, 90, 60, 255},
		false,
	)
	vector.StrokeRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		sw,
		color.RGBA{100, 140, 100, 255},
		false,
	)
	if p := g.client.Player(); p != nil {
		ui.DrawText(
			screen,
			g.font,
			fmt.Sprintf("Upgrade (%dg)", p.UpgradeCost),
			upgrade.X+upgrade.W*0.08,
			upgrade.Y+upgrade.H*0.25,
			color.RGBA{200, 255, 200, 255},
		)
	}

	// Freeze.
	sr = freeze.Screen()
	if g.client.IsShopFrozen() {
		vector.FillRect(
			screen,
			float32(sr.X),
			float32(sr.Y),
			float32(sr.W),
			float32(sr.H),
			color.RGBA{40, 120, 200, 255},
			false,
		)
		vector.StrokeRect(
			screen,
			float32(sr.X),
			float32(sr.Y),
			float32(sr.W),
			float32(sr.H),
			sw,
			color.RGBA{80, 160, 255, 255},
			false,
		)
		ui.DrawText(
			screen,
			g.font,
			"Unfreeze",
			freeze.X+freeze.W*0.08,
			freeze.Y+freeze.H*0.25,
			color.RGBA{200, 230, 255, 255},
		)
	} else {
		vector.FillRect(
			screen,
			float32(sr.X),
			float32(sr.Y),
			float32(sr.W),
			float32(sr.H),
			color.RGBA{40, 60, 90, 255},
			false,
		)
		vector.StrokeRect(
			screen,
			float32(sr.X),
			float32(sr.Y),
			float32(sr.W),
			float32(sr.H),
			sw,
			color.RGBA{80, 100, 140, 255},
			false,
		)
		ui.DrawText(
			screen,
			g.font,
			"Freeze",
			freeze.X+freeze.W*0.08,
			freeze.Y+freeze.H*0.25,
			color.RGBA{150, 200, 255, 255},
		)
	}
}

func (g *Game) drawDiscoverOverlay(screen *ebiten.Image, lay ui.GameLayout, discover []api.Card) {
	ui.FillScreen(screen, color.RGBA{0, 0, 0, 160})

	ui.DrawText(screen, g.font, "DISCOVER — Pick a card",
		lay.Screen.W*0.4, lay.Screen.H*0.5-lay.CardH/2-lay.Screen.H*0.04,
		color.RGBA{255, 215, 0, 255})

	discoverZone := ui.Rect{
		X: 0,
		Y: lay.Screen.H/2 - lay.CardH/2 - lay.Screen.H*0.03,
		W: lay.Screen.W,
		H: lay.CardH + lay.Screen.H*0.06,
	}
	for i, c := range discover {
		r := ui.CardRect(discoverZone, i, len(discover), lay.CardW, lay.CardH, lay.Gap)
		if g.isSpell(c) {
			g.cr.DrawSpellCard(screen, c, r)
		} else {
			g.cr.DrawMinionCard(screen, c, r)
		}
	}
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
	ui.DrawText(
		screen,
		g.font,
		"OPPONENT",
		lay.Opponent.W*0.04,
		lay.Opponent.Y+lay.Opponent.H*0.02,
		color.RGBA{255, 120, 120, 255},
	)
	for i, c := range state.OpponentBoard {
		r := ui.CardRect(lay.Opponent, i, len(state.OpponentBoard), lay.CardW, lay.CardH, lay.Gap)
		g.cr.DrawMinion(screen, c, r, 255, 0)
	}

	// Player board.
	ui.DrawText(
		screen,
		g.font,
		"YOUR BOARD",
		lay.Player.W*0.04,
		lay.Player.Y+lay.Player.H*0.02,
		color.RGBA{120, 255, 120, 255},
	)
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
