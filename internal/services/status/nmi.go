package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	nmiHigh = common.Red
	nmiLow  = common.BrightGreen
)

type Nmi struct {
	state  uint8
	log    *logging.Log
	redraw func(bool)
}

func NewNmi(log *logging.Log, redraw func(bool)) *Nmi {
	return &Nmi{
		log:    log,
		redraw: redraw,
	}
}

func (n *Nmi) NmiHigh() {
	n.state = 1
	n.redraw(false)
}

func (n *Nmi) NmiLow() {
	n.state = 0
	n.redraw(false)
}

func (n *Nmi) NmiBlock() string {
	str := nmiLow
	if n.state == 1 {
		str = nmiHigh
	}
	return fmt.Sprintf("%s%d%s", str, n.state, common.Reset)
}
