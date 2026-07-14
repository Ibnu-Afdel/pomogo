// Package ui provides the Bubble Tea TUI for PomoGo.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/integrations"
	"github.com/Ibnu-Afdel/pomogo/internal/notify"
	"github.com/Ibnu-Afdel/pomogo/internal/restore"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/stats"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// bigDigits[d][row] is row `row` of digit `d`. Each row is exactly 4 visible chars.
var bigDigits = [10][5]string{
	{" ██ ", "█  █", "█  █", "█  █", " ██ "}, // 0
	{"  █ ", "  █ ", "  █ ", "  █ ", "  █ "}, // 1
	{" ██ ", "   █", " ██ ", "█   ", "████"}, // 2
	{" ██ ", "   █", " ██ ", "   █", " ██ "}, // 3
	{"█  █", "█  █", "████", "   █", "   █"}, // 4
	{"████", "█   ", "███ ", "   █", "███ "}, // 5
	{" ██ ", "█   ", "███ ", "█  █", " ██ "}, // 6
	{"████", "  █ ", " █  ", " █  ", " █  "}, // 7
	{" ██ ", "█  █", " ██ ", "█  █", " ██ "}, // 8
	{" ██ ", "█  █", " ███", "   █", " ██ "}, // 9
}

// bigColon[row] is row `row` of the ":" separator. Each row is exactly 2 visible chars.
var bigColon = [5]string{"  ", " ●", "  ", " ●", "  "}

type inputModeType int

const (
	modeNone inputModeType = iota
	modeTaskInput
	modeNoteInput
)

// Model is the main Bubble Tea model for the PomoGo TUI.
type Model struct {
	session *timer.Session
	cfg     *config.Config
	theme   *theme.Theme
	width   int
	height  int

	notifier       *notify.Notifier
	stateManager   *statefile.Manager
	dbStore        *store.Store
	restorePending bool
	showHelp       bool
	statusMessage  string

	textInput      textinput.Model
	inputMode      inputModeType
	currentTask    string
	pendingSession *store.Session
	showStats      bool
	lastTickTime   time.Time
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config) *Model {
	th := theme.Get(cfg.Theme)
	manager, _ := statefile.NewManager()
	st, err := store.New(config.DBFilePath())
	var statusMsg string
	if err != nil {
		statusMsg = fmt.Sprintf("database error: %v", err)
	}

	ti := textinput.New()
	ti.Width = 30

	return &Model{
		session: timer.NewSession(
			cfg.WorkDurationAsDuration(),
			cfg.ShortBreakDurationAsDuration(),
			cfg.LongBreakDurationAsDuration(),
			cfg.SessionsBeforeLongBreak,
		),
		cfg:            cfg,
		theme:          th,
		width:          80,
		height:         24,
		notifier:       notify.NewNotifier(cfg.NotificationsEnabled, cfg.SoundEnabled),
		stateManager:   manager,
		dbStore:        st,
		restorePending: restore.CanRestore(),
		statusMessage:  statusMsg,
		textInput:      ti,
		inputMode:      modeNone,
	}
}

// Init initializes the model (required by tea.Model).
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseCellMotion,
	)
}

