package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

var (
	labels = [...]string{ "N", "V", "B", "D", "I", "Z", "C" }
	bit    = [...]uint8 { 128,  64, - 0,   0,  32,  16,  8  }
)

type Flags struct {
	flags        uint8
	currentFlags uint8
	lastFlags    uint8
	log          *logging.Log
}
func NewFlags(log *logging.Log) *Flags {
	return &Flags{
		log: log,
	}
}

func (f *Flags) SetFlags(status uint8) {
	f.flags = status
	f.currentFlags = (status & 192) >> 4 | (status & 24) >> 3
}
func (f *Flags) CurrentFlags() uint8 {
	return f.currentFlags
}

func (f *Flags) FlagsBlock() string {
	str := ""
	lastColour := ""
	for n, label := range labels {
		isSet  := false
		wasSet := false
		if bit[n] > 0 {
			isSet  = f.flags&bit[n] > 0
			wasSet = f.lastFlags&bit[n] > 0
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
	f.lastFlags = f.flags
	return fmt.Sprintf("%s%s", str, common.Reset)
}
