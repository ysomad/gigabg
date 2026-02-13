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
	Cards    *catalog.Catalog
	Font     *text.GoTextFace
	BoldFont *text.GoTextFace
	Tick     int // incremented each frame, used for animations
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

	// Attack and health badges.
	badgeR := rect.W * 0.09
	r.drawAttackBadge(screen, c.Attack, rect.X+rect.W*0.10, rect.Bottom()-rect.H*0.10, badgeR, 255)
	r.drawHealthBadge(screen, c.Health, rect.Right()-rect.W*0.10, rect.Bottom()-rect.H*0.10, badgeR, 255)
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

	// Keywords (below placeholder).
	r.drawKeywords(screen, c.Keywords, rect, 255)

	// Keyword effects on top of text/art.
	r.drawKeywordEffects(screen, rect, c.Keywords, 255, c.Attack, c.Health)

	// Tier badge on top of everything.
	if t != nil && t.Tier().IsValid() {
		r.drawTierBadge(screen, int(t.Tier()), rect.X+rect.W*0.5, rect.Y-rect.H*0.02, rect.W*0.144, 255)
	}
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

	// Keyword effects on top of text/art.
	r.drawKeywordEffects(screen, rect, c.Keywords, alpha, c.Attack, c.Health)
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

// drawKeywordEffects draws all keyword visual effects on top of text and art.
// Badges are drawn after stealth but before divine shield so they stay visible.
func (r *CardRenderer) drawKeywordEffects(
	screen *ebiten.Image,
	rect ui.Rect,
	keywords game.Keywords,
	alpha uint8,
	attack, health int,
) {
	sr := rect.Screen()
	s := ui.ActiveRes.Scale()

	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	rx := float32(sr.W / 2)
	ry := float32(sr.H / 2)

	if keywords.Has(game.KeywordPoisonous) {
		r.drawPoisonous(screen, cx, cy+ry*0.95, s, alpha)
	}

	if keywords.Has(game.KeywordVenomous) {
		r.drawVenomous(screen, cx, cy+ry*0.95, s, alpha)
	}

	if keywords.Has(game.KeywordStealth) {
		r.drawStealth(screen, cx, cy, rx, ry, s, alpha)
	}

	// Badges on top of effects but below windfury and divine shield.
	badgeR := rect.W * 0.11
	bcx, bcy := rect.X+rect.W*0.5, rect.Y+rect.H*0.5
	brx, bry := rect.W*0.5, rect.H*0.5
	angle := math.Pi * 0.75
	r.drawAttackBadge(screen, attack, bcx+brx*math.Cos(angle), bcy+bry*math.Sin(angle), badgeR, alpha)
	r.drawHealthBadge(screen, health, bcx+brx*math.Cos(math.Pi-angle), bcy+bry*math.Sin(math.Pi-angle), badgeR, alpha)

	// Wind streaks in front of the minion (back pass is in drawEllipseBase).
	if keywords.Has(game.KeywordWindfury) {
		r.drawWindfury(screen, cx, cy, ry, s, alpha, true)
	}

	if keywords.Has(game.KeywordDivineShield) {
		r.drawDivineShield(screen, cx, cy, rx, ry, s, alpha)
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

	// Taunt shield drawn behind the minion body.
	if keywords.Has(game.KeywordTaunt) {
		r.drawTaunt(screen, cx, cy, rx, ry, s, alpha)
	}

	// Wind streaks behind the minion.
	if keywords.Has(game.KeywordWindfury) {
		r.drawWindfury(screen, cx, cy, ry, s, alpha, false)
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
}

// drawWindfury draws two crossing orbits of wind ribbons wrapping the minion in 3D.
// Streaks are split into front/back segments; segment edges taper to zero width
// at front/back boundaries so no seam lines are visible.
func (r *CardRenderer) drawWindfury(screen *ebiten.Image, cx, cy, ry float32, s float64, alpha uint8, front bool) {
	orbitR := float64(ry)*1.05 + 1*s
	const tilt = 0.3
	const deg39 = 39.0 * math.Pi / 180.0

	angle := float64(r.Tick) * 0.04

	const streaksPerOrbit = 2
	const arcLen = math.Pi * 0.55
	const steps = 36
	waveAmp := 1.5 * s
	maxHalf := 1.8 * s

	var a uint8
	if front {
		a = uint8(float64(alpha) * 0.18)
	} else {
		a = uint8(float64(alpha) * 0.05)
	}

	type orbitCfg struct {
		rot float64
		dir float64
	}
	orbits := [2]orbitCfg{
		{rot: -deg39, dir: 1},
		{rot: deg39, dir: -1},
	}

	type sample struct {
		mid          [2]float32
		outer, inner [2]float32
		inFront      bool
	}

	for _, orb := range orbits {
		cosRot := math.Cos(orb.rot)
		sinRot := math.Sin(orb.rot)
		orbAngle := angle * orb.dir

		for i := range streaksPerOrbit {
			base := orbAngle + float64(i)*2*math.Pi/streaksPerOrbit

			samples := make([]sample, steps+1)
			for j := range steps + 1 {
				t := float64(j) / float64(steps)
				theta := base + arcLen*t

				depth := math.Sin(theta)
				ux := orbitR * math.Cos(theta)
				uy := orbitR * depth * tilt
				screenX := float64(cx) + ux*cosRot - uy*sinRot
				screenY := float64(cy) + ux*sinRot + uy*cosRot

				wave := math.Sin(t*1.5*2*math.Pi) * waveAmp
				taper := math.Sin(t * math.Pi)
				halfW := maxHalf * taper

				dx := screenX - float64(cx)
				dy := screenY - float64(cy)
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < 1 {
					dist = 1
				}
				nx := dx / dist
				ny := dy / dist

				midX := screenX + wave*nx
				midY := screenY + wave*ny

				samples[j] = sample{
					mid:     [2]float32{float32(midX), float32(midY)},
					outer:   [2]float32{float32(midX + halfW*nx), float32(midY + halfW*ny)},
					inner:   [2]float32{float32(midX - halfW*nx), float32(midY - halfW*ny)},
					inFront: depth >= 0,
				}
			}

			// Draw per-step quads; assign each quad to the pass of its first sample.
			var path vector.Path
			for k := range steps {
				if samples[k].inFront != front {
					continue
				}
				path.MoveTo(samples[k].outer[0], samples[k].outer[1])
				path.LineTo(samples[k+1].outer[0], samples[k+1].outer[1])
				path.LineTo(samples[k+1].inner[0], samples[k+1].inner[1])
				path.LineTo(samples[k].inner[0], samples[k].inner[1])
				path.Close()
			}

			op := &vector.DrawPathOptions{AntiAlias: true}
			op.ColorScale.ScaleWithColor(color.RGBA{175, 175, 178, a})
			vector.FillPath(screen, &path, nil, op)
		}
	}
}

var (
	bottleFill   = color.RGBA{30, 160, 50, 255}
	bottleStroke = color.RGBA{15, 60, 20, 255}
	bottleCap    = color.RGBA{100, 80, 50, 255}
)

// drawPoisonous draws a rounded potion vial at (cx, cy).
func (r *CardRenderer) drawPoisonous(screen *ebiten.Image, cx, cy float32, s float64, alpha uint8) {
	sf := float32(s)

	h := 21 * sf    // bottle height
	w := 8.4 * sf   // body half-width
	nw := 3.15 * sf // neck half-width
	sw := 4.6 * sf  // stopper half-width

	var stopper vector.Path
	stopper.MoveTo(cx-sw, cy-h*0.5)
	stopper.LineTo(cx+sw, cy-h*0.5)
	stopper.LineTo(cx+sw, cy-h*0.4)
	stopper.LineTo(cx-sw, cy-h*0.4)
	stopper.Close()

	stopOp := &vector.DrawPathOptions{AntiAlias: true}
	stopOp.ColorScale.ScaleWithColor(withAlpha(bottleCap, alpha))
	vector.FillPath(screen, &stopper, nil, stopOp)

	var body vector.Path
	body.MoveTo(cx-nw, cy-h*0.4)
	body.LineTo(cx+nw, cy-h*0.4)
	body.LineTo(cx+nw, cy-h*0.15)
	body.CubicTo(cx+w*1.2, cy-h*0.05, cx+w*1.2, cy+h*0.1, cx+w, cy+h*0.15)
	body.CubicTo(cx+w, cy+h*0.4, cx+w*0.5, cy+h*0.5, cx, cy+h*0.5)
	body.CubicTo(cx-w*0.5, cy+h*0.5, cx-w, cy+h*0.4, cx-w, cy+h*0.15)
	body.CubicTo(cx-w*1.2, cy+h*0.1, cx-w*1.2, cy-h*0.05, cx-nw, cy-h*0.15)
	body.Close()

	fillOp := &vector.DrawPathOptions{AntiAlias: true}
	fillOp.ColorScale.ScaleWithColor(withAlpha(bottleFill, alpha))
	vector.FillPath(screen, &body, nil, fillOp)

	strokeOp := &vector.DrawPathOptions{AntiAlias: true}
	strokeOp.ColorScale.ScaleWithColor(withAlpha(bottleStroke, alpha))
	vector.StrokePath(screen, &body, &vector.StrokeOptions{Width: float32(0.8 * s)}, strokeOp)
}

// drawVenomous draws a narrow elongated vial at (cx, cy).
func (r *CardRenderer) drawVenomous(screen *ebiten.Image, cx, cy float32, s float64, alpha uint8) {
	sf := float32(s)

	h := 23 * sf   // bottle height
	w := 5.9 * sf  // body half-width
	nw := 2.5 * sf // neck half-width
	cw := 3.8 * sf // cap half-width

	var cp vector.Path
	cp.MoveTo(cx-cw, cy-h*0.5)
	cp.LineTo(cx+cw, cy-h*0.5)
	cp.LineTo(cx+cw, cy-h*0.42)
	cp.LineTo(cx-cw, cy-h*0.42)
	cp.Close()

	capOp := &vector.DrawPathOptions{AntiAlias: true}
	capOp.ColorScale.ScaleWithColor(withAlpha(bottleCap, alpha))
	vector.FillPath(screen, &cp, nil, capOp)

	var body vector.Path
	body.MoveTo(cx-nw, cy-h*0.42)
	body.LineTo(cx+nw, cy-h*0.42)
	body.LineTo(cx+nw, cy-h*0.2)
	body.CubicTo(cx+w, cy-h*0.15, cx+w, cy-h*0.1, cx+w, cy+h*0.2)
	body.CubicTo(cx+w, cy+h*0.45, cx+w*0.4, cy+h*0.5, cx, cy+h*0.5)
	body.CubicTo(cx-w*0.4, cy+h*0.5, cx-w, cy+h*0.45, cx-w, cy+h*0.2)
	body.CubicTo(cx-w, cy-h*0.1, cx-w, cy-h*0.15, cx-nw, cy-h*0.2)
	body.Close()

	fillOp := &vector.DrawPathOptions{AntiAlias: true}
	fillOp.ColorScale.ScaleWithColor(withAlpha(bottleFill, alpha))
	vector.FillPath(screen, &body, nil, fillOp)

	strokeOp := &vector.DrawPathOptions{AntiAlias: true}
	strokeOp.ColorScale.ScaleWithColor(withAlpha(bottleStroke, alpha))
	vector.StrokePath(screen, &body, &vector.StrokeOptions{Width: float32(0.8 * s)}, strokeOp)
}

func withAlpha(c color.RGBA, alpha uint8) color.RGBA {
	c.A = alpha
	return c
}

// drawAttackBadge draws a gold circle with bold white number.
func (r *CardRenderer) drawAttackBadge(screen *ebiten.Image, attack int, baseX, baseY, baseR float64, alpha uint8) {
	s := ui.ActiveRes.Scale()
	ox := ui.ActiveRes.OffsetX()
	oy := ui.ActiveRes.OffsetY()

	sx := float32(baseX*s + ox)
	sy := float32(baseY*s + oy)
	sr := float32(baseR * s)

	vector.FillCircle(screen, sx, sy, sr, color.RGBA{180, 140, 10, alpha}, true)
	vector.StrokeCircle(screen, sx, sy, sr, float32(2*s), color.RGBA{80, 80, 100, alpha}, true)

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(sx), float64(sy))
	op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, alpha})
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(screen, strconv.Itoa(attack), r.BoldFont, op)
}