// Update handles messages and updates the model (required by tea.Model).
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		now := time.Now()
		// 1. Suspend check
		if m.cfg.PauseOnSuspend && !m.lastTickTime.IsZero() && m.session.IsRunning && !m.session.IsPaused {
			elapsed := now.Sub(m.lastTickTime)
			if elapsed > 5*time.Second {
				m.session.Pause(timer.RealClock{})
				m.writeState()
				if m.notifier != nil {
					_ = m.notifier.NotifyCustom("PomoGo Resume", "Timer paused due to system suspend.", "normal")
				}
			}
		}
		m.lastTickTime = now

		// 2. Lock check
		if m.cfg.PauseOnLock && m.session.IsRunning && !m.session.IsPaused {
			if locked, err := integrations.IsSessionLocked(); err == nil && locked {
				m.session.Pause(timer.RealClock{})
				m.writeState()
				if m.notifier != nil {
					_ = m.notifier.NotifyCustom("PomoGo Locked", "Timer paused due to screen lock.", "normal")
				}
			}
		}

		// 3. Update terminal title
		var titleCmd tea.Cmd
		if m.cfg.TerminalTitleEnabled {
			titleCmd = m.updateTerminalTitle()
		}

		if m.session.IsRunning && !m.session.IsPaused {
			prevPhase := m.session.Phase
			startedAt := m.session.StartedAt
			duration := m.getDurationForPhase()

			if m.session.Tick(timer.RealClock{}) {
				m.recordSession(prevPhase, startedAt, time.Now(), true, duration)
				m.afterTransition(true)
				return m, titleCmd
			}
		}

		var tickCmd tea.Cmd
		if m.session.IsRunning && !m.session.IsPaused {
			tickCmd = m.tick1s()
		}
		return m, tea.Batch(titleCmd, tickCmd)
	}

	if m.inputMode != modeNone {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				if m.inputMode == modeNoteInput && m.pendingSession != nil {
					if err := m.dbStore.SaveSession(m.pendingSession); err != nil {
						m.statusMessage = fmt.Sprintf("database error: %v", err)
					}
					m.pendingSession = nil
				}
				m.inputMode = modeNone
				m.textInput.Blur()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.textInput.Value())
				if m.inputMode == modeTaskInput {
					m.currentTask = val
				} else if m.inputMode == modeNoteInput && m.pendingSession != nil {
					m.pendingSession.Note = val
					if err := m.dbStore.SaveSession(m.pendingSession); err != nil {
						m.statusMessage = fmt.Sprintf("database error: %v", err)
					}
					m.pendingSession = nil
				}
				m.inputMode = modeNone
				m.textInput.Blur()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeypress(msg)
	case tea.QuitMsg:
		return m, tea.Quit
	}
	return m, nil
}

// View renders the UI (required by tea.Model).
func (m *Model) View() string {
	if m.width < 50 || m.height < 16 {
		return "Terminal too small — minimum 50 × 16.\n"
	}

	if m.inputMode != modeNone {
		return m.renderInputScreen()
	}

	if m.showHelp {
		return m.renderHelpOverlay()
	}

	if m.restorePending {
		return m.renderRestorePrompt()
	}

	if m.showStats {
		return m.renderStatsScreen()
	}

	return m.renderTimerScreen()
}

func (m *Model) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.restorePending {
		return m.handleRestorePrompt(msg)
	}

	if m.showHelp {
		switch msg.String() {
		case "?", "esc":
			m.showHelp = false
		case "q", "ctrl+c":
			m.persistOnQuit()
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "s":
		if !m.session.IsRunning {
			if err := m.session.Start(timer.RealClock{}); err != nil {
				m.statusMessage = err.Error()
				return m, nil
			}
			m.lastTickTime = time.Now()
			m.afterTransition(true)
			return m, m.tick1s()
		}
	case " ": // space — pause / resume
		if m.session.IsRunning {
			if m.session.IsPaused {
				if err := m.session.Resume(timer.RealClock{}); err != nil {
					m.statusMessage = err.Error()
					return m, nil
				}
				m.lastTickTime = time.Now()
				m.afterTransition(false)
				return m, m.tick1s()
			} else {
				if err := m.session.Pause(timer.RealClock{}); err != nil {
					m.statusMessage = err.Error()
					return m, nil
				}
				m.afterTransition(false)
			}
		}
	case "n":
		if m.session.IsRunning || m.session.IsPaused {
			prevPhase := m.session.Phase
			startedAt := m.session.StartedAt
			duration := m.getDurationForPhase()
			m.session.Skip()
			m.recordSession(prevPhase, startedAt, time.Now(), false, duration)
			m.afterTransition(true)
		}
	case "t":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.inputMode = modeTaskInput
			m.textInput.SetValue(m.currentTask)
			m.textInput.Placeholder = "Enter current task..."
			m.textInput.Focus()
		}
	case "tab":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.showStats = !m.showStats
		}
	case "r":
		m.session.Reset()
		m.removeState()
	case "q", "ctrl+c":
		m.persistOnQuit()
		return m, tea.Quit
	case "?":
		m.showHelp = true
	}
	return m, nil
}

