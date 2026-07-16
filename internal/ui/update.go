package ui

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/integrations"
	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{ time time.Time }
type dbusActionMsg string
type clearStatusMsg struct{}

// Update handles messages and updates the model (required by tea.Model).
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.tickCount++
		now := time.Now()
		// 1. Suspend check
		if m.cfg.PauseOnSuspend && !m.lastTickTime.IsZero() && m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
			elapsed := now.Sub(m.lastTickTime)
			if elapsed > 5*time.Second {
				m.runner.Timer.Pause(timer.RealClock{})
				m.writeState()
				if m.notifier != nil {
					_ = m.notifier.NotifyCustom("PomoGo Resume", "Timer paused due to system suspend.", "normal")
				}
			}
		}
		m.lastTickTime = now

		// 2. Lock check
		if m.cfg.PauseOnLock && m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
			if locked, err := integrations.IsSessionLocked(); err == nil && locked {
				m.runner.Timer.Pause(timer.RealClock{})
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

		if m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
			prevPhase := m.runner.Timer.Phase
			startedAt := m.runner.Timer.StartedAt
			duration := m.getDurationForPhase()

			evt, _ := m.runner.Tick(timer.RealClock{})
			if evt.Type != session.EventNone {
				m.recordSession(prevPhase, startedAt, time.Now(), true, duration)
				if evt.Type == session.EventBlockEnded {
					m.finishBlock(true)
					if m.selectedMode == session.ModeDeep && m.cfg.PromptForNotes {
						sess := &store.Session{
							Type:         "work",
							Task:         m.currentTask,
							ProjectID:    m.currentProjectID,
							BlockID:      m.currentBlockID,
							Mode:         string(m.runner.Block.Mode),
							Completed:    true,
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
				if m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
					return m, tea.Batch(titleCmd, m.tick1s())
				}
				return m, titleCmd
			}
		}

		return m, tea.Batch(titleCmd, m.tick1s())
	}

	if m.inputMode != modeNone {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.inputMode == modeRecapScreen && m.keymap.Quit.Matches(keyMsg.String()) {
				m.persistOnQuit()
				return m, tea.Quit
			}
			switch keyMsg.String() {
			case "esc":
				if m.inputMode == modeNoteInput && m.pendingSession != nil {
					if err := m.dbStore.SaveSession(m.pendingSession); err != nil {
						m.statusMessage = fmt.Sprintf("database error: %v", err)
					}
					m.pendingSession = nil
					m.inputMode = modeRecapScreen
					m.textInput.Blur()
					return m, nil
				}
				if m.inputMode == modeRecapScreen {
					m.inputMode = modeNone
					return m, nil
				}
				m.inputMode = modeNone
				m.textInput.Blur()
				return m, nil
			case "down":
				if m.inputMode == modeTaskInput || m.inputMode == modeProjectInput {
					if len(m.filteredSuggestions) > 0 {
						m.suggestionIndex++
						if m.suggestionIndex >= len(m.filteredSuggestions) {
							m.suggestionIndex = len(m.filteredSuggestions) - 1
						}
						return m, nil
					}
				}
			case "up":
				if m.inputMode == modeTaskInput || m.inputMode == modeProjectInput {
					m.suggestionIndex--
					if m.suggestionIndex < -1 {
						m.suggestionIndex = -1
					}
					return m, nil
				}
			case "tab":
				if (m.inputMode == modeTaskInput || m.inputMode == modeProjectInput) && m.suggestionIndex >= 0 {
					m.textInput.SetValue(m.filteredSuggestions[m.suggestionIndex])
					m.suggestionIndex = -1
					m.filterSuggestions()
					return m, nil
				}
			case "ctrl+d":
				if (m.inputMode == modeTaskInput || m.inputMode == modeProjectInput) && m.suggestionIndex >= 0 && m.suggestionIndex < len(m.filteredSuggestions) {
					item := m.filteredSuggestions[m.suggestionIndex]
					if m.inputMode == modeTaskInput {
						_ = m.dbStore.DeleteTaskName(item, m.currentProjectID)
						if list, err := m.dbStore.GetUniqueTasks(m.currentProjectID); err == nil {
							m.suggestions = list
						}
					} else {
						_ = m.dbStore.ArchiveProject(item)
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
					m.suggestionIndex = -1
					m.filterSuggestions()
					return m, nil
				}
			case "enter":
				if m.inputMode == modeRecapScreen {
					m.inputMode = modeNone
					m.runner.Timer.State = timer.StateIdle
					m.runner.Timer.IsRunning = false
					m.runner.Timer.IsPaused = false
					return m, nil
				}
				val := strings.TrimSpace(m.textInput.Value())
				if m.suggestionIndex >= 0 && m.suggestionIndex < len(m.filteredSuggestions) {
					val = m.filteredSuggestions[m.suggestionIndex]
				}

				if m.inputMode == modeTaskInput {
					m.currentTask = val
					m.currentVerbLabel = GetVerbForTask(val)
				} else if m.inputMode == modeNoteInput && m.pendingSession != nil {
					m.pendingSession.Note = val
					if err := m.dbStore.SaveSession(m.pendingSession); err != nil {
						m.statusMessage = fmt.Sprintf("database error: %v", err)
					}
					m.pendingSession = nil
					m.inputMode = modeRecapScreen
					m.textInput.Blur()
				} else if m.inputMode == modeProjectInput {
					if val == "" {
						m.currentProjectID = nil
						m.currentProjectName = ""
					} else {
						if m.dbStore != nil {
							p, err := m.dbStore.GetProjectByName(val)
							if err != nil {
								p = &store.Project{Name: val}
								if err := m.dbStore.CreateProject(p); err == nil {
									m.currentProjectID = &p.ID
									m.currentProjectName = p.Name
								} else {
									m.statusMessage = fmt.Sprintf("database error: %v", err)
								}
							} else {
								if p.Archived {
									_ = m.dbStore.UnarchiveProject(p.Name)
								}
								m.currentProjectID = &p.ID
								m.currentProjectName = p.Name
							}
						} else {
							m.currentProjectName = val
						}
					}
				}
				m.inputMode = modeNone
				m.textInput.Blur()
				m.writeState()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		m.filterSuggestions()
		return m, cmd
	}

	switch msg := msg.(type) {
	case dbusActionMsg:
		switch string(msg) {
		case "skip":
			if m.runner.Timer.IsRunning || m.runner.Timer.IsPaused {
				prevPhase := m.runner.Timer.Phase
				startedAt := m.runner.Timer.StartedAt
				duration := m.getDurationForPhase()
				evt, _ := m.runner.Skip(timer.RealClock{})
				m.recordSession(prevPhase, startedAt, time.Now(), false, duration)
				if evt.Type == session.EventBlockEnded {
					m.finishBlock(false)
				}
				m.afterTransition(true)
			}
		case "start_work":
			if !m.runner.Timer.IsRunning {
				_ = m.runner.Start(timer.RealClock{})
				m.afterTransition(true)
			}
		case "add_5":
			m.runner.Timer.AddTime(5 * time.Minute)
			m.writeState()
		}
		return m, m.listenForDbusActions()
	case clearStatusMsg:
		m.statusMessage = ""
		return m, nil
	case tea.KeyMsg:
		return m.handleKeypress(msg)
	case tea.QuitMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) tick1s() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{t}
	})
}

