package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// SpawnGlow draws a blue glow pillar when a minion is reborn.
// The combatboard reads Progress to compute the spawning minion's opacity
// and clears the spawning flag when the effect completes.
type SpawnGlow struct {
	glowTimer    float64
	glowDuration float64
	fadeStart    float64 // glow progress fraction when opacity fade-in begins
	tick         int     // internal tick counter for sparkle rotation

	pillarColor color.RGBA
	innerColor  color.RGBA
	glowColor   color.RGBA
	sparkColor  color.RGBA
	sparkCount  int
}

var _ Effect = (*SpawnGlow)(nil)

func NewSpawnGlow(glowDuration, fadeStart float64) *SpawnGlow {
	return &SpawnGlow{
		glowTimer:    glowDuration,
		glowDuration: glowDuration,
		fadeStart:    fadeStart,
		pillarColor:  color.RGBA{60, 120, 255, 255},
		innerColor:   color.RGBA{120, 180, 255, 255},
		glowColor:    color.RGBA{100, 160, 255, 255},
		sparkColor:   color.RGBA{180, 220, 255, 255},
		sparkCount:   6,
	}
}

func (e *SpawnGlow) Kind() Kind { return KindSpawnGlow }

func (e *SpawnGlow) Update(elapsed float64) bool {
	e.tick++
	e.glowTimer -= elapsed
	if e.glowTimer < 0 {
		e.glowTimer = 0
	}
	return e.glowTimer <= 0
}

// Progress returns the glow completion fraction (0→1).
func (e *SpawnGlow) Progress() float64 {
	return 1.0 - e.glowTimer/e.glowDuration
}

// FadeProgress returns the opacity fade-in fraction (0→1).
// Returns 0 until glow progress reaches fadeStart, then ramps to 1.
func (e *SpawnGlow) FadeProgress() float64 {
	p := e.Progress()
	if p < e.fadeStart {
		return 0
	}
	fp := (p - e.fadeStart) / (1.0 - e.fadeStart)
	if fp > 1.0 {
		return 1.0
	}
	return fp
}

func (e *SpawnGlow) Modify(*ui.Rect, *uint8, *float64)               {}
func (e *SpawnGlow) DrawFront(*ebiten.Image, ui.Resolution, ui.Rect) {}

func (e *SpawnGlow) DrawBehind(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	if e.glowTimer <= 0 {
		return
	}

	sr := rect.Screen(res)
	cx := float32(sr.X + sr.W/2)
	cy := float32(sr.Y + sr.H/2)
	s := float32(res.Scale())

	t := float32(e.glowTimer / e.glowDuration) // 1 -> 0

	// Vertical light pillar.
	pillarW := float32(sr.W) * 0.3 * t
	pillarH := float32(sr.H) * 1.5
	pillarA := uint8(80 * t)
	pc := e.pillarColor
	pc.A = pillarA
	vector.FillRect(screen, cx-pillarW/2, cy-pillarH/2, pillarW, pillarH, pc, false)

	// Inner brighter pillar.
	innerW := pillarW * 0.4
	innerA := uint8(140 * t)
	ic := e.innerColor
	ic.A = innerA
	vector.FillRect(screen, cx-innerW/2, cy-pillarH/2, innerW, pillarH, ic, false)

	// Glow circle at center.
	glowR := s * 25 * t
	glowA := uint8(float32(100) * t * t)
	gc := e.glowColor
	gc.A = glowA
	vector.FillCircle(screen, cx, cy, glowR, gc, true)

	// Small sparkle particles around the pillar.
	for j := range e.sparkCount {
		angle := float64(j)*math.Pi*2/float64(e.sparkCount) + float64(e.tick)*0.08
		dist := float64(sr.W/2) * 0.5 * float64(t)
		sx := cx + float32(math.Cos(angle)*dist)
		sy := cy + float32(math.Sin(angle)*dist*1.3) - s*10*(1.0-t)
		sparkR := s * 2 * t
		sparkA := uint8(200 * t)
		sc := e.sparkColor
		sc.A = sparkA
		vector.FillCircle(screen, sx, sy, sparkR, sc, true)
	}
}
