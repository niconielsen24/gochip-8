package util

import (
	"fmt"
	"io"
	"log"
	"os"
)

func PrintUsage() {
	fmt.Println("Usage: gochip-8 <ROM file>")
}

func PrintHelp() {
	fmt.Println("GoChip-8 is a CHIP-8 emulator written in Go.")
	fmt.Println("Usage: gochip-8 <ROM file>")
	fmt.Println("Supported ROM formats: .ch8")
	fmt.Println("Disclosure: \tSome roms may not work correctly due to differences in CHIP-8 implementations.")
	fmt.Println("\t\tXO-CHIP and SCHIP roms are not supported.")
	fmt.Println("\n\tAwesome CHIP-8 resources, visit: https://github.com/tobiasvl/awesome-chip-8")
	fmt.Println("\tRaylib resources, visit: https://github.com/gen2brain/raylib-go")
}

func GetRomBytes(path string) ([]byte, error) {
	rom_file, err := os.Open(path)
	if err != nil {
		log.Default().Fatalf("Failed to open ROM file: %v", err)
	}
	defer rom_file.Close()

	// Read the ROM file into a byte slice
	rom_bytes, err := io.ReadAll(rom_file)
	if err != nil {
		return nil, fmt.Errorf("failed to read ROM file: %v", err)
	}
	return rom_bytes, nil
}
