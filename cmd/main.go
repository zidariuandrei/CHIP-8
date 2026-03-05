package main

import (
	"chip8/internal/cpu"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	chip8 *cpu.CPU
}

func (g *Game) Update() error {
	keyMapping := map[ebiten.Key]uint8{
		ebiten.Key1: 0x1, ebiten.Key2: 0x2, ebiten.Key3: 0x3, ebiten.Key4: 0xC,
		ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
		ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
		ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
	}
	for i := range g.chip8.Keypad {
		g.chip8.Keypad[i] = false
	}

	for k, index := range keyMapping {
		if ebiten.IsKeyPressed(k) {
			g.chip8.Keypad[index] = true
		}
	}
	for range 10 {
		g.chip8.Step()
	}
	if g.chip8.DelayTimer > 0 {
		g.chip8.DelayTimer--
	}
	if g.chip8.SoundTimer > 0 {
		g.chip8.SoundTimer--
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	buffer := make([]byte, 64*32*4)
	for i := 0; i < 64*32; i++ {
		var color byte = 0
		if g.chip8.Display[i] == 1 {
			color = 255
		}
		buffer[i*4] = color
		buffer[i*4+1] = color
		buffer[i*4+2] = color
		buffer[i*4+3] = 255
	}
	screen.WritePixels(buffer)
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 64, 32
}
func main() {
	c := cpu.NewCPU()

	err := c.LoadROM("games/SpaceInvaders.ch8")
	if err != nil {
		log.Fatal(err)
	}
	ebiten.RunGame(&Game{chip8: c})
}
