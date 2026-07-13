// Package ui provides the Bubble Tea TUI for PomoGo.
package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// Model is the main Bubble Tea model for the PomoGo TUI.
type Model struct {
	session *timer.Session
	cfg     *config.Config
	theme   *theme.Theme
	width   int
	height  int
	tick    *time.Ticker
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config) *Model {
	th := theme.Get(cfg.Theme)

	return &Model{
		session: timer.NewSession(
			cfg.WorkDurationAsDuration(),
			cfg.ShortBreakDurationAsDuration(),
			cfg.LongBreakDurationAsDuration(),
			cfg.SessionsBeforeLongBreak,
		),
		cfg:    cfg,
		theme:  th,
		width:  80,
		height: 24,
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
	case tea.KeyMsg:
		return m.handleKeypress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		if m.session.IsRunning && !m.session.IsPaused {
			if m.session.Tick(timer.RealClock{}) {
				// Session ended
				return m, m.tick1s()
			}
		}
		if m.session.IsRunning && !m.session.IsPaused {
			return m, m.tick1s()
		}
	case tea.QuitMsg:
		return m, tea.Quit
	}
	return m, nil
}

// View renders the UI (required by tea.Model).
func (m *Model) View() string {
	if m.width < 40 || m.height < 12 {
		return "Terminal too small. Min: 40x12"
	}

	return m.renderTimerScreen()
}

// handleKeypress handles keyboard input.
func (m *Model) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		if !m.session.IsRunning {
			m.session.Start(timer.RealClock{})
			return m, m.tick1s()
		}
	case "space":
		if m.session.IsRunning {
			if m.session.IsPaused {
				m.session.Resume(timer.RealClock{})
				return m, m.tick1s()
			} else {
				m.session.Pause(timer.RealClock{})
			}
		}
	case "n":
		m.session.Skip()
	case "r":
		m.session.Reset()
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?":
		// Help overlay would go here (Phase 1.4+)
	}
	return m, nil
}

// tick1s returns a command that sends a tick message after 1 second.
func (m *Model) tick1s() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

// tickMsg is a message sent on each tick.
type tickMsg struct {
	time time.Time
}

// renderTimerScreen renders the main timer display.
func (m *Model) renderTimerScreen() string {
	// Format remaining time as MM:SS
	mins := int(m.session.RemainingTime.Minutes())
	secs := int(m.session.RemainingTime.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	// Phase display
	phaseStr := m.session.Phase.String()

	// Session count
	sessionCountStr := fmt.Sprintf("Session %d of %d", m.session.SessionCount+1, m.session.SessionsBeforeLongBreak)

	// Progress bar (simple: percentage of session complete)
	totalDuration := m.getDurationForPhase()
	progress := float64(0)
	if totalDuration > 0 {
		elapsedTime := totalDuration - m.session.RemainingTime
		progress = float64(elapsedTime) / float64(totalDuration)
		if progress > 1.0 {
			progress = 1.0
		}
		if progress < 0 {
			progress = 0
		}
	}
	progressWidth := int(float64(m.width-4) * progress)
	if progressWidth < 0 {
		progressWidth = 0
	}
	colorCode := fmt.Sprintf("%d", m.theme.Work.ANSI256())
	progressBar := "[" + lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode)).Render(
		repeatStr("█", progressWidth),
	) + repeatStr(" ", m.width-4-progressWidth) + "]"

	// Keybind hints
	hints := "s: start · space: pause · n: skip · r: reset · ?: help · q: quit"

	// Build the display
	lines := []string{
		"",
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color(colorCode)).Bold(true).Render(centerStr(timeStr, m.width)),
		"",
		centerStr(phaseStr, m.width),
		centerStr(sessionCountStr, m.width),
		"",
		progressBar,
		"",
		"",
		centerStr(hints, m.width),
		"",
	}

	return lipgloss.JoinVertical(lipgloss.Center, lines...)
}

// getDurationForPhase returns the duration for the current phase.
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

// centerStr centers a string within the given width.
func centerStr(s string, width int) string {
	if len(s) >= width {
		return s
	}
	padding := (width - len(s)) / 2
	return repeatStr(" ", padding) + s
}

// repeatStr repeats a string n times.
func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
