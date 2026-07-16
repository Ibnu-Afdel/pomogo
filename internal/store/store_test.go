package store

import (
	"os"
	"path/filepath"
	"strings"
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

func TestProjects(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pomogo-projects-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	st, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to open store: %v", err)
	}
	defer st.Close()

	// 1. Create a project
	p := &Project{
		Name:  "testing-proj",
		Color: "red",
	}
	if err := st.CreateProject(p); err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	if p.ID == 0 {
		t.Errorf("expected populated project ID, got 0")
	}

	// 2. Fetch project by name
	fetched, err := st.GetProjectByName("testing-proj")
	if err != nil {
		t.Fatalf("failed to fetch project: %v", err)
	}
	if fetched.Name != p.Name || fetched.Color != p.Color {
		t.Errorf("fetched project mismatch: %+v", fetched)
	}

	// 3. Save a session linked to the project
	now := time.Now().Truncate(time.Second)
	sess := &Session{
		Type:         "work",
		Task:         "Task with project",
		StartedAt:    now.Add(-25 * time.Minute),
		EndedAt:      now,
		Completed:    true,
		DurationSecs: 1500,
		ProjectID:    &p.ID,
	}
	if err := st.SaveSession(sess); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	// 4. Retrieve session and check project relations
	sessions, err := st.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	got := sessions[0]
	if got.ProjectID == nil || *got.ProjectID != p.ID {
		t.Errorf("expected project ID %d, got %v", p.ID, got.ProjectID)
	}
	if got.ProjectName != p.Name {
		t.Errorf("expected project name %q, got %q", p.Name, got.ProjectName)
	}

	// 5. Get all projects
	list, err := st.GetProjects()
	if err != nil {
		t.Fatalf("failed to get projects: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 project in list, got %d", len(list))
	}

	// 6. Archive project
	if err := st.ArchiveProject("testing-proj"); err != nil {
		t.Fatalf("failed to archive project: %v", err)
	}
	fetchedArchived, err := st.GetProjectByName("testing-proj")
	if err != nil {
		t.Fatalf("failed to get project: %v", err)
	}
	if !fetchedArchived.Archived {
		t.Errorf("expected project to be archived, but it was not")
	}

	// 7. Test tasks queries — scoped to project
	tasks, err := st.GetUniqueTasks(&p.ID)
	if err != nil {
		t.Fatalf("failed to get unique tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0] != "Task with project" {
		t.Errorf("expected tasks list ['Task with project'], got %v", tasks)
	}

	// 7b. Tasks without matching project should return empty
	otherID := int64(9999)
	tasksOther, err := st.GetUniqueTasks(&otherID)
	if err != nil {
		t.Fatalf("failed to get unique tasks for other project: %v", err)
	}
	if len(tasksOther) != 0 {
		t.Errorf("expected no tasks for other project, got %v", tasksOther)
	}

	// 8. Delete task name — scoped to project
	if err := st.DeleteTaskName("Task with project", &p.ID); err != nil {
		t.Fatalf("failed to delete task name: %v", err)
	}
	tasksDeleted, err := st.GetUniqueTasks(&p.ID)
	if err != nil {
		t.Fatalf("failed to get unique tasks: %v", err)
	}
	if len(tasksDeleted) != 0 {
		t.Errorf("expected empty tasks list after deletion, got %v", tasksDeleted)
	}

	// 9. Test exports
	// Create another session to export
	sess2 := &Session{
		Type:         "work",
		Task:         "Export Task",
		Note:         "A note",
		StartedAt:    now.Add(-30 * time.Minute),
		EndedAt:      now.Add(-5 * time.Minute),
		Completed:    true,
		DurationSecs: 1500,
	}
	if err := st.SaveSession(sess2); err != nil {
		t.Fatalf("failed to save export test session: %v", err)
	}

	jsonExport, err := st.ExportSessions("json", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Errorf("json export failed: %v", err)
	}
	if !strings.Contains(jsonExport, "Export Task") {
		t.Errorf("expected json export to contain 'Export Task', got:\n%s", jsonExport)
	}

	csvExport, err := st.ExportSessions("csv", now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Errorf("csv export failed: %v", err)
	}
	if !strings.Contains(csvExport, "Export Task") {
		t.Errorf("expected csv export to contain 'Export Task', got:\n%s", csvExport)
	}

	mdReport, err := st.GenerateMarkdownReport(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Errorf("markdown report failed: %v", err)
	}
	if !strings.Contains(mdReport, "# PomoGo Focus Report") {
		t.Errorf("expected markdown report title, got:\n%s", mdReport)
	}
}

func TestBlocksPersistence(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pomogo-blocks-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	st, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to initialize store: %v", err)
	}
	defer st.Close()

	// Create block
	b := &BlockStore{
		Mode:        "deep",
		PlannedSecs: 3600,
		StartedAt:   time.Now(),
		Completed:   false,
		Pauses:      0,
	}
	err = st.CreateBlock(b)
	if err != nil {
		t.Fatalf("failed to create block: %v", err)
	}
	if b.ID == 0 {
		t.Error("expected non-zero block ID after insertion")
	}

	// Increment pauses
	err = st.IncrementBlockPauses(b.ID)
	if err != nil {
		t.Fatalf("failed to increment block pauses: %v", err)
	}

	// Finish block
	now := time.Now()
	err = st.FinishBlock(b.ID, true, now)
	if err != nil {
		t.Fatalf("failed to finish block: %v", err)
	}

	// Save session referencing this block
	sess := &Session{
		Type:         "work",
		Task:         "Deep Focus Segment",
		StartedAt:    now.Add(-25 * time.Minute),
		EndedAt:      now,
		Completed:    true,
		DurationSecs: 1500,
		Mode:         "deep",
		BlockID:      &b.ID,
	}
	err = st.SaveSession(sess)
	if err != nil {
		t.Fatalf("failed to save session referencing block: %v", err)
	}

	// Fetch sessions and verify
	sessions, err := st.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if *sessions[0].BlockID != b.ID {
		t.Errorf("got block ID %d, want %d", *sessions[0].BlockID, b.ID)
	}
	if sessions[0].Mode != "deep" {
		t.Errorf("got mode %q, want 'deep'", sessions[0].Mode)
	}
}