func (m *Model) clearStatusAfter2s() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) listenForDbusActions() tea.Cmd {
	if m.notifier == nil || m.notifier.DBusNotifier() == nil {
		return nil
	}
	return func() tea.Msg {
		action := <-m.notifier.DBusNotifier().ActionsChan()
		return dbusActionMsg(action)
	}
}

func (m *Model) triggerHooks() {
	if m.lastHookState == m.runner.Timer.State {
		return
	}

	// 1. End events
	if m.lastHookState == timer.StateWork {
		m.runHook(m.cfg.OnWorkEnd)
	} else if m.lastHookState == timer.StateShortBreak || m.lastHookState == timer.StateLongBreak {
		m.runHook(m.cfg.OnBreakEnd)
	}

	// 2. Start events
	if m.runner.Timer.State == timer.StateWork {
		m.runHook(m.cfg.OnWorkStart)
	} else if m.runner.Timer.State == timer.StateShortBreak || m.runner.Timer.State == timer.StateLongBreak {
		m.runHook(m.cfg.OnBreakStart)
	}

	m.lastHookState = m.runner.Timer.State
}

func (m *Model) runHook(cmdStr string) {
	if cmdStr == "" {
		return
	}

	go func() {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("POMOGO_STATE=%s", m.sessionStateString(m.runner.Timer.State)))
		cmd.Env = append(cmd.Env, fmt.Sprintf("POMOGO_PREV_STATE=%s", m.sessionStateString(m.lastHookState)))
		cmd.Env = append(cmd.Env, fmt.Sprintf("POMOGO_TASK=%s", m.currentTask))
		cmd.Env = append(cmd.Env, fmt.Sprintf("POMOGO_PROJECT=%s", m.currentProjectName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("POMOGO_DURATION=%d", int(m.getDurationForPhase().Seconds())))

		cmd.Stdout = nil
		cmd.Stderr = nil
		_ = cmd.Run()
	}()
}

func copyOSC52(text string) tea.Cmd {
	return func() tea.Msg {
		b64 := base64.StdEncoding.EncodeToString([]byte(text))
		seq := fmt.Sprintf("\033]52;c;%s\007", b64)
		if os.Getenv("TMUX") != "" {
			seq = fmt.Sprintf("\033Ptmux;\033%s\033\\", seq)
		}
		fmt.Print(seq)
		return nil
	}
}
