package storage

import "time"

// Session represents a stored typing test session
type Session struct {
	ID            int64
	StartedAt     time.Time
	Mode          string
	Seconds       int
	Words         int
	QuoteID       string
	TargetLen     int
	DurationMs    int64
	CorrectChars  int
	IncorrectChars int
	TotalTyped    int
	Accuracy      float64
	WPM           float64
	RawWPM        float64
}

// SessionSample represents a speed sample for a session
type SessionSample struct {
	SessionID int64
	TimeMs    int64
	WPM       float64
	RawWPM    float64
}

// Stats represents aggregate statistics
type Stats struct {
	TotalTests    int
	TotalTime     time.Duration
	AverageWPM    float64
	BestWPM       float64
	AverageAccuracy float64
	Last7DaysAvgWPM  float64
	Last30DaysAvgWPM float64
	ModeStats     map[string]ModeStats
}

// ModeStats represents statistics for a specific mode
type ModeStats struct {
	TestCount  int
	AverageWPM float64
	BestWPM    float64
}

// Store defines the interface for session storage
type Store interface {
	// SaveSession saves a completed session and its samples
	SaveSession(session *Session, samples []SessionSample) (int64, error)

	// GetSession retrieves a session by ID
	GetSession(id int64) (*Session, error)

	// GetSamples retrieves samples for a session
	GetSamples(sessionID int64) ([]SessionSample, error)

	// ListSessions retrieves recent sessions with optional filtering
	ListSessions(limit int, mode string) ([]Session, error)

	// GetStats calculates aggregate statistics
	GetStats() (*Stats, error)

	// Close closes the storage connection
	Close() error
}

