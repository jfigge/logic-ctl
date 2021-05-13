// https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#colors
package driver

import (
	"fmt"
	"github.com/pkg/term"
	xterm "golang.org/x/term"
	"os"
	"regexp"
	"time"
)

const (
	Black   = "\u001b[30m"
	Red     = "\u001b[31m"
	Green   = "\u001b[32m"
	Yellow  = "\u001b[33m"
	Blue    = "\u001b[34m"
	Magenta = "\u001b[35m"
	Cyan    = "\u001b[36m"
	White   = "\u001b[37m"

	Grey          = "\u001b[90m"
	BrightRed     = "\u001b[91m"
	BrightGreen   = "\u001b[92m"
	BrightYellow  = "\u001b[93m"
	BrightBlue    = "\u001b[94m"
	BrightMagenta = "\u001b[95m"
	BrightCyan    = "\u001b[96m"
	BrightWhite   = "\u001b[97m"

	BGBlack   = "\u001b[40m"
	BGRed     = "\u001b[41m"
	BGGreen   = "\u001b[42m"
	BGYellow  = "\u001b[43m"
	BGBlue    = "\u001b[44m"
	BGMagenta = "\u001b[45m"
	BGCyan    = "\u001b[46m"
	BGWhite   = "\u001b[47m"

	BGGrey          = "\u001b[100m"
	BGBrightRed     = "\u001b[101m"
	BGBrightGreen   = "\u001b[102m"
	BGBrightYellow  = "\u001b[103m"
	BGBrightBlue    = "\u001b[104m"
	BGBrightMagenta = "\u001b[105m"
	BGBrightCyan    = "\u001b[106m"
	BGBrightWhite   = "\u001b[107m"


	Bold      = "\u001b[1m"
	Underline = "\u001b[4m"
	Reset     = "\u001b[0m"

	Up    = "\u001b[%dA" // n rows up
	Down  = "\u001b[%dB" // n rows down
	Right = "\u001b[%dC" // n columns right
	Left  = "\u001b[%dD" // n columns left

	Bell = "\a"

	ClearDown   = "\u001b[0J" // clears from cursor until end of screen
	ClearUp     = "\u001b[1J" // clears from cursor to beginning of screen
	ClearScreen = "\u001b[2J" // clears entire screen

	ClearEnd   = "\u001b[0K" // clears from cursor to end of line
	ClearStart = "\u001b[1K" // clears from cursor to start of line
	ClearLine  = "\u001b[2K" // clears entire line

	SetColumn   = "\u001b[%dG"     // moves cursor to column n
	SetPosition = "\u001b[%d;%dH" // moves cursor to row n column m

	// Cursor keycodes
	CursorUp    = 38
	CursorDown  = 40
	CursorLeft  = 37
	CursorRight = 39

	// Show / Hide cursor
	Show = "\u001b[?25h"
	Hide = "\u001b[?25l"
)

var (
	HEX = [16]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F"}
	rex = regexp.MustCompile("\u001b\\[[0-9]{1,2}m")
)

type Display struct {
	fd    int
	cols  int
	rows  int
	col   int
	row   int
	state *xterm.State
	noticeTimer *time.Timer
	notice string
	dirty  bool
}
type DisplayMessage struct {
	Message string
	IsError bool
}

func NewDisplay() (*Display, error) {
	display := Display{fd: int(os.Stdin.Fd())}

	if w,h,e := xterm.GetSize(int(os.Stdin.Fd())); e != nil {
		return nil, e
	} else {
		display.cols = w
		display.rows = h
	}

	if s, e := xterm.GetState(int(os.Stdin.Fd())); e != nil {
		return nil, e
 	} else {
 		display.state = s
	}

	display.dirty = true
	return &display, nil
}

// Cursors
func (d *Display) Up(n int) {
	if d.row - n >= 1 {
		fmt.Printf(Up, n)
		d.row -= n
	} else {
		d.Bell()
	}
}
func (d *Display) Down(n int) {
	if d.row + n <= d.rows {
		fmt.Printf(Down, n)
		d.row += n
	} else {
		d.Bell()
	}
}
func (d *Display) Left(n int) {
	if d.col - n >= 1 {
		fmt.Printf(Left, n)
		d.col -= n
	} else {
		d.Bell()
	}
}
func (d *Display) Right(n int) {
	if d.col + n <= d.cols {
		fmt.Printf(Right, n)
		d.col += n
	} else {
		d.Bell()
	}
}

// Screen positioning
func (d *Display) At(col int, row int) bool {
	str := Bell
	if col >= 1 && col <= d.cols && row >= 1 && row <= d.rows {
		str = fmt.Sprintf(SetPosition, row, col)
		d.col = col
		d.row = row
	}
	fmt.Printf(str)
	return str != Bell
}
func (d *Display) Start() {
	fmt.Printf(SetColumn, 1)
	d.col = 1
}
func (d *Display) Home() {
	fmt.Printf(SetPosition, 1, 1)
	d.col = 1
	d.row = 1
}

