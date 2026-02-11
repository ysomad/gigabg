package widget

import (
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/internal/ui"
)

type TextInput struct {
	Rect    ui.Rect // base coords
	MaxLen  int
	text    string
	focused bool
}

func (t *TextInput) Update() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		t.focused = t.Rect.Contains(mx, my)
	}

	if !t.focused {
		return
	}

	chars := ebiten.AppendInputChars(nil)

	if t.handlePaste() {
		return
	}

	for _, r := range chars {
		if len(t.text) < t.MaxLen {
			t.text += string(r)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(t.text) > 0 {
		_, size := utf8.DecodeLastRuneInString(t.text)
		t.text = t.text[:len(t.text)-size]
	}
}

func (t *TextInput) handlePaste() bool {
	mod := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if !mod || !inpututil.IsKeyJustPressed(ebiten.KeyV) {
		return false
	}

	clip := readClipboard()
	if clip == "" {
		return false
	}

	remaining := t.MaxLen - len(t.text)
	if remaining <= 0 {
		return true
	}

	if len(clip) > remaining {
		clip = clip[:remaining]
	}
	t.text += clip
	return true
}

func (t *TextInput) Draw(screen *ebiten.Image, font *text.GoTextFace) {
	sr := t.Rect.Screen()
	sw := float32(2 * ui.ActiveRes.Scale())

	borderClr := color.RGBA{60, 60, 80, 255}
	if t.focused {
		borderClr = color.RGBA{100, 150, 255, 255}
	}

	vector.FillRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), color.RGBA{40, 40, 50, 255}, true)
	vector.StrokeRect(screen, float32(sr.X), float32(sr.Y), float32(sr.W), float32(sr.H), sw, borderClr, false)

	display := t.text
	if t.focused && len(display) < t.MaxLen {
		display += "_"
	}

	padX := t.Rect.W * 0.04
	padY := t.Rect.H * 0.25
	ui.DrawText(screen, font, display, t.Rect.X+padX, t.Rect.Y+padY, color.White)
}

func (t *TextInput) Value() string    { return t.text }
func (t *TextInput) Focused() bool    { return t.focused }
func (t *TextInput) SetFocused(b bool) { t.focused = b }
