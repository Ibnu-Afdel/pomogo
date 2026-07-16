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
	ProjectID    *int64    // Optional project ID
	ProjectName  string    // Resolved project name
	Mode         string    // "quick", "deep"
	BlockID      *int64    // Block reference
}

// Project represents a user category/project for focus sessions.
type Project struct {
	ID       int64
	Name     string
	Color    string
	Archived bool
	Icon     string
}

// BlockStore represents a deep focus block.
type BlockStore struct {
	ID          int64
	Mode        string
	PlannedSecs int
	StartedAt   time.Time
	EndedAt     time.Time
	Completed   bool
	Pauses      int
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
	query := `INSERT INTO sessions (type, task, note, started_at, ended_at, completed, duration_secs, project_id, mode, block_id)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	completedInt := 0
	if sess.Completed {
		completedInt = 1
	}

	var projectIDVal interface{}
	if sess.ProjectID != nil {
		projectIDVal = *sess.ProjectID
	}

	var blockIDVal interface{}
	if sess.BlockID != nil {
		blockIDVal = *sess.BlockID
	}

	res, err := s.db.Exec(query, sess.Type, sess.Task, sess.Note, sess.StartedAt, sess.EndedAt, completedInt, sess.DurationSecs, projectIDVal, sess.Mode, blockIDVal)
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
	query := `SELECT s.id, s.type, s.task, s.note, s.started_at, s.ended_at, s.completed, s.duration_secs, s.project_id, p.name, s.mode, s.block_id 
	          FROM sessions s
	          LEFT JOIN projects p ON s.project_id = p.id
	          WHERE s.started_at >= ? AND s.started_at <= ?
	          ORDER BY s.started_at ASC`

	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess Session
		var completedInt int
		var taskNull sql.NullString
		var noteNull sql.NullString
		var projectIDNull sql.NullInt64
		var projectNameNull sql.NullString
		var modeNull sql.NullString
		var blockIDNull sql.NullInt64
		err := rows.Scan(&sess.ID, &sess.Type, &taskNull, &noteNull, &sess.StartedAt, &sess.EndedAt, &completedInt, &sess.DurationSecs, &projectIDNull, &projectNameNull, &modeNull, &blockIDNull)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		sess.Completed = completedInt == 1
		if taskNull.Valid {
			sess.Task = taskNull.String
		}
		if noteNull.Valid {
			sess.Note = noteNull.String
		}
		if projectIDNull.Valid {
			id := projectIDNull.Int64
			sess.ProjectID = &id
		}
		if projectNameNull.Valid {
			sess.ProjectName = projectNameNull.String
		}
		if modeNull.Valid {
			sess.Mode = modeNull.String
		}
		if blockIDNull.Valid {
			id := blockIDNull.Int64
			sess.BlockID = &id
		}
		sessions = append(sessions, &sess)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading sessions rows: %w", err)
	}

	return sessions, nil
}

// CreateProject inserts a new project record.
func (s *Store) CreateProject(p *Project) error {
	query := `INSERT INTO projects (name, color, archived, icon) VALUES (?, ?, ?, ?)`
	archivedInt := 0
	if p.Archived {
		archivedInt = 1
	}

	res, err := s.db.Exec(query, p.Name, p.Color, archivedInt, p.Icon)
	if err != nil {
		return fmt.Errorf("failed to insert project: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		p.ID = id
	}

	return nil
}

// GetProjectByName retrieves a project by its name.
func (s *Store) GetProjectByName(name string) (*Project, error) {
	query := `SELECT id, name, color, archived, icon FROM projects WHERE name = ?`
	var p Project
	var archivedInt int
	err := s.db.QueryRow(query, name).Scan(&p.ID, &p.Name, &p.Color, &archivedInt, &p.Icon)
	if err != nil {
		return nil, err
	}
	p.Archived = archivedInt == 1
	return &p, nil
}

// GetProjects returns all projects.
func (s *Store) GetProjects() ([]*Project, error) {
	query := `SELECT id, name, color, archived, icon FROM projects ORDER BY name ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		var archivedInt int
		if err := rows.Scan(&p.ID, &p.Name, &p.Color, &archivedInt, &p.Icon); err != nil {
			return nil, err
		}
		p.Archived = archivedInt == 1
		projects = append(projects, &p)
	}

	return projects, nil
}