// drawHealthBadge draws a red circle with bold white number.
func (r *CardRenderer) drawHealthBadge(screen *ebiten.Image, health int, baseX, baseY, baseR float64, alpha uint8) {
	s := ui.ActiveRes.Scale()
	ox := ui.ActiveRes.OffsetX()
	oy := ui.ActiveRes.OffsetY()

	sx := float32(baseX*s + ox)
	sy := float32(baseY*s + oy)
	sr := float32(baseR * s)

	vector.FillCircle(screen, sx, sy, sr, color.RGBA{160, 20, 20, alpha}, true)
	vector.StrokeCircle(screen, sx, sy, sr, float32(2*s), color.RGBA{80, 80, 100, alpha}, true)

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(sx), float64(sy))
	op.ColorScale.ScaleWithColor(color.RGBA{255, 255, 255, alpha})
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(screen, strconv.Itoa(health), r.BoldFont, op)
}

// drawTierBadge draws a shield-shaped badge with golden stars (one per tier level).
// Shield has deep purple fill with ellipse-matching outline color.
func (r *CardRenderer) drawTierBadge(screen *ebiten.Image, tier int, baseX, baseY, baseR float64, alpha uint8) {
	s := ui.ActiveRes.Scale()
	ox := ui.ActiveRes.OffsetX()
	oy := ui.ActiveRes.OffsetY()

	cx := float32(baseX*s + ox)
	cy := float32(baseY*s + oy)
	radius := float32(baseR * s)

	starR := radius * 0.38

	// Fixed shield size for all tiers.
	shieldW := radius * 1.3
	shieldH := radius * 2.6

	// Shield shape: rounded top, tapered bottom point.
	topY := cy - shieldH*0.4
	botY := cy + shieldH*0.6

	var shield vector.Path
	shield.MoveTo(cx, topY)
	shield.CubicTo(cx+shieldW*0.8, topY, cx+shieldW, topY+shieldH*0.05, cx+shieldW, topY+shieldH*0.15)
	shield.LineTo(cx+shieldW, topY+shieldH*0.55)
	shield.CubicTo(cx+shieldW, topY+shieldH*0.75, cx+shieldW*0.3, botY-shieldH*0.05, cx, botY)
	shield.CubicTo(cx-shieldW*0.3, botY-shieldH*0.05, cx-shieldW, topY+shieldH*0.75, cx-shieldW, topY+shieldH*0.55)
	shield.LineTo(cx-shieldW, topY+shieldH*0.15)
	shield.CubicTo(cx-shieldW, topY+shieldH*0.05, cx-shieldW*0.8, topY, cx, topY)
	shield.Close()

	// Fill: deep purple/magenta.
	{
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(color.RGBA{55, 15, 75, alpha})
		vector.FillPath(screen, &shield, nil, op)
	}

	// Outline: same color as ellipse border.
	{
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(color.RGBA{80, 80, 100, alpha})
		vector.StrokePath(screen, &shield, &vector.StrokeOptions{Width: float32(1.5 * s)}, op)
	}

	// Star positions per tier (offsets from shield center).
	shieldCY := cy + shieldH*0.05 // visual center of shield content area
	gap := starR * 1.2            // half gap between stars

	switch tier {
	case 1:
		// Single centered star.
		r.drawStar(screen, cx, shieldCY, starR, alpha)
	case 2:
		// Two stars in a row.
		r.drawStar(screen, cx-gap, shieldCY, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY, starR, alpha)
	case 3:
		// Triangle: 1 on top, 2 on bottom.
		r.drawStar(screen, cx, shieldCY-gap, starR, alpha)
		r.drawStar(screen, cx-gap, shieldCY+gap*0.7, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY+gap*0.7, starR, alpha)
	case 4:
		// 2x2 grid.
		r.drawStar(screen, cx-gap, shieldCY-gap*0.7, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY-gap*0.7, starR, alpha)
		r.drawStar(screen, cx-gap, shieldCY+gap*0.7, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY+gap*0.7, starR, alpha)
	case 5:
		// 2 top, 1 middle, 2 bottom.
		r.drawStar(screen, cx-gap, shieldCY-gap*1.1, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY-gap*1.1, starR, alpha)
		r.drawStar(screen, cx, shieldCY, starR, alpha)
		r.drawStar(screen, cx-gap, shieldCY+gap*1.1, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY+gap*1.1, starR, alpha)
	case 6:
		// 3x2 grid.
		r.drawStar(screen, cx-gap, shieldCY-gap*1.1, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY-gap*1.1, starR, alpha)
		r.drawStar(screen, cx-gap, shieldCY, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY, starR, alpha)
		r.drawStar(screen, cx-gap, shieldCY+gap*1.1, starR, alpha)
		r.drawStar(screen, cx+gap, shieldCY+gap*1.1, starR, alpha)
	}
}

