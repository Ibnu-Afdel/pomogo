package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/render"
	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	"github.com/Ibnu-Afdel/pomogo/internal/ui/screens"
)

func TestNewModel(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model == nil {
		t.Fatal("NewModel returned nil")
	}
	if model.cfg != cfg {
		t.Error("Config not set correctly")
	}
	if model.runner == nil {
		t.Fatal("Runner not initialized")
	}
	if model.theme == nil {
		t.Fatal("Theme not initialized")
	}
}

func TestModelInit(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestViewRender(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()

	view := model.View()
	if view == "" {
		t.Error("View returned empty string")
	}
}

func TestViewSmallTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 30
	model.height = 10

	view := model.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("expected 'Terminal too small', got: %q", view)
	}
}

func TestViewTinyTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 20
	model.height = 5

	view := model.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("expected 'Terminal too small', got: %q", view)
	}
}

func TestGoldenClassicRender(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24
	model.currentProjectName = "TestProject"
	model.currentTask = "TestTask"
	model.gitBranch = "feature/deep-work"

	// Run the render
	ds := model.displayState()
	f := render.Frame{Width: 80, Height: 24}
	got := render.Classic(ds, model.theme, f)

	goldenPath := "testdata/classic.golden"
	if os.Getenv("UPDATE_GOLDENS") != "" {
		_ = os.MkdirAll("testdata", 0755)
		_ = os.WriteFile(goldenPath, []byte(got), 0644)
	} else if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		_ = os.MkdirAll("testdata", 0755)
		err := os.WriteFile(goldenPath, []byte(got), 0644)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Log("Created golden file")
		return
	}

	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}
	want := string(wantBytes)

	if got != want {
		t.Errorf("rendered output does not match golden file %s.\nGOT:\n%s\nWANT:\n%s", goldenPath, got, want)
	}
}

func TestGoldenDeepClassicRender(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24
	model.currentProjectName = "DeepProject"
	model.currentTask = "DeepTask"
	model.gitBranch = "feature/deep-work"

	// Select Deep Focus
	block := session.NewDeepBlock(2*time.Hour, 25*time.Minute, 5*time.Minute, true)
	model.runner = session.NewRunner(block)
	model.selectedMode = session.ModeDeep

	// Start the runner so it is running Work
	_ = model.runner.Start(timer.RealClock{})

	// 1. Test Work Golden
	dsWork := model.displayState()
	f := render.Frame{Width: 80, Height: 24}
	gotWork := render.Classic(dsWork, model.theme, f)

	goldenWorkPath := "testdata/deep_work.golden"
	if os.Getenv("UPDATE_GOLDENS") != "" {
		_ = os.MkdirAll("testdata", 0755)
		_ = os.WriteFile(goldenWorkPath, []byte(gotWork), 0644)
	} else if _, err := os.Stat(goldenWorkPath); os.IsNotExist(err) {
		_ = os.MkdirAll("testdata", 0755)
		err := os.WriteFile(goldenWorkPath, []byte(gotWork), 0644)
		if err != nil {
			t.Fatalf("failed to write golden work file: %v", err)
		}
	} else {
		wantBytes, err := os.ReadFile(goldenWorkPath)
		if err != nil {
			t.Fatalf("failed to read golden work file: %v", err)
		}
		want := string(wantBytes)
		if gotWork != want {
			t.Errorf("Deep work output does not match golden file %s.\nGOT:\n%s\nWANT:\n%s", goldenWorkPath, gotWork, want)
		}
	}

	// 2. Skip to break segment
	_, _ = model.runner.Skip(timer.RealClock{})
	dsBreak := model.displayState()
	gotBreak := render.Classic(dsBreak, model.theme, f)

	goldenBreakPath := "testdata/deep_break.golden"
	if os.Getenv("UPDATE_GOLDENS") != "" {
		_ = os.MkdirAll("testdata", 0755)
		_ = os.WriteFile(goldenBreakPath, []byte(gotBreak), 0644)
	} else if _, err := os.Stat(goldenBreakPath); os.IsNotExist(err) {
		_ = os.MkdirAll("testdata", 0755)
		err := os.WriteFile(goldenBreakPath, []byte(gotBreak), 0644)
		if err != nil {
			t.Fatalf("failed to write golden break file: %v", err)
		}
	} else {
		wantBytes, err := os.ReadFile(goldenBreakPath)
		if err != nil {
			t.Fatalf("failed to read golden break file: %v", err)
		}
		want := string(wantBytes)
		if gotBreak != want {
			t.Errorf("Deep break output does not match golden file %s.\nGOT:\n%s\nWANT:\n%s", goldenBreakPath, gotBreak, want)
		}
	}
}

