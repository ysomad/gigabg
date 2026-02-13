package ui

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
)

const baseFontSize = 14

// Overlay is drawn on top of the current scene and blocks scene input.
type Overlay interface {
	Update()
	Draw(screen *ebiten.Image)
}

// App is the root ebiten.Game that owns the SceneManager and font.
type App struct {
	sm             SceneManager
	overlay        Overlay
	fontSource     *text.GoTextFaceSource
	boldFontSource *text.GoTextFaceSource
	font           *text.GoTextFace
	boldFont       *text.GoTextFace
	debug          bool
}

func NewApp() (*App, error) {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return nil, err
	}
	boldSrc, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		return nil, err
	}
	sz := baseFontSize * ActiveRes.Scale()
	return &App{
		fontSource:     src,
		boldFontSource: boldSrc,
		font:           &text.GoTextFace{Source: src, Size: sz},
		boldFont:       &text.GoTextFace{Source: boldSrc, Size: sz},
	}, nil
}

// Font returns the shared font face. Size is updated each frame.
func (a *App) Font() *text.GoTextFace { return a.font }

// BoldFont returns the shared bold font face. Size is updated each frame.
func (a *App) BoldFont() *text.GoTextFace { return a.boldFont }

// SwitchScene transitions to a new scene.
func (a *App) SwitchScene(s Scene) { a.sm.Switch(s) }

// ShowOverlay displays an overlay on top of the current scene.
func (a *App) ShowOverlay(o Overlay) { a.overlay = o }

// HideOverlay removes the current overlay.
func (a *App) HideOverlay() { a.overlay = nil }

// SetDebug enables or disables the debug overlay.
func (a *App) SetDebug(on bool) { a.debug = on }

func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		a.debug = !a.debug
	}
	a.font.Size = baseFontSize * ActiveRes.Scale()
	a.boldFont.Size = baseFontSize * ActiveRes.Scale()
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
	if a.debug {
		a.drawDebug(screen)
	}
}

func (a *App) drawDebug(screen *ebiten.Image) {
	cx, cy := ebiten.CursorPosition()
	s := ActiveRes.Scale()
	bx := (float64(cx) - ActiveRes.OffsetX()) / s
	by := (float64(cy) - ActiveRes.OffsetY()) / s

	ww, wh := ebiten.WindowSize()
	var dbg ebiten.DebugInfo
	ebiten.ReadDebugInfo(&dbg)

	msg := fmt.Sprintf(
		"TPS: %.1f  FPS: %.1f\nWindow: %dx%d  Scale: %.2f\nCursor: %d,%d  Base: %.0f,%.0f\nGPU: %s  VRAM: %d KB",
		ebiten.ActualTPS(), ebiten.ActualFPS(),
		ww, wh, s,
		cx, cy, bx, by,
		dbg.GraphicsLibrary, dbg.TotalGPUImageMemoryUsageInBytes/1024,
	)
	ebitenutil.DebugPrintAt(screen, msg, 4, ActiveRes.Height-64)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	ActiveRes.Width = outsideWidth
	ActiveRes.Height = outsideHeight
	return outsideWidth, outsideHeight
}
