package status

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
)

var (
	labels = [...]string{ "N", "V",   "B", "D", "I", "Z", "C" }
	bit    = [...]uint8 { 128,  64, -  0,   0,   32,  16,  8  }
)

type Flags struct {
	flags        uint8
	currentFlags uint8
	log          *logging.Log
	devFlags     uint8
	terminal     *display.Terminal
	redraw       func(bool)
	cursor       common.Coord
	Ignore       bool
}
func NewFlags(log *logging.Log, terminal *display.Terminal, redraw func(bool)) *Flags {
	return &Flags{
		log:      log,
		Ignore:   false,
		redraw:   redraw,
		terminal: terminal,
		cursor:   common.Coord{X:0, Y:0},
	}
}

func (f *Flags) SetFlags(status uint8) bool {
	f.flags = status
	currentFlags := (status & 192) >> 4 | status & 24 >> 3
	changed := f.currentFlags != currentFlags
	f.currentFlags = currentFlags
	return changed
}
func (f *Flags) SyncFlags() {
	f.log.Info("Set Developer flags to current flags")
	f.devFlags = f.currentFlags
}
func (f *Flags) CurrentFlags() uint8 {
	return f.currentFlags
}
func (f *Flags) DevFlags() uint8 {
	return f.devFlags
}
func (f *Flags) FlagsBlock() string {
	str := ""
	lastColour := ""
	for n, label := range labels {
		isSet  := false
		if bit[n] > 0 {
			isSet  = f.flags & bit[n] > 0
		}
		colour := off
		if isSet {
			colour = on
		}
		if colour == lastColour { colour = "" } else { lastColour = colour }
		str = fmt.Sprintf("%s%s %s ", str, colour, label)

	}
	return fmt.Sprintf("%s%s", str, common.Reset)
}

func (f *Flags) Toggle() {
	f.devFlags ^= 1 << (3 - f.cursor.X)
	f.redraw(true)
}
func (f *Flags) Up() {
	f.devFlags |= 1 << (3 - f.cursor.X)
	f.redraw(true)
}
func (f *Flags) Down() {
	f.devFlags &^= 1 << (3 - f.cursor.X)
	f.redraw(true)
}
func (f *Flags) Left(n int) {
	if f.cursor.X - n >= 0 {
		f.cursor.X -= n
		f.PositionCursor()
		f.redraw(false)
	} else {
		f.terminal.Bell()
	}
}
func (f *Flags) Right(n int) {
	if f.cursor.X + n <= 3 {
		f.cursor.X += n
		f.PositionCursor()
		f.redraw(false)
	} else {
		f.terminal.Bell()
	}
}
func (f *Flags) PositionCursor() {
	f.terminal.At(f.cursor.X * 2 + 17, 20)
}
func (f *Flags) CursorPosition() string {
	return "Dev Flags"
}
func (f *Flags) DevBlock() string {
	c1, c2, c3, c4 := common.White, common.White, common.White, common.White
	if f.devFlags & 8 != 0 { c1 = common.BrightGreen }
	if f.devFlags & 4 != 0 { c2 = common.BrightGreen }
	if f.devFlags & 2 != 0 { c3 = common.BrightGreen }
	if f.devFlags & 1 != 0 { c4 = common.BrightGreen }
	return fmt.Sprintf(" %sN %sV %sZ %sC -> %s%02d ", c1, c2, c3, c4, common.BrightBlue, f.devFlags)
}
func (f *Flags) KeyIntercept(input common.Input) bool {
	if input.KeyCode != 0 {
		switch input.KeyCode {
		case display.CursorUp:
			f.Up()
		case display.CursorDown:
			f.Down()
		case display.CursorLeft:
			f.Left(1)
		case display.CursorRight:
			f.Right(1)
		default:
			return false
		}
	} else if input.Ascii > 0 {
		switch input.Ascii {
		case '1':
			f.Up()
		case '0':
			f.Down()
		case ' ', 13:
			f.Toggle()
		default:
			return false
		}
	}
	return true
}