func TestGetDurationForPhase(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	tests := []struct {
		phase    timer.SessionPhase
		expected time.Duration
	}{
		{timer.PhaseWork, cfg.WorkDurationAsDuration()},
		{timer.PhaseShortBreak, cfg.ShortBreakDurationAsDuration()},
		{timer.PhaseLongBreak, cfg.LongBreakDurationAsDuration()},
	}

	for _, tt := range tests {
		model.runner.Block.CurrentSegment.Duration = tt.expected
		got := model.getDurationForPhase()
		if got != tt.expected {
			t.Errorf("getDurationForPhase(%v) = %v, want %v", tt.phase, got, tt.expected)
		}
	}
}

func TestWindowResize(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.width != 80 || model.height != 24 {
		t.Error("Initial dimensions should be 80x24")
	}

	model.width = 120
	model.height = 40

	if model.width != 120 || model.height != 40 {
		t.Error("Dimensions not updated correctly")
	}
}

func TestSessionTracking(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.runner.Timer.IsRunning {
		t.Error("Session should not be running initially")
	}
}

func TestKeyboardQuitAcrossThemesAndLayouts(t *testing.T) {
	for _, themeName := range theme.List() {
		for layoutName := range render.Registry {
			t.Run(themeName+"/"+layoutName, func(t *testing.T) {
				model := newKeyboardTestModel(t, themeName, layoutName)
				_, cmd := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("q")}))
				if !isQuitCmd(cmd) {
					t.Fatalf("q did not return tea.Quit for theme=%s layout=%s", themeName, layoutName)
				}
			})
		}
	}
}

func TestKeyboardSpacePauseResumeAcrossThemesAndLayouts(t *testing.T) {
	for _, themeName := range theme.List() {
		for layoutName := range render.Registry {
			t.Run(themeName+"/"+layoutName, func(t *testing.T) {
				model := newKeyboardTestModel(t, themeName, layoutName)
				if err := model.runner.Start(timer.RealClock{}); err != nil {
					t.Fatalf("start failed: %v", err)
				}

				_, cmd := model.Update(tea.KeyMsg(tea.Key{Type: tea.KeySpace}))
				if !model.runner.Timer.IsPaused {
					t.Fatalf("space did not pause for theme=%s layout=%s", themeName, layoutName)
				}
				if view := model.View(); view == "" {
					t.Fatal("paused view is empty")
				}

				_, cmd = model.Update(tea.KeyMsg(tea.Key{Type: tea.KeySpace}))
				_ = cmd
				if model.runner.Timer.IsPaused {
					t.Fatalf("second space did not resume for theme=%s layout=%s", themeName, layoutName)
				}
				if view := model.View(); view == "" {
					t.Fatal("resumed view is empty")
				}
			})
		}
	}
}

