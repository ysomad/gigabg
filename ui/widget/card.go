package widget

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
)

// CardRenderer draws cards in different contexts.
type CardRenderer struct {
	Cards *cards.Cards
	Font  *text.GoTextFace
}

// DrawMinionCard renders a hand minion card (rectangle) with full detail.
func (r *CardRenderer) DrawMinionCard(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawRectBase(screen, rect, color.RGBA{40, 40, 60, 255}, c.IsGolden, false, 255)

	name, desc, tribe := r.cardInfo(c)
	t := r.Cards.ByTemplateID(c.TemplateID)

	// Name (top-left).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.04, rect.Y+rect.H*0.04, color.White)

	// Tier (top-right).
	if t != nil && t.Tier.IsValid() {
		ui.DrawText(screen, r.Font, fmt.Sprintf("T%d", t.Tier),
			rect.Right()-rect.W*0.22, rect.Y+rect.H*0.04,
			color.RGBA{180, 180, 180, 255})
	}

	// Description (center).
	ui.DrawText(screen, r.Font, desc, rect.X+rect.W*0.04, rect.Y+rect.H*0.30, color.RGBA{180, 180, 180, 255})

	// Tribe (bottom-center).
	ui.DrawText(screen, r.Font, tribe,
		rect.X+rect.W*0.3, rect.Bottom()-rect.H*0.25,
		color.RGBA{150, 150, 200, 255})

	// Attack (bottom-left, yellow).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Attack),
		rect.X+rect.W*0.04, rect.Bottom()-rect.H*0.12,
		color.RGBA{255, 215, 0, 255})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Health),
		rect.Right()-rect.W*0.15, rect.Bottom()-rect.H*0.12,
		color.RGBA{255, 80, 80, 255})
}

// DrawSpellCard renders a hand spell card (rectangle) with full detail.
func (r *CardRenderer) DrawSpellCard(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawRectBase(screen, rect, color.RGBA{80, 40, 100, 255}, c.IsGolden, true, 255)

	name, desc, _ := r.cardInfo(c)

	// Name (top-left).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.04, rect.Y+rect.H*0.04, color.White)

	// Description (center).
	ui.DrawText(screen, r.Font, desc, rect.X+rect.W*0.04, rect.Y+rect.H*0.30, color.RGBA{180, 180, 180, 255})

	// SPELL label (bottom-center).
	ui.DrawText(screen, r.Font, "SPELL",
		rect.X+rect.W*0.30, rect.Bottom()-rect.H*0.12,
		color.RGBA{200, 150, 255, 255})
}

// DrawShopMinion renders a shop minion (portrait ellipse) with tier, stats, and image placeholder.
func (r *CardRenderer) DrawShopMinion(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawEllipseBase(screen, rect, color.RGBA{35, 40, 70, 255}, c.IsGolden, false, 255)

	t := r.Cards.ByTemplateID(c.TemplateID)

	// Image placeholder (center).
	ui.DrawText(screen, r.Font, c.TemplateID, rect.X+rect.W*0.15, rect.Y+rect.H*0.42, color.RGBA{100, 100, 120, 255})

	// Tier (top, gold).
	if t != nil && t.Tier.IsValid() {
		ui.DrawText(screen, r.Font, fmt.Sprintf("T%d", t.Tier),
			rect.X+rect.W*0.42, rect.Y+rect.H*0.15,
			color.RGBA{255, 215, 0, 255})
	}

	// Keywords (below placeholder).
	if t != nil {
		kwY := rect.Y + rect.H*0.58
		for _, kw := range t.Keywords.List() {
			ui.DrawText(screen, r.Font, kw.String(),
				rect.X+rect.W*0.20, kwY,
				color.RGBA{180, 220, 140, 255})
			kwY += rect.H * 0.08
		}
	}

	// Attack (bottom-left, yellow).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Attack),
		rect.X+rect.W*0.20, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 215, 0, 255})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Health),
		rect.Right()-rect.W*0.30, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 80, 80, 255})
}

