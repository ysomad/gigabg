package effect

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/ui"
)

// Flash is a brief white flash on a minion body after taking damage.
type Flash struct {
	timer    float64
	duration float64
}

var _ Effect = (*Flash)(nil)

func NewFlash(duration float64) *Flash {
	return &Flash{timer: duration, duration: duration}
}

func (e *Flash) Kind() Kind { return KindFlash }

func (e *Flash) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

// Modify sets flashPct so CardRenderer.DrawMinion renders the white overlay.
func (e *Flash) Modify(_ *ui.Rect, _ *uint8, flashPct *float64) {
	if flashPct != nil {
		*flashPct = e.timer / e.duration
	}
}

func (e *Flash) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}
func (e *Flash) DrawFront(*ebiten.Image, ui.Resolution, ui.Rect)  {}
