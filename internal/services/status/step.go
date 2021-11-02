package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

type Steps struct {
	step uint8
	log  *logging.Log
}
func NewSteps(log *logging.Log) *Steps {
	return &Steps {
		log: log,
	}
}

func (s *Steps) SetStep(status uint8) bool {
	step := status & 7
	changed := step != s.step
	s.step = step
	return changed
}
func (s *Steps) CurrentStep() uint8 {
	return s.step
}

func (s *Steps) StepBlock(lastStep uint8) string {
	t, colour, lastColour := s.step, common.Yellow, ""
	str := ""
	for i := uint8(0); i <= 7; i++ {
		colour = step
		if i == t + 1 || (i == 0 && t + 1 == lastStep) {
			colour = currentStep
		}
		if colour == lastColour { colour = "" } else { lastColour = colour }
		str = fmt.Sprintf("%s%s %d ", str, colour, i + 1)
	}
	return str + common.Reset
}

