package widget

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/ysomad/gigabg/ui"
)

// Toast shows centered text that auto-fades after a duration.
type Toast struct {
	font  *text.GoTextFace
	text  string
	timer float64
}

const toastDuration = 1

func NewToast(font *text.GoTextFace) *Toast {
	return &Toast{font: font}
}

// Show displays the toast with the given text.
func (t *Toast) Show(text string) {
	t.text = text
	t.timer = toastDuration
}

// Active returns true if the toast is currently visible.
func (t *Toast) Active() bool {
	return t.timer > 0
}

// Update ticks the toast timer.
func (t *Toast) Update(dt float64) {
	if t.timer > 0 {
		t.timer -= dt
	}
}

// Draw renders the toast inside the given rect.
func (t *Toast) Draw(screen *ebiten.Image, res ui.Resolution, rect ui.Rect) {
	if t.timer <= 0 {
		return
	}

	sr := rect.Screen(res)
	s := res.Scale()

	vector.FillRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		color.RGBA{30, 30, 50, 255},
		false,
	)
	vector.StrokeRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		float32(2*s),
		color.RGBA{80, 80, 120, 255},
		false,
	)

	op := &text.DrawOptions{}
	op.GeoM.Translate((rect.X+rect.W*0.5)*s+res.OffsetX(), (rect.Y+rect.H*0.3)*s+res.OffsetY())
	op.ColorScale.ScaleWithColor(color.White)
	op.PrimaryAlign = text.AlignCenter
	text.Draw(screen, t.text, t.font, op)
}

// Popup is a generic modal overlay that displays a title, message, and an optional button.
type Popup struct {
	font *text.GoTextFace
	rect ui.Rect

	mu      sync.Mutex
	title   string
	message string
	btn     *Button
}

func NewPopup(font *text.GoTextFace, rect ui.Rect, title, message string) *Popup {
	return &Popup{
		font:    font,
		rect:    rect,
		title:   title,
		message: message,
	}
}

// SetTitle updates the popup title.
func (p *Popup) SetTitle(s string) {
	p.mu.Lock()
	p.title = s
	p.mu.Unlock()
}

// SetMessage updates the popup message.
func (p *Popup) SetMessage(s string) {
	p.mu.Lock()
	p.message = s
	p.mu.Unlock()
}

// ShowButton shows a button at the bottom of the popup.
func (p *Popup) ShowButton(text string, onClick func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	btnW := p.rect.W * 0.22
	btnH := p.rect.H * 0.20
	p.btn = &Button{
		Rect:      ui.Rect{X: p.rect.X + p.rect.W/2 - btnW/2, Y: p.rect.Y + p.rect.H*0.72, W: btnW, H: btnH},
		Text:      text,
		Color:     color.RGBA{120, 50, 50, 255},
		BorderClr: color.RGBA{160, 80, 80, 255},
		TextClr:   color.RGBA{255, 255, 255, 255},
		OnClick:   onClick,
	}
}

func (p *Popup) Update(res ui.Resolution) {
	p.mu.Lock()
	btn := p.btn
	p.mu.Unlock()

	if btn != nil {
		btn.Update(res)
	}
}

func (p *Popup) Draw(screen *ebiten.Image, res ui.Resolution) {
	// Semi-transparent overlay covering actual screen.
	ui.FillScreen(screen, res, color.RGBA{0, 0, 0, 160})

	sr := p.rect.Screen(res)
	sw := float32(res.Scale())

	vector.FillRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		color.RGBA{30, 30, 40, 255},
		false,
	)
	vector.StrokeRect(
		screen,
		float32(sr.X),
		float32(sr.Y),
		float32(sr.W),
		float32(sr.H),
		sw,
		color.RGBA{80, 80, 100, 255},
		false,
	)

	p.mu.Lock()
	title := p.title
	message := p.message
	btn := p.btn
	p.mu.Unlock()

	if title != "" {
		ui.DrawText(screen, res, p.font, title, p.rect.X+p.rect.W*0.05, p.rect.Y+p.rect.H*0.15, color.RGBA{220, 200, 60, 255})
	}

	ui.DrawText(screen, res, p.font, message, p.rect.X+p.rect.W*0.05, p.rect.Y+p.rect.H*0.45, color.RGBA{200, 200, 200, 255})

	if btn != nil {
		btn.Draw(screen, res, p.font)
	}
}