func TestGlobalKeysAcrossThemesAndLayouts(t *testing.T) {
	for _, themeName := range theme.List() {
		for layoutName := range render.Registry {
			t.Run(themeName+"/"+layoutName, func(t *testing.T) {
				model := newKeyboardTestModel(t, themeName, layoutName)

				_, _ = model.Update(keyRunes("d"))
				if model.inputMode != modeDurationPicker {
					t.Fatal("d did not open Deep Focus duration picker")
				}
				_, _ = model.Update(keyEsc())
				if model.inputMode != modeNone {
					t.Fatal("esc did not close duration picker")
				}

				_, _ = model.Update(keyRunes("t"))
				if model.inputMode != modeTaskInput {
					t.Fatal("t did not open task input")
				}
				_, _ = model.Update(keyEsc())
				if model.inputMode != modeNone {
					t.Fatal("esc did not close task input")
				}

				_, _ = model.Update(keyRunes("p"))
				if model.inputMode != modeProjectInput {
					t.Fatal("p did not open project input")
				}
				_, _ = model.Update(keyEsc())
				if model.inputMode != modeNone {
					t.Fatal("esc did not close project input")
				}

				_, _ = model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
				if !model.showStats {
					t.Fatal("tab did not open stats")
				}
				_, _ = model.Update(keyEsc())
				if model.showStats {
					t.Fatal("esc did not close stats")
				}

				_, _ = model.Update(keyRunes("?"))
				if !model.showHelp {
					t.Fatal("? did not open help")
				}
				_, _ = model.Update(keyEsc())
				if model.showHelp {
					t.Fatal("esc did not close help")
				}

				oldTheme := model.currentThemeName
				_, _ = model.Update(keyRunes("T"))
				if model.currentThemeName == oldTheme {
					t.Fatal("T did not cycle theme")
				}

				oldLayout := model.currentLayoutName
				_, _ = model.Update(keyRunes("L"))
				if model.currentLayoutName == oldLayout {
					t.Fatal("L did not cycle layout")
				}

				_, _ = model.Update(keyRunes("S"))
				if !model.zenMode {
					t.Fatal("S did not enable zen mode")
				}
				_, _ = model.Update(keyEsc())
				if model.zenMode {
					t.Fatal("esc did not exit zen mode")
				}

				oldEffects := model.currentEffectsName
				_, _ = model.Update(keyRunes("e"))
				if model.currentEffectsName == oldEffects {
					t.Fatal("e did not cycle effects")
				}

				oldVerb := model.currentVerbLabel
				_, _ = model.Update(keyRunes("v"))
				if model.currentVerbLabel == oldVerb {
					t.Fatal("v did not cycle activity verb")
				}

				_, _ = model.Update(keyRunes("s"))
				if !model.runner.Timer.IsRunning {
					t.Fatal("s did not start timer")
				}

				_, _ = model.Update(keyRunes("n"))
				if model.runner.Block.Index == 0 && model.runner.Block.CurrentSegment.Kind == session.SegmentKindWork {
					t.Fatal("n did not skip the current segment")
				}

				_, _ = model.Update(keyRunes("r"))
				if model.runner.Timer.IsRunning || model.runner.Timer.IsPaused {
					t.Fatalf("r did not reset timer: inputMode=%d state=%s phase=%s running=%v paused=%v status=%q",
						model.inputMode,
						model.runner.Timer.State,
						model.runner.Timer.Phase,
						model.runner.Timer.IsRunning,
						model.runner.Timer.IsPaused,
						model.statusMessage,
					)
				}
			})
		}
	}
}

func newKeyboardTestModel(t *testing.T, themeName, layoutName string) *Model {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	cfg := config.Default()
	cfg.Theme = themeName
	cfg.Layout = layoutName
	cfg.PromptForNotes = false
	model := NewModel(cfg)
	model.width = 120
	model.height = 32
	model.restorePending = false
	model.currentThemeName = themeName
	model.theme = theme.Get(themeName)
	model.currentLayoutName = layoutName
	return model
}

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune(s)})
}

func keyEsc() tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyEsc})
}

func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestThemeLoading(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.theme == nil {
		t.Fatal("Theme should be loaded")
	}
	if model.theme.Name == "" {
		t.Error("Theme name should not be empty")
	}
}

func TestMultipleModels(t *testing.T) {
	cfg1 := config.Default()
	cfg2 := config.Default()

	m1 := NewModel(cfg1)
	m2 := NewModel(cfg2)

	if m1 == m2 {
		t.Error("Each NewModel should create a new instance")
	}
	if m1.runner.Timer == m2.runner.Timer {
		t.Error("Each model should have its own session")
	}
}

func TestPhaseLabel(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	// Idle state
	label := phaseLabel(model.runner.Timer)
	if label != "Focus Timer" {
		t.Errorf("idle label = %q, want %q", label, "Focus Timer")
	}

	// Start a work session
	_ = model.runner.Timer.Start(timer.RealClock{})
	label = phaseLabel(model.runner.Timer)
	if label != "Work" {
		t.Errorf("work label = %q, want %q", label, "Work")
	}

	// Skip to short break
	model.runner.Timer.Skip()
	label = phaseLabel(model.runner.Timer)
	if label != "Short Break" {
		t.Errorf("short break label = %q, want %q", label, "Short Break")
	}
}

func TestBigClockRows(t *testing.T) {
	rows := render.BigClockRows("00:00", lipgloss.Color("#ff0000"))
	if len(rows) != 5 {
		t.Errorf("BigClockRows returned %d rows, want 5", len(rows))
	}
	for i, row := range rows {
		if row == "" {
			t.Errorf("row %d is empty", i)
		}
	}
}

