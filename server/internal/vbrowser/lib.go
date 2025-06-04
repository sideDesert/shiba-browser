package vbrowser

type Cursor struct {
	X float32
	Y float32
}

type Keyboard struct {
}

func NewCursor(x float32, y float32) Cursor {
	return Cursor{
		X: x,
		Y: y,
	}
}

type Display struct {
	Height int
	Width  int
	FPS    int
	Port   int
}

func NewDisplay(port, height, width, fps int) *Display {
	return &Display{
		Port:   port,
		Height: height,
		Width:  width,
		FPS:    fps,
	}
}
