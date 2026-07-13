// Package notify provides notification and sound support for PomoGo.
package notify

import (
	"os/exec"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// soundEventComplete is the canberra event ID played when a work session ends (entering a break).
const soundEventComplete = "complete"

// soundEventBell is the canberra event ID played when a break ends (entering a work session).
const soundEventBell = "bell"

// Notifier sends system notifications and plays sounds for Pomodoro events.
type Notifier struct {
	enabled      bool
	soundEnabled bool
}

// NewNotifier creates a Notifier.
func NewNotifier(enabled, soundEnabled bool) *Notifier {
	return &Notifier{enabled: enabled, soundEnabled: soundEnabled}
}

// NotifyTransition sends a notification and plays a sound on a session transition.
// Sound and notifications are independent: each respects its own enabled flag.
func (n *Notifier) NotifyTransition(state timer.SessionState, phase timer.SessionPhase) error {
	n.playSound(state)
	if !n.enabled {
		return nil
	}
	title, message, urgency := n.messageForTransition(state, phase)
	return n.send(title, message, urgency)
}

func (n *Notifier) messageForTransition(state timer.SessionState, phase timer.SessionPhase) (title, message, urgency string) {
	switch state {
	case timer.StateWork:
		title = "PomoGo — Focus Time"
		message = "Break's over. Time to focus."
		urgency = "normal"
	case timer.StateShortBreak:
		title = "PomoGo — Short Break"
		message = "Session done. Stretch, hydrate, rest your eyes."
		urgency = "normal"
	case timer.StateLongBreak:
		title = "PomoGo — Long Break"
		message = "You've earned it. Take 15 minutes to recharge."
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
	return
}

// playSound plays a canberra event sound non-blocking. Silent if sound is disabled
// or canberra-gtk-play is unavailable.
func (n *Notifier) playSound(state timer.SessionState) {
	if !n.soundEnabled {
		return
	}
	var eventID string
	switch state {
	case timer.StateShortBreak, timer.StateLongBreak:
		eventID = soundEventComplete // work done → entering break
	case timer.StateWork:
		eventID = soundEventBell // break done → back to work
	default:
		return
	}

	path, err := exec.LookPath("canberra-gtk-play")
	if err != nil {
		return
	}
	cmd := exec.Command(path, "-i", eventID)
	cmd.Stdout = nil
	cmd.Stderr = nil
	go func() { _ = cmd.Run() }()
}

func (n *Notifier) send(title, message, urgency string) error {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return nil // silent no-op when notify-send is absent
	}
	cmd := exec.Command("notify-send", "-a", "pomogo", "-u", urgency, title, message)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
	return nil
}

// NotifyError sends a critical notification for an error condition.
func (n *Notifier) NotifyError(msg string) error {
	if !n.enabled {
		return nil
	}
	return n.send("PomoGo — Error", msg, "critical")
}

// NotifyCustom sends a custom notification with a validated urgency level.
func (n *Notifier) NotifyCustom(title, message, urgency string) error {
	if !n.enabled {
		return nil
	}
	if !validUrgency(urgency) {
		urgency = "normal"
	}
	return n.send(title, message, urgency)
}

// StateString returns a human-readable state name.
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

func validUrgency(urgency string) bool {
	return urgency == "low" || urgency == "normal" || urgency == "critical"
}
