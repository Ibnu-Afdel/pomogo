// Package integrations provides status bar formatters for external integrations.
package integrations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
)

// WaybarOutput represents the JSON schema expected by Waybar's custom module.
type WaybarOutput struct {
	Text    string `json:"text"`
	Class   string `json:"class"`
	Tooltip string `json:"tooltip"`
}

// FormatStatus formats the session state according to the requested format.
func FormatStatus(state *statefile.State, format string) (string, error) {
	if isIdle(state) {
		switch format {
		case "waybar":
			out := WaybarOutput{
				Text:    "",
				Class:   "pomogo-idle",
				Tooltip: "Idle",
			}
			data, err := json.Marshal(out)
			if err != nil {
				return "", err
			}
			return string(data), nil
		case "tmux":
			return "", nil
		case "json":
			return `{"state":"idle"}`, nil
		default:
			return "Idle", nil
		}
	}

	// Active session format
	var timeStr string
	if state.Mode == "deep" && state.BlockRemainingSecs > 0 {
		hrs := state.BlockRemainingSecs / 3600
		mins := (state.BlockRemainingSecs % 3600) / 60
		secs := state.BlockRemainingSecs % 60
		if hrs > 0 {
			timeStr = fmt.Sprintf("%d:%02d:%02d", hrs, mins, secs)
		} else {
			timeStr = fmt.Sprintf("%02d:%02d", mins, secs)
		}
	} else {
		mins := state.RemainingSecs / 60
		secs := state.RemainingSecs % 60
		timeStr = fmt.Sprintf("%02d:%02d", mins, secs)
	}

	// Determine icon and class
	var icon string
	var class string
	if state.Paused {
		icon = "⏸️"
		class = "pomogo-paused"
	} else if state.SessionType == "work" {
		icon = "🍅"
		class = "pomogo-work"
	} else {
		icon = "☕"
		class = "pomogo-break"
	}

	displayStr := fmt.Sprintf("%s %s", icon, timeStr)
	endTime := time.Unix(state.EndsAt, 0).Local().Format("15:04")
	phaseName := state.SessionType
	if phaseName == "short_break" {
		phaseName = "break"
	} else if phaseName == "long_break" {
		phaseName = "long break"
	}

	tooltip := fmt.Sprintf("%s session ends at %s", capitalize(phaseName), endTime)
	if state.Task != "" {
		tooltip = fmt.Sprintf("Task: %s\n%s", state.Task, tooltip)
	}

	switch format {
	case "waybar":
		out := WaybarOutput{
			Text:    displayStr,
			Class:   class,
			Tooltip: tooltip,
		}
		data, err := json.Marshal(out)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "tmux":
		return displayStr, nil
	case "json":
		data, err := json.Marshal(state)
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		// Default human-readable status line
		res := fmt.Sprintf("%s · %s", displayStr, state.SessionType)
		if state.Paused {
			res += " (paused)"
		}
		if state.Task != "" {
			res += fmt.Sprintf(" [%s]", state.Task)
		}
		return res, nil
	}
}

func isIdle(state *statefile.State) bool {
	if state == nil {
		return true
	}
	if state.SessionState == "idle" {
		return true
	}
	if statefile.IsStale(state) {
		return true
	}
	if statefile.IsExpired(state) {
		return true
	}
	return false
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
