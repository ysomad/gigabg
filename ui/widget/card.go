package widget

import (
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/api"
	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/game/catalog"
	"github.com/ysomad/gigabg/ui"
)

// CardRenderer draws cards in different contexts.
type CardRenderer struct {
	Cards *catalog.Catalog
	Font  *text.GoTextFace
	Tick  int // incremented each frame, used for animations
}

func (r *CardRenderer) isSpell(c api.Card) bool {
	t := r.Cards.ByTemplateID(c.TemplateID)
	return t != nil && t.Kind() == game.CardKindSpell
}

// DrawHandCard renders a hand card (rectangle) with full detail.
// Branches on spell vs minion internally.
func (r *CardRenderer) DrawHandCard(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	if r.isSpell(c) {
		r.drawHandSpell(screen, c, rect)
	} else {
		r.drawHandMinion(screen, c, rect)
	}
}

func (r *CardRenderer) drawHandMinion(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawRectBase(screen, rect, color.RGBA{40, 40, 60, 255}, c.IsGolden, false, 255)

	name, desc, tribe := r.cardInfo(c)
	t := r.Cards.ByTemplateID(c.TemplateID)

	// Name (top-left).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.04, rect.Y+rect.H*0.04, color.White)

	// Tier (top-right).
	if t != nil && t.Tier().IsValid() {
		ui.DrawText(screen, r.Font, "T"+strconv.Itoa(int(t.Tier())),
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
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Attack),
		rect.X+rect.W*0.04, rect.Bottom()-rect.H*0.12,
		color.RGBA{255, 215, 0, 255})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Health),
		rect.Right()-rect.W*0.15, rect.Bottom()-rect.H*0.12,
		color.RGBA{255, 80, 80, 255})
}

func (r *CardRenderer) drawHandSpell(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawRectBase(screen, rect, color.RGBA{80, 40, 100, 255}, c.IsGolden, true, 255)

	name, desc, _ := r.cardInfo(c)

	// Name (top-left).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.04, rect.Y+rect.H*0.04, color.White)

	// Cost (top-right).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Cost),
		rect.Right()-rect.W*0.15, rect.Y+rect.H*0.04,
		color.RGBA{255, 215, 0, 255})

	// Description (center).
	ui.DrawText(screen, r.Font, desc, rect.X+rect.W*0.04, rect.Y+rect.H*0.30, color.RGBA{180, 180, 180, 255})

	// SPELL label (bottom-center).
	ui.DrawText(screen, r.Font, "SPELL",
		rect.X+rect.W*0.30, rect.Bottom()-rect.H*0.12,
		color.RGBA{200, 150, 255, 255})
}

// DrawShopCard renders a shop card (portrait ellipse).
// Branches on spell vs minion internally.
func (r *CardRenderer) DrawShopCard(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	if r.isSpell(c) {
		r.drawShopSpell(screen, c, rect)
	} else {
		r.drawShopMinion(screen, c, rect)
	}
}

func (r *CardRenderer) drawShopMinion(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawEllipseBase(screen, rect, color.RGBA{35, 40, 70, 255}, c.IsGolden, false, 255, c.Keywords)

	t := r.Cards.ByTemplateID(c.TemplateID)

	// Image placeholder (center).
	ui.DrawText(screen, r.Font, c.TemplateID, rect.X+rect.W*0.15, rect.Y+rect.H*0.42, color.RGBA{100, 100, 120, 255})

	// Tier (top, gold).
	if t != nil && t.Tier().IsValid() {
		ui.DrawText(screen, r.Font, "T"+strconv.Itoa(int(t.Tier())),
			rect.X+rect.W*0.42, rect.Y+rect.H*0.15,
			color.RGBA{255, 215, 0, 255})
	}

	// Keywords (below placeholder).
	r.drawKeywords(screen, c.Keywords, rect, 255)

	// Attack (bottom-left, yellow).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Attack),
		rect.X+rect.W*0.20, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 215, 0, 255})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Health),
		rect.Right()-rect.W*0.30, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 80, 80, 255})
}

