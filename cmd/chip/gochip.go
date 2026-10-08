package chip

import (
	"log"
	"math/rand"
)

// Stack has no overflow or underflow checks,
// as the CHIP-8 spec does not define what happens in those cases.
// It is up to the program to ensure that it does not exceed the stack size.
// Otherwise, the program will crash with a runtime error.

type GoChip struct {
	opcode          Opcode
	program_counter MemAddress
	stack_pointer   StackPointer
	memory          [MemorySize]Word
	stack           [StackSize]MemAddress
	registers       [RegisterCount]Word
	index_register  MemAddress
	delay_timer     TimerDelay
	sound_timer     TimerDelay
	Display         [DisplayWidth][DisplayHeight]bool
	Keyboard        [KeyboardSize]bool
}

func NewGoChip() *GoChip {
	return &GoChip{
		opcode:          ZeroOpCode,
		program_counter: ProgramCounterStart, // Text starts at 0x200
		stack_pointer:   StackPointerStart,
		memory:          [MemorySize]Word{},
		stack:           [StackSize]MemAddress{},
		registers:       [RegisterCount]Word{},
		index_register:  IndexRegisterStart,
		delay_timer:     DelayTimerStart,
		sound_timer:     SoundTimerStart,
		Display:         [DisplayWidth][DisplayHeight]bool{},
		Keyboard:        [KeyboardSize]bool{},
	}
}

func (chip *GoChip) Init() {
	// Initialize the font set
	for i, char := range ChipFontSet {
		chip.memory[i] = Word(char)
	}
}

func (chip *GoChip) LoadROM(rom []byte) {
	for i, b := range rom {
		chip.memory[ProgramCounterStart+MemAddress(i)] = Word(b)
	}
}

