package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	// Create a temp directory for a real file-based SQLite test to ensure file paths/directories are resolved.
	tempDir, err := os.MkdirTemp("", "pomogo-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	// Test creation & migration
	st, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}
	defer st.Close()

	// Verify DB file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected DB file to exist at %s, but it does not", dbPath)
	}

	// Test SaveSession
	now := time.Now().Truncate(time.Second) // SQLite might drop nanosecond precision, so truncate to second
	sess := &Session{
		Type:         "work",
		Task:         "Coding PomoGo",
		Note:         "Implemented SQLite storage",
		StartedAt:    now.Add(-25 * time.Minute),
		EndedAt:      now,
		Completed:    true,
		DurationSecs: 1500,
	}

	err = st.SaveSession(sess)
	if err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	if sess.ID == 0 {
		t.Errorf("expected session ID to be populated, got 0")
	}

	// Test GetSessions
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)
	sessions, err := st.GetSessions(start, end)
	if err != nil {
		t.Fatalf("failed to get sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	got := sessions[0]
	if got.ID != sess.ID {
		t.Errorf("expected ID %d, got %d", sess.ID, got.ID)
	}
	if got.Type != sess.Type {
		t.Errorf("expected Type %q, got %q", sess.Type, got.Type)
	}
	if got.Task != sess.Task {
		t.Errorf("expected Task %q, got %q", sess.Task, got.Task)
	}
	if got.Note != sess.Note {
		t.Errorf("expected Note %q, got %q", sess.Note, got.Note)
	}
	if !got.StartedAt.Equal(sess.StartedAt) {
		t.Errorf("expected StartedAt %v, got %v", sess.StartedAt, got.StartedAt)
	}
	if !got.EndedAt.Equal(sess.EndedAt) {
		t.Errorf("expected EndedAt %v, got %v", sess.EndedAt, got.EndedAt)
	}
	if got.Completed != sess.Completed {
		t.Errorf("expected Completed %v, got %v", sess.Completed, got.Completed)
	}
	if got.DurationSecs != sess.DurationSecs {
		t.Errorf("expected DurationSecs %d, got %d", sess.DurationSecs, got.DurationSecs)
	}

	// Test query out of range
	outSessions, err := st.GetSessions(now.Add(2*time.Hour), now.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if len(outSessions) != 0 {
		t.Errorf("expected 0 sessions in out-of-range query, got %d", len(outSessions))
	}
}
