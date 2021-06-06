package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	clockHigh = common.BrightCyan
	clockLow  = common.Cyan
)

type Clock struct {
	state uint8
	tick  func()
	log   *logging.Log
}
func NewClock(log *logging.Log, tick func()) *Clock {
	return &Clock{
		log:  log,
		tick: tick,
	}
}

func (c *Clock) ClockHigh() {
	c.state = 1
	c.tick()
}
func (c *Clock) ClockLow() {
	c.state = 0
	c.tick()
}

func (c *Clock) CurrentState() uint8 {
	return c.state
}

func (c *Clock) Block() string {
	str := clockLow
	if c.state == 1 {
		str = clockHigh
	}
	return fmt.Sprintf("%sΦ%d%s", str, c.state + 1, common.Reset)
}