// ArchiveProject marks a project as archived.
func (s *Store) ArchiveProject(name string) error {
	query := `UPDATE projects SET archived = 1 WHERE name = ?`
	_, err := s.db.Exec(query, name)
	return err
}

// UnarchiveProject marks a project as active (not archived).
func (s *Store) UnarchiveProject(name string) error {
	query := `UPDATE projects SET archived = 0 WHERE name = ?`
	_, err := s.db.Exec(query, name)
	return err
}

// GetUniqueTasks returns distinct task names, scoped to a project when projectID is non-nil.
func (s *Store) GetUniqueTasks(projectID *int64) ([]string, error) {
	var query string
	var args []interface{}
	if projectID != nil {
		query = `SELECT DISTINCT task FROM sessions WHERE task IS NOT NULL AND task != '' AND project_id = ? ORDER BY started_at DESC LIMIT 50`
		args = append(args, *projectID)
	} else {
		query = `SELECT DISTINCT task FROM sessions WHERE task IS NOT NULL AND task != '' AND project_id IS NULL ORDER BY started_at DESC LIMIT 50`
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var task string
		if err := rows.Scan(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// DeleteTaskName clears the task description from sessions matching the task name, scoped to a project.
func (s *Store) DeleteTaskName(task string, projectID *int64) error {
	if projectID != nil {
		query := `UPDATE sessions SET task = NULL WHERE task = ? AND project_id = ?`
		_, err := s.db.Exec(query, task, *projectID)
		return err
	}
	query := `UPDATE sessions SET task = NULL WHERE task = ? AND project_id IS NULL`
	_, err := s.db.Exec(query, task)
	return err
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
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT,
			archived INTEGER NOT NULL DEFAULT 0
		);`,
		`ALTER TABLE sessions ADD COLUMN project_id INTEGER REFERENCES projects(id);`,
		`CREATE TABLE IF NOT EXISTS blocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			mode TEXT NOT NULL,
			planned_secs INTEGER NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME,
			completed INTEGER NOT NULL,
			pauses INTEGER DEFAULT 0
		);`,
		`ALTER TABLE sessions ADD COLUMN mode TEXT;`,
		`ALTER TABLE sessions ADD COLUMN block_id INTEGER REFERENCES blocks(id);`,
		`ALTER TABLE projects ADD COLUMN icon TEXT DEFAULT '';`,
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

// CreateBlock inserts a new block record.
func (s *Store) CreateBlock(b *BlockStore) error {
	query := `INSERT INTO blocks (mode, planned_secs, started_at, completed, pauses) VALUES (?, ?, ?, ?, ?)`
	completedInt := 0
	if b.Completed {
		completedInt = 1
	}
	res, err := s.db.Exec(query, b.Mode, b.PlannedSecs, b.StartedAt, completedInt, b.Pauses)
	if err != nil {
		return fmt.Errorf("failed to insert block: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		b.ID = id
	}
	return nil
}

// FinishBlock updates the block record at completion.
func (s *Store) FinishBlock(id int64, completed bool, endedAt time.Time) error {
	query := `UPDATE blocks SET completed = ?, ended_at = ? WHERE id = ?`
	completedInt := 0
	if completed {
		completedInt = 1
	}
	_, err := s.db.Exec(query, completedInt, endedAt, id)
	return err
}

// IncrementBlockPauses increments the pauses counter of a block.
func (s *Store) IncrementBlockPauses(id int64) error {
	query := `UPDATE blocks SET pauses = pauses + 1 WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// GetLastBlock retrieves the most recently created block.
func (s *Store) GetLastBlock() (*BlockStore, error) {
	row := s.db.QueryRow("SELECT id, mode, planned_secs, started_at, ended_at, completed, pauses FROM blocks ORDER BY id DESC LIMIT 1")
	var b BlockStore
	var endedAt sql.NullTime
	err := row.Scan(&b.ID, &b.Mode, &b.PlannedSecs, &b.StartedAt, &endedAt, &b.Completed, &b.Pauses)
	if err != nil {
		return nil, err
	}
	if endedAt.Valid {
		b.EndedAt = endedAt.Time
	}
	return &b, nil
}
