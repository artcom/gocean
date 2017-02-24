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
var dbg = log.New(os.Stdout, "", 0)

// the only enocean radio telegram packet format we can actually decode
type IQfyDruckSensor struct {
	t *RadioTelegram
}

func (s *IQfyDruckSensor) sensor_id() string {
	//@sensor_id ||= @data[3..5].map{|b| sprintf('%02X', b)}.join(' ')
	return fmt.Sprintf("%x", s.t.data()[3:6])
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
	return fmt.Sprintf("<%s:%v>", s.sensor_id(), s.state())
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

	state_cb      func()
	packetHandler func(s *IQfyDruckSensor)
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
	if pp.packetHandler != nil {
		pp.packetHandler(s)
	}
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
	p.state_cb()
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

// open named device fill and read input stream byte by byte to push function
func loopread(devname string, baud int, push func(byte)) {

	port := openPort(&serial.Config{Name: devname, Baud: baud})
	buf := make([]byte, 128)

	log.Println("start reading device now...")
	n, err := port.Read(buf) // initial read..
	// ..and than loop until EOF
	for ; err == nil; n, err = port.Read(buf) {
		dbg.Printf("(%d) :: %q", n, buf[:n])
		// bytes are pushed one by ony and processed according to
		// parse/packet state
		for _, c := range buf[:n] {
			push(c) // forward byte to the next processing level
		}
	}

	if err != nil {
		log.Println(err)
	}
}

// LogWriter local type definition to suppress/enable debug outputs on demand
/*
type LogWriter struct {
	io.Writer
}

func (w *LogWriter) enable()  { w.Writer = os.Stdout }
func (w *LogWriter) disable() { w.Writer = ioutil.Discard }
*/

/*  list of sensor ids which are reported. Can be controlled by command line
*  option. When no option is given, all received sensors are reported. This
*  makes sense when you need to define a whitelist of sensors which are
*  actually recognized */
type idList []string

func (l *idList) contains(s string) bool {
	for _, e := range *l {
		if e == s {
			return true
		}
	}
	return false
}

// flag.Value interface
func (l *idList) String() string {
	return fmt.Sprint("-->%v<--", *l)
}

// flag.Value interface
func (l *idList) Set(val string) error {
	*l = append(*l, val)
	// dbg.Printf("only id list: %v\n", *l)
	return nil
}

func main() {
	// output control, activate -v for debugging
	verbose := flag.Bool("verbose", false, "verbosity, print debug/info")
	flag.BoolVar(verbose, "v", false, "--verbose (same)")

	// prepend log output with or without timestamp prefix.
	tslp := flag.Bool("ts", false, "timestamp log prefix")

	// serial line control, only baudrate for now,
	baud := flag.Int("baud", 57600, "baudrate")
	//flag.Var(&parity, "parity", "parity mode: none, even, odd")

	var idList idList
	flag.Var(&idList, "id", "comma separeted list of included sensor ids. If given, only sensor ids from this list are reported")

	// scan cmdline
	flag.Parse()

	// enable debug output only on demand, default be quiet
	if !*verbose {
		dbg.SetOutput(ioutil.Discard)
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
	dbg.Println("sensor id list: ", idList)

	// create parser instance in default start state
	pp := PacketParser{}
	pp.reset()
	pp.packetHandler = func(s *IQfyDruckSensor) {
		dbg.Printf("->%s<-\n", s.sensor_id())
		if len(idList) == 0 || idList.contains(s.sensor_id()) {
			log.Println(s) // print current button state on stdout
		}
	}

	// start portreading and pushing bytewise to the parser
	loopread(flag.Arg(0), *baud, func(c byte) { pp.push(c) })
}
