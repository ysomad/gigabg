package ui

// Base resolution — all layout coordinates are defined in this space.
const (
	BaseWidth  = 1280
	BaseHeight = 720
)

// ActiveRes is the current window resolution, updated each frame by App.Layout.
var ActiveRes = Resolution{BaseWidth, BaseHeight}

type Resolution struct {
	Width  int
	Height int
}

// Scale returns the uniform scale factor relative to the base resolution.
// Uses the smaller dimension ratio so content always fits within the window.
func (r Resolution) Scale() float64 {
	sx := float64(r.Width) / BaseWidth
	sy := float64(r.Height) / BaseHeight
	if sx < sy {
		return sx
	}
	return sy
}

// OffsetX returns the horizontal offset to center content in the window.
func (r Resolution) OffsetX() float64 {
	return (float64(r.Width) - BaseWidth*r.Scale()) / 2
}

// OffsetY returns the vertical offset to center content in the window.
func (r Resolution) OffsetY() float64 {
	return (float64(r.Height) - BaseHeight*r.Scale()) / 2
}

// Rect is a positioned rectangle in base coordinate space (1280x720).
type Rect struct {
	X, Y, W, H float64
}

// Contains reports whether screen pixel (px, py) falls within this base-space rect.
// Converts screen pixels to base coordinates before testing.
func (r Rect) Contains(px, py int) bool {
	s := ActiveRes.Scale()
	bx := (float64(px) - ActiveRes.OffsetX()) / s
	by := (float64(py) - ActiveRes.OffsetY()) / s
	return bx >= r.X && bx < r.X+r.W && by >= r.Y && by < r.Y+r.H
}

// Screen converts this base-space rect to screen pixel coordinates.
func (r Rect) Screen() Rect {
	s := ActiveRes.Scale()
	return Rect{
		X: r.X*s + ActiveRes.OffsetX(),
		Y: r.Y*s + ActiveRes.OffsetY(),
		W: r.W * s,
		H: r.H * s,
	}
}

func (r Rect) Right() float64  { return r.X + r.W }
func (r Rect) Bottom() float64 { return r.Y + r.H }

// GameLayout holds computed zone rects in base coordinate space.
type GameLayout struct {
	Screen    Rect
	Header    Rect // turn info + timer
	BtnRow    Rect // refresh / upgrade / freeze
	Shop      Rect // shop card zone
	Board     Rect // board card zone
	Hand      Rect // hand card zone
	PlayerBar Rect // bottom HP bar

	CardW float64
	CardH float64
	Gap   float64
}

// CalcGameLayout computes layout zones in base coordinate space.
func CalcGameLayout() GameLayout {
	w := float64(BaseWidth)
	h := float64(BaseHeight)

	cardW := w * 0.10
	cardH := cardW * 1.2
	gap := w * 0.016

	headerH := h * 0.08
	btnRowH := h * 0.06
	zoneH := cardH + h*0.05 // card + label padding

	return GameLayout{
		Screen:    Rect{0, 0, w, h},
		Header:    Rect{0, 0, w, headerH},
		BtnRow:    Rect{0, headerH, w, btnRowH},
		Shop:      Rect{0, headerH + btnRowH, w, zoneH},
		Board:     Rect{0, headerH + btnRowH + zoneH, w, zoneH},
		Hand:      Rect{0, headerH + btnRowH + 2*zoneH, w, zoneH},
		PlayerBar: Rect{0, h - h*0.05, w, h * 0.05},
		CardW:     cardW,
		CardH:     cardH,
		Gap:       gap,
	}
}

// CardRect returns the Rect of the i-th card in a centered row of n cards within a zone.
func CardRect(zone Rect, i, n int, cardW, cardH, gap float64) Rect {
	totalW := float64(n)*cardW + float64(n-1)*gap
	startX := zone.X + (zone.W-totalW)/2

	labelH := zone.H * 0.2
	cardY := zone.Y + labelH + (zone.H-labelH-cardH)/2

	return Rect{
		X: startX + float64(i)*(cardW+gap),
		Y: cardY,
		W: cardW,
		H: cardH,
	}
}

// ButtonRects returns rects for refresh, upgrade, freeze buttons within a zone.
func ButtonRects(zone Rect) (refresh, upgrade, freeze Rect) {
	btnW := zone.W * 0.10
	btnH := zone.H * 0.7
	gap := zone.W * 0.01
	totalW := 3*btnW + 2*gap
	startX := zone.X + (zone.W-totalW)/2
	y := zone.Y + (zone.H-btnH)/2

	refresh = Rect{startX, y, btnW, btnH}
	upgrade = Rect{startX + btnW + gap, y, btnW, btnH}
	freeze = Rect{startX + 2*(btnW+gap), y, btnW, btnH}
	return
}

// CombatLayout holds computed zone rects in base coordinate space.
type CombatLayout struct {
	Screen   Rect
	Header   Rect
	Opponent Rect // opponent board zone
	Player   Rect // player board zone

	CardW float64
	CardH float64
	Gap   float64
}

// CalcCombatLayout computes layout for the combat animation in base coordinate space.
func CalcCombatLayout() CombatLayout {
	w := float64(BaseWidth)
	h := float64(BaseHeight)

	cardW := w * 0.10
	cardH := cardW * 1.2
	gap := w * 0.016

	headerH := h * 0.12
	zoneH := cardH + h*0.06

	return CombatLayout{
		Screen:   Rect{0, 0, w, h},
		Header:   Rect{0, 0, w, headerH},
		Opponent: Rect{0, headerH, w, zoneH},
		Player:   Rect{0, h - zoneH - h*0.08, w, zoneH},
		CardW:    cardW,
		CardH:    cardH,
		Gap:      gap,
	}
}
