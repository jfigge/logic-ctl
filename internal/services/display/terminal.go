// https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#colors
package display

import (
	"fmt"
	xterm "golang.org/x/term"
	"os"
	"regexp"
)

const (
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

type Terminal struct {
	fd      int
	cols    int
	rows    int
	col     int
	row     int
	state   *xterm.State
}
func New() (*Terminal, error) {
	display := Terminal{
		fd:  int(os.Stdin.Fd()),
	}

	for display.cols ==- 0 {
		if w, h, e := xterm.GetSize(int(os.Stdin.Fd())); e != nil {
			return nil, e
		} else {
			display.cols = w
			display.rows = h
		}
	}

	if s, e := xterm.GetState(int(os.Stdin.Fd())); e != nil {
		return nil, e
 	} else {
 		display.state = s
	}
	display.HideCursor()
	return &display, nil
}

func (t *Terminal) Up(n int) {
	if t.row - n >= 1 {
		fmt.Printf(Up, n)
		t.row -= n
	} else {
		t.Bell()
	}
}
func (t *Terminal) Down(n int) {
	if t.row + n <= t.rows {
		fmt.Printf(Down, n)
		t.row += n
	} else {
		t.Bell()
	}
}
func (t *Terminal) Left(n int) {
	if t.col - n >= 1 {
		fmt.Printf(Left, n)
		t.col -= n
	} else {
		t.Bell()
	}
}
func (t *Terminal) Right(n int) {
	if t.col + n <= t.cols {
		fmt.Printf(Right, n)
		t.col += n
	} else {
		t.Bell()
	}
}

func (t *Terminal) At(col int, row int) bool {
	str := Bell
	if col >= 1 && col <= t.cols && row >= 1 && row <= t.rows {
		str = fmt.Sprintf(SetPosition, row, col)
		t.col = col
		t.row = row
	}
	fmt.Printf(str)
	return str != Bell
}
func (t *Terminal) Start() {
	fmt.Printf(SetColumn, 1)
	t.col = 1
}
func (t *Terminal) Home() {
	fmt.Printf(SetPosition, 1, 1)
	t.col = 1
	t.row = 1
}

func (t *Terminal) PrintAt(text string, col int, row int) bool {
	ok := t.At(col, row)
	if ok {
		t.Print(text)
	}
	return ok
}
func (t *Terminal) Print(text string) {
	bs := []byte(StripFormatting(text))
	if t.col + len(bs) > t.cols {
		bs = bs[:t.cols - t.col]
	}
	fmt.Printf("%s", text)
	t.col += len(bs)
}

func (t *Terminal) Bell() {
	fmt.Printf(Bell)
}
func (t *Terminal) Cll() {
	fmt.Printf(ClearLine)
	t.Start()
}
func (t *Terminal) Cls() {
	fmt.Printf(ClearScreen)
	t.Home()
}
func (t *Terminal) HideCursor() {
	fmt.Printf(Hide)
}
func (t *Terminal) ShowCursor() {
	fmt.Printf(Show)
}

func (t *Terminal) Row() int {
	return t.row
}
func (t *Terminal) Col() int {
	return t.col
}
func (t *Terminal) Rows() int {
	return t.rows
}
func (t *Terminal) Cols() int {
	return t.cols
}

func StripFormatting(text string) string {
	return string(rex.ReplaceAll([]byte(text), []byte{}))
}
func HexData(data uint8) string {
	return fmt.Sprintf("%s%s", HEX[data >> 4], HEX[data & 15])
}
func HexAddress(address uint16) string {
	return fmt.Sprintf("%s%s%s%s", HEX[address >> 12 ], HEX[address >> 8 & 15], HEX[address >> 4 & 15], HEX[address & 15])
}