package scene

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// recruitPhase handles all input and drawing during the recruit phase.
type recruitPhase struct {
	client *client.GameClient
	cr     *widget.CardRenderer

	drag      dragState
	hoverCard *api.Card
	hoverRect ui.Rect

	boardOrder []int
	shopOrder  []int
}

// SyncOrders sends the local board/shop order to the server.
func (r *recruitPhase) SyncOrders() {
	r.client.SyncBoards(r.boardOrder, r.shopOrder)
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
	shop := r.client.Shop()
	if len(r.shopOrder) != len(shop) {
		r.shopOrder = make([]int, len(shop))
		for i := range r.shopOrder {
			r.shopOrder[i] = i
		}
	}
}

func (r *recruitPhase) isSpell(c api.Card) bool {
	t := r.cr.Cards.ByTemplateID(c.TemplateID)
	return t != nil && t.IsSpell()
}

// Update processes recruit-phase input. Returns nil always.
func (r *recruitPhase) Update() error {
	r.syncSizes()

	lay := ui.CalcGameLayout()
	mx, my := ebiten.CursorPosition()

	r.hoverCard = nil

	// Discover overlay blocks all other input.
	if discover := r.client.Discover(); discover != nil {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			r.handleDiscoverClick(lay, discover, mx, my)
		}
		return nil
	}

	// Start drag.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// From board.
		for i := range r.boardOrder {
			rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
			if rect.Contains(mx, my) {
				r.drag.Start(i, true, false, mx, my)
				return nil
			}
		}
		// From hand.
		hand := r.client.Hand()
		for i, c := range hand {
			rect := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
			if rect.Contains(mx, my) {
				if r.isSpell(c) {
					r.client.PlaySpell(i)
					return nil
				}
				r.drag.Start(i, false, false, mx, my)
				return nil
			}
		}
		// From shop.
		shop := r.client.Shop()
		for i := range shop {
			rect := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
			if rect.Contains(mx, my) {
				r.drag.Start(i, false, true, mx, my)
				return nil
			}
		}
	}

	// Update drag position.
	if r.drag.active && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		r.drag.cursorX = mx
		r.drag.cursorY = my
		return nil
	}

	// End drag.
	if r.drag.active && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		r.endDrag(lay, mx, my)
		return nil
	}

	// Click buttons.
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
		switch {
		case refresh.Contains(mx, my):
			r.client.RefreshShop()
			return nil
		case upgrade.Contains(mx, my):
			r.client.UpgradeShop()
			return nil
		case freeze.Contains(mx, my):
			r.client.FreezeShop()
			return nil
		}
	}

	// Hover detection (skip while dragging).
	if !r.drag.active {
		shop := r.client.Shop()
		for i, serverIdx := range r.shopOrder {
			rect := ui.CardRect(lay.Shop, i, len(r.shopOrder), lay.CardW, lay.CardH, lay.Gap)
			if rect.Contains(mx, my) {
				c := shop[serverIdx]
				r.hoverCard = &c
				r.hoverRect = rect
				break
			}
		}
		if r.hoverCard == nil {
			board := r.client.Board()
			for i, serverIdx := range r.boardOrder {
				if serverIdx >= 0 && serverIdx < len(board) {
					rect := ui.CardRect(lay.Board, i, len(r.boardOrder), lay.CardW, lay.CardH, lay.Gap)
					if rect.Contains(mx, my) {
						c := board[serverIdx]
						r.hoverCard = &c
						r.hoverRect = rect
						break
					}
				}
			}
		}
	}

	return nil
}

