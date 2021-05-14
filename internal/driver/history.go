package driver

type History struct {
	history []string
}

func (h *History) Add(text string) {
	h.history = append(h.history, text)
}

func (h *History) Draw(d *Display) {
	d.Cls()
	display.PrintAt(Yellow + "Notification History" + Reset, 1,1)
	max := len(h.history)
	for row := 2; row < d.rows - 1; row++ {
 		if max - row + 2 > 0 {
			d.PrintAt(h.history[max - row + 1], 1, row)
		}
	}
	d.PressAnyKey()
}

