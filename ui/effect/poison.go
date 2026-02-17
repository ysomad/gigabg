package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// PoisonDrip draws a green poison drip overlay on a minion about to die
// from poison. The combatboard starts the death sequence when this effect
// completes (is removed from the list).
type PoisonDrip struct {
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

var _ Effect = (*PoisonDrip)(nil)

func NewPoisonDrip(duration float64, alpha uint8) *PoisonDrip {
	return &PoisonDrip{
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

func (e *PoisonDrip) Kind() Kind { return KindPoisonDrip }

func (e *PoisonDrip) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer <= 0 {
		e.timer = 0
		return true
	}
	return false
}

func (e *PoisonDrip) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *PoisonDrip) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *PoisonDrip) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *PoisonDrip) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
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

	poisonA := uint8(float64(e.alpha) * 0.4 * t)
	fc := e.fillColor
	fc.A = poisonA
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
