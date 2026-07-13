// Package store handles session persistence using SQLite.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Session represents a focus session to be saved in the database.
type Session struct {
	ID           int64
	Type         string    // "work", "short_break", "long_break"
	Task         string    // Name of active task
	Note         string    // Optional user note
	StartedAt    time.Time
	EndedAt      time.Time
	Completed    bool
	DurationSecs int
}

// Store wraps the SQL database connection.
type Store struct {
	db *sql.DB
}

// New opens a connection to the SQLite database at dbPath and runs migrations.
func New(dbPath string) (*Store, error) {
	// Ensure the parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open connection. CGo-free driver name is "sqlite"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ensure database connection is valid
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Store{db: db}
	if err := s.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveSession inserts a new focus session record.
func (s *Store) SaveSession(sess *Session) error {
	query := `INSERT INTO sessions (type, task, note, started_at, ended_at, completed, duration_secs)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	completedInt := 0
	if sess.Completed {
		completedInt = 1
	}

	res, err := s.db.Exec(query, sess.Type, sess.Task, sess.Note, sess.StartedAt, sess.EndedAt, completedInt, sess.DurationSecs)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		sess.ID = id
	}

	return nil
}

// GetSessions returns sessions in the specified date range.
func (s *Store) GetSessions(start, end time.Time) ([]*Session, error) {
	query := `SELECT id, type, task, note, started_at, ended_at, completed, duration_secs 
	          FROM sessions 
	          WHERE started_at >= ? AND started_at <= ?
	          ORDER BY started_at ASC`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var completedInt int
		err := rows.Scan(&sess.ID, &sess.Type, &sess.Task, &sess.Note, &sess.StartedAt, &sess.EndedAt, &completedInt, &sess.DurationSecs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		sess.Completed = completedInt == 1
		sessions = append(sessions, &sess)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading sessions rows: %w", err)
	}

	return sessions, nil
}

func (s *Store) runMigrations() error {
	// First, ensure the schema_migrations table exists
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY
	);`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// List of migrations. The index + 1 is the migration version.
	var migrations = []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			task TEXT,
			note TEXT,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			completed INTEGER NOT NULL,
			duration_secs INTEGER NOT NULL
		);`,
	}

	for i, query := range migrations {
		version := i + 1
		// Check if migration has run
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration version %d: %w", version, err)
		}
		if count > 0 {
			continue
		}

		// Run migration in transaction
		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %d: %w", version, err)
		}

		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", version, err)
		}
	}

	return nil
}
