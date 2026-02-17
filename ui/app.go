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
	Update(res Resolution)
	Draw(screen *ebiten.Image, res Resolution)
}

// App is the root ebiten.Game that owns the SceneManager and font.
type App struct {
	sm             SceneManager
	overlay        Overlay
	fontSource     *text.GoTextFaceSource
	boldFontSource *text.GoTextFaceSource
	font           *text.GoTextFace
	boldFont       *text.GoTextFace
	res            Resolution
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
	res := Resolution{BaseWidth, BaseHeight}
	sz := baseFontSize * res.Scale()
	return &App{
		fontSource:     src,
		boldFontSource: boldSrc,
		font:           &text.GoTextFace{Source: src, Size: sz},
		boldFont:       &text.GoTextFace{Source: boldSrc, Size: sz},
		res:            res,
	}, nil
}

// Font returns the shared font face. Size is updated each frame.
func (a *App) Font() *text.GoTextFace { return a.font }

// BoldFont returns the shared bold font face. Size is updated each frame.
func (a *App) BoldFont() *text.GoTextFace { return a.boldFont }

// Res returns the current resolution.
func (a *App) Res() Resolution { return a.res }

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
	a.font.Size = baseFontSize * a.res.Scale()
	a.boldFont.Size = baseFontSize * a.res.Scale()
	if a.overlay != nil {
		a.overlay.Update(a.res)
		return nil
	}
	return a.sm.Update(a.res)
}

func (a *App) Draw(screen *ebiten.Image) {
	a.sm.Draw(screen, a.res)
	if a.overlay != nil {
		a.overlay.Draw(screen, a.res)
	}
	if a.debug {
		a.drawDebug(screen)
	}
}

func (a *App) drawDebug(screen *ebiten.Image) {
	cx, cy := ebiten.CursorPosition()
	bx, by := a.res.ScreenToBase(cx, cy)

	ww, wh := ebiten.WindowSize()
	var dbg ebiten.DebugInfo
	ebiten.ReadDebugInfo(&dbg)

	msg := fmt.Sprintf(
		"TPS: %.1f  FPS: %.1f\nWindow: %dx%d  Scale: %.2f\nCursor: %d,%d  Base: %.0f,%.0f\nGPU: %s  VRAM: %d KB",
		ebiten.ActualTPS(), ebiten.ActualFPS(),
		ww, wh, a.res.Scale(),
		cx, cy, bx, by,
		dbg.GraphicsLibrary, dbg.TotalGPUImageMemoryUsageInBytes/1024,
	)
	ebitenutil.DebugPrintAt(screen, msg, 4, a.res.Height-64)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	a.res.Width = outsideWidth
	a.res.Height = outsideHeight
	return outsideWidth, outsideHeight
}
