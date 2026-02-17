package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

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
