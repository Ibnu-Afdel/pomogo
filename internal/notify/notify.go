// Package notify provides notification and sound support for PomoGo.
package notify

import (
	"os/exec"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

const (
	defaultSoundStartEvent = "message-new-instant"
	defaultSoundEndEvent   = "complete"
)

// SoundProfile is a paired set of canberra event IDs for focus transitions.
type SoundProfile struct {
	Name        string
	Description string
	StartEvent  string
	EndEvent    string
}

// SoundProfiles returns curated freedesktop/libcanberra event pairs.
func SoundProfiles() []SoundProfile {
	return []SoundProfile{
		{Name: "soft", Description: "Subtle focus cue and completion chime", StartEvent: "message-new-instant", EndEvent: "complete"},
		{Name: "bell", Description: "Classic terminal bell cues", StartEvent: "bell", EndEvent: "bell"},
		{Name: "dialog", Description: "Desktop notification style", StartEvent: "dialog-information", EndEvent: "complete"},
		{Name: "service", Description: "Login/logout style transitions", StartEvent: "service-login", EndEvent: "service-logout"},
	}
}

// Notifier sends system notifications and plays sounds for Pomodoro events.
type Notifier struct {
	enabled         bool
	soundEnabled    bool
	dbusNotifier    *DBusNotifier
	soundStartEvent string
	soundEndEvent   string
}

// NewNotifier creates a Notifier.
func NewNotifier(enabled, soundEnabled bool, soundEvents ...string) *Notifier {
	dn, _ := NewDBusNotifier() // Fallback is clean if D-Bus is unavailable
	startEvent := defaultSoundStartEvent
	endEvent := defaultSoundEndEvent
	if len(soundEvents) > 0 && soundEvents[0] != "" {
		startEvent = soundEvents[0]
	}
	if len(soundEvents) > 1 && soundEvents[1] != "" {
		endEvent = soundEvents[1]
	}
	return &Notifier{
		enabled:         enabled,
		soundEnabled:    soundEnabled,
		dbusNotifier:    dn,
		soundStartEvent: startEvent,
		soundEndEvent:   endEvent,
	}
}

// DBusNotifier returns the underlying DBusNotifier instance, if any.
func (n *Notifier) DBusNotifier() *DBusNotifier {
	return n.dbusNotifier
}

// SetCustomSoundEvent sets a custom transition sound event name.
func (n *Notifier) SetCustomSoundEvent(event string) {
	n.SetSoundEvents(event, event)
}

// SetSoundEvents sets focus-start and focus-end transition sound event names.
func (n *Notifier) SetSoundEvents(startEvent, endEvent string) {
	if startEvent != "" {
		n.soundStartEvent = startEvent
	}
	if endEvent != "" {
		n.soundEndEvent = endEvent
	}
}

// SoundEvents returns the configured focus-start and focus-end event IDs.
func (n *Notifier) SoundEvents() (string, string) {
	return n.soundStartEvent, n.soundEndEvent
}

// PreviewSoundEvent plays one event ID without changing the configured profile.
func (n *Notifier) PreviewSoundEvent(eventID string) {
	if eventID == "" {
		return
	}
	n.playEvent(eventID)
}

// NotifyTransition sends a notification and plays a sound on a session transition.
// Sound and notifications are independent: each respects its own enabled flag.
func (n *Notifier) NotifyTransition(state timer.SessionState, phase timer.SessionPhase) error {
	n.playSound(state)
	if !n.enabled {
		return nil
	}
	title, message, urgency := n.messageForTransition(state, phase)

	if n.dbusNotifier != nil {
		var actions []string
		switch state {
		case timer.StateWork:
			actions = []string{"skip", "Skip Work", "add_5", "+5 Min"}
		case timer.StateShortBreak, timer.StateLongBreak:
			actions = []string{"skip", "Skip Break", "add_5", "+5 Min"}
		case timer.StateIdle:
			actions = []string{"start_work", "Start Work"}
		}

		err := n.dbusNotifier.Send(title, message, urgency, actions)
		if err == nil {
			return nil
		}
		// Fallback to notify-send on error
	}

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
		eventID = n.soundEndEvent // work done → entering break
	case timer.StateWork:
		eventID = n.soundStartEvent // break done → back to work
	default:
		return
	}

	n.playEvent(eventID)
}

func (n *Notifier) playEvent(eventID string) {
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
	if n.dbusNotifier != nil {
		err := n.dbusNotifier.Send(title, message, urgency, nil)
		if err == nil {
			return nil
		}
	}
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
