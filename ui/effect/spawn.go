package effect

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// SpawnGlow draws a blue glow pillar when a minion is reborn and manages
// the fade-in of the spawning minion's opacity.
type SpawnGlow struct {
	glowTimer    float64
	glowDuration float64
	fadeDuration float64
	fadeStart    float64  // glow progress fraction when fade begins
	opacity      *float64 // pointer to minion's opacity
	spawning     *bool    // pointer to minion's spawning flag
	tick         *int     // pointer to tick counter for sparkle rotation

	pillarColor color.RGBA
	innerColor  color.RGBA
	glowColor   color.RGBA
	sparkColor  color.RGBA
	sparkCount  int
}

var _ Effect = (*SpawnGlow)(nil)

func NewSpawnGlow(
	glowDuration, fadeDuration, fadeStart float64,
	opacity *float64,
	spawning *bool,
	tick *int,
) *SpawnGlow {
	return &SpawnGlow{
		glowTimer:    glowDuration,
		glowDuration: glowDuration,
		fadeDuration: fadeDuration,
		fadeStart:    fadeStart,
		opacity:      opacity,
		spawning:     spawning,
		tick:         tick,
		pillarColor:  color.RGBA{60, 120, 255, 255},
		innerColor:   color.RGBA{120, 180, 255, 255},
		glowColor:    color.RGBA{100, 160, 255, 255},
		sparkColor:   color.RGBA{180, 220, 255, 255},
		sparkCount:   6,
	}
}

func (e *SpawnGlow) Kind() Kind { return KindSpawnGlow }

func (e *SpawnGlow) Update(elapsed float64) bool {
	if e.glowTimer > 0 {
		e.glowTimer -= elapsed
		if e.glowTimer < 0 {
			e.glowTimer = 0
		}
	}

	// Fade in opacity once glow is partially done.
	glowProgress := 1.0 - e.glowTimer/e.glowDuration
	if glowProgress >= e.fadeStart {
		*e.opacity += elapsed / e.fadeDuration
		if *e.opacity >= 1.0 {
			*e.opacity = 1.0
			*e.spawning = false
		}
	}

	return e.glowTimer <= 0 && *e.opacity >= 1.0
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
		angle := float64(j)*math.Pi*2/float64(e.sparkCount) + float64(*e.tick)*0.08
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
