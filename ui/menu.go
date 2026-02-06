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

	centerX := float64(ActiveRes.Width) / 2
	centerY := float64(ActiveRes.Height) / 2

	drawText(screen, m.font, "GIGA Battlegrounds", centerX-scf(80), sc(100), color.RGBA{255, 215, 0, 255})

	drawText(screen, m.font, "Lobby ID:", centerX-scf(100), centerY-scf(60), color.RGBA{200, 200, 200, 255})

	// Input field
	inputX := float32(centerX - scf(100))
	inputY := float32(centerY - scf(40))
	inputW := float32(scf(200))
	inputH := float32(scf(30))

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
	drawText(screen, m.font, displayText, float64(inputX)+scf(8), float64(inputY)+scf(8), color.White)

	// Join button
	btnX := float32(centerX - scf(50))
	btnY := float32(centerY + scf(20))
	btnW := float32(scf(100))
	btnH := float32(scf(35))

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
	drawText(screen, m.font, "Join", float64(btnX)+scf(35), float64(btnY)+scf(10), textColor)

	drawText(
		screen,
		m.font,
		"Enter lobby ID and click Join",
		centerX-scf(110),
		float64(ActiveRes.Height)-sc(100),
		color.RGBA{100, 100, 100, 255},
	)
}

func (m *Menu) inputContains(mx, my int) bool {
	centerX := ActiveRes.Width / 2
	centerY := ActiveRes.Height / 2
	inputX := centerX - int(scf(100))
	inputY := centerY - int(scf(40))
	return mx >= inputX && mx <= inputX+int(scf(200)) && my >= inputY && my <= inputY+int(scf(30))
}

func (m *Menu) buttonContains(mx, my int) bool {
	centerX := ActiveRes.Width / 2
	centerY := ActiveRes.Height / 2
	btnX := centerX - int(scf(50))
	btnY := centerY + int(scf(20))
	return mx >= btnX && mx <= btnX+int(scf(100)) && my >= btnY && my <= btnY+int(scf(35))
}
