package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/restore"
	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleKeypress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.zenMode {
		m.zenMode = false
		return m, nil
	}

	if m.restorePending {
		return m.handleRestorePrompt(msg)
	}

	if m.showHelp {
		switch {
		case m.keymap.Help.Matches(msg.String()) || msg.String() == "esc":
			m.showHelp = false
		case m.keymap.Quit.Matches(msg.String()):
			m.persistOnQuit()
			return m, tea.Quit
		}
		return m, nil
	}

	// 1. Duration Picker Input Mode
	if m.inputMode == modeDurationPicker {
		switch msg.String() {
		case "up":
			m.selectedDurationIdx--
			if m.selectedDurationIdx < 0 {
				m.selectedDurationIdx = 4
			}
		case "down":
			m.selectedDurationIdx++
			if m.selectedDurationIdx > 4 {
				m.selectedDurationIdx = 0
			}
		case "esc":
			m.inputMode = modeNone
		case "enter":
			if m.selectedDurationIdx >= 0 && m.selectedDurationIdx <= 3 {
				m.deepDuration = time.Duration(m.selectedDurationIdx+1) * time.Hour
				block := session.NewDeepBlock(
					m.deepDuration,
					m.cfg.DeepFocusWorkDurationAsDuration(),
					m.cfg.DeepFocusShortBreakDurationAsDuration(),
					true,
				)
				m.runner = session.NewRunner(block)
				m.selectedMode = session.ModeDeep
				m.inputMode = modeNone
			} else if m.selectedDurationIdx == 4 {
				m.inputMode = modeCustomDurationInput
				m.textInput.SetValue("")
				m.textInput.Placeholder = "e.g., 1h30m, 90m..."
				m.textInput.Focus()
			}
		}
		return m, nil
	}

	// 2. Custom Duration Input Mode
	if m.inputMode == modeCustomDurationInput {
		switch msg.String() {
		case "esc":
			m.inputMode = modeDurationPicker
		case "enter":
			val := m.textInput.Value()
			d, err := parseCustomDuration(val)
			if err != nil {
				m.statusMessage = fmt.Sprintf("invalid duration: %v", err)
				m.inputMode = modeDurationPicker
				return m, nil
			}
			m.deepDuration = d
			block := session.NewDeepBlock(
				m.deepDuration,
				m.cfg.DeepFocusWorkDurationAsDuration(),
				m.cfg.DeepFocusShortBreakDurationAsDuration(),
				true,
			)
			m.runner = session.NewRunner(block)
			m.selectedMode = session.ModeDeep
			m.inputMode = modeNone
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	// Idle state mode switches
	if !m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
		switch msg.String() {
		case "q":
			block := session.NewQuickBlock(
				m.cfg.QuickFocusWorkDurationAsDuration(),
				m.cfg.QuickFocusShortBreakDurationAsDuration(),
				m.cfg.QuickFocusLongBreakDurationAsDuration(),
				m.cfg.QuickFocusSessionsBeforeLongBreak(),
				m.cfg.QuickFocusAutoAdvance(),
			)
			m.runner = session.NewRunner(block)
			m.selectedMode = session.ModeQuick
			m.statusMessage = "quick focus mode selected"
			return m, m.clearStatusAfter2s()
		case "d":
			m.inputMode = modeDurationPicker
			m.selectedDurationIdx = 0
			return m, nil
		}
	}

	switch {
	case m.keymap.Start.Matches(msg.String()):
		if !m.runner.Timer.IsRunning {
			if m.selectedMode == session.ModeDeep && m.dbStore != nil {
				bStore := &store.BlockStore{
					Mode:        "deep",
					PlannedSecs: int(m.deepDuration.Seconds()),
					StartedAt:   time.Now(),
					Completed:   false,
				}
				if err := m.dbStore.CreateBlock(bStore); err == nil {
					m.currentBlockID = &bStore.ID
				}
			}

			if err := m.runner.Start(timer.RealClock{}); err != nil {
				m.statusMessage = err.Error()
				return m, nil
			}
			m.lastTickTime = time.Now()
			m.afterTransition(true)
			return m, m.tick1s()
		}
	case m.keymap.PauseResume.Matches(msg.String()):
		if m.runner.Timer.IsRunning {
			if m.runner.Timer.IsPaused {
				if err := m.runner.Timer.Resume(timer.RealClock{}); err != nil {
					m.statusMessage = err.Error()
					return m, nil
				}
				m.lastTickTime = time.Now()
				m.afterTransition(false)
				return m, m.tick1s()
			} else {
				if err := m.runner.Timer.Pause(timer.RealClock{}); err != nil {
					m.statusMessage = err.Error()
					return m, nil
				}
				if m.runner.Block.Mode == session.ModeDeep && m.currentBlockID != nil && m.dbStore != nil {
					_ = m.dbStore.IncrementBlockPauses(*m.currentBlockID)
				}
				m.afterTransition(false)
			}
		}
	case m.keymap.Skip.Matches(msg.String()):
		if m.runner.Timer.IsRunning || m.runner.Timer.IsPaused {
			prevPhase := m.runner.Timer.Phase
			startedAt := m.runner.Timer.StartedAt
			duration := m.getDurationForPhase()
			_, blockEnded := m.runner.Skip(timer.RealClock{})

			m.recordSession(prevPhase, startedAt, time.Now(), false, duration)

			if blockEnded {
				m.finishBlock(false)
				if m.selectedMode == session.ModeDeep && m.cfg.PromptForNotes {
					sess := &store.Session{
						Type:         "work",
						Task:         m.currentTask,
						ProjectID:    m.currentProjectID,
						BlockID:      m.currentBlockID,
						Mode:         string(m.runner.Block.Mode),
						Completed:    false,
						StartedAt:    startedAt,
						EndedAt:      time.Now(),
						DurationSecs: int(duration.Seconds()),
					}
					m.pendingSession = sess
					m.inputMode = modeNoteInput
					m.textInput.SetValue("")
					m.textInput.Placeholder = "Enter block notes (Enter to skip)..."
					m.textInput.Focus()
				} else {
					m.inputMode = modeRecapScreen
				}
			} else if m.selectedMode == session.ModeQuick && prevPhase == timer.PhaseLongBreak {
				m.inputMode = modeRecapScreen
			}
			m.afterTransition(true)
		}
	case m.keymap.Task.Matches(msg.String()):
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.inputMode = modeTaskInput
			m.textInput.SetValue(m.currentTask)
			m.textInput.Placeholder = "Enter current task..."
			m.textInput.Focus()
			m.suggestions = nil
			if m.dbStore != nil {
				if list, err := m.dbStore.GetUniqueTasks(m.currentProjectID); err == nil {
					m.suggestions = list
				}
			}
			m.filteredSuggestions = m.suggestions
			m.suggestionIndex = -1
		}
	case m.keymap.Project.Matches(msg.String()):
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.inputMode = modeProjectInput
			m.textInput.SetValue(m.currentProjectName)
			m.textInput.Placeholder = "Enter project name..."
			m.textInput.Focus()
			m.suggestions = nil
			if m.dbStore != nil {
				if list, err := m.dbStore.GetProjects(); err == nil {
					var active []string
					for _, p := range list {
						if !p.Archived {
							active = append(active, p.Name)
						}
					}
					m.suggestions = active
				}
			}
			m.filteredSuggestions = m.suggestions
			m.suggestionIndex = -1
		}
	case m.keymap.ToggleStats.Matches(msg.String()):
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.showStats = !m.showStats
		}
	case m.keymap.CopyStats.Matches(msg.String()):
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			s := m.getStats()
			summary := fmt.Sprintf("PomoGo Stats - Today: %d sessions (%d mins) | Streak: %d days | Month: %d sessions",
				s.TodayCount, s.TodayMinutes, s.CurrentStreak, s.MonthCount)
			m.statusMessage = "Copied stats to clipboard!"
			return m, tea.Batch(copyOSC52(summary), m.clearStatusAfter2s())
		}
	case m.keymap.Reset.Matches(msg.String()):
		m.runner.Timer.Reset()
		m.finishBlock(false)
		m.removeState()
	case m.keymap.Quit.Matches(msg.String()):
		m.persistOnQuit()
		return m, tea.Quit
	case m.keymap.Help.Matches(msg.String()):
		m.showHelp = true
	case msg.String() == "T":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.cycleTheme()
			m.statusMessage = fmt.Sprintf("Theme: %s", m.currentThemeName)
			return m, m.clearStatusAfter2s()
		}
	case msg.String() == "L":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.cycleLayout()
			m.statusMessage = fmt.Sprintf("Layout: %s", m.currentLayoutName)
			return m, m.clearStatusAfter2s()
		}
	case msg.String() == "S":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.zenMode = !m.zenMode
			return m, nil
		}
	case msg.String() == "e":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.cycleEffects()
			m.statusMessage = fmt.Sprintf("Effects: %s", m.currentEffectsName)
			return m, m.clearStatusAfter2s()
		}
	case msg.String() == "v":
		if !m.showHelp && !m.restorePending && m.inputMode == modeNone {
			m.cycleVerbLabel()
			m.statusMessage = fmt.Sprintf("Activity: %s", m.currentVerbLabel)
			return m, m.clearStatusAfter2s()
		}
	}
	return m, nil
}

