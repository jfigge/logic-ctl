package common

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type UI interface {
	Draw(t *display.Terminal)
	SetDirty(initialize bool)
	Process(ascii int, keyCode int) bool
}