func (m *Model) handleRestorePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		session, _, err := restore.RestoreWithDurations(
			m.cfg.WorkDurationAsDuration(),
			m.cfg.ShortBreakDurationAsDuration(),
			m.cfg.LongBreakDurationAsDuration(),
			m.cfg.SessionsBeforeLongBreak,
		)
		if err != nil {
			m.restorePending = false
			m.statusMessage = fmt.Sprintf("restore failed: %v", err)
			return m, nil
		}
		if m.stateManager != nil {
			if st, err := m.stateManager.Read(); err == nil && st != nil {
				m.currentTask = st.Task
			}
		}
		m.session = session
		m.restorePending = false
		m.afterTransition(false)
		if m.session.IsPaused {
			return m, nil
		}
		return m, m.tick1s()
	case "n", "N", "esc":
		m.restorePending = false
		m.removeState()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) tick1s() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

type tickMsg struct{ time time.Time }

// renderTimerScreen renders the main timer display inside a themed border box.
func (m *Model) renderTimerScreen() string {
	color := lipgloss.Color(m.phaseColor().String())
	muted := lipgloss.Color(m.theme.Muted.String())

	// Text area width: total terminal width minus border (2), padding (2×4=8), outer margin (2)
	textW := m.width - 12
	if textW > 68 {
		textW = 68
	}
	if textW < 34 {
		textW = 34
	}

	center := func(s string) string {
		return lipgloss.NewStyle().Width(textW).Align(lipgloss.Center).Render(s)
	}

	// Time display
	remaining := m.displayRemaining()
	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60
	clockRows := bigClockRows(fmt.Sprintf("%02d:%02d", mins, secs), color)

	// Phase and status
	label := lipgloss.NewStyle().Foreground(color).Bold(true).Render(phaseLabel(m.session))
	completed := m.session.SessionCount % m.session.SessionsBeforeLongBreak
	dots := sessionDots(completed, m.session.SessionsBeforeLongBreak, color, muted)
	statusStr := m.sessionStatus()
	if m.statusMessage != "" {
		statusStr = m.statusMessage
	}
	status := lipgloss.NewStyle().Foreground(muted).Render(statusStr)

	// Progress bar
	total := m.getDurationForPhase()
	progress := 0.0
	if total > 0 {
		elapsed := total - remaining
		if elapsed < 0 {
			elapsed = 0
		}
		progress = float64(elapsed) / float64(total)
		if progress > 1 {
			progress = 1
		}
	}
	bar := progressBar(progress, textW, color, muted)

	// Hint line
	hints := lipgloss.NewStyle().Foreground(muted).Render(
		"s start  ·  space pause  ·  n skip  ·  t task  ·  r reset  ·  ? help  ·  q quit",
	)

	taskLine := ""
	if m.currentTask != "" {
		taskLine = lipgloss.NewStyle().Foreground(muted).Italic(true).Render("Task: " + m.currentTask)
	}

	var lines []string
	lines = append(lines, "")
	for _, row := range clockRows {
		lines = append(lines, center(row))
	}
	lines = append(lines, "")
	lines = append(lines, center(label))
	lines = append(lines, center(dots))
	if taskLine != "" {
		lines = append(lines, center(taskLine))
	}
	lines = append(lines, center(status))
	lines = append(lines, "")
	lines = append(lines, bar)
	lines = append(lines, "")
	lines = append(lines, center(hints))
	lines = append(lines, "")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderHelpOverlay() string {
	accent := lipgloss.Color(m.theme.Accent.String())
	muted := lipgloss.Color(m.theme.Muted.String())

	key := func(s string) string {
		return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(s)
	}
	dim := func(s string) string {
		return lipgloss.NewStyle().Foreground(muted).Render(s)
	}

	rows := []string{
		key("s") + "            " + "Start the queued session",
		key("space") + "        " + "Pause / resume",
		key("n") + "            " + "Skip to next phase",
		key("t") + "            " + "Set current task",
		key("tab") + "          " + "Toggle statistics view",
		key("r") + "            " + "Reset and clear state",
		key("q") + " / " + key("ctrl+c") + "   " + "Quit (state saved if running)",
		"",
		dim("Sessions auto-restore after an unexpected close."),
		dim("Press ? or Esc to close this overlay."),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 3).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderRestorePrompt() string {
	color := lipgloss.Color(m.phaseColor().String())
	muted := lipgloss.Color(m.theme.Muted.String())

	rows := []string{
		lipgloss.NewStyle().Foreground(color).Bold(true).Render("Restore previous session?"),
		"",
		lipgloss.NewStyle().Foreground(muted).Render("y resume  ·  n discard  ·  q quit"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) getDurationForPhase() time.Duration {
	switch m.session.Phase {
	case timer.PhaseWork:
		return m.session.WorkDuration
	case timer.PhaseShortBreak:
		return m.session.ShortBreakDuration
	case timer.PhaseLongBreak:
		return m.session.LongBreakDuration
	default:
		return m.session.WorkDuration
	}
}

func (m *Model) displayRemaining() time.Duration {
	if !m.session.IsRunning {
		if m.session.RemainingTime > 0 {
			return m.session.RemainingTime
		}
		return m.getDurationForPhase()
	}
	if m.session.IsPaused {
		return m.session.RemainingTime
	}
	remaining := time.Until(m.session.EndsAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (m *Model) phaseColor() theme.Color {
	if !m.session.IsRunning && m.session.State == timer.StateIdle {
		return m.theme.Idle
	}
	switch m.session.Phase {
	case timer.PhaseShortBreak:
		return m.theme.Break
	case timer.PhaseLongBreak:
		return m.theme.LongBreak
	default:
		return m.theme.Work
	}
}

func (m *Model) sessionStatus() string {
	switch {
	case m.session.IsPaused:
		return "paused"
	case m.session.IsRunning:
		return "running"
	default:
		return "press s to start"
	}
}

func (m *Model) afterTransition(sendNotification bool) {
	m.statusMessage = ""
	m.writeState()
	if sendNotification && m.notifier != nil {
		_ = m.notifier.NotifyTransition(m.session.State, m.session.Phase)
	}
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
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		Completed:    completed,
		DurationSecs: int(duration.Seconds()),
	}

	if m.cfg.PromptForNotes {
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
		return stats.Calculate(nil, now)
	}
	// Query sessions for the last 365 days
	start := now.AddDate(-1, 0, 0)
	sessions, err := m.dbStore.GetSessions(start, now.Add(24*time.Hour))
	if err != nil {
		return stats.Calculate(nil, now)
	}
	return stats.Calculate(sessions, now)
}

func weekBarGraph(weekDays [7]stats.DayStats, color, muted lipgloss.Color) string {
	maxVal := 0
	for _, wd := range weekDays {
		if wd.Count > maxVal {
			maxVal = wd.Count
		}
	}

	var sb strings.Builder
	for i, wd := range weekDays {
		if i > 0 {
			sb.WriteByte('\n')
		}
		
		weekdayName := wd.Date.Format("Mon")
		sb.WriteString(lipgloss.NewStyle().Foreground(muted).Render(weekdayName + "  "))

		barLen := wd.Count
		if barLen > 0 {
			bar := strings.Repeat("█", barLen)
			sb.WriteString(lipgloss.NewStyle().Foreground(color).Render(bar))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(muted).Render("░"))
		}
		
		sb.WriteString(fmt.Sprintf(" %d", wd.Count))
	}
	return sb.String()
}

func (m *Model) renderStatsScreen() string {
	accent := lipgloss.Color(m.theme.Accent.String())
	muted := lipgloss.Color(m.theme.Muted.String())
	color := lipgloss.Color(m.theme.Work.String())

	textW := m.width - 12
	if textW > 68 {
		textW = 68
	}
	if textW < 34 {
		textW = 34
	}

	center := func(s string) string {
		return lipgloss.NewStyle().Width(textW).Align(lipgloss.Center).Render(s)
	}

	now := time.Now()
	var sessions []*store.Session
	if m.dbStore != nil {
		start := now.AddDate(-1, 0, 0)
		if s, err := m.dbStore.GetSessions(start, now.Add(24*time.Hour)); err == nil {
			sessions = s
		}
	}
	s := stats.Calculate(sessions, now)

	// Today stats
	todayStr := fmt.Sprintf("Today: %d sessions (%d mins focused)", s.TodayCount, s.TodayMinutes)
	
	// Streak stats
	streakStr := fmt.Sprintf("Streak: %d days (Best: %d days)", s.CurrentStreak, s.BestStreak)
	
	// Month stats
	monthStr := fmt.Sprintf("This Month: %d sessions (Rate: %.0f%%)", s.MonthCount, s.CompletionRate*100)

	// Week graph title
	graphTitle := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("Weekly Focus Activity")
	graph := weekBarGraph(s.WeekDays, color, muted)

	// Recent sessions
	var recent []*store.Session
	for i := len(sessions) - 1; i >= 0 && len(recent) < 3; i-- {
		if sessions[i].Type == "work" {
			recent = append(recent, sessions[i])
		}
	}

	recentTitle := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("Recent Work Sessions")
	
	var recentLines []string
	if len(recent) == 0 {
		recentLines = append(recentLines, lipgloss.NewStyle().Foreground(muted).Render("No sessions recorded yet"))
	} else {
		for _, r := range recent {
			timeStr := r.StartedAt.Local().Format("15:04")
			taskStr := r.Task
			if taskStr == "" {
				taskStr = "[no task]"
			}
			
			status := "completed"
			statusColor := color
			if !r.Completed {
				status = "skipped"
				statusColor = muted
			}

			line := fmt.Sprintf("%s  %s (%s)", 
				lipgloss.NewStyle().Foreground(muted).Render(timeStr),
				lipgloss.NewStyle().Bold(true).Render(taskStr),
				lipgloss.NewStyle().Foreground(statusColor).Render(status),
			)
			if r.Note != "" {
				line += lipgloss.NewStyle().Foreground(muted).Render(" - " + r.Note)
			}
			recentLines = append(recentLines, line)
		}
	}

	// Hints
	hints := lipgloss.NewStyle().Foreground(muted).Render("Tab timer  ·  ? help  ·  q quit")

	var lines []string
	lines = append(lines, "")
	lines = append(lines, center(lipgloss.NewStyle().Foreground(color).Bold(true).Render("Focus Statistics")))
	lines = append(lines, "")
	lines = append(lines, center(todayStr))
	lines = append(lines, center(streakStr))
	lines = append(lines, center(monthStr))
	lines = append(lines, "")
	lines = append(lines, center(graphTitle))
	lines = append(lines, "")
	
	// Center the bar graph block
	graphLines := strings.Split(graph, "\n")
	maxLineLen := 0
	for _, l := range graphLines {
		length := lipgloss.Width(l)
		if length > maxLineLen {
			maxLineLen = length
		}
	}
	pad := (textW - maxLineLen) / 2
	if pad < 0 {
		pad = 0
	}
	var paddedGraphLines []string
	for _, l := range graphLines {
		paddedGraphLines = append(paddedGraphLines, strings.Repeat(" ", pad)+l)
	}
	lines = append(lines, strings.Join(paddedGraphLines, "\n"))
	
	lines = append(lines, "")
	lines = append(lines, center(recentTitle))
	lines = append(lines, "")

	// Center recent sessions block
	maxRecentLen := 0
	for _, l := range recentLines {
		length := lipgloss.Width(l)
		if length > maxRecentLen {
			maxRecentLen = length
		}
	}
	rPad := (textW - maxRecentLen) / 2
	if rPad < 0 {
		rPad = 0
	}
	var paddedRecentLines []string
	for _, l := range recentLines {
		paddedRecentLines = append(paddedRecentLines, strings.Repeat(" ", rPad)+l)
	}
	lines = append(lines, strings.Join(paddedRecentLines, "\n"))

	lines = append(lines, "")
	lines = append(lines, center(hints))
	lines = append(lines, "")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) renderInputScreen() string {
	color := lipgloss.Color(m.theme.Accent.String())
	title := "Set Current Task"
	if m.inputMode == modeNoteInput {
		color = lipgloss.Color(m.theme.Work.String())
		title = "Session Note"
	}

	rows := []string{
		lipgloss.NewStyle().Foreground(color).Bold(true).Render(title),
		"",
		m.textInput.View(),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Muted.String())).Render("enter save  ·  esc cancel"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Width(40).
		Align(lipgloss.Center).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m *Model) writeState() {
	if m.stateManager == nil {
		return
	}
	if err := m.stateManager.Write(m.session, m.currentTask); err != nil {
		m.statusMessage = fmt.Sprintf("state error: %v", err)
	}
}

func (m *Model) removeState() {
	if m.stateManager == nil {
		return
	}
	_ = m.stateManager.Remove()
}

func (m *Model) persistOnQuit() {
	if m.session.IsRunning {
		m.writeState()
		return
	}
	m.removeState()
}

// phaseLabel returns a human-readable name for the current phase.
func phaseLabel(s *timer.Session) string {
	if !s.IsRunning && !s.IsPaused && s.State == timer.StateIdle {
		return "Focus Timer"
	}
	switch s.Phase {
	case timer.PhaseWork:
		return "Work"
	case timer.PhaseShortBreak:
		return "Short Break"
	case timer.PhaseLongBreak:
		return "Long Break"
	default:
		return "Focus Timer"
	}
}

// sessionDots renders a dot-per-session progress indicator (● ● ○ ○).
func sessionDots(completed, total int, filled, empty lipgloss.Color) string {
	on := lipgloss.NewStyle().Foreground(filled)
	off := lipgloss.NewStyle().Foreground(empty)
	var sb strings.Builder
	for i := 0; i < total; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if i < completed {
			sb.WriteString(on.Render("●"))
		} else {
			sb.WriteString(off.Render("○"))
		}
	}
	return sb.String()
}

// progressBar renders a ━─ style bar of exactly `width` visible characters.
func progressBar(progress float64, width int, filled, empty lipgloss.Color) string {
	n := int(float64(width) * progress)
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	f := lipgloss.NewStyle().Foreground(filled).Render(strings.Repeat("━", n))
	e := lipgloss.NewStyle().Foreground(empty).Render(strings.Repeat("─", width-n))
	return f + e
}

// bigClockRows returns 5 rows of large ASCII-art digits for a "MM:SS" string,
// styled with the given color.
func bigClockRows(timeStr string, color lipgloss.Color) []string {
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	var rows [5]strings.Builder
	first := true
	for _, ch := range timeStr {
		if !first {
			for i := range rows {
				rows[i].WriteByte(' ')
			}
		}
		first = false
		switch {
		case ch >= '0' && ch <= '9':
			d := bigDigits[ch-'0']
			for i, seg := range d {
				rows[i].WriteString(seg)
			}
		case ch == ':':
			for i, seg := range bigColon {
				rows[i].WriteString(seg)
			}
		}
	}
	result := make([]string, 5)
	for i, sb := range rows {
		result[i] = style.Render(sb.String())
	}
	return result
}

func (m *Model) updateTerminalTitle() tea.Cmd {
	return func() tea.Msg {
		title := "Focus Timer"
		if m.session.IsRunning {
			mins := int(m.session.RemainingTime.Minutes())
			secs := int(m.session.RemainingTime.Seconds()) % 60
			stateStr := "work"
			emoji := "🍅"
			if m.session.Phase == timer.PhaseShortBreak || m.session.Phase == timer.PhaseLongBreak {
				stateStr = "break"
				emoji = "☕"
			}
			if m.session.IsPaused {
				title = fmt.Sprintf("⏸️ %02d:%02d · %s", mins, secs, stateStr)
			} else {
				title = fmt.Sprintf("%s %02d:%02d · %s", emoji, mins, secs, stateStr)
			}
		}
		fmt.Printf("\033]2;%s\007", title)
		return nil
	}
}
