package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

const (
	off         = common.Grey
	turnedOff   = common.White
	on          = common.Green
	turnedOn    = common.BrightGreen
	step        = common.Reset + common.Cyan
	currentStep = common.BGCyan + common.Black
)

var (
	labels = [...]string{ "N", "V", "B", "D", "I", "Z", "C" }
	bit    = [...]uint8 { 128,  64, - 0,   0,  32,  16,  8  }
)

type Status struct {
	flags        uint8
	currentFlags uint8
	lastFlags    uint8
	log          *logging.Log
}
func NewStatus(log *logging.Log) *Status {
	return &Status{
		log: log,
	}
}

func (s *Status) SetStatus(status uint8) {
	s.flags = status
	s.currentFlags = (status & 192) >> 4 | (status & 24) >> 3
}
func (s *Status) CurrentStep() uint8 {
	return s.flags & 7
}
func (s *Status) CurrentFlags() uint8 {
	return s.currentFlags
}

func (s *Status) FlagsBlock() string {
	str := ""
	lastColour := ""
	for n, label := range labels {
		isSet  := false
		wasSet := false
		if bit[n] > 0 {
			isSet  = s.flags&bit[n] > 0
			wasSet = s.lastFlags&bit[n] > 0
		}
		colour := off
		if isSet && !wasSet {
			colour = turnedOn
		} else if isSet && wasSet {
			colour = on
		} else if !isSet && wasSet {
			colour = turnedOff
		}
		if colour == lastColour { colour = "" } else { lastColour = colour }
		str = fmt.Sprintf("%s%s %s ", str, colour, label)

	}
	s.lastFlags = s.flags
	return fmt.Sprintf("%s%s", str, common.Reset)
}
func (s *Status) StepBlock() string {
	t, colour, lastColour := s.flags& 7, common.Yellow, ""
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

