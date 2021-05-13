package driver

import (
	"fmt"
	"os"
)

var display *Display

func NewDriver() {
	if d,e := NewDisplay(); e != nil {
		fmt.Printf("Failed to initialize terminal: %v", e)
		os.Exit(1)
	} else {
		display = d
	}

	//display.NotifyDM(ReadInstructions())
}


func Run() {
	loop := true
	for loop {
		display.Draw()
		a,k,e := display.ReadChar()
		if e != nil {
			fmt.Printf("Unexpected error: %v", e)
			os.Exit(1)
		}

		if k != 0 {
			switch k {
			case CursorUp:
				display.Up(1)
			case CursorDown:
				display.Down(1)
			case CursorLeft:
				display.Left(1)
			case CursorRight:
				display.Right(1)
			default:
				display.Warn(fmt.Sprintf("Unknown code: [%v]", k))
			}
		} else {
			switch a {
			case 'q':
				loop = false
			case 'h':
				display.HideCursor()
				display.Info("Hide")
			case 's':
				display.ShowCursor()
				display.Info("Show")
			case 'n':
				N()
			case 't':
				Next()
			case 'e':
				Export()
			//case 'r':
			//	display.NotifyDM(ReadInstructions())
			//case 'w':
			//	display.NotifyDM(WriteInstructions())
			default:
				display.Warn(fmt.Sprintf("Unmapped ascii code: [%c]", a))
			}
		}
	}
}