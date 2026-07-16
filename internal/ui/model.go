// Package ui provides the Bubble Tea TUI for PomoGo.
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/devinfo"
	"github.com/Ibnu-Afdel/pomogo/internal/notify"
	"github.com/Ibnu-Afdel/pomogo/internal/render"
	"github.com/Ibnu-Afdel/pomogo/internal/restore"
	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/stats"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	"github.com/Ibnu-Afdel/pomogo/internal/ui/screens"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputModeType int

const (
	modeNone inputModeType = iota
	modeTaskInput
	modeNoteInput
	modeProjectInput
	modeDurationPicker
	modeCustomDurationInput
	modeRecapScreen
)

// Model is the main Bubble Tea model for the PomoGo TUI.
type Model struct {
	runner *session.Runner
	cfg    *config.Config
	theme  *theme.Theme
	width  int
	height int
	keymap KeyMap

	notifier       *notify.Notifier
	stateManager   *statefile.Manager
	dbStore        *store.Store
	restorePending bool
	showHelp       bool
	statusMessage  string

	textInput           textinput.Model
	inputMode           inputModeType
	currentTask         string
	pendingSession      *store.Session
	showStats           bool
	lastTickTime        time.Time
	currentProjectID    *int64
	currentProjectName  string
	suggestions         []string
	filteredSuggestions []string
	suggestionIndex     int
	lastHookState       timer.SessionState

	// Mode selection fields
	selectedMode        session.Mode
	selectedDurationIdx int
	deepDuration        time.Duration
	currentBlockID      *int64

	// Theme & Layout fields
	currentThemeName   string
	currentLayoutName  string
	currentEffectsName string
	zenMode            bool
	tickCount          int
	currentProjectIcon string
	currentVerbLabel   string
	gitBranch          string
	tmuxSession        string
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config) *Model {
	// Load external themes first so they are available in theme selection/resolution
	_ = theme.LoadExternalThemes()

	resolvedTheme := theme.ResolveThemeName(cfg.Theme)
	resolvedLayout := render.ResolveLayoutName(cfg.Layout)
	resolvedEffects := render.ResolveEffectsName(cfg.Effects)

	th := theme.Get(resolvedTheme)
	manager, _ := statefile.NewManager()
	st, err := store.New(config.DBFilePath())
	var statusMsg string
	if err != nil {
		statusMsg = fmt.Sprintf("database error: %v", err)
	}

	var task string
	var projectID *int64
	var projectName string
	if manager != nil {
		if state, err := manager.Read(); err == nil && state != nil {
			task = state.Task
			projectID = state.ProjectID
			projectName = state.ProjectName
		}
	}

	ti := textinput.New()
	ti.Width = 30

	var gitBranch string
	var tmuxSession string
	if cfg.ShowGit {
		cwd, _ := os.Getwd()
		gitBranch = devinfo.FindGitBranch(cwd)
	}
	if cfg.ShowTmux {
		tmuxSession = devinfo.GetTmuxSession()
	}

	block := session.NewQuickBlock(
		cfg.QuickFocusWorkDurationAsDuration(),
		cfg.QuickFocusShortBreakDurationAsDuration(),
		cfg.QuickFocusLongBreakDurationAsDuration(),
		cfg.QuickFocusSessionsBeforeLongBreak(),
		cfg.QuickFocusAutoAdvance(),
	)

	return &Model{
		runner:              session.NewRunner(block),
		cfg:                 cfg,
		theme:               th,
		width:               80,
		height:              24,
		keymap:              DefaultKeyMap,
		notifier:            notify.NewNotifier(cfg.NotificationsEnabled, cfg.SoundEnabled),
		stateManager:        manager,
		dbStore:             st,
		restorePending:      restore.CanRestore(),
		statusMessage:       statusMsg,
		textInput:           ti,
		inputMode:           modeNone,
		currentTask:         task,
		currentProjectID:    projectID,
		currentProjectName:  projectName,
		selectedMode:        session.ModeQuick,
		selectedDurationIdx: 0,
		deepDuration:        cfg.DeepFocusDefaultDurationAsDuration(),
		currentThemeName:    resolvedTheme,
		currentLayoutName:   resolvedLayout,
		currentEffectsName:  resolvedEffects,
		zenMode:             false,
		tickCount:           0,
		currentVerbLabel:    "Focusing",
		gitBranch:           gitBranch,
		tmuxSession:         tmuxSession,
	}
}

// Init initializes the model (required by tea.Model).
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseCellMotion,
		m.listenForDbusActions(),
		m.tick1s(),
	)
}