func TestProgressBar(t *testing.T) {
	bar := render.ProgressBar(0.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if bar == "" {
		t.Error("ProgressBar returned empty string")
	}

	// Test edge cases
	_ = render.ProgressBar(0.0, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = render.ProgressBar(1.0, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = render.ProgressBar(1.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = render.ProgressBar(-0.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
}

func TestSessionDots(t *testing.T) {
	dots := render.SessionDots(2, 4, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if dots == "" {
		t.Error("SessionDots returned empty string")
	}

	// Test zero completed
	dots = render.SessionDots(0, 4, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if dots == "" {
		t.Error("SessionDots returned empty string for 0 completed")
	}
}

func BenchmarkRender(b *testing.B) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkGetDurationForPhase(b *testing.B) {
	cfg := config.Default()
	model := NewModel(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.getDurationForPhase()
	}
}

func TestRecordSession(t *testing.T) {
	// Set XDG_DATA_HOME to a temp directory so we don't overwrite the real DB
	tempDir, err := os.MkdirTemp("", "pomogo-ui-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDGDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer func() {
		if oldXDGDataHome == "" {
			os.Unsetenv("XDG_DATA_HOME")
		} else {
			os.Setenv("XDG_DATA_HOME", oldXDGDataHome)
		}
	}()

	cfg := config.Default()
	cfg.PromptForNotes = false // Disable notes prompt for immediate save test
	model := NewModel(cfg)
	if model.dbStore == nil {
		t.Fatal("expected dbStore to be initialized, got nil")
	}

	// Record a completed work session (immediate save)
	now := time.Now().Truncate(time.Second)
	model.recordSession(timer.PhaseWork, now.Add(-25*time.Minute), now, true, 25*time.Minute)

	// Retrieve sessions from dbStore
	sessions, err := model.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session to be recorded, got %d", len(sessions))
	}

	got := sessions[0]
	if got.Type != "work" {
		t.Errorf("expected type 'work', got %q", got.Type)
	}
	if !got.Completed {
		t.Errorf("expected completed to be true")
	}
	if got.DurationSecs != 1500 {
		t.Errorf("expected duration_secs to be 1500, got %d", got.DurationSecs)
	}

	// Verify that break sessions are NOT recorded
	model.recordSession(timer.PhaseShortBreak, now, now.Add(5*time.Minute), true, 5*time.Minute)
	sessions2, err := model.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if len(sessions2) != 1 {
		t.Errorf("expected break session to be ignored, got count %d", len(sessions2))
	}

	// Test note prompting
	cfg.PromptForNotes = true
	model2 := NewModel(cfg)

	// Trigger recordSession with PromptForNotes = true
	model2.recordSession(timer.PhaseWork, now.Add(-25*time.Minute), now, true, 25*time.Minute)

	if model2.inputMode != modeNoteInput {
		t.Errorf("expected inputMode to be modeNoteInput, got %v", model2.inputMode)
	}
	if model2.pendingSession == nil {
		t.Fatal("expected pendingSession to be set")
	}

	// Simulate entering a note and pressing enter
	model2.textInput.SetValue("Completed feature X")
	model2.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))

	if model2.inputMode != modeNone {
		t.Errorf("expected inputMode to revert to modeNone, got %v", model2.inputMode)
	}

	// Retrieve sessions from model2's store (which is the same DB file)
	sessions3, err := model2.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}

	if len(sessions3) != 2 {
		t.Fatalf("expected 2 sessions to be recorded, got %d", len(sessions3))
	}
	if sessions3[1].Note != "Completed feature X" {
		t.Errorf("expected note 'Completed feature X', got %q", sessions3[1].Note)
	}
}

func TestProjectInput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pomogo-ui-project-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDGDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer func() {
		if oldXDGDataHome == "" {
			os.Unsetenv("XDG_DATA_HOME")
		} else {
			os.Setenv("XDG_DATA_HOME", oldXDGDataHome)
		}
	}()

	cfg := config.Default()
	cfg.PromptForNotes = false
	model := NewModel(cfg)

	// Simulate pressing 'p'
	model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune("p")}))
	if model.inputMode != modeProjectInput {
		t.Errorf("expected inputMode modeProjectInput, got %v", model.inputMode)
	}

	// Enter project name "frontend" and press Enter
	model.textInput.SetValue("frontend")
	model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyEnter}))

	if model.inputMode != modeNone {
		t.Errorf("expected inputMode modeNone, got %v", model.inputMode)
	}
	if model.currentProjectName != "frontend" {
		t.Errorf("expected currentProjectName 'frontend', got %q", model.currentProjectName)
	}
	if model.currentProjectID == nil {
		t.Errorf("expected currentProjectID to be populated")
	}

	// Record session and check DB association
	now := time.Now().Truncate(time.Second)
	model.recordSession(timer.PhaseWork, now.Add(-25*time.Minute), now, true, 25*time.Minute)

	sessions, err := model.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to get sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].ProjectName != "frontend" {
		t.Errorf("expected ProjectName 'frontend', got %q", sessions[0].ProjectName)
	}
}

