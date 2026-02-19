package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

const hoverDelay = 0.2 // seconds before hover tooltip appears

// hoverTooltip tracks hover state with a delay timer and draws a card tooltip.
type hoverTooltip struct {
	card     *api.Card
	rect     ui.Rect
	prevRect ui.Rect
	elapsed  float64
	ready    bool
}

// Update sets the currently hovered card and accumulates the delay timer.
// Pass nil card when nothing is hovered.
func (h *hoverTooltip) Update(card *api.Card, rect ui.Rect) {
	h.card = card
	h.rect = rect

	if card == nil || rect != h.prevRect {
		h.elapsed = 0
		h.ready = false
	} else {
		h.elapsed += 1.0 / float64(ebiten.TPS())
		if h.elapsed >= hoverDelay {
			h.ready = true
		}
	}
	h.prevRect = rect
}

// Draw renders the hover tooltip adjacent to the hovered card.
func (h *hoverTooltip) Draw(screen *ebiten.Image, lay ui.GameLayout, cr *widget.CardRenderer) {
	if !h.ready || h.card == nil {
		return
	}

	t := cr.Cards.ByTemplateID(h.card.Template)
	if t == nil {
		return
	}

	tipW := lay.CardW * 2.0
	tipH := lay.CardH * 2.0

	// Prefer right side; fall back to left if it would overflow.
	tipX := h.rect.X + h.rect.W + lay.Gap
	if tipX+tipW > ui.BaseWidth {
		tipX = h.rect.X - lay.Gap - tipW
	}

	// Vertically center on the hovered card, clamp to screen.
	tipY := h.rect.Y + h.rect.H/2 - tipH/2
	if tipY < 0 {
		tipY = 0
	}
	if tipY+tipH > ui.BaseHeight {
		tipY = ui.BaseHeight - tipH
	}

	cr.DrawCard(screen, t, ui.Rect{X: tipX, Y: tipY, W: tipW, H: tipH})
}
