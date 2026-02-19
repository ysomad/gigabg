package effect

import (
	"image/color"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// HitType selects the visual variant of a hit indicator.
type HitType uint8

const (
	HitTypeDamage HitType = iota + 1
	HitTypeVenom
)

// HitIndicator displays a hit indicator on a minion body.
// Pops in with overshoot, holds, then fades out.
// Damage variant shows a gold circle with "-X" number.
// Venom variant shows a green circle with a skull icon.
type HitIndicator struct {
	hitType  HitType
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
	contentColor color.RGBA
}

var _ Effect = (*HitIndicator)(nil)

func NewHitIndicator(hitType HitType, duration float64, damage int, boldFont *text.GoTextFace) *HitIndicator {
	h := &HitIndicator{
		hitType:      hitType,
		timer:        duration,
		duration:     duration,
		damage:       damage,
		boldFont:     boldFont,
		popInEnd:     0.2,
		fadeStart:    0.3,
		popScale:     0.3,
		circleRadius: 0.28,
		fontSize:     2.2,
		borderColor:  color.RGBA{80, 80, 100, 255},
		contentColor: color.RGBA{255, 255, 255, 255},
	}

	switch hitType {
	case HitTypeVenom:
		h.bgColor = color.RGBA{30, 160, 50, 255}
	default:
		h.bgColor = color.RGBA{180, 140, 10, 255}
	}

	return h
}

func (e *HitIndicator) Kind() Kind { return KindHitIndicator }

func (e *HitIndicator) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *HitIndicator) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *HitIndicator) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *HitIndicator) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *HitIndicator) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	if e.hitType == HitTypeDamage && e.damage <= 0 {
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

	cc := e.contentColor
	cc.A = a

	switch e.hitType {
	case HitTypeVenom:
		e.drawSkull(screen, cx, cy, radius, scaleMul, cc, s)
	default:
		e.drawDamageText(screen, cx, cy, scaleMul, cc)
	}
}

func (e *HitIndicator) drawDamageText(screen *ebiten.Image, cx, cy float32, scaleMul float64, tc color.RGBA) {
	fontSize := e.boldFont.Size * e.fontSize * scaleMul
	face := *e.boldFont
	face.Size = fontSize

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(cx), float64(cy))
	op.ColorScale.ScaleWithColor(tc)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	text.Draw(screen, "-"+strconv.Itoa(e.damage), &face, op)
}

func (e *HitIndicator) drawSkull(screen *ebiten.Image, cx, cy, radius float32, scaleMul float64, tc color.RGBA, scale float64) {
	r := float64(radius) * 0.55

	// Skull head (circle).
	headR := float32(r)
	headY := cy - float32(r*0.15)
	vector.FillCircle(screen, cx, headY, headR, tc, true)

	// Eye sockets.
	eyeR := float32(r * 0.2)
	eyeY := headY - float32(r*0.1)
	eyeSpread := float32(r * 0.35)
	eyeClr := e.bgColor
	eyeClr.A = tc.A
	vector.FillCircle(screen, cx-eyeSpread, eyeY, eyeR, eyeClr, true)
	vector.FillCircle(screen, cx+eyeSpread, eyeY, eyeR, eyeClr, true)

	// Nose (small triangle approximated by a tiny circle).
	noseY := headY + float32(r*0.15)
	noseR := float32(r * 0.1)
	vector.FillCircle(screen, cx, noseY, noseR, eyeClr, true)

	// Jaw — a rounded rectangle below the head.
	jawW := float32(r * 1.2)
	jawH := float32(r * 0.45)
	jawY := headY + float32(r*0.55)
	// Draw jaw with stroke lines to suggest teeth.
	sw := float32(math.Max(1.0, scale))
	vector.StrokeRect(screen, cx-jawW/2, jawY, jawW, jawH, sw, tc, true)

	// Teeth: 3 vertical lines across the jaw.
	teethGap := jawW / 4
	for i := float32(1); i <= 3; i++ {
		tx := cx - jawW/2 + teethGap*i
		vector.StrokeLine(screen, tx, jawY, tx, jawY+jawH, sw, tc, true)
	}
}
