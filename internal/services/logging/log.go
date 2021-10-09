package logging

import (
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"sync"
	"time"
)

type Log struct {
	history   *History
	sync      sync.Mutex
	activeLMs []activeLM
	redraw    func(bool)
	debug     bool
}
type LogMessage struct {
	Message string
	IsError bool
}
type activeLM struct {
	Message string
	timer *time.Timer
}
func New(redraw func(bool)) *Log {
	return &Log{
		history: &History{ redraw: redraw, silence: true, sync: sync.Mutex{}},
		redraw:  redraw,
		debug:   false,
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

	str := fmt.Sprintf("%s%s%s%s%s", display.ClearLine, colour, text, common.Reset, display.ClearEnd)
	l.history.add(str)

	l.sync.Lock()
	l.activeLMs = append(
		[]activeLM{activeLM{
			Message: str,
			timer: time.AfterFunc(5 * time.Second, func() {
				l.sync.Lock()
				l.activeLMs = l.activeLMs[:len(l.activeLMs) - 1]
				l.sync.Unlock()
				l.redraw(false)
			})}}, l.activeLMs...)
	l.sync.Unlock()
	l.redraw(false)
}
func (l *Log) NotifyLM(lm LogMessage) {
	if lm.IsError {
		l.Error(lm.Message)
	} else {
		l.Info(lm.Message)
	}
}
func (l *Log) SetDebug(enabled bool) {
	if l.debug != enabled {
		if enabled {
			l.Info("Debug output enabled")
			l.debug = true
		} else {
			l.Info("Debug output disabled")
			l.debug = false
		}
	}
}

func (l *Log) Tracef(text string, a...interface{}) {
	l.Trace(fmt.Sprintf(text, a...))
}
func (l *Log) Trace(text string) {
	if l.debug {
		l.history.add(fmt.Sprintf("%s%s%s%s", common.White, text, common.Reset, display.ClearEnd))
	}
}
func (l *Log) Debugf(text string, a...interface{}) {
	l.Debug(fmt.Sprintf(text, a...))
}
func (l *Log) Debug(text string) {
	if l.debug {
		l.Notify(text, common.White)
	} else {
		l.history.add(fmt.Sprintf("%s%s%s%s", common.White, text, common.Reset, display.ClearEnd))
	}
}
func (l *Log) Infof(text string, a...interface{}) {
	l.Info(fmt.Sprintf(text, a...))
}
func (l *Log) Info(text string) {
	l.Notify(text, common.BrightWhite)
}
func (l *Log) Warnf(text string, a...interface{}) {
	l.Warn(fmt.Sprintf(text, a...))
}
func (l *Log) Warn(text string) {
	l.Notify(text, common.BrightYellow)
}
func (l *Log) Errorf(text string, a...interface{}) {
	l.Error(fmt.Sprintf(text, a...))
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
	return l.history
}