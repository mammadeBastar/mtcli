package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/mmdbasi/mtcli/internal/test"
)

// ANSIRenderer implements the Renderer interface using ANSI escape codes
type ANSIRenderer struct {
	width   int
	height  int
	noColor bool
	mu      sync.Mutex
}

// RendererOptions holds configuration for the renderer
type RendererOptions struct {
	Width   int  // 0 means auto-detect
	NoColor bool
}

// NewANSIRenderer creates a new ANSI-based renderer
func NewANSIRenderer(opts RendererOptions) *ANSIRenderer {
	width := opts.Width
	if width == 0 {
		w, _, _ := GetTerminalSize()
		width = w
	}

	_, height, _ := GetTerminalSize()

	return &ANSIRenderer{
		width:   width,
		height:  height,
		noColor: opts.NoColor,
	}
}

// Init initializes the renderer
func (r *ANSIRenderer) Init() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	HideCursor()
	ClearScreen()
	MoveHome()
	return nil
}

// Cleanup restores terminal state
func (r *ANSIRenderer) Cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ShowCursor()
	Reset()
	ClearScreen()
	MoveHome()
}

// GetWidth returns the terminal width
func (r *ANSIRenderer) GetWidth() int {
	return r.width
}

// RenderCountdown renders the countdown before test starts
func (r *ANSIRenderer) RenderCountdown(seconds int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ClearScreen()
	MoveHome()

	// Center the countdown number
	centerRow := r.height / 2

	MoveCursor(centerRow, r.width/2-1)
	if !r.noColor {
		SetYellow()
		SetBold()
	}
	fmt.Printf("%d", seconds)
	Reset()

	return nil
}

// Render renders the current typing test state
func (r *ANSIRenderer) Render(state *RenderState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build the entire frame in memory first
	var frame strings.Builder

	// Clear screen and move home
	frame.WriteString(escClearScreen)
	frame.WriteString(escMoveHome)

	// Header line
	r.writeHeader(&frame, state)
	frame.WriteString("\r\n\r\n")

	// Target text with coloring
	r.writeTarget(&frame, state)
	frame.WriteString("\r\n\r\n")

	// Status line
	r.writeStatus(&frame, state)

	// Output the entire frame at once
	fmt.Print(frame.String())

	return nil
}

// writeHeader writes the header to the buffer
func (r *ANSIRenderer) writeHeader(buf *strings.Builder, state *RenderState) {
	modeStr := string(state.Mode)
	var infoStr string

	switch state.Mode {
	case test.ModeTimer:
		remaining := float64(state.TimeLimit) - state.Elapsed
		if remaining < 0 {
			remaining = 0
		}
		infoStr = fmt.Sprintf("%ds remaining", int(remaining))
	case test.ModeWords:
		wordCount := countWords(string(state.Target))
		infoStr = fmt.Sprintf("%d words", wordCount)
	case test.ModeQuote:
		infoStr = "quote mode"
	}

	buf.WriteString("  ")
	if !r.noColor {
		buf.WriteString(colorCyan)
	}
	buf.WriteString(strings.ToUpper(modeStr))
	buf.WriteString(escReset)
	buf.WriteString(" | ")
	buf.WriteString(infoStr)

	// Right-align exit hint
	hint := "Ctrl+C to exit"
	usedWidth := 2 + len(modeStr) + 3 + len(infoStr)
	padding := r.width - usedWidth - len(hint) - 2
	if padding > 0 {
		buf.WriteString(strings.Repeat(" ", padding))
	}
	if !r.noColor {
		buf.WriteString(escDim)
	}
	buf.WriteString(hint)
	buf.WriteString(escReset)
}

// writeTarget writes the target text with per-character coloring
func (r *ANSIRenderer) writeTarget(buf *strings.Builder, state *RenderState) {
	// Word wrap the target text
	maxWidth := r.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}

	lines := r.wrapText(state.Target, maxWidth)

	charIdx := 0
	for lineNum, line := range lines {
		if lineNum > 0 {
			buf.WriteString("\r\n")
		}
		buf.WriteString("  ") // Left margin

		for _, ch := range line {
			r.writeChar(buf, ch, charIdx, state)
			charIdx++
		}
	}
	buf.WriteString(escReset)
}

// writeChar writes a single character with appropriate coloring
func (r *ANSIRenderer) writeChar(buf *strings.Builder, ch rune, idx int, state *RenderState) {
	if r.noColor {
		buf.WriteRune(ch)
		return
	}

	if idx >= len(state.CharStates) {
		buf.WriteString(colorGray)
		buf.WriteRune(ch)
		return
	}

	switch state.CharStates[idx] {
	case test.CharUnattempted:
		buf.WriteString(colorGray)
	case test.CharCorrect:
		buf.WriteString(colorWhite)
	case test.CharIncorrect:
		buf.WriteString(colorOrange)
	}

	// Handle space visibility for incorrect
	if ch == ' ' && state.CharStates[idx] == test.CharIncorrect {
		buf.WriteRune('·') // Show incorrect space as middle dot
	} else {
		buf.WriteRune(ch)
	}
}

