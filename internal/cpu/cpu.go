package cpu

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const MEMORY_SIZE = 4096
const DISPLAY_SIZE = 64 * 32

type CPU struct {
	rng        *rand.Rand
	Memory     [MEMORY_SIZE]byte
	V          [16]uint8
	I          uint16
	PC         uint16
	SP         uint8
	Stack      [16]uint16
	DelayTimer uint8
	SoundTimer uint8
	Display    [DISPLAY_SIZE]byte
	Keypad     [16]bool
}

func (c *CPU) op00E0() {
	c.Display = [DISPLAY_SIZE]byte{}
}
func (c *CPU) op00EE() {
	if c.SP == 0x0 {
		return
	}
	c.SP--
	c.PC = c.Stack[c.SP]
}
func (c *CPU) op1NNN(NNN uint16) {
	c.PC = NNN
}
func (c *CPU) op2NNN(NNN uint16) {
	if c.SP >= 0x16 {
		return
	}
	c.Stack[c.SP] = c.PC
	c.SP++
	c.PC = NNN
}
func (c *CPU) op3XNN(X uint8, NN uint8) {
	if c.V[X] == NN {
		c.PC += 2
	}
}
func (c *CPU) op4XNN(X uint8, NN uint8) {
	if c.V[X] != NN {
		c.PC += 2
	}
}
func (c *CPU) op5XY0(X uint8, Y uint8) {
	if c.V[X] == c.V[Y] {
		c.PC += 2
	}
}
func (c *CPU) op6XNN(X uint8, NN uint8) {
	c.V[X] = NN
}
func (c *CPU) op7XNN(X uint8, NN uint8) {
	c.V[X] += NN
}
func (c *CPU) op8XY0(X uint8, Y uint8) {
	c.V[X] = c.V[Y]
}
func (c *CPU) op8XY1(X uint8, Y uint8) {
	c.V[X] = c.V[X] | c.V[Y]
}
func (c *CPU) op8XY2(X uint8, Y uint8) {
	c.V[X] = c.V[X] & c.V[Y]
}
func (c *CPU) op8XY3(X uint8, Y uint8) {
	c.V[X] = c.V[X] ^ c.V[Y]
}
func (c *CPU) op8XY4(X uint8, Y uint8) {
	borrow := uint8(0)
	if uint16(c.V[X])+uint16(c.V[Y]) > 0xFF {
		borrow = 1
	} else {
		borrow = 0
	}
	c.V[X] = c.V[X] + c.V[Y]
	c.V[0xF] = borrow
}
func (c *CPU) op8XY5(X uint8, Y uint8) {
	borrow := uint8(0)
	if c.V[X] >= c.V[Y] {
		borrow = 1
	} else {
		borrow = 0
	}
	c.V[X] = c.V[X] - c.V[Y]
	c.V[0xF] = borrow
}
func (c *CPU) op8XY6(X uint8, Y uint8) {
	c.V[0xF] = c.V[X] & 0x01
	c.V[X] >>= 1

}
func (c *CPU) op8XY7(X uint8, Y uint8) {
	borrow := uint8(0)
	if c.V[Y] >= c.V[X] {
		borrow = 1
	} else {
		borrow = 0
	}
	c.V[X] = c.V[Y] - c.V[X]
	c.V[0xF] = borrow
}
func (c *CPU) op8XYE(X uint8, Y uint8) {
	c.V[0xF] = (c.V[X] & 0x80) >> 7
	c.V[X] <<= 1
}
func (c *CPU) op9XY0(X uint8, Y uint8) {
	if c.V[X] != c.V[Y] {
		c.PC += 2
	}
}
func (c *CPU) opANNN(NNN uint16) {
	c.I = NNN
}
func (c *CPU) opBNNN(NNN uint16) {
	c.PC = NNN + uint16(c.V[0])
}
func (c *CPU) opCXNN(X uint8, NN uint8) {
	c.V[X] = uint8(c.rng.Intn(256)) & NN
}
func (c *CPU) opDXYN(X uint8, Y uint8, N uint8) {
	c.V[0xF] = 0
	for i := uint8(0); i < N; i++ {
		sprite := c.Memory[c.I+uint16(i)]
		for j := 0; j < 8; j++ {
			mask := uint8(0x80 >> j)
			if (sprite & mask) != 0 {
				screenX := (uint16(c.V[X]) + uint16(j)) % 64
				screenY := (uint16(c.V[Y]) + uint16(i)) % 32
				if c.Display[64*screenY+screenX] == 1 {
					c.V[0xF] = 1
				}
				c.Display[64*screenY+screenX] ^= 1
			}
		}
	}
}
func (c *CPU) opEX9E(X uint8) {
	key := c.V[X]
	if c.Keypad[key] {
		c.PC += 2
	}
}
func (c *CPU) opEXA1(X uint8) {
	key := c.V[X]
	if !c.Keypad[key] {
		c.PC += 2
	}
}
func (c *CPU) opFX07(X uint8) {
	c.V[X] = c.DelayTimer
}
func (c *CPU) opFX0A(X uint8) {
	for i := 0; i < 16; i++ {
		if c.Keypad[i] {
			c.V[X] = uint8(i)
			return
		}
	}
	c.PC -= 2
}
func (c *CPU) opFX15(X uint8) {
	c.DelayTimer = c.V[X]
}
func (c *CPU) opFX18(X uint8) {
	c.SoundTimer = c.V[X]
}
func (c *CPU) opFX1E(X uint8) {
	c.I += uint16(c.V[X])
}
func (c *CPU) opFX29(X uint8) {
	c.I = uint16(c.V[X]&0x0F) * 5
}
func (c *CPU) opFX33(X uint8) {
	num := c.V[X]
	c.Memory[c.I] = num / 100
	c.Memory[c.I+1] = (num / 10) % 10
	c.Memory[c.I+2] = num % 10
}
func (c *CPU) opFX55(X uint8) {
	for i := uint16(0); i <= uint16(X); i++ {
		c.Memory[c.I+i] = c.V[i]
	}
}
func (c *CPU) opFX65(X uint8) {
	for i := uint16(0); i <= uint16(X); i++ {
		c.V[i] = c.Memory[c.I+i]
	}
}