// View renders the UI (required by tea.Model).
func (m *Model) View() string {
	if m.width < 40 || m.height < 10 {
		return "Terminal too small — minimum 40 × 10.\n"
	}

	if m.inputMode == modeRecapScreen {
		return screens.Recap(m.width, m.height, m.theme, m.getRecapInfo())
	}

	if m.inputMode == modeDurationPicker {
		return screens.DurationPicker(
			m.width,
			m.height,
			m.theme,
			m.selectedDurationIdx,
			m.cfg.DeepFocusDefaultDurationAsDuration(),
			m.cfg.DeepFocusWorkDurationAsDuration(),
			m.cfg.DeepFocusShortBreakDurationAsDuration(),
			m.cfg.DeepFocusLongBreakDurationAsDuration(),
			m.cfg.DeepFocusSessionsBeforeLongBreak(),
		)
	}

	if m.inputMode == modeCustomDurationInput {
		return screens.Input(m.width, m.height, m.theme, "custom_duration", m.textInput.View(), nil, -1)
	}

	if m.inputMode != modeNone {
		var modeStr string
		switch m.inputMode {
		case modeNoteInput:
			modeStr = "note"
		case modeProjectInput:
			modeStr = "project"
		default:
			modeStr = "task"
		}
		return screens.Input(m.width, m.height, m.theme, modeStr, m.textInput.View(), m.filteredSuggestions, m.suggestionIndex)
	}

	if m.showHelp {
		var bindings []screens.HelpBinding
		for _, b := range m.keymap.ShortHelp() {
			bindings = append(bindings, screens.HelpBinding{
				Keys:        strings.Join(b.Keys, ", "),
				Description: b.Description,
			})
		}
		return screens.Help(m.width, m.height, m.theme, bindings)
	}

	if m.restorePending {
		return screens.RestorePrompt(m.width, m.height, m.theme, m.phaseColor())
	}

	if m.showStats {
		now := time.Now()
		var sessions []*store.Session
		if m.dbStore != nil {
			start := now.AddDate(-1, 0, 0)
			if s, err := m.dbStore.GetSessions(start, now.Add(24*time.Hour)); err == nil {
				sessions = s
			}
		}
		s := stats.Calculate(sessions, now, m.currentProjectName)

		var recent []*store.Session
		for i := len(sessions) - 1; i >= 0 && len(recent) < 3; i-- {
			if sessions[i].Type == "work" {
				recent = append(recent, sessions[i])
			}
		}
		return screens.Stats(m.width, m.height, m.theme, s, m.statusMessage, recent)
	}

	return m.renderTimerScreen()
}

func (m *Model) renderTimerScreen() string {
	ds := m.displayState()
	resolvedName, layoutFunc := render.ResolveLayout(m.currentLayoutName, m.width, m.height)
	ds.LayoutName = resolvedName
	ds.ThemeName = m.currentThemeName
	ds.Zen = m.zenMode
	f := render.Frame{Width: m.width, Height: m.height}
	layoutContent := layoutFunc(ds, m.theme, f)
	return render.RenderAmbient(m.currentEffectsName, m.tickCount, f, m.theme, layoutContent)
}

func (m *Model) getDurationForPhase() time.Duration {
	return m.runner.Block.CurrentSegment.Duration
}

func (m *Model) displayRemaining() time.Duration {
	if !m.runner.Timer.IsRunning {
		if m.runner.Block.Mode == session.ModeDeep {
			return m.runner.Block.PlannedTotal
		}
		return m.cfg.WorkDurationAsDuration()
	}
	return m.runner.Block.Remaining(m.runner.Timer.RemainingTime)
}

func (m *Model) phaseColor() theme.Color {
	if !m.runner.Timer.IsRunning {
		return m.theme.Idle
	}
	switch m.runner.Timer.Phase {
	case timer.PhaseWork:
		return m.theme.Work
	case timer.PhaseShortBreak:
		return m.theme.Break
	case timer.PhaseLongBreak:
		return m.theme.LongBreak
	default:
		return m.theme.Idle
	}
}

func (m *Model) sessionStatus() string {
	switch {
	case m.runner.Timer.IsPaused:
		return "paused"
	case m.runner.Timer.IsRunning:
		return "running"
	default:
		if m.runner.Block.Mode == session.ModeDeep {
			hours := int(m.runner.Block.PlannedTotal.Hours())
			mins := int(m.runner.Block.PlannedTotal.Minutes()) % 60
			var durStr string
			if hours > 0 {
				durStr = fmt.Sprintf("%dh%dm", hours, mins)
				if mins == 0 {
					durStr = fmt.Sprintf("%dh", hours)
				}
			} else {
				durStr = fmt.Sprintf("%dm", mins)
			}
			return fmt.Sprintf("deep focus %s · press s to start", durStr)
		}
		return "press s to start"
	}
}

