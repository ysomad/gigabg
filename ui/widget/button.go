package widget

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

type Button struct {
	Rect      ui.Rect // base coords
	Text      string
	Color     color.RGBA
	BorderClr color.RGBA
	TextClr   color.RGBA
	OnClick   func()
}

func (b *Button) Update() {
	if b.OnClick == nil {
		return
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if b.Rect.Contains(mx, my) {
			b.OnClick()
		}
	}
}

func (b *Button) Draw(screen *ebiten.Image, font *text.GoTextFace) {
	sr := b.Rect.Screen()
	sw := float32(ui.ActiveRes.Scale())

	vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), b.Color, false)
	vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, b.BorderClr, false)

	padX := b.Rect.W * 0.08
	padY := b.Rect.H * 0.25
	ui.DrawText(screen, font, b.Text, b.Rect.X+padX, b.Rect.Y+padY, b.TextClr)
}