// writeStatus writes the status line
func (r *ANSIRenderer) writeStatus(buf *strings.Builder, state *RenderState) {
	buf.WriteString("  ")

	if state.Elapsed > 0.5 {
		if !r.noColor {
			buf.WriteString(colorGreen)
			buf.WriteString(escBold)
		}
		buf.WriteString(fmt.Sprintf("%.0f WPM", state.LiveWPM))
		buf.WriteString(escReset)
		buf.WriteString("  ")
	}

	if !r.noColor {
		buf.WriteString(escDim)
	}
	buf.WriteString(fmt.Sprintf("%.1fs", state.Elapsed))
	buf.WriteString(escReset)

	// Progress for words/quote mode
	if state.Mode != test.ModeTimer {
		progress := float64(len(state.Typed)) / float64(len(state.Target)) * 100
		if progress > 100 {
			progress = 100
		}
		buf.WriteString(fmt.Sprintf("  %.0f%%", progress))
	}
}

// wrapText wraps text to fit within the given width
func (r *ANSIRenderer) wrapText(runes []rune, maxWidth int) [][]rune {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	text := string(runes)
	words := strings.Fields(text)
	if len(words) == 0 {
		return [][]rune{runes}
	}

	var lines [][]rune
	var currentLine []rune

	for i, word := range words {
		wordRunes := []rune(word)

		if len(currentLine) > 0 {
			// Check if word fits on current line (with space)
			if len(currentLine)+1+len(wordRunes) <= maxWidth {
				currentLine = append(currentLine, ' ')
				currentLine = append(currentLine, wordRunes...)
			} else {
				// Start new line
				lines = append(lines, currentLine)
				currentLine = wordRunes
			}
		} else {
			// First word on line
			if len(wordRunes) <= maxWidth {
				currentLine = wordRunes
			} else {
				// Word is too long, force break
				currentLine = wordRunes[:maxWidth]
			}
		}

		// Handle last word
		if i == len(words)-1 && len(currentLine) > 0 {
			lines = append(lines, currentLine)
		}
	}

	// Handle case where last line wasn't added
	if len(lines) == 0 && len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

// RenderSummary renders the final results summary
func (r *ANSIRenderer) RenderSummary(result *test.SessionResult, chart string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Build summary with \r\n for raw mode compatibility
	var buf strings.Builder

	buf.WriteString(escClearScreen)
	buf.WriteString(escMoveHome)
	buf.WriteString(escShowCursor)

	// Title
	if !r.noColor {
		buf.WriteString(colorGreen)
		buf.WriteString(escBold)
	}
	buf.WriteString("\r\n")
	buf.WriteString("  ═══════════════════════════════════\r\n")
	buf.WriteString("          TEST COMPLETE!\r\n")
	buf.WriteString("  ═══════════════════════════════════\r\n")
	buf.WriteString(escReset)
	buf.WriteString("\r\n")

	// Main stats
	buf.WriteString("  ")
	if !r.noColor {
		buf.WriteString(colorGreen)
		buf.WriteString(escBold)
	}
	buf.WriteString(fmt.Sprintf("WPM: %.1f", result.WPM))
	buf.WriteString(escReset)

	buf.WriteString("  |  ")
	if !r.noColor {
		buf.WriteString(colorCyan)
	}
	buf.WriteString(fmt.Sprintf("Raw: %.1f", result.RawWPM))
	buf.WriteString(escReset)

	buf.WriteString("  |  ")
	if !r.noColor {
		buf.WriteString(colorYellow)
	}
	buf.WriteString(fmt.Sprintf("Accuracy: %.1f%%", result.Accuracy))
	buf.WriteString(escReset)
	buf.WriteString("\r\n\r\n")

	// Details
	buf.WriteString(fmt.Sprintf("  Time:       %.1fs\r\n", result.Duration.Seconds()))
	buf.WriteString(fmt.Sprintf("  Characters: %d/%d correct\r\n", result.CorrectChars, result.TotalTyped))
	buf.WriteString(fmt.Sprintf("  Mode:       %s\r\n", result.Mode))

	if result.Mode == test.ModeQuote && result.Metadata.Source != "" {
		buf.WriteString(fmt.Sprintf("  Source:     %s\r\n", result.Metadata.Source))
	}

	buf.WriteString("\r\n")

	// Speed chart
	if chart != "" {
		buf.WriteString("  Speed over time:\r\n\r\n")
		// Indent chart lines and convert newlines
		for _, line := range strings.Split(chart, "\n") {
			buf.WriteString("  ")
			buf.WriteString(line)
			buf.WriteString("\r\n")
		}
	}

	buf.WriteString("\r\n")
	if !r.noColor {
		buf.WriteString(escDim)
	}
	buf.WriteString("  Press Enter to continue...")
	buf.WriteString(escReset)

	// Output all at once
	fmt.Print(buf.String())

	// Wait for Enter
	inputBuf := make([]byte, 1)
	os.Stdin.Read(inputBuf)

	return nil
}

// countWords counts words in a string
func countWords(s string) int {
	return len(strings.Fields(s))
}
