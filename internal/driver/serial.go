package driver

import (
	"encoding/binary"
	"fmt"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"go.bug.st/serial"
	"strings"
)

func ListPorts() {
	ports, err := serial.GetPortsList()
	if err != nil {
		fmt.Printf("%v", err)
	}
	if len(ports) == 0 {
		fmt.Printf("%v", err)
	}
	display.Cls()
	display.PrintAt(Yellow + "Serial Ports" + Reset, 1,1)
	line := 2
	for _, port := range ports {
		if strings.HasPrefix(port, "/dev/cu") {
			display.PrintAt(fmt.Sprintf("%sFound port: %v%s", Blue, port, Reset), 4, line)
			line ++
		}
	}
	display.PressAnyKey()

}

func Connect() (serial.Port, DisplayMessage) {
	mode := &serial.Mode{
		DataBits: config.CLIConfig.Serial.DataBits,
		BaudRate: config.CLIConfig.Serial.BaudRate,
		StopBits: ToStopBits(config.CLIConfig.Serial.StopBits),
		Parity:   ToParity(config.CLIConfig.Serial.Parity),
	}
	port, err := serial.Open(config.CLIConfig.Serial.PortName, mode)
	if err != nil {
		return nil, DisplayMessage{ fmt.Sprintf("Failed to open USB port %s: %v", config.CLIConfig.Serial.PortName, err), true }
	}
	return port, DisplayMessage{ fmt.Sprintf("Openned port %s", config.CLIConfig.Serial.PortName), false }
}
func ToStopBits(value int) serial.StopBits {
	switch value {
	case 1: return serial.OneStopBit
	case 2: return serial.OnePointFiveStopBits
	case 3: return serial.TwoStopBits
	default:
		fmt.Println("Invalid StopBit")
		return serial.OneStopBit
	}
}
func ToParity(value int) serial.Parity {
	switch value {
	case 0: return serial.NoParity
	case 1: return serial.OddParity
	case 2: return serial.EvenParity
	case 3: return serial.MarkParity
	case 4: return serial.SpaceParity
	default:
		fmt.Println("Invalid StopBit")
		return serial.NoParity
	}
}

func ReadAddress() (uint16, DisplayMessage) {
	port, dm := Connect()
	if dm.IsError {
		return 0, dm
	}

	defer func() {
		if err := port.Close(); err != nil {
			fmt.Printf("Failed to close port: %v", err)
		}
	}()

	if n, err := port.Write([]byte("a\n")); err != nil {
		return 0, DisplayMessage{ fmt.Sprintf("Failed to send request for address: %v", err), true }
	} else if n != 2 {
		return 0, DisplayMessage{ fmt.Sprintf("Unexpected number of bytes sent.  Expected 2, sent: %d", n), true }
	}

	bs := make([]byte, 2)
	if n, err := port.Read(bs); err != nil {
		return 0, DisplayMessage{ fmt.Sprintf("Failed to receive response for address: %v", err), true }
	} else if n != 2 {
		return 0, DisplayMessage{ fmt.Sprintf("Unexpected number of bytes received.  Expected 2, received: %d", n), true }
	}

	a := binary.LittleEndian.Uint16(bs)

	return a, DisplayMessage{ fmt.Sprintf("Address received: %d", a), false }
}

func ReadData() (uint8, DisplayMessage) {
	port, dm := Connect()
	if dm.IsError {
		return 0, dm
	}

	defer func() {
		if err := port.Close(); err != nil {
			fmt.Printf("Failed to close port: %v", err)
		}
	}()

	if n, err := port.Write([]byte("r\n")); err != nil {
		return 0, DisplayMessage{ fmt.Sprintf("Failed to send request for data: %v", err), true }
	} else if n != 2 {
		return 0, DisplayMessage{ fmt.Sprintf("Unexpected number of bytes sent.  Expected 2, sent: %d", n), true }
	}

	var bs []byte
	if n, err := port.Read(bs); err != nil {
		return 0, DisplayMessage{ fmt.Sprintf("Failed to receive response for data: %v", err), true }
	} else if n != 1 {
		return 0, DisplayMessage{ fmt.Sprintf("Unexpected number of bytes received.  Expected 1, received: %d", n), true }
	}

	d := bs[0]
	return d, DisplayMessage{ fmt.Sprintf("Data received: %d", d), false }
}

