package logging

import (
	"fmt"
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

	t.PrintAt(fmt.Sprintf("%sNotification log%s", common.Yellow, common.Reset), 1,1)
	max := len(h.messages)
	for row := 2; row < t.Rows() - 1; row++ {
		if max - row + 2 > 0 {
			t.PrintAt(h.messages[max - row + 1], 1, row)
		}
	}
	t.PrintAt(fmt.Sprintf("%sPress any key%s", common.Yellow, common.Reset), 1, t.Rows())
	h.dirty = false
}
func (h *History) SetDirty(initialize bool) {
	h.dirty = true
	if initialize {
		h.initialize = true
	}
}
func (h *History) Process(a int, k int) bool {
	return true
}
