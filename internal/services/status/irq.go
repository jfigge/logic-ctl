package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	irqHigh = common.Red
	irqLow  = common.BrightGreen
)

type Irq struct {
	state uint8
	log   *logging.Log
}

func NewIrq(log *logging.Log) *Irq {
	return &Irq{
		log: log,
	}
}

func (c *Irq) IrqHigh() {
	c.state = 1
}

func (c *Irq) IrqLow() {
	c.state = 0
}

func (c *Irq) IrqBlock() string {
	str := irqLow
	if c.state == 1 {
		str = irqHigh
	}
	return fmt.Sprintf("%s%d%s", str, c.state, common.Reset)
}
