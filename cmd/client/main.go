package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/ysomad/gigabg/client"
	"github.com/ysomad/gigabg/game/cards"
	"github.com/ysomad/gigabg/ui"
)

type state uint8

const (
	stateMenu state = iota
	stateConnecting
	stateGame
)

type ClientApp struct {
	state      state
	serverAddr string

	fontSource *text.GoTextFaceSource
	font       *text.GoTextFace
	cards      *cards.Cards
	menu       *ui.Menu
	game       *ui.GameScene
	client     *client.Client
	err        error
}

func main() {
	cardStore, err := cards.New()
	if err != nil {
		log.Fatal(err)
	}

	app := &ClientApp{
		state:      stateMenu,
		serverAddr: "ws://localhost:8080",
		cards:      cardStore,
	}
	app.loadFont()
	app.menu = ui.NewMenu(app.font, app.onJoin)

	ebiten.SetWindowSize(ui.ActiveRes.Width, ui.ActiveRes.Height)
	ebiten.SetWindowTitle("GIGA Battlegrounds")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}

func (a *ClientApp) loadFont() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return
	}
	a.fontSource = src
	a.font = &text.GoTextFace{
		Source: src,
		Size:   14 * ui.ActiveRes.Scale(),
	}
}

func (a *ClientApp) updateFontSize() {
	if a.fontSource == nil {
		return
	}
	a.font.Size = 14 * ui.ActiveRes.Scale()
}

func (a *ClientApp) onJoin(lobbyID string) {
	a.state = stateConnecting
	go a.connect(lobbyID)
}

func (a *ClientApp) connect(lobbyID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.New(ctx, a.serverAddr, lobbyID)
	if err != nil {
		a.err = err
		a.state = stateMenu
		return
	}

	if err := c.WaitForState(ctx); err != nil {
		c.Close()
		a.err = err
		a.state = stateMenu
		return
	}

	a.client = c
	a.game = ui.NewGameScene(c, a.cards, a.font)
	a.state = stateGame
}

func (a *ClientApp) Update() error {
	a.updateFontSize()

	switch a.state {
	case stateMenu:
		a.menu.Update()
	case stateGame:
		if a.game != nil {
			return a.game.Update()
		}
	}
	return nil
}

func (a *ClientApp) Draw(screen *ebiten.Image) {
	switch a.state {
	case stateMenu:
		a.menu.Draw(screen)
	case stateConnecting:
		a.drawConnecting(screen)
	case stateGame:
		if a.game != nil {
			a.game.Draw(screen)
		}
	}
}

func (a *ClientApp) drawConnecting(screen *ebiten.Image) {
	screen.Fill(ui.ColorBackground)
	w := float64(ui.ActiveRes.Width)
	h := float64(ui.ActiveRes.Height)
	ui.DrawTextAt(screen, a.font, "Connecting...", w/2-50, h/2)
}

func (a *ClientApp) Layout(outsideWidth, outsideHeight int) (int, int) {
	ui.ActiveRes.Width = outsideWidth
	ui.ActiveRes.Height = outsideHeight
	return outsideWidth, outsideHeight
}
