package scene

import (
	"fmt"
	"image/color"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

// shopPanel handles shop card rendering, buttons, and shop-specific input.
type shopPanel struct {
	client *client.GameClient
	cr     *widget.CardRenderer
	order  []int
}

func (s *shopPanel) syncSize() {
	shop := s.client.Shop()
	if len(s.order) != len(shop) {
		s.order = make([]int, len(shop))
		for i := range s.order {
			s.order[i] = i
		}
	}
}

func (s *shopPanel) handleStartDrag(res ui.Resolution, lay ui.GameLayout, mx, my int, drag *dragState) bool {
	shop := s.client.Shop()
	for i := range shop {
		rect := ui.CardRect(lay.Shop, i, len(shop), lay.CardW, lay.CardH, lay.Gap)
		if rect.Contains(res, mx, my) {
			drag.Start(i, false, true, mx, my)
			return true
		}
	}
	return false
}

func (s *shopPanel) endDrag(res ui.Resolution, lay ui.GameLayout, mx, my int, drag *dragState) {
	dropPad := lay.CardH * 0.4
	shopZone := ui.Rect{X: lay.Shop.X, Y: lay.Shop.Y - dropPad, W: lay.Shop.W, H: lay.Shop.H + 2*dropPad}

	if shopZone.Contains(res, mx, my) {
		pos := s.getDropPosition(res, lay, mx)
		if pos != drag.index && pos >= 0 && pos <= len(s.order) {
			val := s.order[drag.index]
			s.order = append(s.order[:drag.index], s.order[drag.index+1:]...)
			if pos > drag.index {
				pos--
			}
			s.order = append(s.order[:pos], append([]int{val}, s.order[pos:]...)...)
		}
		return
	}

	_, baseY := res.ScreenToBase(mx, my)
	if baseY > lay.Shop.Y+lay.Shop.H+dropPad {
		if err := s.client.BuyCard(s.order[drag.index]); err != nil {
			slog.Error("buy card", "error", err)
		}
	}
}

func (s *shopPanel) getDropPosition(res ui.Resolution, lay ui.GameLayout, mx int) int {
	baseMx, _ := res.ScreenToBase(mx, 0)
	shop := s.client.Shop()
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

func (s *shopPanel) handleButtonClick(res ui.Resolution, lay ui.GameLayout, mx, my int) bool {
	refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
	switch {
	case refresh.Contains(res, mx, my):
		if err := s.client.RefreshShop(); err != nil {
			slog.Error("refresh shop", "error", err)
		}
		return true
	case upgrade.Contains(res, mx, my):
		if p := s.client.Player(); p == nil || p.ShopTier >= game.Tier6 {
			return true
		}
		if err := s.client.UpgradeShop(); err != nil {
			slog.Error("upgrade shop", "error", err)
		}
		return true
	case freeze.Contains(res, mx, my):
		if err := s.client.FreezeShop(); err != nil {
			slog.Error("freeze shop", "error", err)
		}
		return true
	}
	return false
}

func (s *shopPanel) updateHover(res ui.Resolution, lay ui.GameLayout, mx, my int) (*api.Card, ui.Rect, bool) {
	shop := s.client.Shop()
	for i, serverIdx := range s.order {
		rect := ui.CardRect(lay.Shop, i, len(s.order), lay.CardW, lay.CardH, lay.Gap)
		if rect.Contains(res, mx, my) {
			c := shop[serverIdx]
			return &c, rect, true
		}
	}
	return nil, ui.Rect{}, false
}

func (s *shopPanel) drawCards(screen *ebiten.Image, res ui.Resolution, lay ui.GameLayout, drag *dragState) {
	shop := s.client.Shop()
	frozen := s.client.IsShopFrozen()

	for i, serverIdx := range s.order {
		if drag.active && drag.fromShop && i == drag.index {
			continue
		}
		if serverIdx >= len(shop) {
			continue
		}
		c := shop[serverIdx]
		rect := ui.CardRect(lay.Shop, i, len(s.order), lay.CardW, lay.CardH, lay.Gap)
		s.cr.DrawShopCard(screen, c, rect)
		if frozen {
			sr := rect.Screen(res)
			sc := res.Scale()
			ui.StrokeEllipse(screen,
				float32(sr.X+sr.W/2), float32(sr.Y+sr.H/2),
				float32(sr.W/2), float32(sr.H/2),
				float32(3*sc), color.RGBA{80, 160, 255, 255})
		}
	}
}

func (s *shopPanel) drawButtons(screen *ebiten.Image, res ui.Resolution, font *text.GoTextFace, lay ui.GameLayout) {
	refresh, upgrade, freeze := ui.ButtonRects(lay.BtnRow)
	sc := res.Scale()
	sw := float32(sc)

	// Refresh.
	sr := refresh.Screen(res)
	vector.FillRect(
		screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
		color.RGBA{60, 60, 90, 255}, false,
	)
	vector.StrokeRect(
		screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
		sw, color.RGBA{100, 100, 140, 255}, false,
	)
	refreshCost := game.ShopRefreshCost
	if p := s.client.Player(); p != nil {
		refreshCost = p.RefreshCost
	}
	ui.DrawText(screen, res, font, fmt.Sprintf("Refresh (%dg)", refreshCost),
		refresh.X+refresh.W*0.08, refresh.Y+refresh.H*0.25,
		color.RGBA{200, 200, 255, 255})

	// Upgrade.
	sr = upgrade.Screen(res)
	if p := s.client.Player(); p != nil && p.ShopTier < game.Tier6 {
		vector.FillRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			color.RGBA{60, 90, 60, 255}, false,
		)
		vector.StrokeRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			sw, color.RGBA{100, 140, 100, 255}, false,
		)
		ui.DrawText(screen, res, font,
			fmt.Sprintf("Upgrade (%dg)", p.UpgradeCost),
			upgrade.X+upgrade.W*0.08, upgrade.Y+upgrade.H*0.25,
			color.RGBA{200, 255, 200, 255})
	}

	// Freeze.
	sr = freeze.Screen(res)
	if s.client.IsShopFrozen() {
		vector.FillRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			color.RGBA{40, 120, 200, 255}, false,
		)
		vector.StrokeRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			sw, color.RGBA{80, 160, 255, 255}, false,
		)
		ui.DrawText(screen, res, font, "Unfreeze",
			freeze.X+freeze.W*0.08, freeze.Y+freeze.H*0.25,
			color.RGBA{200, 230, 255, 255})
	} else {
		vector.FillRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			color.RGBA{40, 60, 90, 255}, false,
		)
		vector.StrokeRect(
			screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H),
			sw, color.RGBA{80, 100, 140, 255}, false,
		)
		ui.DrawText(screen, res, font, "Freeze",
			freeze.X+freeze.W*0.08, freeze.Y+freeze.H*0.25,
			color.RGBA{150, 200, 255, 255})
	}
}
