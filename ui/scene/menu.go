package scene

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/ysomad/gigabg/game"
	"github.com/ysomad/gigabg/ui"
	"github.com/ysomad/gigabg/ui/widget"
)

var lobbySizes = [4]int{2, 4, 6, 8}

type menuMode uint8

const (
	modeJoin menuMode = iota
	modeCreate
)

type Menu struct {
	font     *text.GoTextFace
	onJoin   func(player game.PlayerID, lobbyID string)
	onCreate func(player game.PlayerID, lobbySize int)

	playerID *widget.TextInput
	mode     menuMode

	// Tab buttons.
	joinTab   *widget.Button
	createTab *widget.Button

	// Join mode.
	lobbyID   *widget.TextInput
	submitBtn *widget.Button

	// Create mode.
	sizeBtns     [4]*widget.Button
	selectedSize int
	createBtn    *widget.Button
}

func NewMenu(
	font *text.GoTextFace,
	onJoin func(player game.PlayerID, lobbyID string),
	onCreate func(player game.PlayerID, lobbySize int),
) *Menu {
	m := &Menu{
		font:         font,
		onJoin:       onJoin,
		onCreate:     onCreate,
		mode:         modeJoin,
		selectedSize: 2,
	}
	m.buildWidgets()
	return m
}

func (m *Menu) buildWidgets() {
	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)
	inputW := w * 0.2
	inputH := h * 0.05
	cx := w/2 - inputW/2

	// Player ID input.
	m.playerID = &widget.TextInput{
		Rect:   ui.Rect{X: cx, Y: h * 0.30, W: inputW, H: inputH},
		MaxLen: 20,
	}
	m.playerID.SetFocused(true)

	// Tab buttons.
	tabW := w * 0.10
	tabH := h * 0.05
	tabGap := w * 0.02
	tabStartX := w/2 - (2*tabW+tabGap)/2
	tabY := h * 0.42

	m.joinTab = &widget.Button{
		Rect: ui.Rect{X: tabStartX, Y: tabY, W: tabW, H: tabH},
		Text: "Join",
		OnClick: func() {
			m.mode = modeJoin
		},
	}
	m.createTab = &widget.Button{
		Rect: ui.Rect{X: tabStartX + tabW + tabGap, Y: tabY, W: tabW, H: tabH},
		Text: "Create",
		OnClick: func() {
			m.mode = modeCreate
		},
	}

	// Join mode: lobby ID input + submit button.
	contentY := h * 0.55
	m.lobbyID = &widget.TextInput{
		Rect:   ui.Rect{X: cx, Y: contentY, W: inputW, H: inputH},
		MaxLen: 36,
	}

	btnH := h * 0.05
	btnW := w * 0.08
	m.submitBtn = &widget.Button{
		Rect:    ui.Rect{X: w/2 - btnW/2, Y: contentY + inputH + h*0.03, W: btnW, H: btnH},
		Text:    "Join",
		OnClick: m.submitJoin,
	}

	// Create mode: size selector + create button.
	sizeBtnW := w * 0.04
	sizeGap := w * 0.02
	totalSizeW := 4*sizeBtnW + 3*sizeGap
	sizeStartX := w/2 - totalSizeW/2

	for i, size := range lobbySizes {
		sz := size
		m.sizeBtns[i] = &widget.Button{
			Rect: ui.Rect{X: sizeStartX + float64(i)*(sizeBtnW+sizeGap), Y: contentY, W: sizeBtnW, H: btnH},
			Text: strconv.Itoa(sz),
			OnClick: func() {
				m.selectedSize = sz
			},
		}
	}

	createW := w * 0.10
	m.createBtn = &widget.Button{
		Rect:    ui.Rect{X: w/2 - createW/2, Y: contentY + btnH + h*0.03, W: createW, H: btnH},
		Text:    "Create",
		OnClick: m.submitCreate,
	}
}

func (m *Menu) submitJoin() {
	lid := m.lobbyID.Value()
	if lid == "" {
		return
	}
	pid, err := game.ParsePlayerID(m.playerID.Value())
	if err != nil {
		return
	}
	m.onJoin(pid, lid)
}

func (m *Menu) submitCreate() {
	pid, err := game.ParsePlayerID(m.playerID.Value())
	if err != nil {
		return
	}
	m.onCreate(pid, m.selectedSize)
}

