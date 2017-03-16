package main

import (
	"flag"
	"github.com/stianeikeland/go-rpio"
	"log"
)

var (
	gpio int = -1
)

func init() {
	// raspberry GPIO pin signaling only when flag is given
	flag.IntVar(&gpio, "gpio", -1, "GPIO pin #no output, -1 means no gpio")

	AppendHandlerPrepper(register_gpio_handler)
}

//
func register_gpio_handler() {
	// this handler only registers when actually requested on cmd line
	if gpio == -1 {
		return
	}

	// aquire GPIO handle for switching leds
	pin, err := openGpioPin(gpio)
	if err != nil {
		log.Printf(" ## GPIO(%v) can't be opend (ignored): %v", gpio, err)
		return
	}

	h := &GPIOHandler{pin: pin}
	AppendHandler(h)
}

//func (pin rpio.Pin) toggle_gpio_button(s *IQfyDruckSensor) {
type GPIOHandler struct {
	pin rpio.Pin
}

func (h GPIOHandler) handle_sensor_packet(s *IQfyDruckSensor) {
	log.Println(s) // print current button state on stdout
	if s.state() == "down" {
		h.pin.High()
	} else {
		h.pin.Low()
	}
}

// raspberry GPIO connection
func openGpioPin(no int) (pin rpio.Pin, err error) {
	if err := rpio.Open(); err != nil {
		return pin, err
		//fmt.Println(err)
		//os.Exit(1)
	}

	pin = rpio.Pin(no)
	pin.Output()

	return pin, err
}
