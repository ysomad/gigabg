package ui

import (
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Menu struct {
	lobbyID  string
	onJoin   func(lobbyID string)
	selected bool
	font     *text.GoTextFace
}

func NewMenu(font *text.GoTextFace, onJoin func(lobbyID string)) *Menu {
	return &Menu{
		font:   font,
		onJoin: onJoin,
	}
}

func (m *Menu) Update() {
	if m.selected {
		m.handleTextInput()
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		m.selected = m.inputContains(mx, my)

		if m.buttonContains(mx, my) && len(m.lobbyID) > 0 {
			m.onJoin(m.lobbyID)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(m.lobbyID) > 0 {
		m.onJoin(m.lobbyID)
	}
}

func (m *Menu) handleTextInput() {
	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		if len(m.lobbyID) < 32 {
			m.lobbyID += string(r)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(m.lobbyID) > 0 {
		_, size := utf8.DecodeLastRuneInString(m.lobbyID)
		m.lobbyID = m.lobbyID[:len(m.lobbyID)-size]
	}
}

func (m *Menu) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 30, 255})

	centerX := float32(ScreenWidth / 2)
	centerY := float32(ScreenHeight / 2)

	drawText(screen, m.font, "GIGA Battlegrounds", float64(centerX)-80, 100, color.RGBA{255, 215, 0, 255})
	drawText(screen, m.font, "Lobby ID:", float64(centerX)-100, float64(centerY)-60, color.RGBA{200, 200, 200, 255})

	// Input field
	inputX := centerX - 100
	inputY := centerY - 40
	inputW := float32(200)
	inputH := float32(30)

	borderColor := color.RGBA{60, 60, 80, 255}
	if m.selected {
		borderColor = color.RGBA{100, 150, 255, 255}
	}
	vector.FillRect(screen, inputX, inputY, inputW, inputH, color.RGBA{40, 40, 50, 255}, true)
	vector.StrokeRect(screen, inputX, inputY, inputW, inputH, 2, borderColor, false)

	displayText := m.lobbyID
	if m.selected && len(displayText) < 32 {
		displayText += "_"
	}
	drawText(screen, m.font, displayText, float64(inputX)+8, float64(inputY)+8, color.White)

	// Join button
	btnX := centerX - 50
	btnY := centerY + 20
	btnW := float32(100)
	btnH := float32(35)

	btnColor := color.RGBA{50, 120, 50, 255}
	if len(m.lobbyID) == 0 {
		btnColor = color.RGBA{60, 60, 60, 255}
	}
	vector.FillRect(screen, btnX, btnY, btnW, btnH, btnColor, true)
	vector.StrokeRect(screen, btnX, btnY, btnW, btnH, 2, color.RGBA{80, 160, 80, 255}, false)

	var textColor color.Color = color.White
	if len(m.lobbyID) == 0 {
		textColor = color.RGBA{100, 100, 100, 255}
	}
	drawText(screen, m.font, "Join", float64(btnX)+35, float64(btnY)+10, textColor)

	drawText(
		screen,
		m.font,
		"Enter lobby ID and click Join",
		float64(centerX)-110,
		float64(ScreenHeight)-100,
		color.RGBA{100, 100, 100, 255},
	)
}

func (m *Menu) inputContains(mx, my int) bool {
	centerX := ScreenWidth / 2
	centerY := ScreenHeight / 2
	inputX := centerX - 100
	inputY := centerY - 40
	return mx >= inputX && mx <= inputX+200 && my >= inputY && my <= inputY+30
}

func (m *Menu) buttonContains(mx, my int) bool {
	centerX := ScreenWidth / 2
	centerY := ScreenHeight / 2
	btnX := centerX - 50
	btnY := centerY + 20
	return mx >= btnX && mx <= btnX+100 && my >= btnY && my <= btnY+35
}
