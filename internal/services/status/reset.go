package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	nmiHigh = common.BrightGreen
	nmiLow  = common.Red
)

type Nmi struct {
	state uint8
	log   *logging.Log
}

func NewNmi(log *logging.Log) *Nmi {
	return &Nmi{
		log: log,
	}
}

func (c *Nmi) NmiHigh() {
	c.state = 1
}

func (c *Nmi) NmiLow() {
	c.state = 0
}

func (c *Nmi) NmiBlock() string {
	str := nmiLow
	if c.state == 1 {
		str = nmiHigh
	}
	return fmt.Sprintf("%sΦ%d%s", str, c.state, common.Reset)
}
