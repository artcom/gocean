package main

import (
	"flag"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"os"
)

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
		os.Exit(1)
	}
	log.Print("setting parity: '", *s, "'")
	return nil
}

func openPort(c *serial.Config) *serial.Port {
	log.Printf("open device: '%s'", c.Name)
	sp, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	return sp
}

type PacketParser struct {
	bytes         []byte
	data          []byte
	data_size     int
	opt_data      []byte
	opt_data_size int
	package_type  byte

	state_cb func()
}

func checksum_check(bytes []byte, checksum byte) {
}

func (pp *PacketParser) waiting_for_sync() {
	log.Printf("waiting for sync")
	// log.Printf("len=%d cap=%d %v\n", len(p.bytes), cap(p.bytes), p.bytes)
	//log.Printf("len=%d cap=%d '%v'\n", len(p.bytes), cap(p.bytes), p.bytes)
	if pp.bytes[len(pp.bytes)-1] == 0x55 {
		log.Printf("sync!")
		pp.state_cb = pp.reading_header
	} else {
		log.Printf("BAD sync ;(")
		pp.bytes = pp.bytes[:0]
	}
}

func (pp *PacketParser) reading_header() {
	log.Printf("reading_header")
	if len(pp.bytes) != 6 {
		return
	}

	checksum_check(pp.bytes[1:5], pp.bytes[5])

	pp.data_size = (((int(pp.bytes[1]) << 8) & 0xFF00) | int(pp.bytes[2]))
	pp.opt_data_size = int(pp.bytes[3])
	pp.package_type = pp.bytes[4]
	//puts "header complete(#{@package_type}: #{@data_size}, #{@opt_data_size})"
	//puts "(header) #{@bytes[0..4].map{|b| sprintf(' %02X', b)}.join}"

	pp.state_cb = pp.reading_data
}

func (pp *PacketParser) reading_data() {
	if len(pp.bytes) != 6+pp.data_size {
		return
	}
	log.Printf("reading_data")

	pp.data = pp.bytes[6:(6 + pp.data_size)]
	pp.state_cb = pp.reading_opt_data
}

func (pp *PacketParser) reading_opt_data() {
	if len(pp.bytes) != 6+pp.data_size+pp.opt_data_size {
		return
	}
	//#puts "opt data complete"
	pp.opt_data = pp.bytes[(6 + pp.data_size):(6 + pp.data_size + pp.opt_data_size)]
	pp.state_cb = pp.read_package_checksum
}

func (pp *PacketParser) read_package_checksum() {
	checksum_check(pp.bytes[6:(6+pp.data_size+pp.opt_data_size)], pp.bytes[len(pp.bytes)-1])

	log.Printf("package complete & ok: ", pp.bytes[6])
	/*
	   #s1 = @data.map{|b| sprintf(" %02X", b)}.join
	   #s2 = @opt_data.map{|b| sprintf(" %02X", b)}.join
	   #puts "(#{@bytes.size}) #{s1} | #{s2}"

	   @on_package && @on_package.call(
	   RadioTelegram.new(data: @data, opt_data: @opt_data)
	   )
	*/
	pp.reset()
}

func (pp *PacketParser) reset() {
	pp.bytes = pp.bytes[:0]
	pp.state_cb = pp.waiting_for_sync
}

func (p *PacketParser) push(c byte) {
	p.bytes = append(p.bytes, c)
	//log.Printf("len=%d cap=%d %v\n", len(p.bytes), cap(p.bytes), p.bytes)
	p.state_cb()
}

func loopread(port *serial.Port) {
	buf := make([]byte, 128)

	pp := PacketParser{}
	pp.reset()

	n, err := port.Read(buf) // initial read..
	// ..and than loop until EOF
	for ; err == nil; n, err = port.Read(buf) {
		log.Printf("(%d) :: %q", n, buf[:n])
		for _, c := range buf[:n] {
			pp.push(c)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	baud = flag.Int("baud", 57600, "baudrate")
	flag.Var(&parity, "parity", "parity mode: none, even, odd")

	flag.Parse()
	log.Print("baud: ", *baud)
	log.Print("parity: ", &parity)
	log.Print("argv: ", flag.Args())

	cfg := &serial.Config{Name: flag.Arg(0), Baud: *baud}
	port := openPort(cfg)
	loopread(port)
}

/*
func ParseSerialFlags(args []string) *goserial.Config {
	flags := flag.NewFlagSet("serial settings", flag.ExitOnError)
	baud = flags.Int("baud", 57600, "baudrate")
	flags.Parse(args)
	c := &goserial.Config{Name: "", Baud: *baud}
	return c
}

    jjj
	flags := flag.NewFlagSet("serial settings", flag.ExitOnError)
	baud = flags.Int("baud", 57600, "baudrate")
	parity = flags.Int("parity", 0, "parity mode: none, even, odd")
	log.Print("argv: ", os.Args[1:])
	log.Print("arg1: ", os.Args[1])
	log.Print("arg1: ", os.Args[1])
	log.Print("flag.Args: ", flags.Args())
	log.Print("flags.Parse: ", flags.Parse(os.Args))
	log.Print("baud: ", *baud)
	log.Print("argv: ", os.Args[1:])
	log.Print("flag.Args: ", flags.Args())
	log.Print("flag.NArg: ", flags.NArg())
}
*/