func (m *Model) handleRestorePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		runner, err := restore.RestoreRunnerWithDurations(restore.Durations{
			QuickWork:                    m.cfg.QuickFocusWorkDurationAsDuration(),
			QuickShortBreak:              m.cfg.QuickFocusShortBreakDurationAsDuration(),
			QuickLongBreak:               m.cfg.QuickFocusLongBreakDurationAsDuration(),
			QuickSessionsBeforeLongBreak: m.cfg.QuickFocusSessionsBeforeLongBreak(),
			QuickAutoAdvance:             m.cfg.QuickFocusAutoAdvance(),
			DeepWork:                     m.cfg.DeepFocusWorkDurationAsDuration(),
			DeepShortBreak:               m.cfg.DeepFocusShortBreakDurationAsDuration(),
		})
		if err != nil {
			m.restorePending = false
			m.statusMessage = fmt.Sprintf("restore failed: %v", err)
			return m, nil
		}
		if m.stateManager != nil {
			if st, err := m.stateManager.Read(); err == nil && st != nil {
				m.currentTask = st.Task
				m.currentVerbLabel = GetVerbForTask(st.Task)
				m.currentProjectID = st.ProjectID
				m.currentProjectName = st.ProjectName
				m.currentProjectIcon = ""
				if m.dbStore != nil && st.ProjectName != "" {
					if p, err := m.dbStore.GetProjectByName(st.ProjectName); err == nil && p != nil {
						m.currentProjectIcon = p.Icon
					}
				}
				if st.BlockID > 0 {
					id := st.BlockID
					m.currentBlockID = &id
				}
			}
		}
		m.runner = runner
		m.selectedMode = runner.Block.Mode
		m.restorePending = false
		m.afterTransition(false)
		if m.runner.Timer.IsPaused {
			return m, nil
		}
		return m, m.tick1s()
	case "n", "N", "esc":
		m.restorePending = false
		m.runner.Timer.State = timer.StateIdle
		m.runner.Timer.Phase = timer.PhaseWork
		m.runner.Timer.IsRunning = false
		m.runner.Timer.IsPaused = false
		m.writeState()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) finishBlock(completed bool) {
	if m.runner.Block.Mode == session.ModeDeep && m.currentBlockID != nil && m.dbStore != nil {
		_ = m.dbStore.FinishBlock(*m.currentBlockID, completed, time.Now())
	}
}

func parseCustomDuration(val string) (time.Duration, error) {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "" {
		return 0, fmt.Errorf("empty duration")
	}
	d, err := time.ParseDuration(val)
	if err == nil {
		return d, nil
	}
	var mins int
	if _, err := fmt.Sscanf(val, "%d", &mins); err == nil {
		return time.Duration(mins) * time.Minute, nil
	}
	return 0, fmt.Errorf("invalid duration format")
}

func (m *Model) filterSuggestions() {
	val := strings.ToLower(strings.TrimSpace(m.textInput.Value()))
	if val == "" {
		m.filteredSuggestions = m.suggestions
	} else {
		var filtered []string
		for _, s := range m.suggestions {
			if strings.Contains(strings.ToLower(s), val) {
				filtered = append(filtered, s)
			}
		}
		m.filteredSuggestions = filtered
	}

	if m.suggestionIndex >= len(m.filteredSuggestions) {
		m.suggestionIndex = len(m.filteredSuggestions) - 1
	}
	if m.suggestionIndex < -1 {
		m.suggestionIndex = -1
	}
}
