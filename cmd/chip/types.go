package chip

type MemAddress uint16
type Register uint8
type Word uint8
type TimerDelay uint8
type StackPointer uint8

// These all reflect the standard CHIP-8 architecture.
const (
	ZeroOpCode          = 0x0000
	MemorySize          = 4096
	StackSize           = 16
	RegisterCount       = 16
	ProgramCounterStart = 0x200
	StackPointerStart   = 0
	IndexRegisterStart  = 0
	DelayTimerStart     = 0
	SoundTimerStart     = 0
	KeyboardSize        = 16
	DisplayWidth        = 64
	DisplayHeight       = 32
)

// GoChip represents the state of the CHIP-8 emulator.
// CHIP-8 OPCODES :
// 0NNN - SYS addr - Jump to a machine code routine at NNN. (Ignored by modern interpreters)
// 00E0 - CLS - Clear the display.
// 00EE - RET - Return from a subroutine.
// 1NNN - JP addr - Jump to location NNN.
// 2NNN - CALL addr - Call subroutine at NNN.
// 3XNN - SE Vx, byte - Skip next instruction if Vx = NN.
// 4XNN - SNE Vx, byte - Skip next instruction if Vx != NN.
// 5XY0 - SE Vx, Vy - Skip next instruction if Vx = Vy.
// 6XNN - LD Vx, byte - Set Vx = NN.
// 7XNN - ADD Vx, byte - Set Vx = Vx + NN.
// 8XY0 - LD Vx, Vy - Set Vx = Vy.
// 8XY1 - OR Vx, Vy - Set Vx = Vx OR Vy.
// 8XY2 - AND Vx, Vy - Set Vx = Vx AND Vy.
// 8XY3 - XOR Vx, Vy - Set Vx = Vx XOR Vy.
// 8XY4 - ADD Vx, Vy - Set Vx = Vx + Vy, set VF = carry.
// 8XY5 - SUB Vx, Vy - Set Vx = Vx - Vy, set VF = NOT borrow.
// 8XY6 - SHR Vx {, Vy} - Set Vx = Vx SHR 1.
// 8XY7 - SUBN Vx, Vy - Set Vx = Vy - Vx, set VF = NOT borrow.
// 8XYE - SHL Vx {, Vy} - Set Vx = Vx SHL 1.
// 9XY0 - SNE Vx, Vy - Skip next instruction if Vx != Vy.
// ANNN - LD I, addr - Set I = NNN.
// BNNN - JP V0, addr - Jump to location NNN + V0.
// CXNN - RND Vx, byte - Set Vx = random byte AND NN.
// DXYN - DRW Vx, Vy, nibble - Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
// EX9E - SKP Vx - Skip next instruction if key with the value of Vx is pressed.
// EXA1 - SKNP Vx - Skip next instruction if key with the value of Vx is not pressed.
// FX07 - LD Vx, DT - Set Vx = delay timer value.
// FX0A - LD Vx, K - Wait for a key press, store the value of the key in Vx.
// FX15 - LD DT, Vx - Set delay timer = Vx.
// FX18 - LD ST, Vx - Set sound timer = Vx.
// FX1E - ADD I, Vx - Set I = I + Vx.
// FX29 - LD F, Vx - Set I = location of sprite for digit Vx.
// FX33 - LD B, Vx - Store BCD representation of Vx in memory locations I, I+1, and I+2.
// FX55 - LD [I], Vx - Store registers V0 through Vx in memory starting at location I.
// FX65 - LD Vx, [I] - Read registers V0 through Vx from memory starting at location I.

// Opcode base values. Each constant is the full 16-bit opcode with every
// variable nibble (X, Y, N, NN, NNN) set to 0. Compare against the opcode
// after masking out the variable parts.

type Opcode uint16

const (
	// 0x0NNN group
	OpSysCallMachineRoutine Opcode = 0x0000 // 0NNN - SYS addr
	OpClearScreen           Opcode = 0x00E0 // 00E0 - CLS
	OpReturnFromSubroutine  Opcode = 0x00EE // 00EE - RET

	// Flow control
	OpJumpToAddress  Opcode = 0x1000 // 1NNN - JP addr
	OpCallSubroutine Opcode = 0x2000 // 2NNN - CALL addr

	// Conditional skips
	OpSkipIfVxEqualsByte    Opcode = 0x3000 // 3XNN - SE Vx, byte
	OpSkipIfVxNotEqualsByte Opcode = 0x4000 // 4XNN - SNE Vx, byte
	OpSkipIfVxEqualsVy      Opcode = 0x5000 // 5XY0 - SE Vx, Vy

	// Register / constant
	OpLoadByteIntoVx Opcode = 0x6000 // 6XNN - LD Vx, byte
	OpAddByteToVx    Opcode = 0x7000 // 7XNN - ADD Vx, byte

	// 0x8XYN arithmetic / logic group
	OpLoadVyIntoVx       Opcode = 0x8000 // 8XY0 - LD Vx, Vy
	OpOrVxVy             Opcode = 0x8001 // 8XY1 - OR Vx, Vy
	OpAndVxVy            Opcode = 0x8002 // 8XY2 - AND Vx, Vy
	OpXorVxVy            Opcode = 0x8003 // 8XY3 - XOR Vx, Vy
	OpAddVyToVxWithCarry Opcode = 0x8004 // 8XY4 - ADD Vx, Vy
	OpSubVyFromVx        Opcode = 0x8005 // 8XY5 - SUB Vx, Vy
	OpShiftVxRight       Opcode = 0x8006 // 8XY6 - SHR Vx {, Vy}
	OpSubVxFromVy        Opcode = 0x8007 // 8XY7 - SUBN Vx, Vy
	OpShiftVxLeft        Opcode = 0x800E // 8XYE - SHL Vx {, Vy}

	OpSkipIfVxNotEqualsVy Opcode = 0x9000 // 9XY0 - SNE Vx, Vy

	// Index register / jumps / random / draw
	OpLoadAddressIntoIndex Opcode = 0xA000 // ANNN - LD I, addr
	OpJumpToAddressPlusV0  Opcode = 0xB000 // BNNN - JP V0, addr
	OpRandomAndByteIntoVx  Opcode = 0xC000 // CXNN - RND Vx, byte
	OpDrawSprite           Opcode = 0xD000 // DXYN - DRW Vx, Vy, nibble

	// 0xEXNN keyboard group
	OpSkipIfKeyPressed    Opcode = 0xE09E // EX9E - SKP Vx
	OpSkipIfKeyNotPressed Opcode = 0xE0A1 // EXA1 - SKNP Vx

	// 0xFXNN timers / memory group
	OpLoadDelayTimerIntoVx    Opcode = 0xF007 // FX07 - LD Vx, DT
	OpWaitForKeyPress         Opcode = 0xF00A // FX0A - LD Vx, K
	OpSetDelayTimerFromVx     Opcode = 0xF015 // FX15 - LD DT, Vx
	OpSetSoundTimerFromVx     Opcode = 0xF018 // FX18 - LD ST, Vx
	OpAddVxToIndex            Opcode = 0xF01E // FX1E - ADD I, Vx
	OpLoadFontSpriteIntoIndex Opcode = 0xF029 // FX29 - LD F, Vx
	OpStoreBCDOfVx            Opcode = 0xF033 // FX33 - LD B, Vx
	OpStoreV0ToVxInMemory     Opcode = 0xF055 // FX55 - LD [I], Vx
	OpLoadV0ToVxFromMemory    Opcode = 0xF065 // FX65 - LD Vx, [I]
)

var ChipFontSet = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}
