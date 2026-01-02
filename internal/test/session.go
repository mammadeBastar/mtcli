package test

import (
	"time"
)

// Session manages the state of a typing test
type Session struct {
	state        *SessionState
	metrics      *MetricsTracker
	onUpdate     func(*SessionState)
	timerSeconds int
	timerDone    chan struct{}
}

// MetricsTracker tracks typing metrics during the session
type MetricsTracker struct {
	samples        []Sample
	lastSampleTime time.Time
	sampleInterval time.Duration
	totalTyped     int
	correctChars   int
}

// NewMetricsTracker creates a new metrics tracker
func NewMetricsTracker() *MetricsTracker {
	return &MetricsTracker{
		samples:        make([]Sample, 0),
		sampleInterval: 500 * time.Millisecond, // Sample every 500ms for better chart resolution
	}
}

// SessionOptions holds options for creating a session
type SessionOptions struct {
	Target       *Target
	TimerSeconds int // Only used in timer mode
	OnUpdate     func(*SessionState)
}

// NewSession creates a new typing session
func NewSession(opts SessionOptions) *Session {
	targetRunes := []rune(opts.Target.Text)
	charStates := make([]CharState, len(targetRunes))
	for i := range charStates {
		charStates[i] = CharUnattempted
	}

	return &Session{
		state: &SessionState{
			Target:      opts.Target,
			TargetRunes: targetRunes,
			TypedRunes:  make([]rune, 0, len(targetRunes)),
			CharStates:  charStates,
		},
		metrics:      NewMetricsTracker(),
		onUpdate:     opts.OnUpdate,
		timerSeconds: opts.TimerSeconds,
	}
}

// Start begins the session (called when first key is pressed or timer starts)
func (s *Session) Start() {
	s.state.StartedAt = time.Now()
	s.metrics.lastSampleTime = s.state.StartedAt

	// Take initial sample
	s.metrics.samples = append(s.metrics.samples, Sample{
		TimeMs: 0,
		WPM:    0,
		RawWPM: 0,
	})

	// Start timer for timer mode
	if s.state.Target.Mode == ModeTimer && s.timerSeconds > 0 {
		s.timerDone = make(chan struct{})
		go func() {
			timer := time.NewTimer(time.Duration(s.timerSeconds) * time.Second)
			select {
			case <-timer.C:
				s.state.Finished = true
				s.state.EndedAt = time.Now()
			case <-s.timerDone:
				timer.Stop()
			}
		}()
	}
}

// HandleKey processes a key input and updates session state
func (s *Session) HandleKey(keyType int, r rune) {
	if s.state.Finished || s.state.Aborted {
		return
	}

	// Start on first keystroke if not started
	if s.state.StartedAt.IsZero() {
		s.Start()
	}

	switch keyType {
	case KeyTypeRune:
		s.handleRune(r)
	case KeyTypeBackspace:
		s.handleBackspace()
	}

	// Check for completion (words/quote mode)
	if s.state.Target.Mode != ModeTimer {
		if len(s.state.TypedRunes) >= len(s.state.TargetRunes) {
			s.finish()
		}
	}

	// Take sample if interval has passed
	s.maybeTakeSample()

	// Notify listener
	if s.onUpdate != nil {
		s.onUpdate(s.state)
	}
}

// handleRune processes a typed character
func (s *Session) handleRune(r rune) {
	idx := len(s.state.TypedRunes)

	// Don't allow typing beyond target in non-timer mode
	if idx >= len(s.state.TargetRunes) {
		return
	}

	s.state.TypedRunes = append(s.state.TypedRunes, r)
	s.metrics.totalTyped++

	// Update char state
	if r == s.state.TargetRunes[idx] {
		s.state.CharStates[idx] = CharCorrect
		s.metrics.correctChars++
	} else {
		s.state.CharStates[idx] = CharIncorrect
	}
}

// handleBackspace removes the last typed character
func (s *Session) handleBackspace() {
	if len(s.state.TypedRunes) == 0 {
		return
	}

	idx := len(s.state.TypedRunes) - 1

	// Revert char state
	if s.state.CharStates[idx] == CharCorrect {
		s.metrics.correctChars--
	}
	s.state.CharStates[idx] = CharUnattempted

	s.state.TypedRunes = s.state.TypedRunes[:idx]
}

