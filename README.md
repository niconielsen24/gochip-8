# GoChip-8

A simple CHIP-8 emulator written in Go, using [raylib-go](https://github.com/gen2brain/raylib-go) for graphics and input.

> **Disclaimer:** This project exists purely for entertainment, research and study.
> It is not meant to be a complete or accurate emulator, and it may very well break
> on some CHIP-8 ROMs because CHIP-8 implementations differ from one another.
> XO-CHIP and SUPER-CHIP ROMs are not supported.

## Requirements

- Go 1.27+
- The system dependencies raylib needs (see the [raylib-go README](https://github.com/gen2brain/raylib-go#requirements))

## Usage

```sh
go run ./cmd <path/to/rom.ch8>
```

For example:

```sh
go run ./cmd roms/roms/br8kout.ch8
```

Run with `-h` or `--help` to print help.

## Controls

The CHIP-8 hex keypad is mapped to the left side of a QWERTY keyboard:

```
CHIP-8        Keyboard
1 2 3 C       1 2 3 4
4 5 6 D   ->  Q W E R
7 8 9 E       A S D F
A 0 B F       Z X C V
```

## Resources

- [Awesome CHIP-8](https://github.com/tobiasvl/awesome-chip-8)
- [raylib-go](https://github.com/gen2brain/raylib-go)