// drawStar draws a 5-pointed golden star at (cx, cy) with the given outer radius
// and a thin dark outline.
func (r *CardRenderer) drawStar(screen *ebiten.Image, cx, cy, outerR float32, alpha uint8) {
	innerR := outerR * 0.45
	const points = 5

	var path vector.Path
	for i := range points * 2 {
		angle := float64(i)*math.Pi/float64(points) - math.Pi/2
		rad := outerR
		if i%2 == 1 {
			rad = innerR
		}
		px := cx + float32(math.Cos(angle))*rad
		py := cy + float32(math.Sin(angle))*rad
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	path.Close()

	// Gold fill.
	fillOp := &vector.DrawPathOptions{AntiAlias: true}
	fillOp.ColorScale.ScaleWithColor(color.RGBA{255, 215, 0, alpha})
	vector.FillPath(screen, &path, nil, fillOp)

	// Thin dark outline.
	strokeOp := &vector.DrawPathOptions{AntiAlias: true}
	strokeOp.ColorScale.ScaleWithColor(color.RGBA{40, 30, 0, alpha})
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: outerR * 0.15}, strokeOp)
}

// drawDivineShield draws a golden glow over the minion using additive blending.
// Perfect vertical ellipse slightly larger than the portrait.
func (r *CardRenderer) drawDivineShield(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	pad := float32(9 * s)
	erx := rx + pad
	ery := ry + pad

	const k = 0.5522847498
	var path vector.Path
	path.MoveTo(cx, cy-ery)
	path.CubicTo(cx+erx*k, cy-ery, cx+erx, cy-ery*k, cx+erx, cy)
	path.CubicTo(cx+erx, cy+ery*k, cx+erx*k, cy+ery, cx, cy+ery)
	path.CubicTo(cx-erx*k, cy+ery, cx-erx, cy+ery*k, cx-erx, cy)
	path.CubicTo(cx-erx, cy-ery*k, cx-erx*k, cy-ery, cx, cy-ery)
	path.Close()

	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(color.RGBA{40, 35, 10, alpha})
	op.Blend = ebiten.BlendLighter
	vector.FillPath(screen, &path, nil, op)
}

