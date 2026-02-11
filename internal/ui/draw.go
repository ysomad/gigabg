package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var ColorBackground = color.RGBA{20, 20, 30, 255}

// DrawText draws text at base coordinates, converting to screen space internally.
func DrawText(screen *ebiten.Image, font *text.GoTextFace, str string, baseX, baseY float64, clr color.Color) {
	if font == nil {
		return
	}
	s := ActiveRes.Scale()
	op := &text.DrawOptions{}
	op.GeoM.Translate(baseX*s+ActiveRes.OffsetX(), baseY*s+ActiveRes.OffsetY())
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, font, op)
}

// FillScreen fills the entire actual screen (for fullscreen overlays).
func FillScreen(screen *ebiten.Image, clr color.Color) {
	vector.FillRect(screen, 0, 0, float32(ActiveRes.Width), float32(ActiveRes.Height), clr, false)
}

// FillEllipse draws a filled ellipse at screen-space center (cx, cy) with radii (rx, ry).
func FillEllipse(screen *ebiten.Image, cx, cy, rx, ry float32, clr color.Color) {
	const k = 0.5522847498 // 4/3 * (sqrt(2) - 1)

	var path vector.Path
	path.MoveTo(cx, cy-ry)
	path.CubicTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy)
	path.CubicTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry)
	path.CubicTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy)
	path.CubicTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry)
	path.Close()

	op := &vector.DrawPathOptions{}
	op.ColorScale.ScaleWithColor(clr)
	vector.FillPath(screen, &path, nil, op)
}

// StrokeEllipse draws an ellipse outline at screen-space center (cx, cy) with radii (rx, ry).
func StrokeEllipse(screen *ebiten.Image, cx, cy, rx, ry, strokeWidth float32, clr color.Color) {
	const k = 0.5522847498

	var path vector.Path
	path.MoveTo(cx, cy-ry)
	path.CubicTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy)
	path.CubicTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry)
	path.CubicTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy)
	path.CubicTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry)
	path.Close()

	op := &vector.DrawPathOptions{}
	op.ColorScale.ScaleWithColor(clr)
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: strokeWidth}, op)
}

// EaseOut gives a smooth deceleration curve.
func EaseOut(t float64) float64 {
	return 1 - math.Pow(1-t, 2)
}
