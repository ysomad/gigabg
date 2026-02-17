package effect

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// HitDamage displays a damage number indicator on a minion body.
// Pops in with overshoot, holds, then fades out.
type HitDamage struct {
	timer    float64
	duration float64
	damage   int
	boldFont *text.GoTextFace

	// Pop-in/fade timing fractions.
	popInEnd  float64 // fraction of lifetime for pop-in (measured from start)
	fadeStart float64 // fraction of lifetime where fade begins (measured from end)
	popScale  float64 // overshoot scale multiplier during pop-in

	// Visual config.
	circleRadius float64 // as fraction of card width
	fontSize     float64 // multiplier on bold font size
	bgColor      color.RGBA
	borderColor  color.RGBA
	textColor    color.RGBA
}

var _ Effect = (*HitDamage)(nil)

func NewHitDamage(duration float64, damage int, boldFont *text.GoTextFace) *HitDamage {
	return &HitDamage{
		timer:        duration,
		duration:     duration,
		damage:       damage,
		boldFont:     boldFont,
		popInEnd:     0.2,
		fadeStart:    0.3,
		popScale:     0.3,
		circleRadius: 0.28,
		fontSize:     2.2,
		bgColor:      color.RGBA{180, 140, 10, 255},
		borderColor:  color.RGBA{80, 80, 100, 255},
		textColor:    color.RGBA{255, 255, 255, 255},
	}
}

func (e *HitDamage) Kind() Kind { return KindHitDamage }

func (e *HitDamage) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *HitDamage) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *HitDamage) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *HitDamage) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	if e.damage <= 0 {
		return
	}

	sr := rect.Screen(res)
	s := res.Scale()

	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)

	t := e.timer / e.duration // 1.0 -> 0.0

	// Pop-in scale.
	scaleMul := 1.0
	if t > 1.0-e.popInEnd {
		pop := (t - (1.0 - e.popInEnd)) / e.popInEnd
		scaleMul = 1.0 + e.popScale*pop
	}

	// Fade out.
	alpha := 255.0
	if t < e.fadeStart {
		alpha = 255.0 * (t / e.fadeStart)
	}
	a := uint8(alpha)

	// Circle background.
	radius := float32(float64(sr.W) * e.circleRadius * scaleMul)
	bg := e.bgColor
	bg.A = a
	vector.FillCircle(screen, cx, cy, radius, bg, true)

	border := e.borderColor
	border.A = a
	vector.StrokeCircle(screen, cx, cy, radius, float32(2*s), border, true)

	// Damage text.
	fontSize := e.boldFont.Size * e.fontSize * scaleMul
	face := *e.boldFont
	face.Size = fontSize

	tc := e.textColor
	tc.A = a

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(cx), float64(cy))
	op.ColorScale.ScaleWithColor(tc)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(screen, "-"+strconv.Itoa(e.damage), &face, op)
}
