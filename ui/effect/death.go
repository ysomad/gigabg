package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/ui"
)

// DeathFade tracks a dying minion's fade-out progress.
// The combatboard reads Progress to compute opacity (1 - progress).
type DeathFade struct {
	timer    float64
	duration float64
}

var _ Effect = (*DeathFade)(nil)

func NewDeathFade(duration float64) *DeathFade {
	return &DeathFade{
		duration: duration,
		timer:    duration,
	}
}

func (e *DeathFade) Kind() Kind { return KindDeathFade }

func (e *DeathFade) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *DeathFade) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *DeathFade) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *DeathFade) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}
func (e *DeathFade) DrawFront(*ebiten.Image, ui.Resolution, ui.Rect)  {}

// DeathTint draws a colored tint overlay on a dying minion.
// Progress tracks how far the tint has advanced; the combatboard passes
// the minion's current opacity when drawing.
type DeathTint struct {
	timer     float64
	duration  float64
	tintColor color.RGBA
	threshold float64
}

var _ Effect = (*DeathTint)(nil)

func NewDeathTint(duration float64, deathReason game.DeathReason) *DeathTint {
	tint := color.RGBA{180, 30, 30, 0}
	if deathReason == game.DeathReasonVenom {
		tint = color.RGBA{30, 160, 50, 0}
	}
	return &DeathTint{
		timer:     duration,
		duration:  duration,
		tintColor: tint,
		threshold: 0.3,
	}
}

func (e *DeathTint) Kind() Kind { return KindDeathTint }

func (e *DeathTint) Update(elapsed float64) bool {
	e.timer -= elapsed
	if e.timer < 0 {
		e.timer = 0
	}
	return e.timer <= 0
}

func (e *DeathTint) Progress() float64 {
	return 1.0 - e.timer/e.duration
}

func (e *DeathTint) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *DeathTint) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

// DrawFront draws the tint overlay. The tint alpha is derived from progress.
func (e *DeathTint) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	opacity := 1.0 - e.Progress()
	if opacity <= e.threshold {
		return
	}

	sr := rect.Screen(res)
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	rx := float32(sr.W / 2)
	ry := float32(sr.H / 2)

	const k = 0.5522847498
	var path vector.Path
	path.MoveTo(cx, cy-ry)
	path.CubicTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy)
	path.CubicTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry)
	path.CubicTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy)
	path.CubicTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry)
	path.Close()

	a := uint8(float64(80) * e.Progress())
	clr := e.tintColor
	clr.A = a
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(screen, &path, nil, op)
}

// Particles manages particle burst on minion death. Unlike minion-attached
// effects, this lives on the combatBoard level since particles persist after
// the minion is removed.
type Particles struct {
	particles []particle
	duration  float64
}

type particle struct {
	x, y   float64
	vx, vy float64
	life   float64
	size   float64
	clr    color.RGBA
}

var _ Effect = (*Particles)(nil)

func NewParticles(cx, cy float64, count int, baseClr color.RGBA, duration float64) *Particles {
	pp := make([]particle, count)
	for j := range count {
		angle := float64(j) * 2 * math.Pi / float64(count)
		speed := 40.0 + float64(j%3)*20.0
		pp[j] = particle{
			x:    cx,
			y:    cy,
			vx:   math.Cos(angle) * speed,
			vy:   math.Sin(angle)*speed - 20,
			life: 1.0,
			size: 2.0 + float64(j%3),
			clr:  baseClr,
		}
	}
	return &Particles{particles: pp, duration: duration}
}

func (e *Particles) Kind() Kind { return Kind(0) }

func (e *Particles) Update(elapsed float64) bool {
	decay := elapsed / e.duration
	n := 0
	for i := range e.particles {
		e.particles[i].x += e.particles[i].vx * elapsed
		e.particles[i].y += e.particles[i].vy * elapsed
		e.particles[i].vy += 80 * elapsed
		e.particles[i].life -= decay
		if e.particles[i].life > 0 {
			e.particles[n] = e.particles[i]
			n++
		}
	}
	e.particles = e.particles[:n]
	return len(e.particles) == 0
}

func (e *Particles) Progress() float64 { return 0 }

func (e *Particles) Modify(*ui.Rect, *uint8, *float64)                {}
func (e *Particles) DrawBehind(*ebiten.Image, ui.Resolution, ui.Rect) {}

// DrawFront draws particles at absolute positions (rect is ignored).
func (e *Particles) DrawFront(screen *ebiten.Image, res ui.Resolution, _ ui.Rect) {
	s := res.Scale()
	for _, p := range e.particles {
		sx := float32(p.x*s + res.OffsetX())
		sy := float32(p.y*s + res.OffsetY())
		size := float32(p.size * s)
		a := uint8(255.0 * p.life)
		clr := p.clr
		clr.A = a
		vector.FillCircle(screen, sx, sy, size, clr, true)
	}
}