// drawStealth draws a dark semi-transparent overlay on the minion, giving a shadowy appearance.
func (r *CardRenderer) drawStealth(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	shrink := float32(2 * s)
	srx := rx - shrink
	sry := ry - shrink

	const k = 0.5522847498
	var path vector.Path
	path.MoveTo(cx, cy-sry)
	path.CubicTo(cx+srx*k, cy-sry, cx+srx, cy-sry*k, cx+srx, cy)
	path.CubicTo(cx+srx, cy+sry*k, cx+srx*k, cy+sry, cx, cy+sry)
	path.CubicTo(cx-srx*k, cy+sry, cx-srx, cy+sry*k, cx-srx, cy)
	path.CubicTo(cx-srx, cy-sry*k, cx-srx*k, cy-sry, cx, cy-sry)
	path.Close()

	a := uint8(float64(alpha) * 0.55)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(color.RGBA{3, 3, 5, a})
	vector.FillPath(screen, &path, nil, op)
}

// drawTaunt draws an organic shield silhouette behind the minion ellipse.
// All curves, no straight segments. Dome top, bulging shoulders (widest),
// tapering sides, rounded bottom tip.
func (r *CardRenderer) drawTaunt(screen *ebiten.Image, cx, cy, rx, ry float32, s float64, alpha uint8) {
	sf := float32(s)
	pad := 16 * sf

	maxW := rx + pad*0.55     // reference half-width (narrower shoulders)
	topY := cy - ry - pad*0.5 // dome peak (tight to portrait top)
	botY := cy + ry + pad*1.8 // bottom tip
	h := botY - topY          // total height

	// Shoulder: widest point, nearly level with top.
	shoulderY := topY + h*0.06
	shoulderW := maxW

	var path vector.Path

	// Start at dome peak.
	path.MoveTo(cx, topY)

	// Right dome: flat top, tight turn at shoulder.
	path.CubicTo(
		cx+shoulderW*0.85, topY,
		cx+shoulderW, shoulderY-h*0.015,
		cx+shoulderW, shoulderY,
	)

	// Right side → bottom tip. Straight down then tight turn at tip.
	// P2 at tip height offset right — mirrors how top dome corners work.
	path.CubicTo(
		cx+shoulderW, cy+ry,
		cx+shoulderW*0.08, botY-h*0.01,
		cx, botY,
	)

	// Left side: bottom tip → shoulder (mirror).
	path.CubicTo(
		cx-shoulderW*0.08, botY-h*0.01,
		cx-shoulderW, cy+ry,
		cx-shoulderW, shoulderY,
	)

	// Left dome: tight shoulder, flat top (mirror).
	path.CubicTo(
		cx-shoulderW, shoulderY-h*0.015,
		cx-shoulderW*0.85, topY,
		cx, topY,
	)

	path.Close()

	// Base fill: mid-tone steel.
	{
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(color.RGBA{72, 76, 88, alpha})
		vector.FillPath(screen, &path, nil, op)
	}

	// Thin outer border.
	{
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(color.RGBA{135, 140, 155, alpha})
		vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: sf}, op)
	}
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
