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
func (t *Toast) Update() {
	if t.timer > 0 {
		t.timer -= 1.0 / 60.0
	}
}

// Draw renders the toast as a colored rectangle with centered text.
func (t *Toast) Draw(screen *ebiten.Image) {
	if t.timer <= 0 {
		return
	}

	boxW := float64(ui.BaseWidth) * 0.3
	boxH := float64(ui.BaseHeight) * 0.1
	boxX := float64(ui.BaseWidth)/2 - boxW/2
	boxY := float64(ui.BaseHeight)/2 - boxH/2

	r := ui.Rect{X: boxX, Y: boxY, W: boxW, H: boxH}
	sr := r.Screen()
	s := ui.ActiveRes.Scale()

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

	ui.DrawText(screen, t.font, t.text, boxX+boxW*0.5-float64(len(t.text))*4, boxY+boxH*0.3, color.White)
}

// Popup is a generic modal overlay that displays a title, message, and an optional button.
type Popup struct {
	font *text.GoTextFace

	mu      sync.Mutex
	title   string
	message string
	btn     *Button
}

func NewPopup(font *text.GoTextFace, title, message string) *Popup {
	return &Popup{
		font:    font,
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

// popupBox returns the base-space rect for the popup box.
func popupBox() (boxX, boxY, boxW, boxH float64) {
	boxW = float64(ui.BaseWidth) * 0.40
	boxH = float64(ui.BaseHeight) * 0.25
	boxX = float64(ui.BaseWidth)/2 - boxW/2
	boxY = float64(ui.BaseHeight)/2 - boxH/2
	return
}

// ShowButton shows a button at the bottom of the popup.
func (p *Popup) ShowButton(text string, onClick func()) {
	p.mu.Lock()
	defer p.mu.Unlock()

	boxX, boxY, boxW, boxH := popupBox()

	btnW := boxW * 0.22
	btnH := boxH * 0.20
	p.btn = &Button{
		Rect:      ui.Rect{X: boxX + boxW/2 - btnW/2, Y: boxY + boxH*0.72, W: btnW, H: btnH},
		Text:      text,
		Color:     color.RGBA{120, 50, 50, 255},
		BorderClr: color.RGBA{160, 80, 80, 255},
		TextClr:   color.RGBA{255, 255, 255, 255},
		OnClick:   onClick,
	}
}

func (p *Popup) Update() {
	p.mu.Lock()
	btn := p.btn
	p.mu.Unlock()

	if btn != nil {
		btn.Update()
	}
}

func (p *Popup) Draw(screen *ebiten.Image) {
	// Semi-transparent overlay covering actual screen.
	ui.FillScreen(screen, color.RGBA{0, 0, 0, 160})

	// Centered box in base coords.
	boxX, boxY, boxW, boxH := popupBox()
	boxRect := ui.Rect{X: boxX, Y: boxY, W: boxW, H: boxH}
	sr := boxRect.Screen()
	sw := float32(ui.ActiveRes.Scale())

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
		ui.DrawText(screen, p.font, title, boxX+boxW*0.05, boxY+boxH*0.15, color.RGBA{220, 200, 60, 255})
	}

	ui.DrawText(screen, p.font, message, boxX+boxW*0.05, boxY+boxH*0.45, color.RGBA{200, 200, 200, 255})

	if btn != nil {
		btn.Draw(screen, p.font)
	}
}
