package input

import (
	"bufio"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

// RawReader reads keyboard input in raw terminal mode
type RawReader struct {
	oldState *term.State
	reader   *bufio.Reader
}

// NewRawReader creates a new raw input reader
func NewRawReader() *RawReader {
	return &RawReader{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Init puts the terminal in raw mode
func (r *RawReader) Init() error {
	var err error
	r.oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
	return err
}

// Cleanup restores the terminal to its original state
func (r *RawReader) Cleanup() error {
	if r.oldState != nil {
		return term.Restore(int(os.Stdin.Fd()), r.oldState)
	}
	return nil
}

// ReadKey reads a single key event from stdin
func (r *RawReader) ReadKey() (KeyEvent, error) {
	buf := make([]byte, 4) // UTF-8 can be up to 4 bytes

	// Read first byte
	n, err := r.reader.Read(buf[:1])
	if err != nil {
		return KeyEvent{Type: KeyUnknown}, err
	}
	if n == 0 {
		return KeyEvent{Type: KeyUnknown}, nil
	}

	b := buf[0]

	// Handle control characters
	switch b {
	case 3: // Ctrl+C
		return KeyEvent{Type: KeyCtrlC}, nil
	case 27: // Escape or escape sequence
		// Check if there's more data (escape sequence)
		if r.reader.Buffered() > 0 {
			// Read escape sequence
			seq := make([]byte, 2)
			r.reader.Read(seq)
			// For now, ignore escape sequences (arrow keys, etc.)
			return KeyEvent{Type: KeyUnknown}, nil
		}
		return KeyEvent{Type: KeyEscape}, nil
	case 13: // Enter/Return
		return KeyEvent{Type: KeyEnter}, nil
	case 127, 8: // Backspace (127 = DEL on most terminals, 8 = BS)
		return KeyEvent{Type: KeyBackspace}, nil
	}

	// Handle printable ASCII
	if b >= 32 && b < 127 {
		return KeyEvent{Type: KeyRune, Rune: rune(b)}, nil
	}

	// Handle UTF-8 multi-byte sequences
	if b >= 0xC0 {
		// Determine how many bytes we need
		var runeLen int
		if b < 0xE0 {
			runeLen = 2
		} else if b < 0xF0 {
			runeLen = 3
		} else {
			runeLen = 4
		}

		// Read remaining bytes
		buf[0] = b
		for i := 1; i < runeLen; i++ {
			n, err := r.reader.Read(buf[i : i+1])
			if err != nil || n == 0 {
				return KeyEvent{Type: KeyUnknown}, err
			}
		}

		// Decode the rune
		ru, _ := utf8.DecodeRune(buf[:runeLen])
		if ru != utf8.RuneError {
			return KeyEvent{Type: KeyRune, Rune: ru}, nil
		}
	}

	return KeyEvent{Type: KeyUnknown}, nil
}

