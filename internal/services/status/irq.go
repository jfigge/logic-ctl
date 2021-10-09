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
	state  uint8
	log    *logging.Log
	redraw func(bool)
}

func NewIrq(log *logging.Log, redraw func(bool)) *Irq {
	return &Irq{
		log:    log,
		redraw: redraw,
	}
}

func (i *Irq) IrqHigh() {
	i.state = 1
	i.redraw(false)
}

func (i *Irq) IrqLow() {
	i.state = 0
	i.redraw(false)
}

func (i *Irq) IrqBlock() string {
	str := irqLow
	if i.state == 1 {
		str = irqHigh
	}
	return fmt.Sprintf("%s%d%s", str, i.state, common.Reset)
}
