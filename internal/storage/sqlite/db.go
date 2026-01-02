package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mmdbasi/mtcli/internal/config"
)

const currentSchemaVersion = 1

// Store represents the SQLite storage
type Store struct {
	db *sql.DB
}

// Open opens or creates the SQLite database
func Open() (*Store, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}

	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// getDBPath returns the path to the SQLite database file
func getDBPath() (string, error) {
	dataDir, err := config.GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, "mtcli.db"), nil
}

// migrate runs database migrations
func (s *Store) migrate() error {
	// Create schema_version table if it doesn't exist
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		)
	`)
	if err != nil {
		return err
	}

	// Get current version
	var version int
	err = s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return err
	}

	// Apply migrations
	if version < 1 {
		if err := s.migrateV1(); err != nil {
			return err
		}
	}

	return nil
}

// migrateV1 creates the initial schema
func (s *Store) migrateV1() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create sessions table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			started_at DATETIME NOT NULL,
			mode TEXT NOT NULL,
			seconds INTEGER DEFAULT 0,
			words INTEGER DEFAULT 0,
			quote_id TEXT DEFAULT '',
			target_len INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			correct_chars INTEGER NOT NULL,
			incorrect_chars INTEGER NOT NULL DEFAULT 0,
			total_typed INTEGER NOT NULL,
			accuracy REAL NOT NULL,
			wpm REAL NOT NULL,
			raw_wpm REAL NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create samples table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			time_ms INTEGER NOT NULL,
			wpm REAL NOT NULL,
			raw_wpm REAL NOT NULL,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes
	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_started_at ON sessions(started_at)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_mode ON sessions(mode)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX IF NOT EXISTS idx_samples_session_id ON samples(session_id)`)
	if err != nil {
		return err
	}

	// Update schema version
	_, err = tx.Exec(`INSERT INTO schema_version (version) VALUES (1)`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

