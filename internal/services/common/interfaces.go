package common

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type UI interface {
	Draw(t *display.Terminal, connected bool, initialize bool)
	Process(input Input) bool
}

type Intercept interface {
	KeyIntercept(input Input) bool
	CursorPosition() string
	PositionCursor()
}

type Input struct {
	Ascii     int
	KeyCode   int
	Connected bool
}

type Coord struct {
	X,Y int
}