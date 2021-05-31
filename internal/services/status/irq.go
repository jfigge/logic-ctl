package timing

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	clockHigh = common.BrightGreen
	clockLow  = common.Red
)

type Clock struct {
	state uint8
	log   *logging.Log
}

func New(log *logging.Log) *Clock {
	return &Clock{
		log: log,
	}
}

func (c *Clock) ClockHigh() {
	c.state = 1
}

func (c *Clock) ClockLow() {
	c.state = 0
}

func (c *Clock) Block() string {
	str := clockLow
	if c.state == 1 {
		str = clockHigh
	}
	return fmt.Sprintf("%sΦ%d%s", str, c.state, common.Reset)
}
