package main

import (
	"fmt"
	"time"
)

func main() {
	digital := PiioDigital{}
	digital.Init(1, 16)

	digital.Config(0, 0) // set pin 0 to output
	digital.Config(1, 1) // set pin 1 to input

	digital.Config(2, 1) // set pin 2 to input
	digital.Pullup(2, 1) // enable pullup resistor on pin 2

	// Read input pin and display the results
	fmt.Printf("Pin 2 = %d", digital.Input(2)>>3)

	// Python speed test on output 0 toggling at max speed
	fmt.Println("Starting blinky on pin 0 (C-c to quit)")

	for {
		digital.Output(0, 1) // Pin 0: High
		time.Sleep(1)
		digital.Output(0, 0) // Pin 0: Low
		time.Sleep(1)
	}
}
