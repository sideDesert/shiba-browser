package vbrowser

type MousePosition struct {
	X int
	Y int
}

func Pos(x int, y int) MousePosition {
	return MousePosition{
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
