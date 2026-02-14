package scene

import (
	"fmt"
	"image/color"
	"log/slog"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// recruitPhase handles all input and drawing during the recruit phase.
type recruitPhase struct {
	client *client.GameClient
	cr     *widget.CardRenderer
	shop   *shopPanel

	drag      dragState
	hoverCard *api.Card
	hoverRect ui.Rect

	boardOrder []int
}

// ReorderCards sends the local board/shop order to the server.
func (r *recruitPhase) ReorderCards() {
	if err := r.client.ReorderCards(r.boardOrder, r.shop.order); err != nil {
		slog.Error("reorder cards", "error", err)
	}
}

// syncSizes keeps local orders in sync with server array sizes.
func (r *recruitPhase) syncSizes() {
	board := r.client.Board()
	if len(r.boardOrder) != len(board) {
		r.boardOrder = make([]int, len(board))
		for i := range r.boardOrder {
			r.boardOrder[i] = i
		}
	}
	r.shop.syncSize()
}

// Update processes recruit-phase input.
func (r *recruitPhase) Update(lay ui.GameLayout) error {
	r.syncSizes()

	mx, my := ebiten.CursorPosition()

	r.hoverCard = nil

	// Discover overlay blocks all other input.
	if discover := r.client.Discovers(); discover != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			r.handleDiscoverClick(lay, discover, mx, my)
		}
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if r.handleStartDrag(lay, mx, my) {
			return nil
		}
	}

	if r.drag.active && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		r.drag.cursorX = mx
		r.drag.cursorY = my
		return nil
	}

	if r.drag.active && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		r.endDrag(lay, mx, my)
		return nil
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if r.shop.handleButtonClick(lay, mx, my) {
			return nil
		}
	}

	if !r.drag.active {
		r.updateHover(lay, mx, my)
	}

	return nil
}

// handleStartDrag checks if the click starts a drag from board, hand, or shop.
func (r *recruitPhase) handleStartDrag(lay ui.GameLayout, mx, my int) bool {
	for i := range r.boardOrder {
		rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
		if rect.Contains(mx, my) {
			r.drag.Start(i, true, false, mx, my)
			return true
		}
	}

	hand := r.client.Hand()
	for i, c := range hand {
		rect := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
		if !rect.Contains(mx, my) {
			continue
		}
		if t := r.cr.Cards.ByTemplateID(c.TemplateID); t != nil && t.Kind() == game.CardKindSpell {
			if err := r.client.PlaySpell(i); err != nil {
				slog.Error("play spell", "error", err)
			}
			return true
		}
		r.drag.Start(i, false, false, mx, my)
		return true
	}

	return r.shop.handleStartDrag(lay, mx, my, &r.drag)
}

// updateHover detects card hover for tooltip display.
func (r *recruitPhase) updateHover(lay ui.GameLayout, mx, my int) {
	if card, rect, ok := r.shop.updateHover(lay, mx, my); ok {
		r.hoverCard = card
		r.hoverRect = rect
		return
	}

	board := r.client.Board()
	for i, serverIdx := range r.boardOrder {
		if serverIdx < 0 || serverIdx >= len(board) {
			continue
		}
		rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
		if rect.Contains(mx, my) {
			c := board[serverIdx]
			r.hoverCard = &c
			r.hoverRect = rect
			return
		}
	}
}

func (r *recruitPhase) endDrag(lay ui.GameLayout, mx, my int) {
	defer r.drag.Reset()

	switch {
	case r.drag.fromShop:
		r.shop.endDrag(lay, mx, my, &r.drag)
	case r.drag.fromBoard:
		r.endBoardDrag(lay, mx, my)
	default:
		r.endHandDrag(lay, mx, my)
	}
}

func (r *recruitPhase) endBoardDrag(lay ui.GameLayout, mx, my int) {
	dropPad := lay.CardH * 0.4
	shopZone := ui.Rect{X: lay.Shop.X, Y: lay.Shop.Y - dropPad, W: lay.Shop.W, H: lay.Shop.H + 2*dropPad}

	if shopZone.Contains(mx, my) {
		if err := r.client.SellMinion(r.boardOrder[r.drag.index]); err != nil {
			slog.Error("sell minion", "error", err)
		}
		return
	}

	boardZone := ui.Rect{X: lay.Board.X, Y: lay.Board.Y - dropPad, W: lay.Board.W, H: lay.Board.H + 2*dropPad}
	if !boardZone.Contains(mx, my) {
		return
	}

	pos := r.getBoardDropPosition(lay, mx)
	if pos == r.drag.index || pos < 0 || pos > len(r.boardOrder) {
		return
	}

	val := r.boardOrder[r.drag.index]
	r.boardOrder = append(r.boardOrder[:r.drag.index], r.boardOrder[r.drag.index+1:]...)
	if pos > r.drag.index {
		pos--
	}
	r.boardOrder = append(r.boardOrder[:pos], append([]int{val}, r.boardOrder[pos:]...)...)
}

