package notify

import (
	"github.com/godbus/dbus/v5"
)

// DBusNotifier uses D-Bus Session Bus to send desktop notifications with custom buttons.
type DBusNotifier struct {
	conn        *dbus.Conn
	actionsChan chan string
	lastNotifID uint32
}

// NewDBusNotifier creates and initializes a DBusNotifier.
// It returns an error if the Session Bus is unavailable or if registering match rules fails.
func NewDBusNotifier() (*DBusNotifier, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	dn := &DBusNotifier{
		conn:        conn,
		actionsChan: make(chan string, 10),
	}

	// Register match rule to listen for ActionInvoked signals from the notification daemon
	rule := "type='signal',interface='org.freedesktop.Notifications',member='ActionInvoked'"
	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatchRule", 0, rule)
	if call.Err != nil {
		conn.Close()
		return nil, call.Err
	}

	// Listen for incoming signals in a background goroutine
	go dn.listenForSignals()

	return dn, nil
}

// ActionsChan returns the read-only channel where action keys are sent.
func (dn *DBusNotifier) ActionsChan() <-chan string {
	return dn.actionsChan
}

// Close closes the underlying D-Bus connection.
func (dn *DBusNotifier) Close() {
	if dn.conn != nil {
		dn.conn.Close()
	}
}

func (dn *DBusNotifier) listenForSignals() {
	c := make(chan *dbus.Signal, 10)
	dn.conn.Signal(c)

	for sig := range c {
		if sig.Name == "org.freedesktop.Notifications.ActionInvoked" {
			if len(sig.Body) >= 2 {
				id, ok1 := sig.Body[0].(uint32)
				actionKey, ok2 := sig.Body[1].(string)
				if ok1 && ok2 && id == dn.lastNotifID {
					dn.actionsChan <- actionKey
				}
			}
		}
	}
}

// Send sends a desktop notification with action buttons.
func (dn *DBusNotifier) Send(title, message, urgency string, actions []string) error {
	obj := dn.conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	hints := map[string]dbus.Variant{
		"urgency": dbus.MakeVariant(urgencyLevel(urgency)),
	}

	var notifID uint32
	call := obj.Call("org.freedesktop.Notifications.Notify", 0,
		"pomogo",    // app_name
		uint32(0),   // replaces_id
		"",          // app_icon
		title,       // summary
		message,     // body
		actions,     // actions (format: ["action_key", "Button Label", ...])
		hints,       // hints
		int32(-1),   // expire_timeout (-1 for default daemon timeout)
	)
	if call.Err != nil {
		return call.Err
	}

	err := call.Store(&notifID)
	if err == nil {
		dn.lastNotifID = notifID
	}
	return err
}

func urgencyLevel(urgency string) byte {
	switch urgency {
	case "low":
		return 0
	case "normal":
		return 1
	case "critical":
		return 2
	default:
		return 1
	}
}
