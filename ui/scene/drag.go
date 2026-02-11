package scene

import "github.com/ysomad/gigabg/ui"

// dragState tracks card drag-and-drop input.
type dragState struct {
	active    bool
	index     int
	fromBoard bool
	fromShop  bool
	cursorX   int
	cursorY   int
}

func (d *dragState) Start(index int, fromBoard, fromShop bool, mx, my int) {
	d.active = true
	d.index = index
	d.fromBoard = fromBoard
	d.fromShop = fromShop
	d.cursorX = mx
	d.cursorY = my
}

func (d *dragState) Reset() {
	d.active = false
}

// screenToBase converts screen pixel coords to base coords.
func screenToBase(screenX, screenY int) (float64, float64) {
	s := ui.ActiveRes.Scale()
	bx := (float64(screenX) - ui.ActiveRes.OffsetX()) / s
	by := (float64(screenY) - ui.ActiveRes.OffsetY()) / s
	return bx, by
}