func TestSuggestionsAndNavigation(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	// Set suggestions manually
	model.suggestions = []string{"backend", "frontend", "sprint-dev"}
	model.textInput.SetValue("ba")
	model.filterSuggestions()

	if len(model.filteredSuggestions) != 1 || model.filteredSuggestions[0] != "backend" {
		t.Errorf("expected filteredSuggestions ['backend'], got %v", model.filteredSuggestions)
	}

	// Pressing Down should highlight 'backend'
	model.inputMode = modeProjectInput
	model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyDown}))
	if model.suggestionIndex != 0 {
		t.Errorf("expected suggestionIndex 0, got %d", model.suggestionIndex)
	}

	// Pressing Tab should select 'backend'
	model.Update(tea.KeyMsg(tea.Key{Type: tea.KeyTab}))
	if model.textInput.Value() != "backend" {
		t.Errorf("expected textInput value 'backend', got %q", model.textInput.Value())
	}
	if model.suggestionIndex != -1 {
		t.Errorf("expected suggestionIndex reset to -1, got %d", model.suggestionIndex)
	}
}

func TestCombinationSweepNoPanics(t *testing.T) {
	cfg := config.Default()
	themes := []string{
		"tokyo-night", "catppuccin", "catppuccin-latte", "gruvbox",
		"rose-pine", "everforest", "nord", "dracula", "kanagawa", "carbon",
		"night-owl", "one-dark", "ayu-mirage", "solarized-dark", "oxocarbon",
	}
	layouts := []string{"classic", "minimal", "centered", "compact", "retro", "dashboard", "monolith", "tinybar", "terminal-rice"}
	sizes := []struct{ w, h int }{
		{60, 18},
		{200, 50},
	}

	for _, themeName := range themes {
		for _, layoutName := range layouts {
			for _, sz := range sizes {
				model := NewModel(cfg)
				model.width = sz.w
				model.height = sz.h
				model.currentThemeName = themeName
				model.theme = theme.Get(themeName)
				model.currentLayoutName = layoutName

				// Test Quick Focus Running
				_ = model.runner.Start(timer.RealClock{})
				view := model.View()
				if view == "" {
					t.Errorf("empty render for theme %s, layout %s at %dx%d", themeName, layoutName, sz.w, sz.h)
				}

				// Test Deep Focus Running
				blockD := session.NewDeepBlock(2*time.Hour, 25*time.Minute, 5*time.Minute, true)
				model.runner = session.NewRunner(blockD)
				model.selectedMode = session.ModeDeep
				_ = model.runner.Start(timer.RealClock{})
				viewD := model.View()
				if viewD == "" {
					t.Errorf("empty render (deep) for theme %s, layout %s at %dx%d", themeName, layoutName, sz.w, sz.h)
				}
			}
		}
	}
}

