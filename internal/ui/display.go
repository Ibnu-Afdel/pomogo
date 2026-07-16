package ui

import (
	"fmt"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/render"
	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

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

// displayState builds the render.DisplayState snapshot from the current Model state.
func (m *Model) displayState() render.DisplayState {
	var label string
	var blockRemaining time.Duration
	var progress float64
	var segIndex int
	var segCount int

	isWorkPhase := false
	if m.runner.Block.Mode == session.ModeDeep {
		isWorkPhase = m.runner.Block.CurrentSegment.Kind == session.SegmentKindWork
	} else {
		isWorkPhase = m.runner.Timer.Phase == timer.PhaseWork
	}

	if isWorkPhase {
		label = m.currentVerbLabel
	} else if m.runner.Block.Mode == session.ModeDeep {
		label = "Break"
	} else {
		label = phaseLabel(m.runner.Timer)
	}

	if m.runner.Block.Mode == session.ModeDeep {
		blockRemaining = m.runner.Block.Remaining(m.runner.Timer.RemainingTime)
		progress = m.runner.Block.Progress(m.runner.Timer.RemainingTime)
		segIndex = 0
		segCount = 0 // Hides session dots in classic layout
	} else {
		blockRemaining = 0

		remaining := m.displayRemaining()
		total := m.getDurationForPhase()
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
		if m.runner.Timer.SessionsBeforeLongBreak > 0 {
			segIndex = m.runner.Timer.SessionCount % m.runner.Timer.SessionsBeforeLongBreak
			segCount = m.runner.Timer.SessionsBeforeLongBreak
		} else {
			segIndex = 0
			segCount = 4
		}
	}

	statusStr := m.sessionStatus()
	if m.runner.Timer.IsRunning && !m.runner.Timer.IsPaused {
		if m.runner.Block.Mode == session.ModeDeep {
			seg := m.runner.Block.CurrentSegment
			mins := int((m.runner.Timer.RemainingTime + 59*time.Second) / time.Minute)
			if mins < 1 {
				mins = 1
			}
			if seg.Kind == session.SegmentKindWork {
				hasBreak := false
				for i := m.runner.Block.Index + 1; i < len(m.runner.Block.Segments); i++ {
					if m.runner.Block.Segments[i].Kind != session.SegmentKindWork {
						hasBreak = true
						break
					}
				}
				if hasBreak {
					statusStr = fmt.Sprintf("focus · break in %dm", mins)
				} else {
					statusStr = "focus · final stretch"
				}
			} else {
				statusStr = fmt.Sprintf("break · focus in %dm", mins)
			}
		}
	}
	if m.statusMessage != "" {
		statusStr = m.statusMessage
	}

	projDisplay := m.currentProjectName
	if m.currentProjectIcon != "" && projDisplay != "" {
		projDisplay = m.currentProjectIcon + " " + projDisplay
	}

	return render.DisplayState{
		ModeLabel:        label,
		Project:          projDisplay,
		Task:             m.currentTask,
		PhaseKind:        m.runner.Timer.Phase,
		SegmentRemaining: m.runner.Timer.RemainingTime,
		BlockRemaining:   blockRemaining,
		Progress:         progress,
		SegmentIndex:     segIndex,
		SegmentCount:     segCount,
		Paused:           m.runner.Timer.IsPaused,
		Running:          m.runner.Timer.IsRunning,
		Idle:             !m.runner.Timer.IsRunning,
		StatusMessage:    statusStr,
		HintsVisibility:  true,
		ThemeName:        m.cfg.Theme,
		LayoutName:       "classic",
		GitBranch:        m.gitBranch,
		TmuxSession:      m.tmuxSession,
	}
}