func (r *recruitPhase) endHandDrag(lay ui.GameLayout, mx, my int) {
	dropPad := lay.CardH * 0.4
	boardZone := ui.Rect{X: lay.Board.X, Y: lay.Board.Y - dropPad, W: lay.Board.W, H: lay.Board.H + 2*dropPad}

	if !boardZone.Contains(mx, my) {
		return
	}

	pos := r.getBoardDropPosition(lay, mx)
	if err := r.client.PlaceMinion(r.drag.index, pos); err != nil {
		slog.Error("place minion", "error", err)
	}
}

func (r *recruitPhase) getBoardDropPosition(lay ui.GameLayout, mx int) int {
	baseMx, _ := screenToBase(mx, 0)
	board := r.client.Board()
	if len(board) == 0 {
		return 0
	}
	for i := range board {
		rect := ui.CardRect(lay.Board, i, len(board), lay.CardW, lay.CardH, lay.Gap)
		if baseMx < rect.X+rect.W/2 {
			return i
		}
	}
	return len(board)
}

func (r *recruitPhase) handleDiscoverClick(lay ui.GameLayout, discover []api.Card, mx, my int) {
	discoverZone := ui.Rect{
		X: 0,
		Y: lay.Screen.H/2 - lay.CardH/2 - lay.Screen.H*0.03,
		W: lay.Screen.W,
		H: lay.CardH + lay.Screen.H*0.06,
	}
	for i := range discover {
		rect := ui.CardRect(discoverZone, i, len(discover), lay.CardW, lay.CardH, lay.Gap)
		if rect.Contains(mx, my) {
			if err := r.client.DiscoverPick(i); err != nil {
				slog.Error("discover pick", "error", err)
			}
			return
		}
	}
}

// Draw renders the recruit phase.
func (r *recruitPhase) Draw(
	screen *ebiten.Image,
	font *text.GoTextFace,
	lay ui.GameLayout,
	turn int,
	timeRemaining time.Duration,
) {
	r.drawHeader(screen, font, lay, turn, timeRemaining)
	r.drawPlayerStats(screen, font, lay)
	r.shop.drawButtons(screen, font, lay)

	ui.DrawText(screen, font, "SHOP",
		lay.Shop.X+lay.Shop.W*0.04, lay.Shop.Y+lay.Shop.H*0.02,
		color.RGBA{150, 150, 150, 255})
	r.shop.drawCards(screen, lay, &r.drag)

	ui.DrawText(screen, font, "BOARD",
		lay.Board.X+lay.Board.W*0.04, lay.Board.Y+lay.Board.H*0.02,
		color.RGBA{150, 150, 150, 255})
	r.drawBoardCards(screen, lay)
	r.drawMagnetizeHighlight(screen, lay)

	ui.DrawText(screen, font, "HAND",
		lay.Hand.X+lay.Hand.W*0.04, lay.Hand.Y+lay.Hand.H*0.02,
		color.RGBA{150, 150, 150, 255})
	r.drawHandCards(screen, lay)

	r.drawDraggedCard(screen, lay)
	r.drawHoverTooltip(screen, lay)

	if discover := r.client.Discovers(); discover != nil {
		r.drawDiscoverOverlay(screen, font, lay, discover)
	}
}

func (r *recruitPhase) drawHeader(
	screen *ebiten.Image, font *text.GoTextFace, lay ui.GameLayout, turn int, timeRemaining time.Duration,
) {
	header := fmt.Sprintf("Turn %d", turn)
	ui.DrawText(screen, font, header,
		lay.Header.X+lay.Header.W*0.04, lay.Header.H*0.5,
		color.RGBA{200, 200, 200, 255})

	secs := int(timeRemaining.Seconds())
	timer := fmt.Sprintf("%d:%02d", secs/60, secs%60)
	ui.DrawText(screen, font, timer,
		lay.Header.X+lay.Header.W*0.9, lay.Header.H*0.5,
		color.RGBA{255, 255, 255, 255})

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
}

func (r *recruitPhase) drawPlayerStats(screen *ebiten.Image, font *text.GoTextFace, lay ui.GameLayout) {
	p := r.client.Player()
	if p == nil {
		return
	}
	stats := fmt.Sprintf("Gold: %d/%d | Tier: %d", p.Gold, p.MaxGold, p.ShopTier)
	ui.DrawText(screen, font, stats,
		lay.BtnRow.X+lay.BtnRow.W*0.04, lay.BtnRow.Y+lay.BtnRow.H*0.15,
		color.RGBA{255, 215, 0, 255})
}

