package test

import "time"

// Mode represents the type of typing test
type Mode string

const (
	ModeTimer Mode = "timer"
	ModeWords Mode = "words"
	ModeQuote Mode = "quote"
)

// CharState represents the state of a character in the target text
type CharState int

const (
	CharUnattempted CharState = iota
	CharCorrect
	CharIncorrect
)

// Target represents the text to be typed
type Target struct {
	Text     string
	Mode     Mode
	Metadata TargetMetadata
}

// TargetMetadata holds mode-specific metadata
type TargetMetadata struct {
	WordCount int    // for words mode
	Seconds   int    // for timer mode
	QuoteID   string // for quote mode
	Source    string // quote source/author
}

// SessionState represents the current state of a typing session
type SessionState struct {
	Target      *Target
	TargetRunes []rune
	TypedRunes  []rune
	CharStates  []CharState
	StartedAt   time.Time
	EndedAt     time.Time
	Finished    bool
	Aborted     bool
}

// SessionResult holds the final results of a typing session
type SessionResult struct {
	Mode         Mode
	StartedAt    time.Time
	Duration     time.Duration
	TargetLen    int
	TotalTyped   int
	CorrectChars int
	WPM          float64
	RawWPM       float64
	Accuracy     float64
	Samples      []Sample
	Metadata     TargetMetadata
}

// Sample represents a point-in-time speed measurement
type Sample struct {
	TimeMs int64   // milliseconds since start
	WPM    float64 // net WPM at this point
	RawWPM float64 // raw WPM at this point
}