func (r *CardRenderer) drawShopSpell(screen *ebiten.Image, c api.Card, rect ui.Rect) {
	r.drawEllipseBase(screen, rect, color.RGBA{70, 35, 100, 255}, c.IsGolden, true, 255, 0)

	t := r.Cards.ByTemplateID(c.TemplateID)
	name, _, _ := r.cardInfo(c)

	// Name (top, centered).
	ui.DrawText(screen, r.Font, name, rect.X+rect.W*0.20, rect.Y+rect.H*0.15, color.White)

	// Tier (center-left).
	if t != nil && t.Tier().IsValid() {
		ui.DrawText(screen, r.Font, "T"+strconv.Itoa(int(t.Tier())),
			rect.X+rect.W*0.20, rect.Y+rect.H*0.40,
			color.RGBA{180, 180, 180, 255})
	}

	// Cost (center-right, gold).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Cost)+"g",
		rect.Right()-rect.W*0.40, rect.Y+rect.H*0.40,
		color.RGBA{255, 215, 0, 255})

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
	r.drawEllipseBase(screen, rect, bg, c.IsGolden, false, alpha, c.Keywords)

	// Image placeholder (center).
	ui.DrawText(screen, r.Font, c.TemplateID, rect.X+rect.W*0.15, rect.Y+rect.H*0.42, color.RGBA{100, 100, 120, alpha})

	// Keywords (below placeholder).
	r.drawKeywords(screen, c.Keywords, rect, alpha)

	// Attack (bottom-left, yellow).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Attack),
		rect.X+rect.W*0.20, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 215, 0, alpha})

	// Health (bottom-right, red).
	ui.DrawText(screen, r.Font, strconv.Itoa(c.Health),
		rect.Right()-rect.W*0.30, rect.Bottom()-rect.H*0.22,
		color.RGBA{255, 80, 80, alpha})
}

// drawKeywords renders keyword labels below the card center.
func (r *CardRenderer) drawKeywords(screen *ebiten.Image, kw game.Keywords, rect ui.Rect, alpha uint8) {
	kwY := rect.Y + rect.H*0.58
	for _, k := range kw.All() {
		ui.DrawText(screen, r.Font, k.String(),
			rect.X+rect.W*0.20, kwY,
			color.RGBA{180, 220, 140, alpha})
		kwY += rect.H * 0.08
	}
}

func (r *CardRenderer) drawRectBase(
	screen *ebiten.Image, rect ui.Rect, bg color.RGBA, golden, spell bool, alpha uint8,
) {
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

func (r *CardRenderer) drawEllipseBase(
	screen *ebiten.Image, rect ui.Rect, bg color.RGBA,
	golden, spell bool, alpha uint8, keywords game.Keywords,
) {
	sr := rect.Screen()
	s := ui.ActiveRes.Scale()

	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	rx := float32(sr.W / 2)
	ry := float32(sr.H / 2)

	if keywords.Has(game.KeywordTaunt) {
		r.drawTaunt(screen, cx, cy, rx, ry, s, alpha)
	}

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

	if keywords.Has(game.KeywordWindfury) {
		r.drawWindfury(screen, cx, cy, rx, ry, s, alpha)
	}

	if keywords.Has(game.KeywordDivineShield) {
		r.drawDivineShield(screen, cx, cy, rx, ry, s, alpha)
	}
}

// drawWindfury draws animated wind streaks orbiting the minion.
func (r *CardRenderer) drawWindfury(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	pad := float32(2 * s)
	orbRX := rx + pad
	orbRY := ry + pad

	// Rotation angle based on tick (60 TPS assumed).
	angle := float64(r.Tick) * 0.03
	clr := color.RGBA{80, 180, 255, alpha}
	strokeW := float32(1.5 * s)

	// 3 wind streaks, evenly spaced.
	const streaks = 3
	const arcLen = 0.6 // radians per streak
	const steps = 10

	for i := range streaks {
		base := angle + float64(i)*2*math.Pi/streaks

		var path vector.Path
		for j := range steps + 1 {
			t := base + arcLen*float64(j)/float64(steps)
			x := cx + orbRX*float32(math.Cos(t))
			y := cy + orbRY*float32(math.Sin(t))
			if j == 0 {
				path.MoveTo(x, y)
			} else {
				path.LineTo(x, y)
			}
		}

		op := &vector.DrawPathOptions{}
		op.ColorScale.ScaleWithColor(clr)
		vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: strokeW}, op)
	}
}

