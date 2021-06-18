package driver

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type ErrorPage struct {
	dirty      bool
	initialize bool
	message    string
}
func NewErrorPage() *ErrorPage {
	return &ErrorPage{}
}

func (e *ErrorPage) Draw(t *display.Terminal, connected bool) {
	if !e.dirty && !e.initialize {
		return
	} else if e.initialize {
		t.Cls()
		e.initialize = false
	}

	t.PrintAtf(1,1, "%sBe right back%s", common.Yellow, common.Reset)
	t.PrintAtf(2,1, "%s%s%s", common.BrightRed, e.message, common.Reset)
	e.dirty = false
}
func (e *ErrorPage) SetDirty(initialize bool) {
	e.dirty = true
	if initialize {
		e.initialize = true
	}
}
func (e *ErrorPage) Process(a int, k int, connected bool) bool {
	if k != 0 {
		switch k {

		}
	} else {
		switch a {

		}
	}
	return false
}
func (e *ErrorPage) ErrorViewer(message string) common.UI {
	e.message = message
	e.initialize = true
	return e
}