func TestGoldenLayouts(t *testing.T) {
	cfg := config.Default()
	layouts := []string{"minimal", "centered", "compact", "retro"}

	for _, lName := range layouts {
		t.Run(lName, func(t *testing.T) {
			model := NewModel(cfg)
			model.width = 80
			model.height = 24
			model.currentLayoutName = lName
			model.currentProjectName = "Prj"
			model.currentTask = "Tsk"
			model.gitBranch = "feature/deep-work"

			// Idle state
			dsIdle := model.displayState()
			_, layoutFunc := render.ResolveLayout(lName, 80, 24)
			gotIdle := layoutFunc(dsIdle, model.theme, render.Frame{Width: 80, Height: 24})
			verifyGolden(t, lName+"_idle.golden", gotIdle)

			// Running work state
			_ = model.runner.Start(timer.RealClock{})
			dsRun := model.displayState()
			gotRun := layoutFunc(dsRun, model.theme, render.Frame{Width: 80, Height: 24})
			verifyGolden(t, lName+"_work.golden", gotRun)

			// Paused state
			_ = model.runner.Timer.Pause(timer.RealClock{})
			dsPause := model.displayState()
			gotPause := layoutFunc(dsPause, model.theme, render.Frame{Width: 80, Height: 24})
			verifyGolden(t, lName+"_paused.golden", gotPause)
		})
	}
}

func verifyGolden(t *testing.T, filename string, got string) {
	goldenPath := filepath.Join("testdata", filename)
	if os.Getenv("UPDATE_GOLDENS") != "" {
		_ = os.MkdirAll("testdata", 0755)
		_ = os.WriteFile(goldenPath, []byte(got), 0644)
		return
	}
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) {
		_ = os.MkdirAll("testdata", 0755)
		_ = os.WriteFile(goldenPath, []byte(got), 0644)
		return
	}
	wantBytes, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden: %v", err)
	}
	want := string(wantBytes)
	if got != want {
		t.Errorf("output does not match golden %s", filename)
	}
}

func TestGoldenZenLayouts(t *testing.T) {
	cfg := config.Default()
	layouts := []string{"classic", "minimal", "centered", "compact", "retro"}

	for _, lName := range layouts {
		t.Run(lName, func(t *testing.T) {
			model := NewModel(cfg)
			model.width = 80
			model.height = 24
			model.currentLayoutName = lName
			model.currentProjectName = "Prj"
			model.currentTask = "Tsk"
			model.zenMode = true
			model.gitBranch = "feature/deep-work"

			// Running state in Zen mode
			_ = model.runner.Start(timer.RealClock{})
			dsRun := model.displayState()
			dsRun.Zen = true
			_, layoutFunc := render.ResolveLayout(lName, 80, 24)
			gotRun := layoutFunc(dsRun, model.theme, render.Frame{Width: 80, Height: 24})
			verifyGolden(t, lName+"_zen_work.golden", gotRun)
		})
	}
}

func TestRecapScreen(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24
	model.selectedMode = session.ModeQuick

	// Create mock recap info
	info := screens.RecapInfo{
		TotalFocused: 1*time.Hour + 30*time.Minute,
		Segments:     3,
		Breaks:       2,
		Pauses:       1,
		Streak:       5,
		IsDeep:       true,
		FocusScore:   8,
	}

	got := screens.Recap(80, 24, model.theme, info)
	verifyGolden(t, "recap.golden", got)
}

func TestVerbLabelTaskKeywords(t *testing.T) {
	tests := []struct {
		task string
		want string
	}{
		{"build docker image", "Building"},
		{"fix compilation error", "Fixing"},
		{"debug session runner", "Debugging"},
		{"read specifications", "Reading"},
		{"write tests", "Writing"},
		{"learn bubble tea patterns", "Learning"},
		{"design layouts system", "Designing"},
		{"research particle fields", "Researching"},
		{"random coding", "Focusing"},
	}

	for _, tt := range tests {
		got := GetVerbForTask(tt.task)
		if got != tt.want {
			t.Errorf("GetVerbForTask(%q) = %q, want %q", tt.task, got, tt.want)
		}
	}
}

func TestGoldenProjectIcon(t *testing.T) {
	cfg := config.Default()
	layouts := []string{"classic", "minimal", "centered", "compact", "retro"}

	for _, lName := range layouts {
		t.Run(lName, func(t *testing.T) {
			model := NewModel(cfg)
			model.width = 80
			model.height = 24
			model.currentLayoutName = lName
			model.currentProjectName = "Fawz"
			model.currentProjectIcon = "🐹"
			model.currentTask = "build auth"
			model.currentVerbLabel = "Building"
			model.gitBranch = "feature/deep-work"

			ds := model.displayState()
			_, layoutFunc := render.ResolveLayout(lName, 80, 24)
			got := layoutFunc(ds, model.theme, render.Frame{Width: 80, Height: 24})
			verifyGolden(t, lName+"_project_icon.golden", got)
		})
	}
}
