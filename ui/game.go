package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	pb "github.com/ysomad/gigabg/proto"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
)

var ColorBackground = color.RGBA{20, 20, 30, 255}

// StateProvider provides game state to the UI.
type StateProvider interface {
	PlayerID() string
	PlayerCount() int
	Players() []*pb.Player
}

// GameScene renders the game UI.
type GameScene struct {
	state StateProvider
	font  *text.GoTextFace
}

func NewGameScene(state StateProvider, font *text.GoTextFace) *GameScene {
	return &GameScene{
		state: state,
		font:  font,
	}
}

func (g *GameScene) Update() error {
	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBackground)

	playerCount := g.state.PlayerCount()
	if playerCount == 0 {
		g.drawConnecting(screen)
		return
	}

	g.drawHeader(screen, playerCount)
	g.drawPlayers(screen)
}

func (g *GameScene) drawConnecting(screen *ebiten.Image) {
	drawText(screen, g.font, "Connecting...", ScreenWidth/2-50, ScreenHeight/2, color.White)
}

func (g *GameScene) drawHeader(screen *ebiten.Image, playerCount int) {
	myID := g.state.PlayerID()

	header := fmt.Sprintf("You are Player %s | Waiting for players... %d/8", myID, playerCount)
	drawText(screen, g.font, header, 50, 50, color.RGBA{200, 200, 200, 255})

	// Divider
	vector.StrokeLine(screen, 40, 80, ScreenWidth-40, 80, 1, color.RGBA{60, 60, 80, 255}, false)
}

func (g *GameScene) drawPlayers(screen *ebiten.Image) {
	drawText(screen, g.font, "PLAYERS", 50, 110, color.RGBA{150, 150, 150, 255})

	players := g.state.Players()
	myID := g.state.PlayerID()

	y := float64(140)
	for _, p := range players {
		label := fmt.Sprintf("Player %s: %d HP", p.Id, p.Hp)

		clr := color.RGBA{255, 255, 255, 255}
		if p.Id == myID {
			label += " (you)"
			clr = color.RGBA{100, 255, 100, 255}
		}

		drawText(screen, g.font, label, 70, y, clr)
		y += 30
	}
}

func drawText(screen *ebiten.Image, font *text.GoTextFace, str string, x, y float64, clr color.Color) {
	if font == nil {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, font, op)
}

func DrawTextAt(screen *ebiten.Image, font *text.GoTextFace, str string, x, y float64) {
	drawText(screen, font, str, x, y, color.White)
}
