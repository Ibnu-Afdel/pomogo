// Package notify provides notification support via notify-send.
package notify

import (
	"os/exec"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// Notifier sends system notifications for Pomodoro events.
type Notifier struct {
	enabled bool
}

// NewNotifier creates a new Notifier with the given enabled state.
func NewNotifier(enabled bool) *Notifier {
	return &Notifier{enabled: enabled}
}

// NotifyTransition sends a notification on session state transition.
func (n *Notifier) NotifyTransition(state timer.SessionState, phase timer.SessionPhase) error {
	if !n.enabled {
		return nil
	}

	title, message, urgency := n.messageForTransition(state, phase)
	return n.send(title, message, urgency)
}

// messageForTransition returns the notification message for a state transition.
func (n *Notifier) messageForTransition(state timer.SessionState, phase timer.SessionPhase) (title, message, urgency string) {
	switch state {
	case timer.StateWork:
		title = "PomoGo — Work Session"
		message = "Time to focus! Dive into your task."
		urgency = "normal"
	case timer.StateShortBreak:
		title = "PomoGo — Short Break"
		message = "Take a quick 5-minute break. Stretch, hydrate, rest your eyes."
		urgency = "normal"
	case timer.StateLongBreak:
		title = "PomoGo — Long Break"
		message = "You've earned a longer break! Take 15 minutes to recharge."
		urgency = "normal"
	case timer.StateIdle:
		title = "PomoGo — Session Complete"
		message = "Great work! Ready for the next session?"
		urgency = "low"
	default:
		title = "PomoGo"
		message = "Session state changed."
		urgency = "low"
	}

	return title, message, urgency
}

// send sends a notification via notify-send.
func (n *Notifier) send(title, message, urgency string) error {
	// Check if notify-send exists
	_, err := exec.LookPath("notify-send")
	if err != nil {
		// notify-send not found — silent no-op per spec
		return nil
	}

	cmd := exec.Command(
		"notify-send",
		"-a", "pomogo",
		"-u", urgency,
		title,
		message,
	)

	// Suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Run and ignore errors (notification failure shouldn't crash the app)
	_ = cmd.Run()
	return nil
}

// NotifyError sends a notification for an error condition.
func (n *Notifier) NotifyError(msg string) error {
	if !n.enabled {
		return nil
	}

	return n.send("PomoGo — Error", msg, "critical")
}

// StateString returns a human-readable state name for notifications.
func StateString(state timer.SessionState) string {
	switch state {
	case timer.StateWork:
		return "Work"
	case timer.StateShortBreak:
		return "Short Break"
	case timer.StateLongBreak:
		return "Long Break"
	case timer.StateIdle:
		return "Idle"
	default:
		return "Unknown"
	}
}

// NotifyCustom sends a custom notification.
func (n *Notifier) NotifyCustom(title, message, urgency string) error {
	if !n.enabled {
		return nil
	}

	if !validUrgency(urgency) {
		urgency = "normal"
	}

	return n.send(title, message, urgency)
}

// validUrgency checks if the urgency level is valid for notify-send.
func validUrgency(urgency string) bool {
	return urgency == "low" || urgency == "normal" || urgency == "critical"
}
