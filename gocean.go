package main

import (
	"flag"
	"fmt"
	"github.com/sigurn/crc8"
	"github.com/tarm/serial"
	"io/ioutil"
	"log"
	"os"
	//"reflect"
)

// debug output only, can be switched on with -verbose cmdline flag
var dbg = log.New(ioutil.Discard, "", 0)

// the only enocean radio telegram packet format we can actually decode
type IQfyDruckSensor struct {
	t *RadioTelegram
}

func (s *IQfyDruckSensor) sensor_id() []byte {
	//@sensor_id ||= @data[3..5].map{|b| sprintf('%02X', b)}.join(' ')
	return s.t.data()[3:6]
}

// returns "up" or "down" depending on sonsor state
func (s *IQfyDruckSensor) state() string {
	// return (s.t.data()[1]>>4&0x01 == 1)
	b := "up"
	if s.t.data()[1]>>4&0x01 == 1 {
		b = "down"
	}
	return b
}
func (s *IQfyDruckSensor) String() string {
	//"<#{sensor_id}:#{state}>"
	return fmt.Sprintf("<%x:%v>", s.sensor_id(), s.state())
}

// the basic packet structure for messages comming from the enocean usb serial
type RadioTelegram struct {
	raw []byte
}

func (t *RadioTelegram) String() string {
	return fmt.Sprintf(
		"type(%v), size(%v), opt size(%v), choice(%v)",
		t.packet_type(), t.data_size(), t.opt_data_size(), t.choice())
}
func (t *RadioTelegram) push(c byte) *RadioTelegram {
	t.raw = append(t.raw, c)
	return t
}
func (t *RadioTelegram) data_size() int {
	return (((int(t.raw[1]) << 8) & 0xFF00) | int(t.raw[2]))
}
func (t *RadioTelegram) packet_type() byte {
	return t.raw[4]
}
func (t *RadioTelegram) data() []byte {
	return t.raw[6 : 6+t.data_size()]
}
func (t *RadioTelegram) choice() byte {
	return t.data()[0]
}
func (t *RadioTelegram) opt_data_size() int {
	return int(t.raw[3])
}
func (t *RadioTelegram) opt_data() []byte {
	return t.raw[6+t.data_size() : 6+t.data_size()+t.opt_data_size()]
}
func (t *RadioTelegram) tail() []byte {
	return t.raw[6 : 6+t.data_size()+t.opt_data_size()]
	//pp.bytes[6:(6+pp.data_size+pp.opt_data_size)]
}
func (t *RadioTelegram) packet_checksum() byte {
	return t.raw[len(t.raw)-1]
}

// when this telegram is actually a coming from a sensor a sensor struct is
// created and returned, otherwise you get nil
func (t *RadioTelegram) IQfyDruckSensor() *IQfyDruckSensor {
	if t.choice() == 0xF6 {
		return &IQfyDruckSensor{t}
	}

	log.Printf("warn: #{choice} unknown packet choice: %v", t.choice())
	return nil
}

// packet parser knows about the structure of the telegram and decodes the
// sync and checksum bytes from the serial binary byte stream coming from the
// usb device into a stream of RadioTelegram structs
type PacketParser struct {
	telegram *RadioTelegram

	bytes         []byte
	data          []byte
	data_size     int
	opt_data      []byte
	opt_data_size int
	packet_type   byte

	state_cb func()
}

func checksum_check(bytes []byte, expected byte) bool {
	table := crc8.MakeTable(crc8.CRC8)
	crc := crc8.Checksum(bytes, table)
	if crc != expected {
		dbg.Printf("checksum error: %X != %X\n", crc, expected)
	}
	return crc == expected
}

func (pp *PacketParser) waiting_for_sync() {
	dbg.Printf("waiting for sync")
	if pp.bytes[0] == 0x55 {
		dbg.Printf("sync!")
		pp.state_cb = pp.reading_header
	} else {
		dbg.Printf("BAD sync ;(")
		pp.reset()
	}
}

func (pp *PacketParser) reading_header() {
	dbg.Printf("reading_header")
	if len(pp.bytes) != 6 {
		return
	}

	dbg.Printf("checking header bytes")
	bytes, checksum := pp.bytes[1:5], pp.bytes[5]
	if !checksum_check(bytes, checksum) {
		log.Printf("error in header checksum: %X\n", checksum)
		return
	}
	//dbg.Print("packet: ", pp.telegram)

	// advaning to next parsing state
	pp.state_cb = pp.reading_data
}

func (pp *PacketParser) reading_data() {
	if len(pp.telegram.raw) != 6+pp.telegram.data_size() {
		return
	}
	dbg.Printf("reading_data")

	pp.state_cb = pp.reading_opt_data
}

