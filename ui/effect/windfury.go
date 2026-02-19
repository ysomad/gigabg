package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// Windfury draws two crossing orbits of wind ribbons wrapping the minion in 3D.
// Unlike transient effects, Windfury is persistent — Update never returns true.
// The caller is responsible for removing it when the keyword is lost.
type Windfury struct {
	angle float64 // current rotation angle in radians
	Alpha uint8   // card alpha, set by the caller before each draw
}

var _ Effect = (*Windfury)(nil)

func NewWindfury() *Windfury {
	return &Windfury{Alpha: 255}
}

func (e *Windfury) Kind() Kind { return KindWindfury }

// Update advances the rotation angle. Never completes (returns false).
func (e *Windfury) Update(elapsed float64) bool {
	e.angle += elapsed * 2.4
	return false
}

func (e *Windfury) Progress() float64                                    { return 0 }
func (e *Windfury) Modify(*ui.Rect, *uint8, *float64)                   {}
func (e *Windfury) DrawBehind(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	e.draw(screen, res, rect, false)
}
func (e *Windfury) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	e.draw(screen, res, rect, true)
}

func (e *Windfury) draw(screen *ebiten.Image, res ui.Resolution, rect ui.Rect, front bool) {
	sr := rect.Screen(res)
	s := res.Scale()

	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	ry := float32(sr.H / 2)

	orbitR := float64(ry)*1.05 + 1*s
	const tilt = 0.3
	const deg39 = 39.0 * math.Pi / 180.0

	const streaksPerOrbit = 2
	const arcLen = math.Pi * 0.55
	const steps = 36
	waveAmp := 1.5 * s
	maxHalf := 1.8 * s

	var a uint8
	if front {
		a = uint8(float64(e.Alpha) * 0.18)
	} else {
		a = uint8(float64(e.Alpha) * 0.05)
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
		orbAngle := e.angle * orb.dir

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