// DrawShopSpell renders a shop spell (portrait ellipse) with spell info.
func (r *CardRenderer) DrawShopSpell(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawEllipseBase(screen, rect, color.RGBA{70, 35, 100, 255}, c.IsGolden, true, 255)

	name, desc, _ := r.cardInfo(c)

	// Name (top, centered).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.20, rect.Y+rect.H*0.15, color.White)

	// Description (center).
	ui.DrawText(screen, r.Font, desc, rect.X+rect.W*0.15, rect.Y+rect.H*0.40, color.RGBA{180, 180, 180, 255})

	// SPELL label (bottom).
	ui.DrawText(screen, r.Font, "SPELL",
		rect.X+rect.W*0.30, rect.Bottom()-rect.H*0.22,
		color.RGBA{200, 150, 255, 255})
}

// DrawMinion renders a minion (portrait ellipse) for board and combat contexts.
// Use alpha < 255 and flashPct > 0 for combat effects (death fade, damage flash).
func (r *CardRenderer) DrawMinion(screen *ebiten.Image, c api.Card, rect ui.Rect, alpha uint8, flashPct float64) {
	bg := color.RGBA{35, 35, 55, alpha}
	if flashPct > 0.7 {
		bg = color.RGBA{200, 200, 220, alpha}
	}
	r.drawEllipseBase(screen, rect, bg, c.IsGolden, false, alpha)

	// Image placeholder (center).
	ui.DrawText(screen, r.Font, c.TemplateID, rect.X+rect.W*0.15, rect.Y+rect.H*0.42, color.RGBA{100, 100, 120, alpha})

	// Keywords (below placeholder).
	if t := r.Cards.ByTemplateID(c.TemplateID); t != nil {
		kwY := rect.Y + rect.H*0.58
		for _, kw := range t.Keywords.List() {
			ui.DrawText(screen, r.Font, kw.String(),
				rect.X+rect.W*0.20, kwY,
				color.RGBA{180, 220, 140, alpha})
			kwY += rect.H * 0.08
		}
	}

	// Attack (bottom-left, yellow).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Attack),
		rect.X+rect.W*0.20, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 215, 0, alpha})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, fmt.Sprintf("%d", c.Health),
		rect.Right()-rect.W*0.30, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 80, 80, alpha})
}

func (r *CardRenderer) drawRectBase(screen *ebiten.Image, rect ui.Rect, bg color.RGBA, golden, spell bool, alpha uint8) {
	sr := rect.Screen()
	s := ui.ActiveRes.Scale()

	bg.A = alpha
	vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), bg, false)

	border := color.RGBA{80, 80, 100, alpha}
	borderW := float32(2 * s)
	if golden {
		border = color.RGBA{255, 215, 0, alpha}
		borderW = float32(3 * s)
	} else if spell {
		border = color.RGBA{140, 80, 180, alpha}
	}
	vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), borderW, border, false)
}

func (r *CardRenderer) drawEllipseBase(screen *ebiten.Image, rect ui.Rect, bg color.RGBA, golden, spell bool, alpha uint8) {
	sr := rect.Screen()
	s := ui.ActiveRes.Scale()

	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	rx := float32(sr.W / 2)
	ry := float32(sr.H / 2)

	bg.A = alpha
	ui.FillEllipse(screen, cx, cy, rx, ry, bg)

	border := color.RGBA{80, 80, 100, alpha}
	borderW := float32(2 * s)
	if golden {
		border = color.RGBA{255, 215, 0, alpha}
		borderW = float32(3 * s)
	} else if spell {
		border = color.RGBA{140, 80, 180, alpha}
	}
	ui.StrokeEllipse(screen, cx, cy, rx, ry, borderW, border)
}

func (r *CardRenderer) cardInfo(c api.Card) (name, desc, tribe string) {
	name = c.TemplateID
	if t := r.Cards.ByTemplateID(c.TemplateID); t != nil {
		name = t.Name
		desc = t.Description
		tribe = t.Tribe.String()
	}
	return
}
