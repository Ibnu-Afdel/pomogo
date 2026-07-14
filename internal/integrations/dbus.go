// Package integrations provides system-level service integrations.
package integrations

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
)

// IsSessionLocked queries systemd-logind via D-Bus to check if the user session is currently locked.
// It resolves the session path using the current process PID.
func IsSessionLocked() (bool, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return false, fmt.Errorf("failed to connect to system bus: %w", err)
	}
	defer conn.Close()

	// Get session path by current PID
	var sessionPath dbus.ObjectPath
	obj := conn.Object("org.freedesktop.login1", "/org/freedesktop/login1")
	err = obj.Call("org.freedesktop.login1.Manager.GetSessionByPID", 0, uint32(os.Getpid())).Store(&sessionPath)
	if err != nil {
		return false, fmt.Errorf("failed to get session path: %w", err)
	}

	// Read LockedHint property
	sessionObj := conn.Object("org.freedesktop.login1", sessionPath)
	variant, err := sessionObj.GetProperty("org.freedesktop.login1.Session.LockedHint")
	if err != nil {
		return false, fmt.Errorf("failed to get LockedHint property: %w", err)
	}

	locked, ok := variant.Value().(bool)
	if !ok {
		return false, fmt.Errorf("LockedHint property is not a boolean")
	}

	return locked, nil
}