func (r *recruitPhase) drawBoardCards(screen *ebiten.Image, lay ui.GameLayout) {
	board := r.client.Board()
	for i, serverIdx := range r.boardOrder {
		if r.drag.active && r.drag.fromBoard && i == r.drag.index {
			continue
		}
		if serverIdx >= 0 && serverIdx < len(board) {
			rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
			r.cr.DrawMinion(screen, board[serverIdx], rect, 255, 0)
		}
	}
}

func (r *recruitPhase) drawHandCards(screen *ebiten.Image, lay ui.GameLayout) {
	hand := r.client.Hand()
	for i, c := range hand {
		if r.drag.active && !r.drag.fromBoard && !r.drag.fromShop && i == r.drag.index {
			continue
		}
		rect := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
		r.cr.DrawHandCard(screen, c, rect)
	}
}

func (r *recruitPhase) drawDraggedCard(screen *ebiten.Image, lay ui.GameLayout) {
	if !r.drag.active {
		return
	}

	shop := r.client.Shop()
	board := r.client.Board()
	hand := r.client.Hand()

	var c api.Card
	switch {
	case r.drag.fromShop:
		if idx := r.shop.order[r.drag.index]; idx < len(shop) {
			c = shop[idx]
		}
	case r.drag.fromBoard:
		serverIdx := r.boardOrder[r.drag.index]
		if serverIdx >= 0 && serverIdx < len(board) {
			c = board[serverIdx]
		}
	default:
		c = hand[r.drag.index]
	}

	bx, by := screenToBase(r.drag.cursorX, r.drag.cursorY)
	dragRect := ui.Rect{
		X: bx - lay.CardW/2,
		Y: by - lay.CardH/2,
		W: lay.CardW,
		H: lay.CardH,
	}

	switch {
	case r.drag.fromShop:
		r.cr.DrawShopCard(screen, c, dragRect)
	case r.drag.fromBoard:
		r.cr.DrawMinion(screen, c, dragRect, 255, 0)
	default:
		r.cr.DrawHandCard(screen, c, dragRect)
	}
}

func (r *recruitPhase) drawHoverTooltip(screen *ebiten.Image, lay ui.GameLayout) {
	if r.hoverCard == nil {
		return
	}

	tooltipY := r.hoverRect.Y - lay.CardH - 8
	if tooltipY < 0 {
		tooltipY = 0
	}
	tooltipRect := ui.Rect{
		X: r.hoverRect.X + r.hoverRect.W/2 - lay.CardW/2,
		Y: tooltipY,
		W: lay.CardW,
		H: lay.CardH,
	}
	r.cr.DrawHandCard(screen, *r.hoverCard, tooltipRect)
}

func (r *recruitPhase) drawMagnetizeHighlight(screen *ebiten.Image, lay ui.GameLayout) {
	if !r.drag.fromHand() {
		return
	}
	hand := r.client.Hand()
	if r.drag.index < 0 || r.drag.index >= len(hand) {
		return
	}
	if !hand[r.drag.index].Keywords.Has(game.KeywordMagnetic) {
		return
	}

	board := r.client.Board()
	for i, serverIdx := range r.boardOrder {
		if serverIdx < 0 || serverIdx >= len(board) {
			continue
		}
		bc := board[serverIdx]
		if bc.Tribe != game.TribeMech && bc.Tribe != game.TribeAll {
			continue
		}
		rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
		sr := rect.Screen()
		lineX := float32(sr.X)
		vector.StrokeLine(screen,
			lineX, float32(sr.Y), lineX, float32(sr.Y+sr.H),
			2*float32(ui.ActiveRes.Scale()),
			color.RGBA{0, 150, 255, 200}, false)
	}
}

func (r *recruitPhase) drawDiscoverOverlay(
	screen *ebiten.Image,
	font *text.GoTextFace,
	lay ui.GameLayout,
	discover []api.Card,
) {
	ui.FillScreen(screen, color.RGBA{0, 0, 0, 160})

	ui.DrawText(screen, font, "DISCOVER — Pick a card",
		lay.Screen.W*0.4, lay.Screen.H*0.5-lay.CardH/2-lay.Screen.H*0.04,
		color.RGBA{255, 215, 0, 255})

	discoverZone := ui.Rect{
		X: 0,
		Y: lay.Screen.H/2 - lay.CardH/2 - lay.Screen.H*0.03,
		W: lay.Screen.W,
		H: lay.CardH + lay.Screen.H*0.06,
	}
	for i, c := range discover {
		rect := ui.CardRect(discoverZone, i, len(discover), lay.CardW, lay.CardH, lay.Gap)
		r.cr.DrawHandCard(screen, c, rect)
	}
}
