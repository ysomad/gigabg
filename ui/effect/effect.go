package effect

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/ysomad/gigabg/ui"
)

// Kind identifies an effect type for querying.
type Kind uint8

const (
	KindFlash Kind = iota + 1
	KindShake
	KindHitDamage
	KindPoisonDrip
	KindShieldBreak
	KindDeathFade
	KindDeathTint
	KindSpawnGlow
)

// Effect is a visual effect attached to a combat minion.
//
// Lifecycle: created, added to a minion's effect list, updated each frame,
// drawn each frame, and removed when Update returns true.
type Effect interface {
	// Kind returns the effect's type for querying (e.g. "has shake?").
	Kind() Kind

	// Update advances state by elapsed seconds. Returns true when done.
	Update(elapsed float64) bool

	// Modify lets the effect alter card draw parameters before rendering.
	// Effects that don't modify draw params leave all pointers unchanged.
	Modify(rect *ui.Rect, alpha *uint8, flashPct *float64)

	// DrawBehind draws the effect layer behind the minion card.
	DrawBehind(screen *ebiten.Image, res ui.Resolution, rect ui.Rect)

	// DrawFront draws the effect layer on top of the minion card.
	DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect)
}

// List is a list of active effects on a minion.
type List []Effect

// Add appends an effect.
func (l *List) Add(e Effect) {
	*l = append(*l, e)
}

// Update ticks all effects and removes finished ones.
func (l *List) Update(elapsed float64) {
	n := 0
	for i := range *l {
		if !(*l)[i].Update(elapsed) {
			(*l)[n] = (*l)[i]
			n++
		}
	}
	*l = (*l)[:n]
}

// Has returns true if any effect of the given kind is active.
func (l List) Has(kind Kind) bool {
	for _, e := range l {
		if e.Kind() == kind {
			return true
		}
	}
	return false
}

// HasAny returns true if any effect is active.
func (l List) HasAny() bool {
	return len(l) > 0
}

// Modify applies all effect modifications to draw params.
func (l List) Modify(rect *ui.Rect, alpha *uint8, flashPct *float64) {
	for _, e := range l {
		e.Modify(rect, alpha, flashPct)
	}
}

// DrawBehind draws all behind-card effect layers.
func (l List) DrawBehind(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	for _, e := range l {
		e.DrawBehind(screen, res, rect)
	}
}

// DrawFront draws all front-of-card effect layers.
func (l List) DrawFront(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	for _, e := range l {
		e.DrawFront(screen, res, rect)
	}
}
