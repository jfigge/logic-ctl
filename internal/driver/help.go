package driver

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type HelpPage struct {
}
func NewHelpPage() *HelpPage {
	return &HelpPage{}
}
func (h *HelpPage) Help() common.UI {
	return h
}
func (h *HelpPage) Draw(t *display.Terminal, connected bool, initialize bool) {
	if initialize {
		t.Cls()
	}

	t.PrintAtf( 1, 1, "%sData Bus%s", common.Yellow, common.Reset)
	t.PrintAtf( 1, 2, "%s1%s Accumulator%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 3, "%s2%s Processor Status%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 4, "%s3%s Special Bus%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 5, "%s4%s PC High%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 6, "%s5%s PC Low%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1, 7, "%s6%s Input%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(21, 1, "%sAddress bus low%s", common.Yellow, common.Reset)
	t.PrintAtf(21, 2, "%s0%s Input%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 3, "%s1%s Program counter%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 4, "%s2%s Constants%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 5, "%s3%s Stack pointer%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 6, "%s4%s ALU%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21, 7, "%s5%s PC Low Register%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(41, 1, "%sAddress bus high%s", common.Yellow, common.Reset)
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
	t.PrintAtf( 1,10, "%sd%s Debug enabled%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21,10, "%sc%s Copies line lines%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41,10, "%sf%s Toggle flag usage%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61,10, "%sl%s Show log history%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(81,10, "%stab%s Toggle editors%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf( 1,11, "%sD%s Debug disabled%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21,11, "%sC%s Copies all lines%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41,11, "%sh%s Show this page%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61,11, "%sp%s Show ports%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(81,11, "%sq%s Quit%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf( 1,13, "%s0%s Deactivate line%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(21,13, "%s1%s Activate line%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(41,13, "%sspace%s Toggle line%s", common.Yellow, common.White, common.Reset)
	t.PrintAtf(61,13, "%sdelete%s Reset line%s", common.Yellow, common.White, common.Reset)

	t.PrintAtf(1, t.Rows(), "%sPress any key to exit%s", common.Yellow, common.Reset)
}
func (h *HelpPage) Process(keyboard common.Input) bool {
	return true
}
