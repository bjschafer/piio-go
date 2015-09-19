package main

import (
	"log"
	"os"

	"github.com/kidoman/embd"
)

// Pin definitions
const (
	Mcp23017IoDirA = 0x00
	Mcp23017IoDirB = 0x01
	Mcp23017GpioA  = 0x12
	Mcp23017GpioB  = 0x13
	Mcp23017GppuA  = 0x0C
	Mcp23017GppuB  = 0x0D
	Mcp23017OlatA  = 0x14
	Mcp23017OlatB  = 0x15
	Mcp23008GpioA  = 0x09
	Mcp23008GppuA  = 0x06
	Mcp23008OlatA  = 0x0A
)

// error log - to quit on err
var (
	Error *log.Logger
)

type PiioDigital struct {
	i2c       embd.I2CBus
	address   byte
	numGpios  uint
	direction byte
}

func (pd *PiioDigital) Init(address byte, numGpios int) {
	// initialize the Error!
	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	var err error

	if numGpios >= 0 && numGpios <= 16 {
		Error.Fatal("Number of Gpios must be between 0 and 16.")
	}

	pd.i2c = embd.NewI2CBus(address)
	pd.address = address
	pd.numGpios = uint(numGpios)

	// set defaults
	if numGpios <= 8 {
		if err := pd.i2c.WriteByte(Mcp23017IoDirA, 0xFF); err != nil { // all inputs on port A
			Error.Fatal("Unable to write byte.")
		}
		if pd.direction, err = pd.i2c.ReadByte(Mcp23017IoDirA); err != nil {
			Error.Fatal("Unable to write byte.")
		}
		if err := pd.i2c.WriteByte(Mcp23008GppuA, 0x00); err != nil {
			Error.Fatal("Unable to write byte.")
		}
	} else if numGpios > 8 {
		if err := pd.i2c.WriteByte(Mcp23017IoDirA, 0XFF); err != nil { // all inputs on port A
			Error.Fatal("Unable to write byte.")
		}
		if err := pd.i2c.WriteByte(Mcp23017IoDirB, 0XFF); err != nil { // all inpots on port B
			Error.Fatal("Unable to write byte.")
		}
		var t1, t2 byte
		if t1, err = pd.i2c.ReadByte(Mcp23017IoDirA); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		if t2, err = pd.i2c.ReadByte(Mcp23017IoDirB); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		pd.direction = t1 | (t2 << 8)
		if err := pd.i2c.WriteByte(Mcp23017GppuA, 0x00); err != nil {
			Error.Fatal("Unable to write byte.")
		}
		if err := pd.i2c.WriteByte(Mcp23017GppuB, 0x00); err != nil {
			Error.Fatal("Unable to write byte.")
		}
	}
}

func (pd *PiioDigital) changebit(bitmap byte, bit uint, value uint) byte {
	if value == 0 {
		return bitmap & ^(1 << bit)
	} else if value == 1 {
		return bitmap | (1 << bit)
	} else {
		Error.Fatalf("Value is %d, must be 0 or 1", value)
		return bitmap
	}
}

func (pd *PiioDigital) readAndChangePin(port byte, pin uint, value uint) byte {
	var err error
	if pin < 0 || pin >= pd.numGpios {
		Error.Fatalf("Pin number %d is invalid, only 0-%d are valid", pin, pd.numGpios)
	}
	var currvalue byte
	if currvalue, err = pd.i2c.ReadByte(port); err != nil {
		Error.Fatal("Unable to read byte.")
	}
	newvalue := pd.changebit(currvalue, pin, value)
	if err := pd.i2c.WriteByte(port, newvalue); err != nil {
		Error.Fatal("Unable to write byte.")
	}
	return newvalue
}

func (pd *PiioDigital) readAndChangePinWithCurrVal(port byte, pin uint, value uint, currval byte) byte {
	if pin < 0 || pin >= pd.numGpios {
		Error.Fatalf("Pin number %d is invalid, only 0-%d are valid", pin, pd.numGpios)
	}
	newvalue := pd.changebit(currval, pin, value)
	if err := pd.i2c.WriteByte(port, newvalue); err != nil {
		Error.Fatal("Unable to write byte.")
	}
	return newvalue
}

func (pd *PiioDigital) Pullup(pin uint, value uint) byte {
	if pd.numGpios <= 8 {
		return pd.readAndChangePin(Mcp23008GppuA, pin, value)
	} else {
		// if pin < 8 {
		// 	return -1
		// }
		return pd.readAndChangePin(Mcp23017GppuB, pin-8, value) << 8
	}
}

// set pin to either input or output mode
func (pd *PiioDigital) Config(pin uint, mode uint) byte {
	if pd.numGpios <= 8 {
		pd.direction = pd.readAndChangePin(Mcp23017IoDirA, pin, mode)
	} else if pd.numGpios <= 16 {
		if pin < 8 {
			pd.direction = pd.readAndChangePin(Mcp23017IoDirA, pin, mode)
		} else {
			pd.direction |= pd.readAndChangePin(Mcp23017IoDirB, pin-8, mode) << 8
		}
	}

	return pd.direction
}

func (pd *PiioDigital) output(pin uint, value uint) byte {
	var tmp byte
	var err error
	if pd.numGpios <= 8 {
		if tmp, err = pd.i2c.ReadByte(Mcp23008OlatA); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		return pd.readAndChangePinWithCurrVal(Mcp23008GpioA, pin, value, tmp)
	} else {
		if pin < 8 {
			if tmp, err = pd.i2c.ReadByte(Mcp23017OlatA); err != nil {
				Error.Fatal("Unable to read byte.")
			}
			return pd.readAndChangePinWithCurrVal(Mcp23017GpioA, pin, value, tmp)
		}
		if tmp, err = pd.i2c.ReadByte(Mcp23017OlatB); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		return pd.readAndChangePinWithCurrVal(Mcp23017GpioB, pin-8, value, tmp) << 8
	}
	// self.outputvalue = self.readandchangepin(Mcp23017IoDirA, pin, value, self.outputvalue)
}

func (pd *PiioDigital) Output(pin uint, value uint) byte {
	return pd.readAndChangePinWithCurrVal(Mcp23017IoDirA, pin, value, pd.output(pin, value))
}

func (pd *PiioDigital) Input(pin uint) byte {
	var err error
	if pin < 0 || pin >= pd.numGpios {
		Error.Fatalf("Pin number %d is invalid, only 0-%d are valid", pin, pd.numGpios)
	}
	if pd.direction&(1<<pin) != 0 {
		Error.Fatalf("Pin %d not set to input", pin)
	}

	var value, value2 byte

	if pd.numGpios <= 8 {
		if value, err = pd.i2c.ReadByte(Mcp23008GpioA); err != nil {
			Error.Fatal("Unable to read byte.")
		}
	} else if pd.numGpios > 8 && pd.numGpios <= 16 {
		if value, err = pd.i2c.ReadByte(Mcp23017GpioA); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		if value2, err = pd.i2c.ReadByte(Mcp23017GpioB); err != nil {
			Error.Fatal("Unable to read byte.")
		}
		value = value | (value2 << 8)
	}

	return value & (1 << pin)
}
