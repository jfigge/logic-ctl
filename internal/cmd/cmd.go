package cmd

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/spf13/cobra"
	"github.td.teradata.com/sandbox/logic-ctl/internal/config"
	"github.td.teradata.com/sandbox/logic-ctl/internal/driver"
	"github.td.teradata.com/sandbox/logic-ctl/internal/log"
	"os"
)

var cfgFile string
var romFile string

var rootCmd = &cobra.Command{
	Use:   "logic",
	Short: "logic is logic 1 breadboard cpu driver",
	RunE: func(cmd *cobra.Command, args []string) error {

		driver.NewDriver()

		//setup log config
		logConfig := log.NewLogConfigurator()
		log.Setup(logConfig)

		// Load 6502 rom
		if config.CLIConfig.RomFile == "" {
			log.Error("No rom specified.  Use -r/--rom <file> to specify")
			os.Exit(1)
		} else {
			driver.LoadRom(config.CLIConfig.RomFile)
		}

		driver.Run()
		return nil
	},
}

// Execute bootstraps the viper
func Execute() error {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "configuration file for logic")
	rootCmd.PersistentFlags().StringVarP(&romFile, "rom",    "r", "", "rom file for logic simulation")
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	if err := initConfigE(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
		return
	}
}

func initConfigE() error {
	defer func() {
		config.CLIConfig.RomFile = romFile
	}()
	return config.NewConfig(cfgFile)
}


























// go get github.com/tarm/serial
// go get github.com/jacobsa/go-serial/serial
// go get github.com/spf13/cobra

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	options := serial.OpenOptions{
		PortName:        "/dev/ttyACM0",
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	port, err := serial.Open(options)
	check(err)
	n, err := port.Write([]byte("read\n"))
	check(err)
	fmt.Println("written", n)
	buf := make([]byte, 100)
	n, err = port.Read(buf)
	check(err)
	fmt.Println("Readen", n)
	fmt.Println(string(buf))
}


/*
String inputString = "";         // a String to hold incoming data
boolean stringComplete = false;  // whether the string is complete
String state = "off";

void setup() {
    // initialize serial:
    Serial.begin(19200);
    // reserve 200 bytes for the inputString:
    inputString.reserve(200);
    pinMode(13, OUTPUT);
}

void loop() {
    // print the string when a newline arrives:
    if (stringComplete) {
        blink();
        if (inputString == "on\n") {
        state = "on";
        } else if (inputString == "off\n") {
        state = "off";
        } else if (inputString == "read\n") {
        Serial.println(state  );
        }
        // clear the string:
        inputString = "";
        stringComplete = false;
    }
}


void blink() {
    digitalWrite(13, HIGH);   // set the LED on
    delay(1000);              // wait for a second
    digitalWrite(13, LOW);    // set the LED off
    delay(1000);              // wait for a second
}

void serialEvent() {
    while (Serial.available()) {
        // get the new byte:
        char inChar = (char)Serial.read();
        // add it to the inputString:
        inputString += inChar;
        // if the incoming character is a newline, set a flag so the main loop can
        // do something about it:
        if (inChar == '\n') {
        stringComplete = true;
        }
    }
}
*/