func (pp *PacketParser) reading_opt_data() {
	dsize := pp.telegram.data_size() + pp.telegram.opt_data_size()
	if len(pp.telegram.raw) != 6+dsize {
		return
	}

	pp.state_cb = pp.read_packet_checksum
}

func (pp *PacketParser) read_packet_checksum() {
	if !checksum_check(pp.telegram.tail(), pp.telegram.packet_checksum()) {
		log.Println("error in packet checksum, dumping packet")
		pp.reset()
		return
	}

	s := pp.telegram.IQfyDruckSensor()
	dbg.Print("packet complete & ok: ", pp.telegram) // pp.bytes[6])
	pp.reset()

	if s == nil {
		dbg.Print("warn: IQfyDruckSensorrno???")
		return
	}

	// print current button state on stdout
	log.Println(s)
}

// reset: starting next packet parse and clearing whatever state we had until
// now
func (pp *PacketParser) reset() {
	pp.bytes = pp.bytes[:0]
	pp.state_cb = pp.waiting_for_sync
	pp.telegram = &RadioTelegram{}
}

// each time a new bytes arrives the state callback is invoked
func (p *PacketParser) push(c byte) {
	p.telegram.push(c)
	p.bytes = append(p.bytes, c)
	//log.Print(p.telegram)
	p.state_cb()
}

func loopread(port *serial.Port) {
	buf := make([]byte, 128)

	pp := PacketParser{}
	pp.reset()

	log.Println("start reading device now...")
	n, err := port.Read(buf) // initial read..
	// ..and than loop until EOF
	for ; err == nil; n, err = port.Read(buf) {
		dbg.Printf("(%d) :: %q", n, buf[:n])
		// bytes are pushed one by ony and processed according to
		// parse/packet state
		for _, c := range buf[:n] {
			pp.push(c)
		}
	}

	if err != nil {
		log.Println(err)
	}
}

// take config and return an open serial port ready for reading
func openPort(c *serial.Config) *serial.Port {
	dbg.Printf("open device: '%s'", c.Name)

	sp, err := serial.OpenPort(c)
	if err != nil {
		fmt.Fprintln(os.Stderr, " ## FATAL:", err)
		os.Exit(-1)
	}

	return sp
}

// LogWriter local type definition to suppress/enable debug outputs on demand
/*
type LogWriter struct {
	io.Writer
}

func (w *LogWriter) enable()  { w.Writer = os.Stdout }
func (w *LogWriter) disable() { w.Writer = ioutil.Discard }
*/

func main() {

	// output control, activate -v for debugging
	verbose := flag.Bool("verbose", false, "verbosity, print debug/info")
	flag.BoolVar(verbose, "v", false, "--verbose (same)")

	// prepend log output with or without timestamp prefix.
	tslp := flag.Bool("ts", false, "timestamp log prefix")

	// serial line control, only baudrate for now,
	baud := flag.Int("baud", 57600, "baudrate")
	//flag.Var(&parity, "parity", "parity mode: none, even, odd")

	// scan cmdline
	flag.Parse()

	// enable debug output only on demand, default be quiet
	if *verbose {
		dbg.SetOutput(os.Stdout)
	}
	// debug output always be prefixed to be recognizable
	dbg.SetPrefix("(debug) ")

	// enable timestamp logging prefix only when -ts option is present
	if *tslp {
		dbg.SetFlags(log.LstdFlags)
		log.SetFlags(log.LstdFlags)
	} else {
		dbg.SetFlags(0)
		log.SetFlags(0)
	}

	// debug output of actuall settings
	dbg.Println("argv: ", flag.Args())
	dbg.Println("baud: ", *baud)
	dbg.Println("verbosity: ", *verbose)
	// log.Print("parity: ", &parity)

	port := openPort(&serial.Config{Name: flag.Arg(0), Baud: *baud})
	loopread(port)
}

/* parity flag option and code not actually use, so commented it


// Create a new type for parity to handle cmd line parsing
type ParityFlag serial.Parity

var (
	baud   *int       = nil
	parity ParityFlag = ParityFlag(serial.ParityNone)
)

// print parity
func (p *ParityFlag) String() string {
	s := fmt.Sprintf("ooops, inconsistent internal parity value: '%v'", *p)
	switch serial.Parity(*p) {
	case serial.ParityNone:
		s = "none"
	case serial.ParityOdd:
		s = "odd"
	case serial.ParityEven:
		s = "even"
	}
	return s
}

// parse parity
func (s *ParityFlag) Set(value string) error {
	log.Print("setting parity: '", value, "'")
	switch value {
	case "none", "n", "0":
		*s = ParityFlag(serial.ParityNone)
	case "odd", "o":
		*s = ParityFlag(serial.ParityOdd)
	case "even", "e":
		*s = ParityFlag(serial.ParityEven)
	default:
		flag.PrintDefaults()
		//os.Exit(1)
	}
	log.Print("setting parity: '", *s, "'")
	return nil
}
*/
