package input

// KeyType represents the type of key pressed
type KeyType int

const (
	KeyRune      KeyType = iota // Regular printable character
	KeyBackspace               // Backspace/Delete
	KeyEnter                   // Enter/Return
	KeyEscape                  // Escape
	KeyCtrlC                   // Ctrl+C
	KeyUnknown                 // Unknown/unhandled key
)

// KeyEvent represents a keyboard input event
type KeyEvent struct {
	Type KeyType
	Rune rune // Only valid when Type == KeyRune
}

// Reader defines the interface for reading keyboard input
type Reader interface {
	// Init puts the terminal in raw mode
	Init() error

	// ReadKey reads a single key event (blocking)
	ReadKey() (KeyEvent, error)

	// Cleanup restores terminal state
	Cleanup() error
}

