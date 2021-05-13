package driver

import "fmt"

const (
	off         = Grey
	turnedOff   = White
	on          = Green
	turnedOn    = BrightGreen
	step        = Reset + Cyan
	currentStep = BGCyan + Black
)

var (
	flags uint8 = 195
	lastFlags uint8 = 83
	labels = []string{ "N", "V", "B", "D", "I", "Z", "C" }
	bit    = []uint8 { 128,  64,  0,   0,  32,  16,  8  }
)

func N() {
	if flags & 128 > 0 {
		flags &= 127
	} else {
		flags |= 128
	}
	display.SetDirty()
}

func Next() {
	t := (flags & 7) + 1
	if t > 7 {t = 1}

	flags &= 248
	flags |= t

	display.SetDirty()
}

func CurrentStep() uint8 {
	return flags & 7
}

func Flags() string {
	str := ""
	lastColour := ""
	for n, label := range labels {
		isSet  := false
		wasSet := false
		if bit[n] > 0 {
			isSet  = flags&bit[n] > 0
			wasSet = lastFlags&bit[n] > 0
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
	lastFlags = flags
	return fmt.Sprintf("%s%s", str, Reset)
}

func Step() string {
	t, colour, lastColour := flags & 7, Yellow, ""
	str := ""
	for i := uint8(1); i < 8; i++ {
		colour = step
		if i == t {
			colour = currentStep
		}
		if colour == lastColour { colour = "" } else { lastColour = colour }
		str = fmt.Sprintf("%s%s %d ", str, colour, i)
	}
	return str + Reset
}