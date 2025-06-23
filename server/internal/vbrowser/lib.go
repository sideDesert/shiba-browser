package vbrowser

type Cursor struct {
	X    float32
	Y    float32
	Port int
}

type Keyboard struct {
	Port int
}

func NewCursor(x float32, y float32, port int) Cursor {
	return Cursor{
		X:    x,
		Y:    y,
		Port: port,
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
