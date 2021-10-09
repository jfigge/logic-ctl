package driver

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type ErrorPage struct {
	//dirty      bool
	//initialize bool
	message    string
}
func NewErrorPage() *ErrorPage {
	return &ErrorPage{}
}

func (e *ErrorPage) Draw(t *display.Terminal, connected bool, initialize bool) {
	if !initialize {
		t.Cls()
		//		e.initialize = false
	}

	t.PrintAtf(1, 1, "%sBe right back%s", common.Yellow, common.Reset)
	t.PrintAtf(2, 1, "%s%s%s", common.BrightRed, e.message, common.Reset)
}

func (e *ErrorPage) SetDirty(initialize bool) {
	fmt.Println()
}
func (e *ErrorPage) Process(input common.Input) bool {
	return false
}
func (e *ErrorPage) ErrorViewer(message string) common.UI {
	e.message = message
	return e
}