// drawDivineShield draws a golden glow over the minion using additive blending.
// Perfect vertical ellipse slightly larger than the portrait.
func (r *CardRenderer) drawDivineShield(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	pad := float32(3 * s)
	w := int((rx+pad)*2 + 2)
	h := int((ry+pad)*2 + 2)

	tmp := ebiten.NewImage(w, h)

	localCX := float32(w) / 2
	localCY := float32(h) / 2

	ui.FillEllipse(tmp, localCX, localCY, rx+pad, ry+pad, color.RGBA{40, 35, 10, alpha})
	ui.StrokeEllipse(tmp, localCX, localCY, rx+pad, ry+pad, float32(1.5*s), color.RGBA{70, 60, 15, alpha})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(cx)-float64(localCX), float64(cy)-float64(localCY))
	op.Blend = ebiten.BlendLighter
	screen.DrawImage(tmp, op)
}

// drawTaunt draws a grey shield ring (taunt frame) behind the minion ellipse.
// Shape: shallow arc on top, slightly bulged rounded sides, rounded/flattened point at bottom.
func (r *CardRenderer) drawTaunt(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	pad := float32(5 * s)

	halfW := rx + pad         // half-width at shoulders (widest)
	topY := cy - ry - pad     // top edge
	botY := cy + ry + pad*2.2 // bottom tip (flattened point, not sharp)

	var path vector.Path

	// Start at top-left shoulder.
	path.MoveTo(cx-halfW, topY)

	// Top edge: shallow upward arc.
	path.CubicTo(
		cx-halfW*0.35, topY-pad*0.7,
		cx+halfW*0.35, topY-pad*0.7,
		cx+halfW, topY,
	)

	// Right side: slight outward bulge at upper portion, then curves in toward bottom.
	path.CubicTo(
		cx+halfW*1.04, cy-ry*0.15,
		cx+halfW*0.85, cy+ry*0.5,
		cx+halfW*0.35, cy+ry*0.9,
	)

	// Right side to bottom tip: smooth rounded point.
	path.CubicTo(
		cx+halfW*0.12, botY-pad*1.0,
		cx+pad*0.3, botY-pad*0.2,
		cx, botY,
	)

	// Left side from bottom tip: mirror of right.
	path.CubicTo(
		cx-pad*0.3, botY-pad*0.2,
		cx-halfW*0.12, botY-pad*1.0,
		cx-halfW*0.35, cy+ry*0.9,
	)

	// Left side upper: mirror bulge back to top-left.
	path.CubicTo(
		cx-halfW*0.85, cy+ry*0.5,
		cx-halfW*1.04, cy-ry*0.15,
		cx-halfW, topY,
	)

	path.Close()

	// Fill (dark grey).
	fillOp := &vector.DrawPathOptions{}
	fillOp.ColorScale.ScaleWithColor(color.RGBA{65, 68, 78, alpha})
	vector.FillPath(screen, &path, nil, fillOp)

	// Stroke border (lighter grey ring).
	strokeOp := &vector.DrawPathOptions{}
	strokeOp.ColorScale.ScaleWithColor(color.RGBA{135, 135, 150, alpha})
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: float32(1 * s)}, strokeOp)
}

func (r *CardRenderer) cardInfo(c api.Card) (string, string, string) {
	name := c.TemplateID
	var desc, tribe string
	if t := r.Cards.ByTemplateID(c.TemplateID); t != nil {
		name = t.Name()
		desc = t.Description()
		tribe = t.Tribe().String()
	}
	return name, desc, tribe
}
