package logging

import (
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
)

type History struct {
	messages   []string
	dirty      bool
	initialize bool
}

func (h* History) add(message string) {
	h.messages = append(h.messages, message)
}
func (h *History) Draw(t *display.Terminal) {
	if !h.dirty && !h.initialize {
		return
	} else if h.initialize {
		t.Cls()
		h.initialize = false
	}

	t.PrintAtf(1,1, "%sNotification log%s", common.Yellow, common.Reset)
	max := len(h.messages)
	for row := 2; row < t.Rows() - 1; row++ {
		if max - row + 2 > 0 {
			t.PrintAt(1, row, h.messages[max - row + 1])
		}
	}
	t.PrintAtf(1, t.Rows(), "%s'Press 'c' to clear, any other key to exit%s", common.Yellow, common.Reset)
	h.dirty = false
}
func (h *History) SetDirty(initialize bool) {
	h.dirty = true
	if initialize {
		h.initialize = true
	}
}
func (h *History) Process(a int, k int) bool {
	if k != 0 {
		switch k {

		}
	} else {
		switch a {
		case 'c':
			h.messages = []string{}
			h.SetDirty(true)
			return false
		}
	}
	return true
}
