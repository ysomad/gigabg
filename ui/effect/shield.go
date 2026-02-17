package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// DivineShieldBreak draws a golden shard burst when divine shield is removed.
type DivineShieldBreak struct {
	timer    float64
	duration float64

	shardCount int
	shardSize  float64
	flashColor color.RGBA
}

var _ Effect = (*DivineShieldBreak)(nil)

func NewDivineShieldBreak(duration float64) *DivineShieldBreak {
	return &DivineShieldBreak{
		timer:      duration,
		duration:   duration,
		shardCount: 8,
		shardSize:  4.0,
		flashColor: color.RGBA{255, 215, 0, 255},
	}
}

func (e *DivineShieldBreak) Kind() Kind { return KindDivineShieldBreak }

func (e *DivineShieldBreak) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *DivineShieldBreak) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *DivineShieldBreak) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *DivineShieldBreak) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *DivineShieldBreak) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	sr := rect.Screen(res)
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := float32(res.Scale())

	t := e.Progress()

	// Central golden flash (fades quickly).
	if t < 0.3 {
		flashA := uint8(255 * (1.0 - t/0.3))
		flashR := s * 15 * float32(1.0+t*3)
		fc := e.flashColor
		fc.A = flashA
		vector.FillCircle(screen, cx, cy, flashR, fc, true)
	}

	// Golden shards flying outward.
	for j := range e.shardCount {
		angle := float64(j) * 2 * math.Pi / float64(e.shardCount)
		dist := float64(s) * (5.0 + t*40.0)
		sx := cx + float32(math.Cos(angle)*dist)
		sy := cy + float32(math.Sin(angle)*dist)

		shardR := s * float32(e.shardSize) * float32(1.0-t)
		shardA := uint8(255 * (1.0 - t))

		sc := e.flashColor
		sc.A = shardA
		vector.FillCircle(screen, sx, sy, shardR, sc, true)
	}
}
