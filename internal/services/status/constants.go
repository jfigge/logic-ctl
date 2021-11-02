package status

import "github.td.teradata.com/sandbox/logic-ctl/internal/services/common"

const (
	off         = common.Grey
	on          = common.BrightGreen
	step        = common.Reset + common.Cyan
	currentStep = common.BGCyan + common.Black
)