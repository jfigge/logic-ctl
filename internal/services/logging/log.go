package logging

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"sync"
	"time"
)

type Log struct {
	history    *History
	sync       sync.Mutex
	activeLMs  []activeLM
	setDirty   func()
}
type LogMessage struct {
	Message string
	IsError bool
}
type activeLM struct {
	Message string
	timer *time.Timer
}
func New(setDirty func()) *Log {
	return &Log{
		history:  &History{},
		setDirty: setDirty,
	}
}

func (l *Log) Progress(text string, percent int) {
	//if percent < 0 {
	//	percent = 0
	//} else if percent > 100 {
	//	percent = 100
	//}
	//d.At(1, d.rows)
	//max := d.cols - len(text) - 3
	//bar := max * percent / 100
	//
	//str := ""
	//for i := 0; i < max; i++ {
	//	if i < bar {
	//		str += "#"
	//	} else {
	//		str += " "
	//	}
	//}
	//
	//d.Print(fmt.Sprintf("%s [%s]", text, str))
}
func (l *Log) Notify(text string, colour string) {

	str := fmt.Sprintf("%s%s%s", colour, text, common.Reset)
	l.history.add(str)

	l.sync.Lock()
	l.activeLMs = append(
		[]activeLM{activeLM{
			Message: str,
			timer: time.AfterFunc(5 * time.Second, func() {
				l.sync.Lock()
				l.activeLMs = l.activeLMs[:len(l.activeLMs) - 1]
				l.sync.Unlock()
				l.setDirty()
			})}}, l.activeLMs...)
	l.sync.Unlock()
	l.setDirty()
}
func (l *Log) NotifyLM(lm LogMessage) {
	if lm.IsError {
		l.Error(lm.Message)
	} else {
		l.Info(lm.Message)
	}
}
func (l *Log) Info(text string) {
	l.Notify(text, common.BrightWhite)
}
func (l *Log) Warn(text string) {
	l.Notify(text, common.BrightYellow)
}
func (l *Log) Error(text string) {
	l.Notify(text, common.BrightRed)
}

func (l *Log) LogBlock() []string {
	var tmp []string
	l.sync.Lock()
	defer l.sync.Unlock()
	for _, alm := range l.activeLMs {
		tmp = append(tmp, alm.Message)
	}
	return tmp
}
func (l *Log) Dump() {
	lines := l.LogBlock()
	for i := len(lines) -1; i >= 0; i-- {
		fmt.Printf("%s\n", lines[i])
	}
}

func (l *Log) HistoryViewer() common.UI {
	l.history.initialize = true
	return l.history
}

