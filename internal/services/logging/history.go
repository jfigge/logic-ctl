package logging

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"sync"
)

type History struct {
	messages  []string
	redraw    func(bool)
	offset    int
	offsetMax int
	pageSize  int
	silence   bool
	sync      sync.Mutex
}

func (h* History) add(message string) {
	h.sync.Lock()
	defer h.sync.Unlock()

	h.messages = append(h.messages, message)
	if len(h.messages) > 1000 {
		h.messages = h.messages[1:]
	}
}

func (h *History) Draw(t *display.Terminal, connected bool, initialize bool) {
	if initialize {
		t.Cls()
	}

	max := len(h.messages)
	scroll := h.scrollBar(t.Rows(), t.Cols())

	t.PrintAtf(1,1, "%sNotification log%s%s", common.Yellow, common.Reset, scroll)
	for row := 2; row < t.Rows() - 1; row++ {
		if max - row + 2 > 0 {
			t.PrintAt(1, row, h.messages[max - h.offset - row + 1])
		}
	}
	t.PrintAtf(1, t.Rows(), "%sPress 'c' to clear, arrows to scroll, 'q/a' to page up/down, any other key to exit%s", common.Yellow, common.Reset)
	t.HideCursor()
}
func (h *History) Process(input common.Input) bool {
	if input.KeyCode != 0 {
		switch input.KeyCode {
		case display.CursorUp:
			if h.offset > 0 {
				h.offset--
				h.redraw(true)
			} else {
				h.bell()
			}
			return false

		case display.CursorDown:
			if h.offset < h.offsetMax {
				h.offset++
				h.redraw(true)
			} else {
				h.bell()
			}
			return false

		default:
			fmt.Printf("")
		}
	} else {
		switch input.Ascii {
		case 's':
			h.silence = !h.silence
			if !h.silence {
				h.bell()
			}
			return false
		case 'q':
			if h.offset > h.pageSize {
				h.offset -= h.pageSize
				h.redraw(true)
			} else if h.offset > 0 {
				h.offset = 0
				h.redraw(true)
			} else {
				h.bell()
			}
			return false

		case 'a':
			if h.offset < h.offsetMax -h.pageSize {
				h.offset += h.pageSize
				h.redraw(true)
			} else if h.offset < h.offsetMax {
				h.offset = h.offsetMax
				h.redraw(true)
			} else {
				h.bell()
			}
			return false

		case 'c':
			h.messages = []string{}
			h.redraw(true)
			return false
		}
	}
	return true
}

func (h *History) bell() {
	if !h.silence {
		fmt.Printf(display.Bell)
	}
}
func (h *History) scrollBar(rows int, cols int) string {
	max := len(h.messages)
	h.pageSize = rows - 4

	h.offsetMax = max - rows + 2
	if h.offsetMax < 0 {
		h.offsetMax = 0
	}
	if h.offset > h.offsetMax {
		h.offset = h.offsetMax
	}
	format := fmt.Sprintf("%%%ds", cols - 16)
	status := ""
	switch {
	case h.offset == 0:
		status = "top"
	case h.offset == h.offsetMax:
		status = "bottom"
	default:
		status = fmt.Sprintf("%d/%d", h.offset, h.offsetMax)
	}

	return fmt.Sprintf(format, status)
}