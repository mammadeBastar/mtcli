package metrics

import (
	"time"
)

// NewTracker creates a new metrics tracker
func NewTracker() *Tracker {
	return &Tracker{
		samples:        make([]Sample, 0),
		sampleInterval: time.Second,
	}
}

// Start initializes the tracker with a start time
func (t *Tracker) Start(startTime time.Time) {
	t.startTime = startTime
	t.lastSampleAt = startTime
}

// Update updates the metrics with new counts
func (t *Tracker) Update(totalTyped, correctChars int) {
	t.totalTyped = totalTyped
	t.correctChars = correctChars
}

// MaybeSample takes a sample if the interval has passed
func (t *Tracker) MaybeSample() {
	if t.startTime.IsZero() {
		return
	}

	now := time.Now()
	if now.Sub(t.lastSampleAt) >= t.sampleInterval {
		t.TakeSample(now)
	}
}

// TakeSample records a sample at the given time
func (t *Tracker) TakeSample(at time.Time) {
	elapsed := at.Sub(t.startTime)
	sample := t.calculateSample(elapsed)
	t.samples = append(t.samples, sample)
	t.lastSampleAt = at
}

// calculateSample calculates WPM metrics at a point in time
func (t *Tracker) calculateSample(elapsed time.Duration) Sample {
	minutes := elapsed.Minutes()
	if minutes < 0.001 {
		minutes = 0.001
	}

	rawWPM := (float64(t.totalTyped) / 5.0) / minutes
	netWPM := (float64(t.correctChars) / 5.0) / minutes

	return Sample{
		TimeMs: elapsed.Milliseconds(),
		WPM:    netWPM,
		RawWPM: rawWPM,
	}
}

// Finalize calculates final metrics and returns the result
func (t *Tracker) Finalize(endTime time.Time) *Result {
	// Take final sample
	if !t.startTime.IsZero() {
		t.TakeSample(endTime)
	}

	duration := endTime.Sub(t.startTime)
	minutes := duration.Minutes()
	if minutes < 0.001 {
		minutes = 0.001
	}

	var accuracy float64
	if t.totalTyped > 0 {
		accuracy = float64(t.correctChars) / float64(t.totalTyped) * 100
	}

	rawWPM := (float64(t.totalTyped) / 5.0) / minutes
	netWPM := (float64(t.correctChars) / 5.0) / minutes

	return &Result{
		Duration:     duration,
		TotalTyped:   t.totalTyped,
		CorrectChars: t.correctChars,
		WPM:          netWPM,
		RawWPM:       rawWPM,
		Accuracy:     accuracy,
		Samples:      t.samples,
	}
}

// GetLiveWPM returns the current net WPM
func (t *Tracker) GetLiveWPM() float64 {
	if t.startTime.IsZero() {
		return 0
	}

	elapsed := time.Since(t.startTime)
	if elapsed < time.Second {
		return 0
	}

	minutes := elapsed.Minutes()
	return (float64(t.correctChars) / 5.0) / minutes
}

// GetLiveRawWPM returns the current raw WPM
func (t *Tracker) GetLiveRawWPM() float64 {
	if t.startTime.IsZero() {
		return 0
	}

	elapsed := time.Since(t.startTime)
	if elapsed < time.Second {
		return 0
	}

	minutes := elapsed.Minutes()
	return (float64(t.totalTyped) / 5.0) / minutes
}

// GetSamples returns all recorded samples
func (t *Tracker) GetSamples() []Sample {
	return t.samples
}