func (m *Menu) Update(res ui.Resolution) error {
	m.playerID.Update(res)
	m.joinTab.Update(res)
	m.createTab.Update(res)

	switch m.mode {
	case modeJoin:
		m.lobbyID.Update(res)
		m.submitBtn.Update(res)
	case modeCreate:
		for _, btn := range m.sizeBtns {
			btn.Update(res)
		}
		m.createBtn.Update(res)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if m.mode == modeJoin {
			if m.playerID.Focused() {
				m.playerID.SetFocused(false)
				m.lobbyID.SetFocused(true)
			} else {
				m.lobbyID.SetFocused(false)
				m.playerID.SetFocused(true)
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		switch m.mode {
		case modeJoin:
			m.submitJoin()
		case modeCreate:
			m.submitCreate()
		}
	}
	return nil
}

var (
	clrTabActive    = color.RGBA{50, 120, 50, 255}
	clrTabBorderAct = color.RGBA{80, 160, 80, 255}
	clrTabInactive  = color.RGBA{50, 50, 50, 255}
	clrTabBorderIn  = color.RGBA{80, 80, 80, 255}
	clrTextBright   = color.RGBA{255, 255, 255, 255}
	clrTextDim      = color.RGBA{160, 160, 160, 255}
	clrBtnEnabled   = color.RGBA{50, 120, 50, 255}
	clrBtnBorderEn  = color.RGBA{80, 160, 80, 255}
	clrBtnDisabled  = color.RGBA{60, 60, 60, 255}
	clrBtnTextDis   = color.RGBA{100, 100, 100, 255}
	clrSizeActive   = color.RGBA{50, 120, 50, 255}
	clrSizeBorderAc = color.RGBA{80, 160, 80, 255}
	clrSizeInactive = color.RGBA{60, 60, 60, 255}
	clrSizeBorderIn = color.RGBA{100, 100, 100, 255}
	clrLabel        = color.RGBA{200, 200, 200, 255}
	clrTitle        = color.RGBA{255, 215, 0, 255}
)

func (m *Menu) Draw(screen *ebiten.Image, res ui.Resolution) {
	screen.Fill(ui.ColorBackground)

	w := float64(ui.BaseWidth)
	h := float64(ui.BaseHeight)

	// Title.
	ui.DrawText(screen, res, m.font, "GIGA Battlegrounds", w*0.38, h*0.14, clrTitle)

	// Player ID.
	ui.DrawText(screen, res, m.font, "Player ID:", m.playerID.Rect.X, m.playerID.Rect.Y-h*0.03, clrLabel)
	m.playerID.Draw(screen, res, m.font)

	// Tab buttons.
	m.styleTab(m.joinTab, m.mode == modeJoin)
	m.styleTab(m.createTab, m.mode == modeCreate)
	m.joinTab.Draw(screen, res, m.font)
	m.createTab.Draw(screen, res, m.font)

	// Mode content.
	switch m.mode {
	case modeJoin:
		m.drawJoinMode(screen, res, h)
	case modeCreate:
		m.drawCreateMode(screen, res, h)
	}
}

func (m *Menu) drawJoinMode(screen *ebiten.Image, res ui.Resolution, h float64) {
	ui.DrawText(screen, res, m.font, "Lobby ID:", m.lobbyID.Rect.X, m.lobbyID.Rect.Y-h*0.03, clrLabel)
	m.lobbyID.Draw(screen, res, m.font)

	canSubmit := m.playerID.Value() != "" && m.lobbyID.Value() != ""
	m.styleSubmitBtn(m.submitBtn, canSubmit)
	m.submitBtn.Draw(screen, res, m.font)
}

func (m *Menu) drawCreateMode(screen *ebiten.Image, res ui.Resolution, h float64) {
	ui.DrawText(screen, res, m.font, "Lobby Size:", m.sizeBtns[0].Rect.X, m.sizeBtns[0].Rect.Y-h*0.03, clrLabel)

	for _, btn := range m.sizeBtns {
		size, err := strconv.Atoi(btn.Text)
		if err != nil {
			continue
		}
		if size == m.selectedSize {
			btn.Color = clrSizeActive
			btn.BorderClr = clrSizeBorderAc
			btn.TextClr = clrTextBright
		} else {
			btn.Color = clrSizeInactive
			btn.BorderClr = clrSizeBorderIn
			btn.TextClr = clrTextDim
		}
		btn.Draw(screen, res, m.font)
	}

	canSubmit := m.playerID.Value() != ""
	m.styleSubmitBtn(m.createBtn, canSubmit)
	m.createBtn.Draw(screen, res, m.font)
}

func (m *Menu) styleTab(btn *widget.Button, active bool) {
	if active {
		btn.Color = clrTabActive
		btn.BorderClr = clrTabBorderAct
		btn.TextClr = clrTextBright
	} else {
		btn.Color = clrTabInactive
		btn.BorderClr = clrTabBorderIn
		btn.TextClr = clrTextDim
	}
}

func (m *Menu) styleSubmitBtn(btn *widget.Button, enabled bool) {
	if enabled {
		btn.Color = clrBtnEnabled
		btn.BorderClr = clrBtnBorderEn
		btn.TextClr = clrTextBright
	} else {
		btn.Color = clrBtnDisabled
		btn.BorderClr = clrBtnDisabled
		btn.TextClr = clrBtnTextDis
	}
}

func (m *Menu) OnEnter() {}
func (m *Menu) OnExit()  {}