func (r *recruitPhase) endDrag(lay ui.GameLayout, mx, my int) {
	defer r.drag.Reset()

	dropPad := lay.CardH * 0.4

	// Shop card drag: reorder if dropped on shop, buy if dropped below.
	if r.drag.fromShop {
		shopZone := ui.Rect{X: lay.Shop.X, Y: lay.Shop.Y - dropPad, W: lay.Shop.W, H: lay.Shop.H + 2*dropPad}
		if shopZone.Contains(mx, my) {
			pos := r.getShopDropPosition(lay, mx)
			if pos != r.drag.index && pos >= 0 && pos <= len(r.shopOrder) {
				val := r.shopOrder[r.drag.index]
				r.shopOrder = append(r.shopOrder[:r.drag.index], r.shopOrder[r.drag.index+1:]...)
				if pos > r.drag.index {
					pos--
				}
				r.shopOrder = append(r.shopOrder[:pos], append([]int{val}, r.shopOrder[pos:]...)...)
			}
		} else {
			_, baseY := screenToBase(mx, my)
			if baseY > lay.Shop.Y+lay.Shop.H+dropPad {
				r.client.BuyCard(r.shopOrder[r.drag.index])
			}
		}
		return
	}

	if r.drag.fromBoard {
		shopZone := ui.Rect{X: lay.Shop.X, Y: lay.Shop.Y - dropPad, W: lay.Shop.W, H: lay.Shop.H + 2*dropPad}
		if shopZone.Contains(mx, my) {
			r.client.SellMinion(r.boardOrder[r.drag.index])
			return
		}
	}

	boardZone := ui.Rect{X: lay.Board.X, Y: lay.Board.Y - dropPad, W: lay.Board.W, H: lay.Board.H + 2*dropPad}
	if !boardZone.Contains(mx, my) {
		return
	}

	pos := r.getBoardDropPosition(lay, mx)
	if r.drag.fromBoard {
		if pos != r.drag.index && pos >= 0 && pos <= len(r.boardOrder) {
			val := r.boardOrder[r.drag.index]
			r.boardOrder = append(r.boardOrder[:r.drag.index], r.boardOrder[r.drag.index+1:]...)
			if pos > r.drag.index {
				pos--
			}
			r.boardOrder = append(r.boardOrder[:pos], append([]int{val}, r.boardOrder[pos:]...)...)
		}
	} else {
		r.client.PlaceMinion(r.drag.index, pos)
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

func (r *recruitPhase) getShopDropPosition(lay ui.GameLayout, mx int) int {
	baseMx, _ := screenToBase(mx, 0)
	shop := r.client.Shop()
	if len(shop) == 0 {
		return 0
	}
	for i := range shop {
		rect := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
		if baseMx < rect.X+rect.W/2 {
			return i
		}
	}
	return len(shop)
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
			r.client.DiscoverPick(i)
			return
		}
	}
}

// Draw renders the recruit phase.
func (r *recruitPhase) Draw(screen *ebiten.Image, font *text.GoTextFace, turn int, timeRemaining int64) {
	lay := ui.CalcGameLayout()

	// Header.
	header := fmt.Sprintf("Turn %d | RECRUIT", turn)
	ui.DrawText(screen, font, header, lay.Header.W*0.04, lay.Header.H*0.5, color.RGBA{100, 200, 100, 255})

	timer := fmt.Sprintf("%d:%02d", timeRemaining/60, timeRemaining%60)
	ui.DrawText(screen, font, timer, lay.Header.W*0.9, lay.Header.H*0.5, color.RGBA{255, 255, 255, 255})

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
	if p := r.client.Player(); p != nil {
		stats := fmt.Sprintf("HP: %d | Gold: %d/%d | Tier: %d", p.HP, p.Gold, p.MaxGold, p.ShopTier)
		ui.DrawText(screen, font, stats, lay.BtnRow.W*0.04, lay.BtnRow.Y+lay.BtnRow.H*0.15, color.RGBA{255, 215, 0, 255})
	}

	// Buttons.
	r.drawButtons(screen, font, lay)

	// Shop.
	ui.DrawText(screen, font, "SHOP", lay.Shop.W*0.04, lay.Shop.Y+lay.Shop.H*0.02, color.RGBA{150, 150, 150, 255})

	shop := r.client.Shop()
	frozen := r.client.IsShopFrozen()
	for i, serverIdx := range r.shopOrder {
		if r.drag.active && r.drag.fromShop && i == r.drag.index {
			continue
		}
		if serverIdx >= len(shop) {
			continue
		}
		c := shop[serverIdx]
		rect := ui.CardRect(lay.Shop, i, len(r.shopOrder), lay.CardW, lay.CardH, lay.Gap)
		if r.isSpell(c) {
			r.cr.DrawShopSpell(screen, c, rect)
		} else {
			r.cr.DrawShopMinion(screen, c, rect)
		}
		if frozen {
			sr := rect.Screen()
			s := ui.ActiveRes.Scale()
			ui.StrokeEllipse(screen,
				float32(sr.X+sr.W/2), float32(sr.Y+sr.H/2),
				float32(sr.W/2), float32(sr.H/2),
				float32(3*s), color.RGBA{80, 160, 255, 255})
		}
	}

	// Board.
	ui.DrawText(screen, font, "BOARD", lay.Board.W*0.04, lay.Board.Y+lay.Board.H*0.02, color.RGBA{150, 150, 150, 255})

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

	// Hand.
	ui.DrawText(screen, font, "HAND", lay.Hand.W*0.04, lay.Hand.Y+lay.Hand.H*0.02, color.RGBA{150, 150, 150, 255})

	hand := r.client.Hand()
	for i, c := range hand {
		if r.drag.active && !r.drag.fromBoard && !r.drag.fromShop && i == r.drag.index {
			continue
		}
		rect := ui.CardRect(lay.Hand, i, len(hand), lay.CardW, lay.CardH, lay.Gap)
		if r.isSpell(c) {
			r.cr.DrawSpellCard(screen, c, rect)
		} else {
			r.cr.DrawMinionCard(screen, c, rect)
		}
	}

	// Dragged card at cursor.
	if r.drag.active {
		var c api.Card
		if r.drag.fromShop {
			if idx := r.shopOrder[r.drag.index]; idx < len(shop) {
				c = shop[idx]
			}
		} else if r.drag.fromBoard {
			serverIdx := r.boardOrder[r.drag.index]
			if serverIdx >= 0 && serverIdx < len(board) {
				c = board[serverIdx]
			}
		} else {
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
			if r.isSpell(c) {
				r.cr.DrawShopSpell(screen, c, dragRect)
			} else {
				r.cr.DrawShopMinion(screen, c, dragRect)
			}
		case r.drag.fromBoard:
			r.cr.DrawMinion(screen, c, dragRect, 255, 0)
		case r.isSpell(c):
			r.cr.DrawSpellCard(screen, c, dragRect)
		default:
			r.cr.DrawMinionCard(screen, c, dragRect)
		}
	}

	// Hover tooltip.
	if r.hoverCard != nil {
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
		if r.isSpell(*r.hoverCard) {
			r.cr.DrawSpellCard(screen, *r.hoverCard, tooltipRect)
		} else {
			r.cr.DrawMinionCard(screen, *r.hoverCard, tooltipRect)
		}
	}

	// Discover overlay.
	if discover := r.client.Discover(); discover != nil {
		r.drawDiscoverOverlay(screen, font, lay, discover)
	}
}

