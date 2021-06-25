package status

import "github.td.teradata.com/sandbox/logic-ctl/internal/services/common"

const (
	off         = common.Grey
	turnedOff   = common.White
	on          = common.Green
	turnedOn    = common.BrightGreen
	step        = common.Reset + common.Cyan
	currentStep = common.BGCyan + common.Black
)