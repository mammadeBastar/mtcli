package ui

import "github.com/mmdbasi/mtcli/internal/test"

// RenderState holds the state needed for rendering
type RenderState struct {
	Target      []rune
	Typed       []rune
	CharStates  []test.CharState
	Mode        test.Mode
	Elapsed     float64 // seconds
	LiveWPM     float64
	TimeLimit   int // for timer mode
	Countdown   int // countdown seconds remaining (-1 if started)
	Finished    bool
}

// Renderer defines the interface for UI rendering
type Renderer interface {
	// Init initializes the renderer (clear screen, hide cursor, etc.)
	Init() error

	// Render renders the current state
	Render(state *RenderState) error

	// RenderCountdown renders the countdown before test starts
	RenderCountdown(seconds int) error

	// RenderSummary renders the final summary
	RenderSummary(result *test.SessionResult, chart string) error

	// Cleanup restores terminal state
	Cleanup()

	// GetWidth returns the terminal width
	GetWidth() int
}

