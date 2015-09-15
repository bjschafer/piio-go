package piio

import (
	"Error"
	"fmt"
	"os"

	"github.com/kidoman/embd/host/generic"
)

const (
	MCP23017_IODIRA = 0x00
	MCP23017_IODIRB = 0x01
	MCP23017_GPIOA  = 0x12
	MCP23017_GPIOB  = 0x13
	MCP23017_GPPUA  = 0x0C
	MCP23017_GPPUB  = 0x0D
	MCP23017_OLATA  = 0x14
	MCP23017_OLATB  = 0x15
	MCP23008_GPIOA  = 0x09
	MCP23008_GPPUA  = 0x06
	MCP23008_OLATA  = 0x0A
)

var (
	Error *Log.Logger
)

type piio_digital struct {
	i2c       i2cBus
	address   byte
	numGpios  int
	direction byte
}

func (this piio_digital) Init(address byte, numGpios int) {
	// initialize the Error!
	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	if numGpios >= 0 && numGpios <= 16 {
		Error.Fatal("Number of GPIOs must be between 0 and 16.")
	}

	this.i2c = NewI2CBus(0)
	this.i2c.address = address
	this.address = address
	this.numGpios = numGpios

	// set defaults
	if numGpios <= 8 {
		if _, err := self.i2c.WriteByte(MCP23017_IODIRA, 0xFF); err != nil { // all inputs on port A
			Error.Fatal("Unable to write byte.")
		}
		if this.direction, err = this.i2c.ReadByte(MCP23017_IODIRA); err != nil {
			Error.Fatal("Unable to write byte.")
		}
		if _, err := this.i2c.WriteByte(MCP23008_GPPUA, 0x00); err != nil {
			Error.Fatal("Unable to write byte.")
		}
	} else if numGpios > 8 {
		this.i2c.WriteByte(MCP23017_IODIRA, 0XFF) // all inputs on port A
		this.i2c.WriteByte(MCP23017_IODIRB, 0XFF) // all inpots on port B
		this.direction = this.i2c.ReadByte(MCP23017_IODIRA) | (this.i2c.ReadByte(MCP23017_IODIRB) << 8)
		this.i2c.WriteByte(MCP23017_GPPUA, 0x00)
		this.i2c.WriteByte(MCP23017_GPPUB, 0x00)
	}
}

func (this piio_digital) changebit(bitmap int, bit int, value int) {
	if value == 0 {
		return bitmap & ^(1 << bit)
	} else if value == 1 {
		return bitmap | (1 << bit)
	} else {
		Error.Fatalf("Value is %d, must be 0 or 1", value)
	}
}

func (this piio_digital) readAndChangePin(port int, pin int, value int, currvalue int) int {
	if pin < 0 || pin >= this.numGpios {
		Error.Fatalf("Pin number %d is invalid, only 0-%d are valid", pin, this.numGpios)
	}

	if currvalue == -1 {
		currvalue = this.i2c.ReadByte(port)
	}

	newvalue := this.changebit(currvalue, pin, value)
	this.i2c.WriteByte(port, newvalue)
	return newvalue
}

func (this piio_digital) Pullup(pin int, value int) {
	if this.numGpios <= 8 {
		return this.readAndChangePin(MCP23008_GPPUA, pin, value, -1)
	} else if this.numGpios <= 16 {
		lvalue := this.readAndChangePin(MCP23017_GPPUA, pin, value, -1)
		if pin < 8 {
			return -1
		} else {
			return this.readAndChangePin(MCP23017_GPPUB, pin-8, value) << 8
		}
	}
}

// set pin to either input or output mode
func (this piio_digital) Config(pin int, mode int) int {
	if this.numGpios <= 8 {
		this.direction = this.readAndChangePin(MCP23017_IODIRA, pin, mode)
	} else if this.numGpios <= 16 {
		if pin < 8 {
			this.direction = this.readAndChangePin(MCP23017_IODIRA, pin, mode)
		} else {
			this.direction |= this.readAndChangePin(MCP23017_IODIRB, pin-8, mode) << 8
		}
	}

	return this.direction
}

func (this piio_digital) Output(pin int, value int) int {
	if this.numGpios <= 8 {
		return this.readAndChangePin(MCP23008_GPIOA, pin, value, this.i2c.ReadByte(MCP23008_OLATA))
	} else if this.numGpios <= 16 {
		if pin < 8 {
			return this.readAndChangePin(MCP23017_GPIOA, pin, value, this.i2c.ReadByte(MCP23017_OLATA))
		} else {
			return this.readAndChangePin(MCP23017_GPIOB, pin-8, value, this.i2c.ReadByte(MCP23017_OLATB)) << 8
		}
	}
	// self.outputvalue = self._readandchangepin(MCP23017_IODIRA, pin, value, self.outputvalue)
}

func (this piio_digital) Input(pin int) int {
	if pin < 0 || pin >= this.numGpios {
		Error.Fatalf("Pin number %d is invalid, only 0-%d are valid", pin, this.numGpios)
	}
	if this.direction&(1<<pin) != 0 {
		Error.Fatalf("Pin %d not set to input", pin)
	}

	var value int

	if this.numGpios <= 8 {
		value = this.i2c.ReadByte(MCP23008_GPIOA)
	} else if this.numGpios > 8 && this.numGpios <= 16 {
		value = this.i2c.ReadByte(MCP23017_GPIOA)
		value |= this.i2c.ReadByte(MCP23017_GPIOB) << 8
	}

	return value & (1 << pin)
}