// main loop for the chip, this will be called every frame
func (chip *GoChip) Cycle() {
	// Fetch the opcode from memory
	upperByte := chip.memory[chip.program_counter]
	lowerByte := chip.memory[chip.program_counter+1]
	opcode := Opcode(upperByte)<<8 | Opcode(lowerByte)

	chip.opcode = opcode

	switch decode(opcode) {
	case OpClearScreen:
		chip.clearScreen()
		chip.advanceProgramCounter()
	case OpReturnFromSubroutine:
		// Return to SP - 1 because otherwise we would be returning to the next instruction after the call
		chip.stack_pointer--
		chip.program_counter = chip.stack[chip.stack_pointer]
		chip.advanceProgramCounter()
	case OpSysCallMachineRoutine:
		// Ignore 0x0NNN (SYS addr) as it's not used in modern interpreters
		chip.advanceProgramCounter()
	case OpJumpToAddress:
		// Jump to address NNN
		address := opcode & 0x0FFF
		chip.program_counter = MemAddress(address)
	case OpCallSubroutine:
		chip.stack[chip.stack_pointer] = chip.program_counter
		chip.stack_pointer++
		chip.program_counter = MemAddress(opcode & 0x0FFF)
	case OpSkipIfVxEqualsByte:
		x := (opcode & 0x0F00) >> 8
		if chip.registers[x] == (Word)(opcode&0x00FF) {
			// Skip the next instruction
			chip.advanceProgramCounter()
		}
		// Advance the program counter to the next instruction
		chip.advanceProgramCounter()
	case OpSkipIfVxNotEqualsByte:
		x := (opcode & 0x0F00) >> 8
		if chip.registers[x] != (Word)(opcode&0x00FF) {
			// Skip the next instruction
			chip.advanceProgramCounter()
		}
		// Advance the program counter to the next instruction
		chip.advanceProgramCounter()
	case OpSkipIfVxEqualsVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		if chip.registers[x] == chip.registers[y] {
			// Skip the next instruction
			chip.advanceProgramCounter()
		}
		// Advance the program counter to the next instruction
		chip.advanceProgramCounter()
	case OpLoadByteIntoVx:
		x := (opcode & 0x0F00) >> 8
		chip.registers[x] = Word(opcode & 0x00FF)
		chip.advanceProgramCounter()
	case OpAddByteToVx:
		x := (opcode & 0x0F00) >> 8
		chip.registers[x] += Word(opcode & 0x00FF)
		chip.advanceProgramCounter()
	case OpLoadVyIntoVx:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		chip.registers[x] = chip.registers[y]
		chip.advanceProgramCounter()
	case OpOrVxVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		chip.registers[x] |= chip.registers[y]
		chip.advanceProgramCounter()
	case OpAndVxVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		chip.registers[x] &= chip.registers[y]
		chip.advanceProgramCounter()
	case OpXorVxVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		chip.registers[x] ^= chip.registers[y]
		chip.advanceProgramCounter()
	case OpAddVyToVxWithCarry:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		sum := uint16(chip.registers[x]) + uint16(chip.registers[y])
		chip.registers[x] = Word(sum & 0xFF)
		// VF is written last so the flag wins when X is 0xF
		if sum > 0xFF {
			chip.registers[0xF] = 1 // Set carry flag
		} else {
			chip.registers[0xF] = 0 // Clear carry flag
		}
		chip.advanceProgramCounter()
	case OpSubVyFromVx:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		noBorrow := chip.registers[x] >= chip.registers[y]
		chip.registers[x] -= chip.registers[y]
		if noBorrow {
			chip.registers[0xF] = 1 // Set NOT borrow flag
		} else {
			chip.registers[0xF] = 0 // Clear NOT borrow flag
		}
		chip.advanceProgramCounter()
	case OpSubVxFromVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		noBorrow := chip.registers[y] >= chip.registers[x]
		chip.registers[x] = chip.registers[y] - chip.registers[x]
		if noBorrow {
			chip.registers[0xF] = 1 // Set NOT borrow flag
		} else {
			chip.registers[0xF] = 0 // Clear NOT borrow flag
		}
		chip.advanceProgramCounter()
	case OpShiftVxRight:
		x := (opcode & 0x0F00) >> 8
		lsb := chip.registers[x] & 0x1
		chip.registers[x] >>= 1
		chip.registers[0xF] = lsb // Store least significant bit in VF
		chip.advanceProgramCounter()
	case OpShiftVxLeft:
		x := (opcode & 0x0F00) >> 8
		msb := chip.registers[x] >> 7
		chip.registers[x] <<= 1
		chip.registers[0xF] = msb // Store most significant bit in VF
		chip.advanceProgramCounter()
	case OpSkipIfVxNotEqualsVy:
		x := (opcode & 0x0F00) >> 8
		y := (opcode & 0x00F0) >> 4
		if chip.registers[x] != chip.registers[y] {
			// Skip the next instruction
			chip.advanceProgramCounter()
		}
		chip.advanceProgramCounter()
	case OpLoadAddressIntoIndex:
		chip.index_register = MemAddress(opcode & 0x0FFF)
		chip.advanceProgramCounter()
	case OpJumpToAddressPlusV0:
		address := opcode & 0x0FFF
		chip.program_counter = MemAddress(address) + MemAddress(chip.registers[0])
	case OpRandomAndByteIntoVx:
		x := (opcode & 0x0F00) >> 8
		nn := Word(opcode & 0x00FF)
		randomValue := Word(rand.Intn(256))
		chip.registers[x] = randomValue & nn
		chip.advanceProgramCounter()
	case OpDrawSprite:
		chip.registers[0xF] = 0 // Reset collision flag
		n := opcode & 0x000F
		x := chip.registers[(opcode&0x0F00)>>8]
		y := chip.registers[(opcode&0x00F0)>>4]

		for i := range n {
			b := chip.memory[(chip.index_register+MemAddress(i))%MemorySize]
			for j := range 8 {
				pixelX := (x + Word(j)) % DisplayWidth
				pixelY := (y + Word(i)) % DisplayHeight

				// Sprites are XORed onto the display, 0 bits leave the pixel unchanged
				if (b>>(7-j))&0x1 == 1 {
					if chip.Display[pixelX][pixelY] {
						chip.registers[0xF] = 1 // Set collision flag
					}
					chip.Display[pixelX][pixelY] = !chip.Display[pixelX][pixelY]
				}
			}
		}

		chip.advanceProgramCounter()
	case OpSkipIfKeyPressed:
		x := (opcode & 0x0F00) >> 8
		if chip.Keyboard[chip.registers[x]&0xF] {
			chip.advanceProgramCounter()
		}
		chip.advanceProgramCounter()
	case OpSkipIfKeyNotPressed:
		x := (opcode & 0x0F00) >> 8
		if !chip.Keyboard[chip.registers[x]&0xF] {
			chip.advanceProgramCounter()
		}
		chip.advanceProgramCounter()
	// TODO : Implement the 0xFXNN opcodes
	case OpLoadDelayTimerIntoVx:
		x := (opcode & 0x0F00) >> 8
		chip.registers[x] = Word(chip.delay_timer)
		chip.advanceProgramCounter()
	case OpWaitForKeyPress:
		x := (opcode & 0x0F00) >> 8
		// Wait for a key press and store the value of the key in Vx.
		// The program counter only advances once a key is down, so this
		// instruction runs again every cycle until then without blocking.
		for i, pressed := range chip.Keyboard {
			if pressed {
				chip.registers[x] = Word(i)
				chip.advanceProgramCounter()
				break
			}
		}
	case OpSetDelayTimerFromVx:
		x := (opcode & 0x0F00) >> 8
		chip.delay_timer = TimerDelay(chip.registers[x])
		chip.advanceProgramCounter()
	case OpSetSoundTimerFromVx:
		x := (opcode & 0x0F00) >> 8
		chip.sound_timer = TimerDelay(chip.registers[x])
		chip.advanceProgramCounter()
	case OpAddVxToIndex:
		x := (opcode & 0x0F00) >> 8
		chip.index_register += MemAddress(chip.registers[x])
		chip.advanceProgramCounter()
	case OpLoadFontSpriteIntoIndex:
		x := (opcode & 0x0F00) >> 8
		chip.index_register = MemAddress(chip.registers[x]) * 0x5
		chip.advanceProgramCounter()
	case OpStoreBCDOfVx:
		x := (opcode & 0x0F00) >> 8
		value := chip.registers[x]
		chip.memory[chip.index_register] = Word(value / 100)
		chip.memory[chip.index_register+1] = Word((value / 10) % 10)
		chip.memory[chip.index_register+2] = Word(value % 10)
		chip.advanceProgramCounter()
	case OpStoreV0ToVxInMemory:
		x := (opcode & 0x0F00) >> 8
		for i := range x + 1 {
			chip.memory[chip.index_register+MemAddress(i)] = chip.registers[i]
		}
		chip.index_register += MemAddress(x) + 1
		chip.advanceProgramCounter()
	case OpLoadV0ToVxFromMemory:
		x := (opcode & 0x0F00) >> 8
		for i := range x + 1 {
			chip.registers[i] = chip.memory[chip.index_register+MemAddress(i)]
		}
		chip.index_register += MemAddress(x) + 1
		chip.advanceProgramCounter()
	default:
		// Unknown or unimplemented opcode, skip it
		// We just igore it and hope it does not break anything.
		// Hope is the last thing to lose, right?
		log.Printf("Unknown or unimplemented opcode: 0x%04X", opcode)
		chip.advanceProgramCounter()
	}
}

// UpdateTimers decrements the delay and sound timers, it must be called at 60Hz
func (chip *GoChip) UpdateTimers() {
	if chip.delay_timer > 0 {
		chip.delay_timer--
	}
	if chip.sound_timer > 0 {
		chip.sound_timer--
	}
}

func (chip *GoChip) advanceProgramCounter() {
	chip.program_counter += 2
}

// decode masks out the variable parts (X, Y, N, NN, NNN) of an opcode so it
// can be compared directly against the Op* constants.
func decode(opcode Opcode) Opcode {
	switch opcode & 0xF000 {
	case 0x0000:
		if opcode == OpClearScreen || opcode == OpReturnFromSubroutine {
			return opcode
		}
		return OpSysCallMachineRoutine
	case 0x8000:
		return opcode & 0xF00F
	case 0xE000, 0xF000:
		return opcode & 0xF0FF
	default:
		return opcode & 0xF000
	}
}

func (chip *GoChip) clearScreen() {
	for x := range DisplayWidth {
		for y := range DisplayHeight {
			chip.Display[x][y] = false
		}
	}
}
