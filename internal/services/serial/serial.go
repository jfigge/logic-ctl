package serial

import (
	"encoding/binary"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/common"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/display"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/logging"
	"github.td.teradata.com/sandbox/logic-ctl/internal/services/status"
	srl "go.bug.st/serial"
	"strings"
	"sync"
	"time"
)

type Serial struct {
	sync 	   sync.Mutex
	port       srl.Port
	buffer     chan byte
	address    chan []byte
	opCode     chan byte
	data       chan byte
	status     chan byte
	terminated bool
	clock      *status.Clock
	irq        *status.Irq
	nmi        *status.Nmi
	reset      *status.Reset
	log        *logging.Log
	setDirty   func()
	setStatus  func(uint8)
	tick       func(synchronized bool)
	dirty      bool
	initialize bool
}
func New(log *logging.Log, clock *status.Clock, irq *status.Irq, nmi *status.Nmi, reset *status.Reset, setDirty func(), setStatus  func(uint8), tick func(synchronized bool)) *Serial {
	s := &Serial{
		clock:     clock,
		irq:       irq,
		nmi:       nmi,
		reset:     reset,
		log:       log,
		setDirty:  setDirty,
		setStatus: setStatus,
		tick:      tick,
	}
	return s
}

func (s *Serial) Connect(reconnect bool) bool {
	s.sync.Lock()
	defer s.sync.Unlock()
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
		s.port = nil
		if !reconnect {
			s.log.Errorf("Failed to open USB port %s: %v", config.CLIConfig.Serial.PortName, err)
		}
		return false
	}

	// Wait for reset
	time.Sleep(1500 * time.Millisecond)

	s.terminated = false
	s.buffer  = make(chan byte, 20)
	s.address = make(chan []byte)
	s.data    = make(chan byte)
	s.status  = make(chan byte)
	s.opCode  = make(chan byte)
	go s.reader()
	go s.driver()

	s.log.Infof("Openned port %s", config.CLIConfig.Serial.PortName)
	return true
}
func (s *Serial) Disconnect() bool {
	s.sync.Lock()
	defer s.sync.Unlock()

	if s.port != nil {
		s.terminated = true
		if err := s.port.Close(); err != nil {
			s.log.Errorf("Failed to close USB port %s: %v", config.CLIConfig.Serial.PortName, err)
			return false
		}
		s.port = nil
	}
	s.log.Info("Port closed")
	return true
}
func (s *Serial) Reconnect() {
	s.Connect(true)
	if s.IsConnected() {
		s.SetDirty(true)
		s.tick(true)
	}
}
func (s *Serial) IsConnected() bool {
	return s.port != nil
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

func (s *Serial) ReadAddress() (uint16, bool) {
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte("a\n")); err != nil {
		s.log.Errorf("Failed to send request for address: %v", err)
		return 0, false
	} else if n != 2 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 2, sent: %d", n)
		return 0, false
	}

	// Receive address
	a := binary.LittleEndian.Uint16(<-s.address)
	return a, true
}
func (s *Serial) ReadOpCode() (uint8, bool) {
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte("o\n")); err != nil {
		s.log.Errorf("Failed to send request for OpCode: %v", err)
		return 0, false
	} else if n != 2 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 2, sent: %d", n)
		return 0, false
	}

	// Receive opCode
	select {
	case b := <-s.opCode:
		return b, true
	case <- time.After(10 * time.Second):
		s.log.Warnf("OpCode not received")
		return 0, false
	}
}
func (s *Serial) ReadData() (uint8, bool) {
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte("d\n")); err != nil {
		s.log.Errorf("Failed to send request for data: %v", err)
		return 0, false
	} else if n != 2 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 2, sent: %d", n)
		return 0, false
	}

	// Receive data
	select {
	case b := <-s.data:
		return b, true
	case <- time.After(10 * time.Second):
		s.log.Warnf("Data not received")
		return 0, false
	}
}
func (s *Serial) ReadStatus() (uint8, bool) {
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return 0, false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte("s\n")); err != nil {
		s.log.Errorf("Failed to send request for status: %v", err)
		return 0, false
	} else if n != 2 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 2, sent: %d", n)
		return 0, false
	}

	// Receive status
	select {
		case b := <-s.status:
			return b, true
		case <- time.After(10 * time.Second):
			s.log.Warnf("Status not received")
			return 0, false
	}
}
func (s *Serial) SetData(data uint8) bool {
	s.sync.Lock()
	defer s.sync.Unlock()
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return false
		}
	}

	// Send command
	if n, err := s.port.Write([]byte{0x44,data,0x0A}); err != nil {
		s.log.Errorf("Failed to send request to set data: %v", err)
		return false
	} else if n != 3 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 3, sent: %d", n)
		return false
	}
	return true
}
func (s *Serial) SetLines(data uint64) bool {
	s.sync.Lock()
	defer s.sync.Unlock()
	if s.port == nil {
		if ok := s.Connect(true); !ok {
			return false
		}
	}

	// Send command
	bs := []byte{'L', uint8(data >> 40), uint8(data >> 32), uint8(data >> 24), uint8(data >> 16), uint8(data >> 8), uint8(data), 0x0A}
	s.log.Debugf("%c [%s %s %s %s %s %s] %s", bs[0], display.HexData(bs[1]), display.HexData(bs[2]), display.HexData(bs[3]), display.HexData(bs[4]), display.HexData(bs[5]), display.HexData(bs[6]), display.HexData(bs[7]) )

	if n, err := s.port.Write(bs); err != nil {
		s.log.Errorf("Failed to send request to set lines: %v", err)
		return false
	} else if n != 8 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 8, sent: %d", n)
		return false
	}
	return true
}

