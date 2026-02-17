package effect

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/ui"
)

// Shake is a damage shake on a minion body.
type Shake struct {
	timer     float64
	duration  float64
	intensity float64 // shake amplitude
	freq      float64 // shake frequency multiplier
}

var _ Effect = (*Shake)(nil)

func NewShake(duration, intensity, freq float64) *Shake {
	return &Shake{
		timer:     duration,
		duration:  duration,
		intensity: intensity,
		freq:      freq,
	}
}

func (e *Shake) Kind() Kind { return KindShake }

func (e *Shake) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *Shake) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

// Modify offsets the card rect to create a shake effect.
func (e *Shake) Modify(rect *ui.Rect, _ *uint8, _ *float64) {
	if rect == nil || e.timer <= 0 {
		return
	}
	t := e.timer / e.duration
	amp := t * e.intensity
	f := e.timer * e.freq
	rect.X += math.Sin(f) * amp
	rect.Y += math.Cos(f*0.8) * amp * 0.6
}

func (e *Shake) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}
func (e *Shake) DrawFront(*ebiten.Image, ui.Resolution, ui.Rect)  {}
