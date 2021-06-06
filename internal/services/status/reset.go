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
	state uint8
	log   *logging.Log
}

func NewReset(log *logging.Log) *Reset {
	return &Reset{
		log: log,
	}
}

func (c *Reset) ResetHigh() {
	c.state = 1
}

func (c *Reset) ResetLow() {
	c.state = 0
}

func (c *Reset) ResetBlock() string {
	str := resetLow
	if c.state == 1 {
		str = resetHigh
	}
	return fmt.Sprintf("%s%d%s", str, c.state, common.Reset)
}
