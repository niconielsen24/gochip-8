package main

import (
	"gochip-8/cmd/chip"
	"gochip-8/cmd/util"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Size in screen pixels of one CHIP-8 pixel
	// Default is 10, which makes the window 640x320 pixels
	pixelScale = 10
	// CHIP-8 runs at roughly 600-700 instructions per second, so at 60 FPS
	// we need to run several cycles per frame
	cyclesPerFrame = 10
)

// keyMap maps each CHIP-8 key (the index) to a key on a QWERTY keyboard:
//
//	1 2 3 C      1 2 3 4
//	4 5 6 D  ->  Q W E R
//	7 8 9 E      A S D F
//	A 0 B F      Z X C V
var keyMap = [chip.KeyboardSize]int32{
	rl.KeyX, rl.KeyOne, rl.KeyTwo, rl.KeyThree,
	rl.KeyQ, rl.KeyW, rl.KeyE, rl.KeyA,
	rl.KeyS, rl.KeyD, rl.KeyZ, rl.KeyC,
	rl.KeyFour, rl.KeyR, rl.KeyF, rl.KeyV,
}

func main() {
	rom_path := ""
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			util.PrintHelp()
			os.Exit(0)
		}
		rom_path = os.Args[1]
	}

	rom_bytes, err := util.GetRomBytes(rom_path)
	if err != nil {
		util.PrintUsage()
		os.Exit(1)
	}

	c := chip.NewGoChip()
	c.Init()
	c.LoadROM(rom_bytes)

	rl.InitWindow(chip.DisplayWidth*pixelScale, chip.DisplayHeight*pixelScale, "GoChip-8")
	rl.SetTargetFPS(60)
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		for i, key := range keyMap {
			c.Keyboard[i] = rl.IsKeyDown(key)
		}

		for range cyclesPerFrame {
			c.Cycle()
		}
		// The loop runs at 60 FPS, which is the rate the timers tick at
		c.UpdateTimers()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		for x, column := range c.Display {
			for y, on := range column {
				if on {
					rl.DrawRectangle(int32(x*pixelScale), int32(y*pixelScale), pixelScale, pixelScale, rl.White)
				}
			}
		}

		rl.EndDrawing()
	}

}
