package ui

import (
	"bytes"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const baseFontSize = 14

// Overlay is drawn on top of the current scene and blocks scene input.
type Overlay interface {
	Update()
	Draw(screen *ebiten.Image)
}

// App is the root ebiten.Game that owns the SceneManager and font.
type App struct {
	sm         SceneManager
	overlay    Overlay
	fontSource *text.GoTextFaceSource
	font       *text.GoTextFace
}

func NewApp() (*App, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return nil, err
	}
	font := &text.GoTextFace{
		Source: src,
		Size:   baseFontSize * ActiveRes.Scale(),
	}
	return &App{
		fontSource: src,
		font:       font,
	}, nil
}

// Font returns the shared font face. Size is updated each frame.
func (a *App) Font() *text.GoTextFace { return a.font }

// SwitchScene transitions to a new scene.
func (a *App) SwitchScene(s Scene) { a.sm.Switch(s) }

// ShowOverlay displays an overlay on top of the current scene.
func (a *App) ShowOverlay(o Overlay) { a.overlay = o }

// HideOverlay removes the current overlay.
func (a *App) HideOverlay() { a.overlay = nil }

func (a *App) Update() error {
	a.font.Size = baseFontSize * ActiveRes.Scale()
	if a.overlay != nil {
		a.overlay.Update()
		return nil
	}
	return a.sm.Update()
}

func (a *App) Draw(screen *ebiten.Image) {
	a.sm.Draw(screen)
	if a.overlay != nil {
		a.overlay.Draw(screen)
	}
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	ActiveRes.Width = outsideWidth
	ActiveRes.Height = outsideHeight
	return outsideWidth, outsideHeight
}