func (s *Serial) reader() {
	bs := make([]byte,1)
	defer func() {
		if r := recover(); r != nil {
			s.log.Errorf("Recovered Read panic: %v", r)
		}
	}()
	for !s.terminated {
		if n, err := s.port.Read(bs); err != nil {
			s.log.Errorf("Read failed: %v", err)
			s.terminated = true
		} else if n == 1 {
			s.buffer <- bs[0]
		} else {
			s.log.Warnf("Unexpected number of bytes received: Wanted 1, Received %d", n)
		}
	}
	close(s.buffer)
}
func (s *Serial) driver() {
	for  {
		b:= <- s.buffer
		s.log.Debugf("Inbound data: %s", string(b))
		switch b {
		case 'a':
			s.address <- []byte {<-s.buffer,<-s.buffer}
		case 'd':
			s.data <- <-s.buffer
		case 'o':
			s.opCode <- <-s.buffer
		case 's':
			s.status <- <-s.buffer
		case 'c':
			s.clock.ClockLow()
			s.setDirty()
		case 'C':
			s.clock.ClockHigh()
			s.setDirty()
		case 'k':
			s.setStatus(<-s.buffer)
			s.clock.ClockLow()
		case 'K':
			s.setStatus(<-s.buffer)
			s.clock.ClockHigh()
		case 'i':
			s.irq.IrqLow()
		case 'I':
			s.irq.IrqHigh()
		case 'n':
			s.nmi.NmiLow()
		case 'N':
			s.nmi.NmiHigh()
		case 'r':
			s.reset.ResetLow()
		case 'R':
			s.reset.ResetHigh()
		default:
			s.log.Debugf("Unknown byte: %v", display.HexData(b))
			if _, e := s.port.GetModemStatusBits(); e != nil {
				s.Disconnect()
				s.terminated = true
				return
			}
		}
	}
	s.log.Warn("Stopped receiving")
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
		s.log.Errorf("%v", err)
		return
	}
	if len(ports) == 0 {
		s.log.Errorf("%v", err)
		return
	}
	t.PrintAtf(1,1, "%sSerial Ports%s", common.Yellow, common.Reset)
	line := 2
	for _, port := range ports {
		if strings.HasPrefix(port, "/dev/cu") {
			t.PrintAtf(4, line, "%s%v%s", common.Blue, port, common.Reset)
			line ++
		}
	}
	t.PrintAtf(1, t.Rows(), "%sPress any key%s", common.Yellow, common.Reset)
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
