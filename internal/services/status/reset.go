package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	resetHigh = common.Red
	resetLow  = common.BrightGreen
)

type Reset struct {
	state  uint8
	log    *logging.Log
	reload func()
	redraw func(bool)
}
func NewReset(log *logging.Log, redraw func(bool), reload func()) *Reset {
	return &Reset{
		log:    log,
		reload: reload,
		redraw: redraw,
	}
}

func (r *Reset) ResetHigh() {
	r.state = 1
	r.redraw(false)
}
func (r *Reset) ResetLow() {
	r.state = 0
	r.reload()
	r.redraw(false)
}

func (r *Reset) ResetBlock() string {
	str := resetLow
	if r.state == 1 {
		str = resetHigh
	}
	return fmt.Sprintf("%s%d%s", str, r.state, common.Reset)
}