func (c *CPU) Step() {
	op := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])
	T := uint8((op & 0xF000) >> 12)
	X := uint8((op & 0x0F00) >> 8)
	Y := uint8((op & 0x00F0) >> 4)
	N := uint8(op & 0x000F)
	NN := uint8(op & 0x00FF)
	NNN := op & 0x0FFF

	c.PC += 2
	switch T {
	//MACHINE
	case 0x0:
		switch NN {
		case 0xE0:
			c.op00E0()
		case 0xEE:
			c.op00EE()
		}
	//JUMP
	case 0x1:
		c.op1NNN(NNN)
	case 0xB:
		c.opBNNN(NNN)
	//CALL
	case 0x2:
		c.op2NNN(NNN)
	//SKIP
	case 0x3:
		c.op3XNN(X, NN)
	case 0x4:
		c.op4XNN(X, NN)
	case 0x5:
		c.op5XY0(X, Y)
	case 0x9:
		c.op9XY0(X, Y)
	case 0xE:
		switch NN {
		case 0x9E:
			c.opEX9E(X)
		case 0xA1:
			c.opEXA1(X)
		}
	//SET REG
	case 0x6:
		c.op6XNN(X, NN)
	case 0xC:
		c.opCXNN(X, NN)
	//SET ADDR
	case 0xA:
		c.opANNN(NNN)
	//DRAW
	case 0xD:
		c.opDXYN(X, Y, N)
	//ADD
	case 0x7:
		c.op7XNN(X, NN)
	//MATH
	case 0x8:
		switch N {
		case 0x0:
			c.op8XY0(X, Y)
		case 0x1:
			c.op8XY1(X, Y)
		case 0x2:
			c.op8XY2(X, Y)
		case 0x3:
			c.op8XY3(X, Y)
		case 0x4:
			c.op8XY4(X, Y)
		case 0x5:
			c.op8XY5(X, Y)
		case 0x6:
			c.op8XY6(X, Y)
		case 0x7:
			c.op8XY7(X, Y)
		case 0xE:
			c.op8XYE(X, Y)
		}
	case 0xF:
		switch NN {
		case 0x07:
			c.opFX07(X)
		case 0x0A:
			c.opFX0A(X)
		case 0x15:
			c.opFX15(X)
		case 0x18:
			c.opFX18(X)
		case 0x1E:
			c.opFX1E(X)
		case 0x29:
			c.opFX29(X)
		case 0x33:
			c.opFX33(X)
		case 0x55:
			c.opFX55(X)
		case 0x65:
			c.opFX65(X)
		}
	}
}

func (c *CPU) LoadROM(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) > (MEMORY_SIZE - 0x200) {
		return fmt.Errorf("ROM is too large for memory")
	}

	copy(c.Memory[0x200:], data)
	return nil
}

func NewCPU() *CPU {
	c := &CPU{
		PC: 0x200,
	}

	source := rand.NewSource(time.Now().UnixNano())
	c.rng = rand.New(source)

	font := []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0,
		0x20, 0x60, 0x20, 0x20, 0x70,
		0xF0, 0x10, 0xF0, 0x80, 0xF0,
		0xF0, 0x10, 0xF0, 0x10, 0xF0,
		0x90, 0x90, 0xF0, 0x10, 0x10,
		0xF0, 0x80, 0xF0, 0x10, 0xF0,
		0xF0, 0x80, 0xF0, 0x90, 0xF0,
		0xF0, 0x10, 0x20, 0x40, 0x40,
		0xF0, 0x90, 0xF0, 0x90, 0xF0,
		0xF0, 0x90, 0xF0, 0x10, 0xF0,
		0xF0, 0x90, 0xF0, 0x90, 0x90,
		0xE0, 0x90, 0xE0, 0x90, 0xE0,
		0xF0, 0x80, 0x80, 0x80, 0xF0,
		0xE0, 0x90, 0x90, 0x90, 0xE0,
		0xF0, 0x80, 0xF0, 0x80, 0xF0,
		0xF0, 0x80, 0xF0, 0x80, 0x80,
	}

	copy(c.Memory[:], font)

	return c
}