// maybeTakeSample takes a metrics sample if the interval has passed
func (s *Session) maybeTakeSample() {
	if s.state.StartedAt.IsZero() {
		return
	}
	now := time.Now()
	if now.Sub(s.metrics.lastSampleTime) >= s.metrics.sampleInterval {
		elapsed := now.Sub(s.state.StartedAt)
		sample := s.calculateSample(elapsed)
		s.metrics.samples = append(s.metrics.samples, sample)
		s.metrics.lastSampleTime = now
	}
}

// TakeSample forces a sample to be taken (called from ticker)
func (s *Session) TakeSample() {
	if s.state.StartedAt.IsZero() || s.state.Finished || s.state.Aborted {
		return
	}
	s.maybeTakeSample()
}

// calculateSample calculates current WPM metrics
func (s *Session) calculateSample(elapsed time.Duration) Sample {
	minutes := elapsed.Minutes()
	if minutes < 0.001 {
		minutes = 0.001 // Avoid division by zero
	}

	// WPM = (chars / 5) / minutes
	rawWPM := (float64(s.metrics.totalTyped) / 5.0) / minutes
	netWPM := (float64(s.metrics.correctChars) / 5.0) / minutes

	return Sample{
		TimeMs: elapsed.Milliseconds(),
		WPM:    netWPM,
		RawWPM: rawWPM,
	}
}

// Abort cancels the session
func (s *Session) Abort() {
	s.state.Aborted = true
	s.state.EndedAt = time.Now()
	if s.timerDone != nil {
		close(s.timerDone)
	}
}

// finish completes the session normally
func (s *Session) finish() {
	s.state.Finished = true
	s.state.EndedAt = time.Now()
	if s.timerDone != nil {
		close(s.timerDone)
	}
}

// IsFinished returns whether the session has ended
func (s *Session) IsFinished() bool {
	return s.state.Finished || s.state.Aborted
}

// IsAborted returns whether the session was aborted
func (s *Session) IsAborted() bool {
	return s.state.Aborted
}

// GetState returns the current session state
func (s *Session) GetState() *SessionState {
	return s.state
}

// GetResult calculates and returns the final session result
func (s *Session) GetResult() *SessionResult {
	// Take final sample
	if !s.state.StartedAt.IsZero() && !s.state.EndedAt.IsZero() {
		elapsed := s.state.EndedAt.Sub(s.state.StartedAt)
		finalSample := s.calculateSample(elapsed)
		s.metrics.samples = append(s.metrics.samples, finalSample)
	}

	duration := s.state.EndedAt.Sub(s.state.StartedAt)
	minutes := duration.Minutes()
	if minutes < 0.001 {
		minutes = 0.001
	}

	totalTyped := s.metrics.totalTyped
	correctChars := s.metrics.correctChars

	var accuracy float64
	if totalTyped > 0 {
		accuracy = float64(correctChars) / float64(totalTyped) * 100
	}

	rawWPM := (float64(totalTyped) / 5.0) / minutes
	netWPM := (float64(correctChars) / 5.0) / minutes

	return &SessionResult{
		Mode:         s.state.Target.Mode,
		StartedAt:    s.state.StartedAt,
		Duration:     duration,
		TargetLen:    len(s.state.TargetRunes),
		TotalTyped:   totalTyped,
		CorrectChars: correctChars,
		WPM:          netWPM,
		RawWPM:       rawWPM,
		Accuracy:     accuracy,
		Samples:      s.metrics.samples,
		Metadata:     s.state.Target.Metadata,
	}
}

// GetElapsed returns time elapsed since session start
func (s *Session) GetElapsed() time.Duration {
	if s.state.StartedAt.IsZero() {
		return 0
	}
	if !s.state.EndedAt.IsZero() {
		return s.state.EndedAt.Sub(s.state.StartedAt)
	}
	return time.Since(s.state.StartedAt)
}

// GetLiveWPM returns the current WPM (net)
func (s *Session) GetLiveWPM() float64 {
	elapsed := s.GetElapsed()
	if elapsed < time.Second {
		return 0
	}
	minutes := elapsed.Minutes()
	return (float64(s.metrics.correctChars) / 5.0) / minutes
}

// KeyType constants for the session (matching input package)
const (
	KeyTypeRune = iota
	KeyTypeBackspace
	KeyTypeEnter
	KeyTypeEscape
	KeyTypeCtrlC
	KeyTypeUnknown
)

