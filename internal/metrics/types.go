package metrics

import "time"

// Tracker tracks typing metrics during a session
type Tracker struct {
	startTime    time.Time
	samples      []Sample
	lastSampleAt time.Time
	sampleInterval time.Duration

	totalTyped   int
	correctChars int
}

// Sample represents a point-in-time measurement
type Sample struct {
	TimeMs int64
	WPM    float64
	RawWPM float64
}

// Result holds the final calculated metrics
type Result struct {
	Duration     time.Duration
	TotalTyped   int
	CorrectChars int
	WPM          float64
	RawWPM       float64
	Accuracy     float64
	Samples      []Sample
}

