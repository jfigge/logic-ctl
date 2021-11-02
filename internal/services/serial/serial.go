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
	port         srl.Port
	buffer       chan byte
	address      chan []byte
	opCode       chan byte
	data         chan byte
	status       chan byte
	terminated   bool
	connected    bool
	steps        *status.Steps
	flags        *status.Flags
	clock        *status.Clock
	irq          *status.Irq
	nmi          *status.Nmi
	reset        *status.Reset
	log          *logging.Log
	connStatus   func(bool)
	startCapture func()
	stopCapture  func()
	mode         *srl.Mode
}
func New(log *logging.Log, clock *status.Clock, irq *status.Irq, nmi *status.Nmi, reset *status.Reset, flags *status.Flags, steps *status.Steps, connStatus func(bool), wg *sync.WaitGroup) *Serial {
	s := &Serial{
		clock:      clock,
		irq:        irq,
		nmi:         nmi,
		reset:      reset,
		steps:      steps,
		flags:      flags,
		log:        log,
		terminated: false,
		connected:  false,
		connStatus: connStatus,
		buffer:     make(chan byte),
		address:    make(chan []byte),
		data:       make(chan byte),
		status:     make(chan byte),
		opCode:     make(chan byte),
		mode:       &srl.Mode {
			DataBits: config.CLIConfig.Serial.DataBits,
			BaudRate: config.CLIConfig.Serial.BaudRate,
			StopBits: toStopBits(config.CLIConfig.Serial.StopBits),
			Parity:   toParity(config.CLIConfig.Serial.Parity),
		},
	}

	go s.driver(wg)
	go s.portMonitor(wg)

	return s
}

func (s *Serial) Terminate() {
	s.terminated = true
	if s.port != nil {
		s.port.Close()
	}
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
	if !s.connected {
		return 0, false
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
	a := uint16(0)
	bs := make([]byte, 0, 2)
	select {
	case bs = <-s.address:
		a = binary.LittleEndian.Uint16(bs)
		s.log.Tracef("Received address: %s", display.HexAddress(a))
	case <- time.After(5 * time.Second):
		s.log.Warnf("Address not received")
		return 0, false
	}
	return a, true
}
func (s *Serial) ReadOpCode() (uint8, bool) {
	if !s.connected {
		return 0, false
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
		s.log.Tracef("OpCode received %s", display.HexData(b))
		return b, true
	case <- time.After(5 * time.Second):
		s.log.Warnf("OpCode not received")
		return 0, false
	}
}
func (s *Serial) ReadData() (uint8, bool) {
	if !s.connected {
		return 0, false
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
		e := <-s.data
		s.log.Tracef("Received data: %s. Error code: %s", display.HexData(b), display.HexData(e))
		return b, e == 0
	case <- time.After(5 * time.Second):
		s.log.Warnf("Data not received")
		return 0, false
	}
}
func (s *Serial) ReadStatus() (uint8, bool) {
	if !s.connected {
		return 0, false
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
			s.log.Tracef("Status received %s", display.BinData(b))
			return b, true
		case <- time.After(5 * time.Second):
			s.log.Warnf("Status not received")
			return 0, false
	}
}
func (s *Serial) SetData(data uint8) bool {
	if !s.connected {
		return false
	}
	// Send command
	s.log.Debugf("Sending data: %s", display.HexData(data))
	if n, err := s.port.Write([]byte{0x44,data,0x0A}); err != nil {
		s.log.Errorf("Failed to send request to set data: %v", err)
		return false
	} else if n != 3 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 3, sent: %d", n)
		return false
	}

	// Read response
	e := <-s.data
	return e == 0
}
func (s *Serial) SetLines(data uint64) (uint8, bool) {
	if !s.connected {
		return 0, false
	}

	// Send command
	bs := []byte{'L', uint8(data >> 40), uint8(data >> 32), uint8(data >> 24), uint8(data >> 16), uint8(data >> 8), uint8(data), 0x0A}
	s.log.Debugf("%c [%s %s %s %s %s %s] %s", bs[0], display.HexData(bs[1]), display.HexData(bs[2]), display.HexData(bs[3]), display.HexData(bs[4]), display.HexData(bs[5]), display.HexData(bs[6]), display.HexData(bs[7]) )
	if n, err := s.port.Write(bs); err != nil {
		s.log.Errorf("Failed to send request to set lines: %v", err)
		return 0, false
	} else if n != 8 {
		s.log.Errorf("Unexpected number of bytes sent.  Expected 8, sent: %d", n)
		return 0, false
	}

	return s.ReadStatus()
}

func (s *Serial) portMonitor(wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		fmt.Println("PortMonitor Done")
		wg.Done()
	}()

	var err error
	tick := time.NewTicker(200 * time.Millisecond)
	for !s.terminated {
		select {
		case <- tick.C:
			if s.port == nil {
				if s.port, err = srl.Open(config.CLIConfig.Serial.PortName, s.mode); err != nil {
					s.port = nil
					if s.connected {
						s.connected = false
						s.connStatus(false)
					}
				} else {
					s.log.Infof("Opened port %s", config.CLIConfig.Serial.PortName)
					if !s.connected {
						s.connStatus(true)
						s.connected = true
					}
					s.readPort()
				}
			}
		}
	}
	tick.Stop()
	close(s.buffer)
}
func (s *Serial) readPort() {
	bs := make([]byte,100)
	defer func() {
		if r := recover(); r != nil {
			s.log.Errorf("Recovered Read panic: %v", r)
		}
	}()
	for !s.terminated {
		if n, err := s.port.Read(bs); err != nil {
			s.port.Close()
			s.port = nil
			s.log.Infof("Lost port %s", config.CLIConfig.Serial.PortName)
			return
		} else {
			for i := 0; i < n; i++ {
				s.buffer <- bs[i]
				bs[i] = 0
			}
			//s.log.Warnf("Unexpected number of bytes received: Wanted 1, Received %d", n)
		}
	}
	close(s.buffer)
}
func (s *Serial) driver(wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		fmt.Println("Driver Done")
		wg.Done()
	}()
	b, ok := byte(0), true
	for {
		if b, ok = <- s.buffer; !ok {
			break
		}
		s.log.Tracef("Inbound data: %s", string(b))
		switch b {
		case 'a':
			s.address <- []byte{<-s.buffer, <-s.buffer}
		case 'd':
			s.data <- <-s.buffer  // data
			s.data <- <-s.buffer  // error code
			s.log.Tracef("Forwarding 'd' data complete")
		case 'D':
			s.data <- <-s.buffer  // error code
			s.log.Tracef("Forwarding 'D' data complete")
		case 'o':
			s.opCode <- <-s.buffer
			s.log.Tracef("Forwarding 'o' data complete")
		case 's':
			s.status <- <-s.buffer
			s.log.Tracef("Forwarding 's' data complete")
		case 'c':
			s.clock.ClockLow()
		case 'C':
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
			s.log.Warnf("Unknown byte: %v", display.HexData(b))
		}
	}
	s.log.Warn("Stopped receiving")
}
func (s *Serial) ResetChannels() {
	for {
		select {
		case <-s.buffer:
		case <-s.address:
		case <-s.opCode:
		case <-s.data:
		default:
			if s.port != nil {
				s.port.Close()
			}
			return
		}
	}
}

func (s *Serial) Draw(t *display.Terminal, connected bool, initialize bool) {
	if initialize {
		t.Cls()
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
}
func (s *Serial) Process(input common.Input) bool {
	return true
}
func (s *Serial) PortViewer() common.UI {
	return s
}