// Display text
func (d *Display) PrintAt(text string, col int, row int) bool {
	ok := d.At(col, row)
	if ok {
		d.Print(text)
	}
	return ok
}
func (d *Display) Print(text string) {
	bs := []byte(StripFormatting(text))
	if d.col + len(bs) > d.cols {
		bs = bs[:d.cols - d.col]
	}
	fmt.Printf("%s", text)
	d.col += len(bs)
}

// Screen control
func (d *Display) Bell() {
	fmt.Printf(Bell)
}
func (d *Display) Cll() {
	fmt.Printf(ClearLine)
	d.Start()
}
func (d *Display) Cls() {
	fmt.Printf(ClearScreen)
	d.Home()
}
func (d *Display) HideCursor() {
	fmt.Printf(Hide)
}
func (d *Display) ShowCursor() {
	fmt.Printf(Show)
}

// Notifications
func (d *Display) Progress(text string, percent int) {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}
	d.At(1, d.rows)
	max := d.cols - len(text) - 3
	bar := max * percent / 100

	str := ""
	for i := 0; i < max; i++ {
		if i < bar {
			str += "#"
		} else {
			str += " "
		}
	}

	d.Print(fmt.Sprintf("%s [%s]", text, str))
}
func (d *Display) Notify(text string, colour string) {

	if d.noticeTimer != nil {
		d.noticeTimer.Stop()
		d.noticeTimer = nil
	}

	d.At(1, d.rows)
	d.Cll()
	str := fmt.Sprintf("%s%s%s", colour, text, Reset)
	d.Print(str)

	d.notice = str
	d.noticeTimer = time.NewTimer(5 * time.Second)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic occured: %s", err)
			}
		}()
		<-d.noticeTimer.C
		d.notice = ""
		d.SetDirty()
		if d.noticeTimer.Stop() {
			d.noticeTimer = nil
		}
	}()
}
func (d *Display) NotifyDM(dm *DisplayMessage) {
	if dm.IsError {
		d.Error(dm.Message)
	} else {
		d.Info(dm.Message)
	}
}
func (d *Display) Info(text string) {
	d.Notify(text, BrightWhite)
}
func (d *Display) Warn(text string) {
	d.Notify(text, BrightYellow)
}
func (d *Display) Error(text string) {
	d.Notify(text, BrightRed)
}

// Redraw
func (d *Display) Draw() {
	c := d.col
	r := d.row

	if !d.dirty {
		return
	}

	d.Cls()

	// Memory
	lines := MemoryBlock(0)
	for row, line := range lines {
		if ok := d.PrintAt(line, 1, row + 1); !ok {
			break
		}
	}
	xOffset := len(StripFormatting(lines[0])) + 3

	// FLags
	d.PrintAt(Yellow + "Flags", xOffset + 6, 1)
	d.PrintAt(Flags(), xOffset, 2)

	// Timing
	d.PrintAt(Yellow + "Step", xOffset + 6, 4)
	d.PrintAt(Step(), xOffset, 5)

	// Instr
	d.PrintAt(Yellow + "Instructions", xOffset + 3, 7)
	d.PrintAt(Step(), xOffset, 5)
	lines = InstructionsBlock(2, 11)
	for i := 0 ; i < 11; i++ {
		d.PrintAt(lines[uint16(i)], xOffset, 8 + i)
	}

	// Control lines
	d.PrintAt(Yellow + "Control Lines", 1, 20)
	for i := uint8(0); i < 7; i++ {
		colour := Red
		if i+1 == CurrentStep() {
			colour = Cyan
		}
		d.PrintAt(fmt.Sprintf("%sT%d %s 1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1  1 0 1 0 1 1 1 1 %s", Yellow, i+1, colour, Reset), 1, 21 + int(i))
	}


	// Notifications
	if d.notice != "" {
		d.PrintAt(d.notice, 1, d.rows)
	}

	d.At(c, r)
	d.dirty = false
}
func (d *Display) SetDirty() {
	d.dirty = true
	d.Draw()
}

// Input
func (d *Display)ReadChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		d.Error("Input error.  Resetting")
		return 0,0,nil
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	return
}

// Miscellaneous static functions
func StripFormatting(text string) string {
	return string(rex.ReplaceAll([]byte(text), []byte{}))
}
func HexData(data uint8) string {
	return fmt.Sprintf("%s%s", HEX[data >> 4], HEX[data & 15])
}
func HexAddress(address uint16) string {
	return fmt.Sprintf("%s%s%s%s", HEX[address >> 12 ], HEX[address >> 8 & 15], HEX[address >> 4 & 15], HEX[address & 15])
}