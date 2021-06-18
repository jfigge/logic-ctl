package common

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type UI interface {
	Draw(t *display.Terminal, connected bool)
	SetDirty(initialize bool)
	Process(ascii int, keyCode int, connected bool) bool
}

type Intercept interface {
	KeyIntercept(ascii int, keyCode int, connected bool) bool
}