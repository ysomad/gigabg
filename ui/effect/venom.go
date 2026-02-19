package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// VenomDrip draws a green venom drip overlay on a minion about to die
// from Venomous. The combatboard starts the death sequence when this effect
// completes (is removed from the list).
type VenomDrip struct {
	timer    float64
	duration float64
	alpha    uint8 // card alpha for compositing

	// Visual config.
	fillColor   color.RGBA
	dropColor   color.RGBA
	skullBG     color.RGBA
	skullBone   color.RGBA
	skullFadeIn float64 // progress fraction where skull starts fading in
}

var _ Effect = (*VenomDrip)(nil)

func NewVenomDrip(duration float64, alpha uint8) *VenomDrip {
	return &VenomDrip{
		timer:       duration,
		duration:    duration,
		alpha:       alpha,
		fillColor:   color.RGBA{20, 180, 40, 255},
		dropColor:   color.RGBA{30, 200, 50, 255},
		skullBG:     color.RGBA{30, 120, 40, 255},
		skullBone:   color.RGBA{230, 220, 200, 255},
		skullFadeIn: 0.6,
	}
}

func (e *VenomDrip) Kind() Kind { return KindVenomDrip }

func (e *VenomDrip) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer <= 0 {
		e.timer = 0
		return true
	}
	return false
}

func (e *VenomDrip) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *VenomDrip) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *VenomDrip) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *VenomDrip) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	sr := rect.Screen(res)
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	rx := float32(sr.W / 2)
	ry := float32(sr.H / 2)

	t := e.Progress()

	fillH := ry * 2 * float32(t)
	_ = fillH

	// Ellipse clipping path.
	const k = 0.5522847498
	var path vector.Path
	path.MoveTo(cx, cy-ry)
	path.CubicTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy)
	path.CubicTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry)
	path.CubicTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy)
	path.CubicTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry)
	path.Close()

	venomA := uint8(float64(e.alpha) * 0.4 * t)
	fc := e.fillColor
	fc.A = venomA
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(fc)
	vector.FillPath(screen, &path, nil, op)

	// Drip droplets.
	dropCount := int(t * 5)
	for j := range dropCount {
		dropX := cx + float32(math.Sin(float64(j)*1.7))*rx*0.6
		dropY := cy - ry + ry*2*float32(t) + float32(j)*ry*0.15
		dropR := float32(3*res.Scale()) * float32(0.5+t*0.5)
		dropA := uint8(float64(e.alpha) * 0.6 * t)
		dc := e.dropColor
		dc.A = dropA
		vector.FillCircle(screen, dropX, dropY, dropR, dc, true)
	}

	// Skull preview.
	if t > e.skullFadeIn {
		skullAlpha := uint8(float64(e.alpha) * (t - e.skullFadeIn) / (1.0 - e.skullFadeIn) * 0.8)
		sbg := e.skullBG
		sbg.A = skullAlpha
		sbone := e.skullBone
		sbone.A = skullAlpha
		drawSkull(screen, cx, cy, rx*0.4, sbg, sbone)
	}
}

func drawSkull(screen *ebiten.Image, cx, cy, size float32, bg, bone color.RGBA) {
	crX := size * 0.96
	crY := size * 1.04
	crCY := cy - size*0.08
	ui.FillEllipse(screen, cx, crCY, crX, crY, bone)

	jrX := crX * 0.68
	jrY := crY * 0.32
	ui.FillEllipse(screen, cx, crCY+crY*0.7, jrX, jrY, bone)

	eyeR := crX * 0.22
	eyeOff := crX * 0.38
	eyeY := crCY - crY*0.08
	vector.FillCircle(screen, cx-eyeOff, eyeY, eyeR, bg, true)
	vector.FillCircle(screen, cx+eyeOff, eyeY, eyeR, bg, true)

	noseY := crCY + crY*0.22
	noseH := crY * 0.18
	noseW := crX * 0.14
	var nose vector.Path
	nose.MoveTo(cx, noseY+noseH)
	nose.LineTo(cx-noseW, noseY)
	nose.LineTo(cx+noseW, noseY)
	nose.Close()
	noseOp := &vector.DrawPathOptions{AntiAlias: true}
	noseOp.ColorScale.ScaleWithColor(bg)
	vector.FillPath(screen, &nose, nil, noseOp)

	teethY := crCY + crY*0.42
	teethH := crY * 0.28
	toothW := size * 0.08
	for _, off := range []float32{-0.22, 0, 0.22} {
		tx := cx + crX*off
		vector.FillRect(screen, tx-toothW*0.5, teethY, toothW, teethH, bg, false)
	}
}

