package main

import (
	"flag"
	"fmt"
	"github.com/huin/goserial"
	"log"
	"os"
)

// Create a new type for parity to handle cmd line parsing
type ParityFlag goserial.ParityMode

var (
	baud   *int = nil
	parity ParityFlag
)

// print parity
func (p *ParityFlag) String() string {
	s := fmt.Sprintf("ooops, inconsistent internal parity value: '%v'", *p)
	switch goserial.ParityMode(*p) {
	//	*s = ParityFlag(goserial.ParityEven)
	case goserial.ParityNone:
		s = "none"
	case goserial.ParityOdd:
		s = "odd"
	case goserial.ParityEven:
		s = "even"
	}
	return s
}

// parse parity
func (s *ParityFlag) Set(value string) error {
	log.Print("setting parity: '", value, "'")
	switch value {
	case "none", "n", "0":
		*s = ParityFlag(goserial.ParityNone)
	case "odd", "o":
		*s = ParityFlag(goserial.ParityOdd)
	case "even", "e":
		*s = ParityFlag(goserial.ParityEven)
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
	log.Print("setting parity: '", *s, "'")
	return nil
}

func ParseSerialFlags(args []string) *goserial.Config {
	flags := flag.NewFlagSet("serial settings", flag.ExitOnError)
	baud = flags.Int("baud", 57600, "baudrate")
	flags.Parse(args)
	c := &goserial.Config{Name: "", Baud: *baud}
	return c
}

func main() {
	baud = flag.Int("baud", 57600, "baudrate")
	flag.Var(&parity, "parity", "parity mode: none, even, odd")

	flag.Parse()
	log.Print("baud: ", *baud)
	log.Print("parity: ", &parity)
	log.Print("argv: ", flag.Args())
}

/*
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
