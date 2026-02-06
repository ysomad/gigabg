package ui

type Resolution struct {
	Width  int
	Height int
}

func (r Resolution) Scale() float64 {
	return float64(r.Width) / 1280.0
}

var ActiveRes = Resolution{1280, 720}
