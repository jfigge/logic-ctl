package serial

import (
	"encoding/binary"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/timing"
	srl "go.bug.st/serial"
	"strings"
	"time"
)

type Serial struct {
	port       srl.Port
	buffer     chan byte
	address    chan []byte
	data       chan byte
	terminated bool
	clock      *timing.Clock
	log        *logging.Log
	setDirty   func()
	dirty      bool
	initialize bool
}
func New(log *logging.Log, clock *timing.Clock, setDirty func()) *Serial {
	s := &Serial{
		clock: clock,
		log: log,
		setDirty: setDirty,
	}
	return s
}

func (s *Serial) Connect() bool {

	if s.port != nil {
		if ok := s.Disconnect(); !ok {
			return false
		}
	}

	mode := &srl.Mode{
		DataBits: config.CLIConfig.Serial.DataBits,
		BaudRate: config.CLIConfig.Serial.BaudRate,
		StopBits: toStopBits(config.CLIConfig.Serial.StopBits),
		Parity:   toParity(config.CLIConfig.Serial.Parity),
	}

	var err error
	if s.port, err = srl.Open(config.CLIConfig.Serial.PortName, mode); err != nil {
		s.log.Error(fmt.Sprintf("Failed to open USB port %s: %v", config.CLIConfig.Serial.PortName, err))
		return false
	}

	// Wait for reset
	time.Sleep(1500 * time.Millisecond)

	s.terminated = false
	s.buffer  = make(chan byte, 20)
	s.address = make(chan []byte)
	s.data    = make(chan byte)
	go s.reader()
	go s.driver()

	s.log.Info(fmt.Sprintf("Openned port %s", config.CLIConfig.Serial.PortName))
	return true
}
func (s *Serial) Disconnect() bool {
	if s.port != nil {
		s.terminated = true
		if err := s.port.Close(); err != nil {
			s.log.Error(fmt.Sprintf("Failed to close USB port %s: %v", config.CLIConfig.Serial.PortName, err))
			return false
		}
	}
	s.log.Info("Port closed")
	return true
}

func (s *Serial) ReadAddress() (uint16, bool) {
	if s.port == nil {
		if ok := s.Connect(); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte{0x61,0x0A}); err != nil {
		s.log.Error(fmt.Sprintf("Failed to send request for address: %v", err))
		return 0, false
	} else if n != 2 {
		s.log.Error(fmt.Sprintf("Unexpected number of bytes sent.  Expected 2, sent: %d", n))
		return 0, false
	}

	// Receive address
	a := binary.LittleEndian.Uint16(<-s.address)
	s.log.Info(fmt.Sprintf("Address received: %d", a))
	return a, true
}
func (s *Serial) ReadData() (uint8, bool) {
	if s.port == nil {
		if ok := s.Connect(); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte("r\n")); err != nil {
		s.log.Error(fmt.Sprintf("Failed to send request for data: %v", err))
		return 0, false
	} else if n != 2 {
		s.log.Error(fmt.Sprintf("Unexpected number of bytes sent.  Expected 2, sent: %d", n))
		return 0, false
	}

	// Receive address
	b := <-s.data
	s.log.Info(fmt.Sprintf("Data received: %d", b))
	return b, true
}

func toStopBits(value int) srl.StopBits {
	switch value {
	case 1: return srl.OneStopBit
	case 2: return srl.OnePointFiveStopBits
	case 3: return srl.TwoStopBits
	default:
		fmt.Println("Invalid StopBit")
		return srl.OneStopBit
	}
}
func toParity(value int) srl.Parity {
	switch value {
	case 0: return srl.NoParity
	case 1: return srl.OddParity
	case 2: return srl.EvenParity
	case 3: return srl.MarkParity
	case 4: return srl.SpaceParity
	default:
		fmt.Println("Invalid StopBit")
		return srl.NoParity
	}
}

func (s *Serial) reader() {
	bs := make([]byte,1)
	for !s.terminated {
		if n, err := s.port.Read(bs); err != nil {
			fmt.Println(err)
			s.terminated = true
		} else if n == 1 {
			s.buffer <- bs[0]
		}
	}
	close(s.buffer)
}
func (s *Serial) driver() {

	for b := range s.buffer {
		switch b {
		case 'a':
			s.address <- []byte {<-s.buffer,<-s.buffer}
		case 'd':
			s.data <- <-s.buffer
		case 'c':
			s.clock.ClockLow()
			s.setDirty()
		case 'C':
			s.clock.ClockHigh()
			s.setDirty()
		default:
			s.log.Error(fmt.Sprintf("Unknown byte: %v", display.HexData(b)))
		}
	}
	fmt.Println("Stopped receiving")
}

func (s *Serial) Draw(t *display.Terminal) {
	if !s.dirty && !s.initialize {
		return
	}  else if s.initialize {
		t.Cls()
		s.initialize = false
	}
	ports, err := srl.GetPortsList()
	if err != nil {
		s.log.Error(fmt.Sprintf("%v", err))
		return
	}
	if len(ports) == 0 {
		s.log.Error(fmt.Sprintf("%v", err))
		return
	}
	t.PrintAt(fmt.Sprintf("%sSerial Ports%s", common.Yellow, common.Reset), 1,1)
	line := 2
	for _, port := range ports {
		if strings.HasPrefix(port, "/dev/cu") {
			t.PrintAt(fmt.Sprintf("%s%v%s", common.Blue, port, common.Reset), 4, line)
			line ++
		}
	}
	t.PrintAt(fmt.Sprintf("%sPress any key%s", common.Yellow, common.Reset), 1, t.Rows())
	s.dirty = false
}
func (s *Serial) SetDirty(initialize bool) {
	s.dirty = true
	if initialize {
		s.initialize = true
	}
}
func (s *Serial) Process(a int, k int) bool {
	return true
}
func (s *Serial) PortViewer() common.UI {
	s.initialize = true
	return s
}