func (m *Model) afterTransition(sendNotification bool) {
	m.statusMessage = ""
	m.writeState()
	if sendNotification && m.notifier != nil {
		_ = m.notifier.NotifyTransition(m.runner.Timer.State, m.runner.Timer.Phase)
	}
	m.triggerHooks()
}

func (m *Model) recordSession(phase timer.SessionPhase, startedAt, endedAt time.Time, completed bool, duration time.Duration) {
	if m.dbStore == nil {
		return
	}
	if phase != timer.PhaseWork {
		return
	}

	sess := &store.Session{
		Type:         "work",
		Task:         m.currentTask,
		ProjectID:    m.currentProjectID,
		BlockID:      m.currentBlockID,
		Mode:         string(m.runner.Block.Mode),
		Completed:    completed,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		DurationSecs: int(duration.Seconds()),
	}

	// For deep focus, note prompt is only at the block end, handled in events.
	if m.runner.Block.Mode == session.ModeQuick && m.cfg.PromptForNotes {
		m.pendingSession = sess
		m.inputMode = modeNoteInput
		m.textInput.SetValue("")
		m.textInput.Placeholder = "Enter session note (Enter to skip)..."
		m.textInput.Focus()
	} else {
		if err := m.dbStore.SaveSession(sess); err != nil {
			m.statusMessage = fmt.Sprintf("database error: %v", err)
		}
	}
}

func (m *Model) getStats() *stats.Stats {
	now := time.Now()
	if m.dbStore == nil {
		return stats.Calculate(nil, now, m.currentProjectName)
	}
	start := now.AddDate(-1, 0, 0)
	sessions, err := m.dbStore.GetSessions(start, now.Add(24*time.Hour))
	if err != nil {
		return stats.Calculate(nil, now, m.currentProjectName)
	}
	return stats.Calculate(sessions, now, m.currentProjectName)
}

func (m *Model) writeState() {
	if m.stateManager == nil {
		return
	}
	if err := m.stateManager.Write(m.runner, m.currentTask, m.currentProjectID, m.currentProjectName, m.currentBlockID); err != nil {
		m.statusMessage = fmt.Sprintf("statefile error: %v", err)
	}
}

func (m *Model) removeState() {
	if m.stateManager == nil {
		return
	}
	_ = m.stateManager.Remove()
}

func (m *Model) persistOnQuit() {
	if m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
		m.runner.Timer.Pause(timer.RealClock{})
	}
	m.writeState()
}

func (m *Model) updateTerminalTitle() tea.Cmd {
	return func() tea.Msg {
		title := "Focus Timer"
		if m.runner.Timer.IsRunning {
			mins := int(m.runner.Timer.RemainingTime.Minutes())
			secs := int(m.runner.Timer.RemainingTime.Seconds()) % 60
			stateStr := "work"
			emoji := "🍅"
			if m.runner.Timer.Phase == timer.PhaseShortBreak || m.runner.Timer.Phase == timer.PhaseLongBreak {
				stateStr = "break"
				emoji = "☕"
			}
			if m.runner.Timer.IsPaused {
				title = fmt.Sprintf("⏸️ %02d:%02d · %s", mins, secs, stateStr)
			} else {
				title = fmt.Sprintf("%s %02d:%02d · %s", emoji, mins, secs, stateStr)
			}
		}
		fmt.Printf("\033]2;%s\007", title)
		return nil
	}
}

// SetProjectByName binds a project to the model by name.
func (m *Model) SetProjectByName(name string) {
	if name == "" {
		m.currentProjectID = nil
		m.currentProjectName = ""
		m.currentProjectIcon = ""
		return
	}
	if m.dbStore != nil {
		p, err := m.dbStore.GetProjectByName(name)
		if err != nil {
			p = &store.Project{Name: name}
			if err := m.dbStore.CreateProject(p); err == nil {
				m.currentProjectID = &p.ID
				m.currentProjectName = p.Name
				m.currentProjectIcon = p.Icon
			}
		} else {
			if p.Archived {
				_ = m.dbStore.UnarchiveProject(p.Name)
			}
			m.currentProjectID = &p.ID
			m.currentProjectName = p.Name
			m.currentProjectIcon = p.Icon
		}
	} else {
		m.currentProjectName = name
		m.currentProjectIcon = ""
	}

	if m.cfg.ShowGit {
		cwd, _ := os.Getwd()
		m.gitBranch = devinfo.FindGitBranch(cwd)
	}
}

// SetCustomSoundEvent overrides the notifier sound behavior.
func (m *Model) SetCustomSoundEvent(event string) {
	if m.notifier != nil {
		m.notifier.SetCustomSoundEvent(event)
	}
}

func (m *Model) sessionStateString(state timer.SessionState) string {
	return state.String()
}