/*
Instructions
------------
U = Unofficial
X = Freezes CPU, so not tested
? = Inconsistent/unknown behavior, so not tested

00   BRK #n
01   ORA (z,X)
02 X KIL
03 U SLO (z,X)
04 U DOP z
05   ORA z
06   ASL z
07 U SLO z
08   PHP
09   ORA #n
0A   ASL A
0B U AAC #n
0C U TOP abs
0D   ORA a
0E   ASL a
0F U SLO abs
10   BPL r
11   ORA (z),Y
12 X KIL
13 U SLO (z),Y
14 U DOP z,X
15   ORA z,X
16   ASL z,X
17 U SLO z,X
18   CLC
19   ORA a,Y
1A U NOP
1B U SLO abs,Y
1C U TOP abs,X
1D   ORA a,X
1E   ASL a,X
1F U SLO abs,X
20   JSR a
21   AND (z,X)
22 X KIL
23 U RLA (z,X)
24   BIT z
25   AND z
26   ROL z
27 U RLA z
28   PLP
29   AND #n
2A   ROL A
2B U AAC #n
2C   BIT a
2D   AND a
2E   ROL a
2F U RLA abs
30   BMI r
31   AND (z),Y
32 X KIL
33 U RLA (z),Y
34 U DOP z,X
35   AND z,X
36   ROL z,X
37 U RLA z,X
38   SEC
39   AND a,Y
3A U NOP
3B U RLA abs,Y
3C U TOP abs,X
3D   AND a,X
3E   ROL a,X
3F U RLA abs,X
40   RTI
41   EOR (z,X)
42 X KIL
43 U SRE (z,X)
44 U DOP z
45   EOR z
46   LSR z
47 U SRE z
48   PHA
49   EOR #n
4A   LSR A
4B U ASR #n
4C   JMP a
4D   EOR a
4E   LSR a
4F U SRE abs
50   BVC r
51   EOR (z),Y
52 X KIL
53 U SRE (z),Y
54 U DOP z,X
55   EOR z,X
56   LSR z,X
57 U SRE z,X
58   CLI
59   EOR a,Y
5A U NOP
5B U SRE abs,Y
5C U TOP abs,X
5D   EOR a,X
5E   LSR a,X
5F U SRE abs,X
60   RTS
61   ADC (z,X)
62 X KIL
63 U RRA (z,X)
64 U DOP z
65   ADC z
66   ROR z
67 U RRA z
68   PLA
69   ADC #n
6A   ROR A
6B U ARR #n
6C   JMP (a)
6D   ADC a
6E   ROR a
6F U RRA abs
70   BVS r
71   ADC (z),Y
72 X KIL
73 U RRA (z),Y
74 U DOP z,X
75   ADC z,X
76   ROR z,X
77 U RRA z,X
78   SEI
79   ADC a,Y
7A U NOP
7B U RRA abs,Y
7C U TOP abs,X
7D   ADC a,X
7E   ROR a,X
7F U RRA abs,X
80 U DOP #n
81   STA (z,X)
82 U DOP #n
83 U AAX (z,X)
84   STY z
85   STA z
86   STX z
87 U AAX z
88   DEY
89 U DOP #n
8A   TXA
8B ? XAA #n
8C   STY a
8D   STA a
8E   STX a
8F U AAX abs
90   BCC r
91   STA (z),Y
92 X KIL
93 ? AXA (z),Y
94   STY z,X
95   STA z,X
96   STX z,Y
97 U AAX z,Y
98   TYA
99   STA a,Y
9A   TXS
9B ? XAS abs,Y
9C U SYA abs,X
9D   STA a,X
9E U SXA abs,Y
9F ? AXA abs,Y
A0   LDY #n
A1   LDA (z,X)
A2   LDX #n
A3 U LAX (z,X)
A4   LDY z
A5   LDA z
A6   LDX z
A7 U LAX z
A8   TAY
A9   LDA #n
AA   TAX
AB U ATX #n
AC   LDY a
AD   LDA a
AE   LDX a
AF U LAX abs
B0   BCS r
B1   LDA (z),Y
B2 X KIL
B3 U LAX (z),Y
B4   LDY z,X
B5   LDA z,X
B6   LDX z,Y
B7 U LAX z,Y
B8   CLV
B9   LDA a,Y
BA   TSX
BB ? LAR abs,Y
BC   LDY a,X
BD   LDA a,X
BE   LDX a,Y
BF U LAX abs,Y
C0   CPY #n
C1   CMP (z,X)
C2 U DOP #n
C3 U DCP (z,X)
C4   CPY z
C5   CMP z
C6   DEC z
C7 U DCP z
C8   INY
C9   CMP #n
CA   DEX
CB U AXS #n
CC   CPY a
CD   CMP a
CE   DEC a
CF U DCP abs
D0   BNE r
D1   CMP (z),Y
D2 X KIL
D3 U DCP (z),Y
D4 U DOP z,X
D5   CMP z,X
D6   DEC z,X
D7 U DCP z,X
D8   CLD
D9   CMP a,Y
DA U NOP
DB U DCP abs,Y
DC U TOP abs,X
DD   CMP a,X
DE   DEC a,X
DF U DCP abs,X
E0   CPX #n
E1   SBC (z,X)
E2 U DOP #n
E3 U ISC (z,X)
E4   CPX z
E5   SBC z
E6   INC z
E7 U ISC z
E8   INX
E9   SBC #n
EA   NOP
EB U SBC #n
EC   CPX a
ED   SBC a
EE   INC a
EF U ISC abs
F0   BEQ r
F1   SBC (z),Y
F2 X KIL
F3 U ISC (z),Y
F4 U DOP z,X
F5   SBC z,X
F6   INC z,X
F7 U ISC z,X
F8   SED
F9   SBC a,Y
FA U NOP
FB U ISC abs,Y
FC U TOP abs,X
FD   SBC a,X
FE   INC a,X
FF U ISC abs,X


*/