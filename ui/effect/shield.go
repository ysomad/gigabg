package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// ShieldBreak draws a golden shard burst when divine shield is removed.
type ShieldBreak struct {
	timer    float64
	duration float64

	shardCount int
	ringColor  color.RGBA
	shardColor color.RGBA
	flashColor color.RGBA
}

var _ Effect = (*ShieldBreak)(nil)

func NewShieldBreak(duration float64) *ShieldBreak {
	return &ShieldBreak{
		timer:      duration,
		duration:   duration,
		shardCount: 8,
		ringColor:  color.RGBA{255, 215, 0, 255},
		shardColor: color.RGBA{255, 215, 0, 255},
		flashColor: color.RGBA{255, 255, 200, 255},
	}
}

func (e *ShieldBreak) Kind() Kind { return KindShieldBreak }

func (e *ShieldBreak) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *ShieldBreak) Modify(*ui.Rect, *uint8, *float64)               {}
func (e *ShieldBreak) DrawFront(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *ShieldBreak) DrawBehind(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	sr := rect.Screen(res)
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := float32(res.Scale())

	t := 1.0 - e.timer/e.duration // 0 -> 1

	// Expanding golden ring.
	ringR := float32(sr.W/2) * float32(1.0+t*0.8)
	ringA := uint8(255 * (1.0 - t) * 0.6)
	ringW := s * 3 * float32(1.0-t)
	if ringW < 0.5 {
		ringW = 0.5
	}
	rc := e.ringColor
	rc.A = ringA
	ui.StrokeEllipse(screen, cx, cy, ringR, ringR*1.2, ringW, rc)

	// Shard particles flying outward.
	for j := range e.shardCount {
		angle := float64(j) * 2 * math.Pi / float64(e.shardCount)
		dist := float64(sr.W/2) * (0.3 + t*1.2)
		sx := cx + float32(math.Cos(angle)*dist)
		sy := cy + float32(math.Sin(angle)*dist*0.8)

		shardSize := s * 4 * float32(1.0-t*0.7)
		shardA := uint8(255 * (1.0 - t))

		var shard vector.Path
		shard.MoveTo(sx, sy-shardSize)
		shard.LineTo(sx+shardSize*0.5, sy)
		shard.LineTo(sx, sy+shardSize)
		shard.LineTo(sx-shardSize*0.5, sy)
		shard.Close()

		sc := e.shardColor
		sc.A = shardA
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(sc)
		vector.FillPath(screen, &shard, nil, op)
	}

	// Central flash.
	flashA := uint8(255 * (1.0 - t) * (1.0 - t))
	flashR := float32(sr.W/2) * float32(0.5+t*0.3)
	fc := e.flashColor
	fc.A = flashA
	vector.FillCircle(screen, cx, cy, flashR, fc, true)
}
