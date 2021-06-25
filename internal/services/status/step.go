package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

type Steps struct {
	step        uint8
	log          *logging.Log
}
func NewSteps(log *logging.Log) *Steps {
	return &Steps {
		log: log,
	}
}

func (s *Steps) SetStatus(status uint8) {
	s.step = status & 7
}
func (s *Steps) CurrentStep() uint8 {
	return s.step
}

func (s *Steps) StepBlock() string {
	t, colour, lastColour := s.step, common.Yellow, ""
	str := ""
	for i := uint8(0); i < 8; i++ {
		colour = step
		if i == t {
			colour = currentStep
		}
		if colour == lastColour { colour = "" } else { lastColour = colour }
		str = fmt.Sprintf("%s%s %d ", str, colour, i)
	}
	return str + common.Reset
}