func (m *Model) cycleTheme() {
	themes := theme.List()
	if len(themes) == 0 {
		return
	}
	idx := -1
	for i, t := range themes {
		if t == m.currentThemeName {
			idx = i
			break
		}
	}
	nextIdx := (idx + 1) % len(themes)
	m.currentThemeName = themes[nextIdx]
	m.theme = theme.Get(m.currentThemeName)
}

func (m *Model) cycleLayout() {
	layouts := []string{"classic", "minimal", "centered", "compact", "retro", "dashboard", "monolith", "tinybar", "terminal-rice"}
	idx := -1
	for i, l := range layouts {
		if l == m.currentLayoutName {
			idx = i
			break
		}
	}
	nextIdx := (idx + 1) % len(layouts)
	m.currentLayoutName = layouts[nextIdx]
}

func (m *Model) cycleEffects() {
	effects := []string{"none", "stars", "snow", "rain"}
	idx := -1
	for i, eff := range effects {
		if eff == m.currentEffectsName {
			idx = i
			break
		}
	}
	nextIdx := (idx + 1) % len(effects)
	m.currentEffectsName = effects[nextIdx]
}

func (m *Model) getRecapInfo() screens.RecapInfo {
	streak := 0
	var sessions []*store.Session
	if m.dbStore != nil {
		start := time.Now().AddDate(-1, 0, 0)
		if s, err := m.dbStore.GetSessions(start, time.Now().Add(24*time.Hour)); err == nil {
			sessions = s
		}
	}
	s := stats.Calculate(sessions, time.Now(), m.currentProjectName)
	streak = s.CurrentStreak

	var totalFocused time.Duration
	pauses := 0
	segments := 0
	breaks := 0
	isDeep := m.selectedMode == session.ModeDeep
	focusScore := 10

	if isDeep && m.currentBlockID != nil && m.dbStore != nil {
		if b, err := m.dbStore.GetLastBlock(); err == nil && b != nil {
			pauses = b.Pauses
			totalFocused = time.Duration(b.PlannedSecs) * time.Second
		}
		completedCount := 0
		abandoned := 0
		skippedBreaks := 0
		for _, sess := range sessions {
			if sess.BlockID != nil && *sess.BlockID == *m.currentBlockID {
				if sess.Type == "work" {
					if sess.Completed {
						completedCount++
					} else {
						abandoned++
					}
				} else if (sess.Type == "short_break" || sess.Type == "long_break") && !sess.Completed {
					skippedBreaks++
				}
			}
		}
		segments = completedCount
		breaks = countBreakSegments(m.runner.Block.Segments)
		focusScore = stats.CalculateFocusScore(pauses, skippedBreaks, abandoned)
	} else {
		segments = m.cfg.QuickFocusSessionsBeforeLongBreak()
		breaks = segments
		totalFocused = m.cfg.QuickFocusWorkDurationAsDuration() * time.Duration(segments)
		pauses = 0
		focusScore = stats.CalculateFocusScore(pauses, 0, 0)
	}

	return screens.RecapInfo{
		TotalFocused: totalFocused,
		Segments:     segments,
		Breaks:       breaks,
		Pauses:       pauses,
		Streak:       streak,
		IsDeep:       isDeep,
		FocusScore:   focusScore,
	}
}

func countBreakSegments(segments []session.Segment) int {
	count := 0
	for _, seg := range segments {
		if seg.Kind == session.SegmentKindShortBreak || seg.Kind == session.SegmentKindLongBreak {
			count++
		}
	}
	return count
}

func (m *Model) cycleVerbLabel() {
	verbs := []string{"Focusing", "Building", "Fixing", "Debugging", "Reading", "Writing", "Learning", "Designing", "Researching"}
	idx := -1
	for i, v := range verbs {
		if v == m.currentVerbLabel {
			idx = i
			break
		}
	}
	nextIdx := (idx + 1) % len(verbs)
	m.currentVerbLabel = verbs[nextIdx]
}

func GetVerbForTask(task string) string {
	taskLower := strings.ToLower(task)
	if strings.Contains(taskLower, "build") {
		return "Building"
	}
	if strings.Contains(taskLower, "fix") {
		return "Fixing"
	}
	if strings.Contains(taskLower, "debug") {
		return "Debugging"
	}
	if strings.Contains(taskLower, "read") {
		return "Reading"
	}
	if strings.Contains(taskLower, "write") {
		return "Writing"
	}
	if strings.Contains(taskLower, "learn") {
		return "Learning"
	}
	if strings.Contains(taskLower, "design") {
		return "Designing"
	}
	if strings.Contains(taskLower, "research") {
		return "Researching"
	}
	return "Focusing"
}