// VenomBreak draws a bottle-shatter animation when the venomous keyword is
// consumed. Glass shards fly outward, green liquid splashes down, and a
// brief green flash appears at the vial's position.
type VenomBreak struct {
	timer    float64
	duration float64

	shardCount  int
	dropCount   int
	glassColor  color.RGBA // shard color
	liquidColor color.RGBA // green splash
	flashColor  color.RGBA // central flash
}

var _ Effect = (*VenomBreak)(nil)

func NewVenomBreak(duration float64) *VenomBreak {
	return &VenomBreak{
		timer:       duration,
		duration:    duration,
		shardCount:  6,
		dropCount:   5,
		glassColor:  color.RGBA{150, 200, 160, 255},
		liquidColor: color.RGBA{30, 180, 50, 255},
		flashColor:  color.RGBA{60, 220, 80, 255},
	}
}

func (e *VenomBreak) Kind() Kind { return KindVenomBreak }

func (e *VenomBreak) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *VenomBreak) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *VenomBreak) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *VenomBreak) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *VenomBreak) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	sr := rect.Screen(res)
	s := float32(res.Scale())

	// Vial position: bottom of card ellipse (matches cy + ry*0.95 from card.go).
	cx := float32(sr.X + sr.W/2)
	ry := float32(sr.H / 2)
	cy := float32(sr.Y+sr.H/2) + ry*0.95

	t := e.Progress()

	// Central green flash (fades quickly).
	if t < 0.4 {
		flashA := uint8(200 * (1.0 - t/0.4))
		flashR := s * 8 * float32(1.0+t*2)
		fc := e.flashColor
		fc.A = flashA
		vector.FillCircle(screen, cx, cy, flashR, fc, true)
	}

	// Glass shards flying outward and rotating.
	for j := range e.shardCount {
		angle := float64(j)*2*math.Pi/float64(e.shardCount) + 0.3 // offset so shards don't align vertically
		dist := float64(s) * (3.0 + t*25.0)
		sx := cx + float32(math.Cos(angle)*dist)
		sy := cy + float32(math.Sin(angle)*dist) - s*float32(t*6*(1.0-t)) // slight arc

		shardSize := s * 3.5 * float32(1.0-t*0.5)
		shardA := uint8(255 * (1.0 - t))

		// Triangular shard.
		rot := t * 3.0 * float64(1+j%2) // each shard rotates at different speed
		sin, cos := math.Sincos(rot)
		dx1 := float32(cos)*shardSize - float32(sin)*shardSize*0.3
		dy1 := float32(sin)*shardSize + float32(cos)*shardSize*0.3
		dx2 := float32(cos)*(-shardSize*0.5) - float32(sin)*shardSize*0.6
		dy2 := float32(sin)*(-shardSize*0.5) + float32(cos)*shardSize*0.6

		var shard vector.Path
		shard.MoveTo(sx+dx1, sy+dy1)
		shard.LineTo(sx+dx2, sy+dy2)
		shard.LineTo(sx-dx1*0.6, sy-dy1*0.8)
		shard.Close()

		gc := e.glassColor
		gc.A = shardA
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(gc)
		vector.FillPath(screen, &shard, nil, op)
	}

	// Green liquid droplets splashing downward.
	for j := range e.dropCount {
		spread := float64(j) - float64(e.dropCount-1)/2.0
		dx := float32(spread * float64(s) * 5.0)
		// Drops fall with gravity: fast initial burst, then downward.
		dropT := t * (1.0 + 0.3*float64(j%3))
		if dropT > 1.0 {
			dropT = 1.0
		}
		dy := s * float32(dropT*dropT*30.0) // quadratic fall
		dropX := cx + dx*float32(1.0+t*0.5)
		dropY := cy + dy - s*float32(8.0*dropT*(1.0-dropT)) // arc up then down

		dropR := s * float32(2.5-t*1.0)
		if dropR < 0.5 {
			dropR = 0.5
		}
		dropA := uint8(220 * (1.0 - t))

		lc := e.liquidColor
		lc.A = dropA
		vector.FillCircle(screen, dropX, dropY, dropR, lc, true)
	}
}