func (r *recruitPhase) drawButtons(screen *ebiten.Image, font *text.GoTextFace, lay ui.GameLayout) {
	refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
	s := ui.ActiveRes.Scale()
	sw := float32(s)

	// Refresh.
	sr := refresh.Screen()
	vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), color.RGBA{60, 60, 90, 255}, false)
	vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, color.RGBA{100, 100, 140, 255}, false)
	ui.DrawText(screen, font, "Refresh (1g)", refresh.X+refresh.W*0.08, refresh.Y+refresh.H*0.25, color.RGBA{200, 200, 255, 255})

	// Upgrade.
	sr = upgrade.Screen()
	vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), color.RGBA{60, 90, 60, 255}, false)
	vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, color.RGBA{100, 140, 100, 255}, false)
	if p := r.client.Player(); p != nil {
		ui.DrawText(screen, font, fmt.Sprintf("Upgrade (%dg)", p.UpgradeCost), upgrade.X+upgrade.W*0.08, upgrade.Y+upgrade.H*0.25, color.RGBA{200, 255, 200, 255})
	}

	// Freeze.
	sr = freeze.Screen()
	if r.client.IsShopFrozen() {
		vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), color.RGBA{40, 120, 200, 255}, false)
		vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, color.RGBA{80, 160, 255, 255}, false)
		ui.DrawText(screen, font, "Unfreeze", freeze.X+freeze.W*0.08, freeze.Y+freeze.H*0.25, color.RGBA{200, 230, 255, 255})
	} else {
		vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), color.RGBA{40, 60, 90, 255}, false)
		vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, color.RGBA{80, 100, 140, 255}, false)
		ui.DrawText(screen, font, "Freeze", freeze.X+freeze.W*0.08, freeze.Y+freeze.H*0.25, color.RGBA{150, 200, 255, 255})
	}
}

func (r *recruitPhase) drawDiscoverOverlay(screen *ebiten.Image, font *text.GoTextFace, lay ui.GameLayout, discover []api.Card) {
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
		if r.isSpell(c) {
			r.cr.DrawSpellCard(screen, c, rect)
		} else {
			r.cr.DrawMinionCard(screen, c, rect)
		}
	}
}
