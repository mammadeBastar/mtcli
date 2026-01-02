package ui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ANSI escape codes
const (
	escClearScreen   = "\033[2J"
	escClearLine     = "\033[2K"
	escMoveCursor    = "\033[%d;%dH" // row, col (1-indexed)
	escMoveHome      = "\033[H"
	escHideCursor    = "\033[?25l"
	escShowCursor    = "\033[?25h"
	escReset         = "\033[0m"
	escBold          = "\033[1m"
	escDim           = "\033[2m"
)

// Color codes (256-color mode)
const (
	colorGray   = "\033[38;5;245m" // Unattempted text
	colorWhite  = "\033[38;5;255m" // Correct text
	colorOrange = "\033[38;5;208m" // Incorrect text
	colorGreen  = "\033[38;5;114m" // Success/WPM
	colorCyan   = "\033[38;5;80m"  // Info
	colorYellow = "\033[38;5;220m" // Warning/highlight
)

// ClearScreen clears the entire terminal screen
func ClearScreen() {
	fmt.Print(escClearScreen)
}

// ClearLine clears the current line
func ClearLine() {
	fmt.Print(escClearLine)
}

// MoveCursor moves the cursor to the specified position (1-indexed)
func MoveCursor(row, col int) {
	fmt.Printf(escMoveCursor, row, col)
}

// MoveHome moves the cursor to the top-left corner
func MoveHome() {
	fmt.Print(escMoveHome)
}

// HideCursor hides the terminal cursor
func HideCursor() {
	fmt.Print(escHideCursor)
}

// ShowCursor shows the terminal cursor
func ShowCursor() {
	fmt.Print(escShowCursor)
}

// Reset resets all text attributes
func Reset() {
	fmt.Print(escReset)
}

// SetGray sets the text color to gray (for unattempted)
func SetGray() {
	fmt.Print(colorGray)
}

// SetWhite sets the text color to white (for correct)
func SetWhite() {
	fmt.Print(colorWhite)
}

// SetOrange sets the text color to orange (for incorrect)
func SetOrange() {
	fmt.Print(colorOrange)
}

// SetGreen sets the text color to green
func SetGreen() {
	fmt.Print(colorGreen)
}

// SetCyan sets the text color to cyan
func SetCyan() {
	fmt.Print(colorCyan)
}

// SetYellow sets the text color to yellow
func SetYellow() {
	fmt.Print(colorYellow)
}

// SetBold sets text to bold
func SetBold() {
	fmt.Print(escBold)
}

// SetDim sets text to dim
func SetDim() {
	fmt.Print(escDim)
}

// GetTerminalSize returns the current terminal width and height
func GetTerminalSize() (width, height int, err error) {
	width, height, err = term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Fallback to reasonable defaults
		return 80, 24, nil
	}
	return width, height, nil
}

// ColoredString returns a string with ANSI color codes
func ColoredString(s string, color string) string {
	return color + s + escReset
}

// GrayString returns a gray-colored string
func GrayString(s string) string {
	return colorGray + s + escReset
}

// WhiteString returns a white-colored string
func WhiteString(s string) string {
	return colorWhite + s + escReset
}

// OrangeString returns an orange-colored string
func OrangeString(s string) string {
	return colorOrange + s + escReset
}

// GreenString returns a green-colored string
func GreenString(s string) string {
	return colorGreen + s + escReset
}

// CyanString returns a cyan-colored string
func CyanString(s string) string {
	return colorCyan + s + escReset
}

// YellowString returns a yellow-colored string
func YellowString(s string) string {
	return colorYellow + s + escReset
}

