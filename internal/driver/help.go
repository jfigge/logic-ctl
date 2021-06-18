package driver

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type HelpPage struct {
	dirty      bool
	initialize bool
}
func NewHelpPage() *HelpPage {
	return &HelpPage{}
}
func (h *HelpPage) Help() common.UI {
	h.initialize = true
	return h
}
func (h *HelpPage) Draw(t *display.Terminal, connected bool) {
	if !h.dirty && !h.initialize {
		return
	} else if h.initialize {
		t.Cls()
		h.initialize = false
	}

	t.PrintAtf( 1, 1, "%sDatabase%s", common.Yellow, common.Reset)
	t.PrintAtf( 1, 2, "%s1%s Accumulator%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 3, "%s2%s Processor Status%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 4, "%s3%s Special Bus%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 5, "%s4%s PC High%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 6, "%s5%s PC Low%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 7, "%s6%s Input%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(21, 1, "%sAddress low bus%s", common.Yellow, common.Reset)
	t.PrintAtf(21, 2, "%s0%s Input%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 3, "%s1%s Program counter%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 4, "%s2%s Constants%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 5, "%s3%s Stack pointer%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 6, "%s4%s ALU%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(41, 1, "%sAddress high bus%s", common.Yellow, common.Reset)
	t.PrintAtf(41, 2, "%s0%s Input%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41, 3, "%s1%s Constants%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41, 4, "%s2%s Program counter%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41, 5, "%s3%s Special bus%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(61, 1, "%sSpecial bus%s", common.Yellow, common.Reset)
	t.PrintAtf(61, 2, "%s0%s Accumulator%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 3, "%s1%s Register Y%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 4, "%s2%s Register X%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 5, "%s3%s ALU%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 6, "%s4%s Stack pointer%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 7, "%s5%s Data bus%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61, 8, "%s6%s Address high bus%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf( 1,10, "%sKey mappings%s", common.Yellow, common.Reset)
	t.PrintAtf( 1,10, "%sd%s Debug disabled%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21,10, "%sf%s Toggle flag usage%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41,10, "%sh%s Show this page%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61,10, "%sl%s Show log history%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(81,10, "%sq%s Quit%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1,11, "%sD%s Debug enabled%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21,11, "%sm%s Toggle mnemonics%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41,11, "%sH%s Toggle line help%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61,11, "%sp%s Show ports%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(1, t.Rows(), "%s'Press any key to exit%s", common.Yellow, common.Reset)
	h.dirty = false
}
func (h *HelpPage) SetDirty(initialize bool) {
	h.dirty = true
	if initialize {
		h.initialize = true
	}
}
func (h *HelpPage) Process(a int, k int, connected bool) bool {
	if k != 0 {
		switch k {
		}
	} else {
		switch a {
		}
	}
	return true
}
