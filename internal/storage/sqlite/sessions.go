package sqlite

import (
	"database/sql"
	"time"
)

// Session represents a stored typing test session
type Session struct {
	ID             int64
	StartedAt      time.Time
	Mode           string
	Seconds        int
	Words          int
	QuoteID        string
	TargetLen      int
	DurationMs     int64
	CorrectChars   int
	IncorrectChars int
	TotalTyped     int
	Accuracy       float64
	WPM            float64
	RawWPM         float64
}

// SessionSample represents a speed sample for a session
type SessionSample struct {
	ID        int64
	SessionID int64
	TimeMs    int64
	WPM       float64
	RawWPM    float64
}

// SaveSession saves a completed session and its samples
func (s *Store) SaveSession(session *Session, samples []SessionSample) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Insert session
	result, err := tx.Exec(`
		INSERT INTO sessions (
			started_at, mode, seconds, words, quote_id, target_len,
			duration_ms, correct_chars, incorrect_chars, total_typed,
			accuracy, wpm, raw_wpm
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		session.StartedAt,
		session.Mode,
		session.Seconds,
		session.Words,
		session.QuoteID,
		session.TargetLen,
		session.DurationMs,
		session.CorrectChars,
		session.IncorrectChars,
		session.TotalTyped,
		session.Accuracy,
		session.WPM,
		session.RawWPM,
	)
	if err != nil {
		return 0, err
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert samples
	for _, sample := range samples {
		_, err = tx.Exec(`
			INSERT INTO samples (session_id, time_ms, wpm, raw_wpm)
			VALUES (?, ?, ?, ?)
		`, sessionID, sample.TimeMs, sample.WPM, sample.RawWPM)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return sessionID, nil
}

// GetSession retrieves a session by ID
func (s *Store) GetSession(id int64) (*Session, error) {
	session := &Session{}
	err := s.db.QueryRow(`
		SELECT id, started_at, mode, seconds, words, quote_id, target_len,
		       duration_ms, correct_chars, incorrect_chars, total_typed,
		       accuracy, wpm, raw_wpm
		FROM sessions WHERE id = ?
	`, id).Scan(
		&session.ID,
		&session.StartedAt,
		&session.Mode,
		&session.Seconds,
		&session.Words,
		&session.QuoteID,
		&session.TargetLen,
		&session.DurationMs,
		&session.CorrectChars,
		&session.IncorrectChars,
		&session.TotalTyped,
		&session.Accuracy,
		&session.WPM,
		&session.RawWPM,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetSamples retrieves samples for a session
func (s *Store) GetSamples(sessionID int64) ([]SessionSample, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, time_ms, wpm, raw_wpm
		FROM samples WHERE session_id = ?
		ORDER BY time_ms
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []SessionSample
	for rows.Next() {
		var sample SessionSample
		err := rows.Scan(&sample.ID, &sample.SessionID, &sample.TimeMs, &sample.WPM, &sample.RawWPM)
		if err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}

	return samples, rows.Err()
}

// ListSessions retrieves recent sessions with optional mode filter
func (s *Store) ListSessions(limit int, mode string) ([]Session, error) {
	var rows *sql.Rows
	var err error

	if mode != "" {
		rows, err = s.db.Query(`
			SELECT id, started_at, mode, seconds, words, quote_id, target_len,
			       duration_ms, correct_chars, incorrect_chars, total_typed,
			       accuracy, wpm, raw_wpm
			FROM sessions
			WHERE mode = ?
			ORDER BY started_at DESC
			LIMIT ?
		`, mode, limit)
	} else {
		rows, err = s.db.Query(`
			SELECT id, started_at, mode, seconds, words, quote_id, target_len,
			       duration_ms, correct_chars, incorrect_chars, total_typed,
			       accuracy, wpm, raw_wpm
			FROM sessions
			ORDER BY started_at DESC
			LIMIT ?
		`, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var session Session
		err := rows.Scan(
			&session.ID,
			&session.StartedAt,
			&session.Mode,
			&session.Seconds,
			&session.Words,
			&session.QuoteID,
			&session.TargetLen,
			&session.DurationMs,
			&session.CorrectChars,
			&session.IncorrectChars,
			&session.TotalTyped,
			&session.Accuracy,
			&session.WPM,
			&session.RawWPM,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// Stats represents aggregate statistics
type Stats struct {
	TotalTests       int
	TotalTimeMs      int64
	AverageWPM       float64
	BestWPM          float64
	AverageAccuracy  float64
	Last7DaysAvgWPM  float64
	Last30DaysAvgWPM float64
	ModeStats        map[string]ModeStats
}

// ModeStats represents statistics for a specific mode
type ModeStats struct {
	TestCount  int
	AverageWPM float64
	BestWPM    float64
}

// GetStats calculates aggregate statistics
func (s *Store) GetStats() (*Stats, error) {
	stats := &Stats{
		ModeStats: make(map[string]ModeStats),
	}

	// Overall stats
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(duration_ms), 0), 
		       COALESCE(AVG(wpm), 0), COALESCE(MAX(wpm), 0), 
		       COALESCE(AVG(accuracy), 0)
		FROM sessions
	`).Scan(
		&stats.TotalTests,
		&stats.TotalTimeMs,
		&stats.AverageWPM,
		&stats.BestWPM,
		&stats.AverageAccuracy,
	)
	if err != nil {
		return nil, err
	}

	// Last 7 days average
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	err = s.db.QueryRow(`
		SELECT COALESCE(AVG(wpm), 0)
		FROM sessions
		WHERE started_at >= ?
	`, sevenDaysAgo).Scan(&stats.Last7DaysAvgWPM)
	if err != nil {
		return nil, err
	}

	// Last 30 days average
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = s.db.QueryRow(`
		SELECT COALESCE(AVG(wpm), 0)
		FROM sessions
		WHERE started_at >= ?
	`, thirtyDaysAgo).Scan(&stats.Last30DaysAvgWPM)
	if err != nil {
		return nil, err
	}

	// Per-mode stats
	rows, err := s.db.Query(`
		SELECT mode, COUNT(*), COALESCE(AVG(wpm), 0), COALESCE(MAX(wpm), 0)
		FROM sessions
		GROUP BY mode
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mode string
		var modeStats ModeStats
		err := rows.Scan(&mode, &modeStats.TestCount, &modeStats.AverageWPM, &modeStats.BestWPM)
		if err != nil {
			return nil, err
		}
		stats.ModeStats[mode] = modeStats
	}

	return stats, rows.Err()
}

// DeleteSession deletes a session and its samples
func (s *Store) DeleteSession(id int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM samples WHERE session_id = ?", id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

