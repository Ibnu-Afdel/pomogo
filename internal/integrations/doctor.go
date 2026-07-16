package integrations

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/godbus/dbus/v5"
)

// Diagnostic represents the status of a system health check.
type Diagnostic struct {
	Name    string
	Passed  bool
	Message string
}

// RunDoctor Checks system configurations, files, permissions, and dependencies.
func RunDoctor() []Diagnostic {
	var diagnostics []Diagnostic

	// 1. Config Check
	cfgDiag := Diagnostic{Name: "Configuration File Valid", Passed: true, Message: "Valid config.toml"}
	cfgPath := config.ConfigFilePath()
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgDiag.Passed = false
		cfgDiag.Message = fmt.Sprintf("Config file does not exist at %s (use 'pomogo config init' to create it)", cfgPath)
	} else {
		_, err := config.Load()
		if err != nil {
			cfgDiag.Passed = false
			cfgDiag.Message = fmt.Sprintf("Error loading config: %v", err)
		}
	}
	diagnostics = append(diagnostics, cfgDiag)

	// 2. Database Check
	dbDiag := Diagnostic{Name: "SQLite Database Health", Passed: true, Message: "Database is readable and writable"}
	dbPath := config.DBFilePath()
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		dbDiag.Passed = false
		dbDiag.Message = fmt.Sprintf("Database file does not exist at %s (it will be created on next startup)", dbPath)
	} else {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			dbDiag.Passed = false
			dbDiag.Message = fmt.Sprintf("Failed to open SQLite: %v", err)
		} else {
			defer db.Close()
			err = db.Ping()
			if err != nil {
				dbDiag.Passed = false
				dbDiag.Message = fmt.Sprintf("SQLite ping failed: %v", err)
			} else {
				// Check write permission
				_, err = db.Exec("CREATE TABLE IF NOT EXISTS _doctor_temp (id INTEGER PRIMARY KEY); DROP TABLE _doctor_temp;")
				if err != nil {
					dbDiag.Passed = false
					dbDiag.Message = fmt.Sprintf("Database is read-only or locked: %v", err)
				}
			}
		}
	}
	diagnostics = append(diagnostics, dbDiag)

	// 3. State File check
	stateDiag := Diagnostic{Name: "Runtime State Directory", Passed: true, Message: "Runtime state directory is writable"}
	mgr, err := statefile.NewManager()
	if err != nil {
		stateDiag.Passed = false
		stateDiag.Message = fmt.Sprintf("Failed to initialize state manager: %v", err)
	} else {
		stateDir := filepath.Dir(mgr.StatePath())
		if err := os.MkdirAll(stateDir, 0700); err != nil {
			stateDiag.Passed = false
			stateDiag.Message = fmt.Sprintf("State directory %s not writable: %v", stateDir, err)
		}
	}
	diagnostics = append(diagnostics, stateDiag)

	// 4. D-Bus Session Bus
	sessBusDiag := Diagnostic{Name: "D-Bus Session Bus (Notifications)", Passed: true, Message: "Connected successfully"}
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		sessBusDiag.Passed = false
		sessBusDiag.Message = fmt.Sprintf("Cannot connect to Session Bus (actionable notifications disabled): %v", err)
	} else {
		conn.Close()
	}
	diagnostics = append(diagnostics, sessBusDiag)

	// 5. D-Bus System Bus
	sysBusDiag := Diagnostic{Name: "D-Bus System Bus (Lock Detection)", Passed: true, Message: "Connected successfully"}
	sConn, err := dbus.ConnectSystemBus()
	if err != nil {
		sysBusDiag.Passed = false
		sysBusDiag.Message = fmt.Sprintf("Cannot connect to System Bus (suspend/lock detection may be limited): %v", err)
	} else {
		sConn.Close()
	}
	diagnostics = append(diagnostics, sysBusDiag)

	// 6. Notify-send Check
	nsDiag := Diagnostic{Name: "Notification Utility (notify-send)", Passed: true, Message: "notify-send is installed"}
	if _, err := exec.LookPath("notify-send"); err != nil {
		nsDiag.Passed = false
		nsDiag.Message = "notify-send command not found in PATH (system notifications will be disabled)"
	}
	diagnostics = append(diagnostics, nsDiag)

	// 7. Canberra Check
	soundDiag := Diagnostic{Name: "Canberra Sound Player (canberra-gtk-play)", Passed: true, Message: "canberra-gtk-play is installed"}
	if _, err := exec.LookPath("canberra-gtk-play"); err != nil {
		soundDiag.Passed = false
		soundDiag.Message = "canberra-gtk-play not found in PATH (transition sounds will be disabled)"
	}
	diagnostics = append(diagnostics, soundDiag)

	// 8. External Themes Check
	themeDiag := Diagnostic{Name: "External Themes", Passed: true, Message: "All external themes loaded successfully"}
	if malformed := theme.CheckExternalThemes(); len(malformed) > 0 {
		themeDiag.Passed = false
		themeDiag.Message = fmt.Sprintf("Malformed theme files found: %s", strings.Join(malformed, ", "))
	} else if lowContrast := theme.CheckExternalThemeContrast(); len(lowContrast) > 0 {
		themeDiag.Passed = false
		themeDiag.Message = fmt.Sprintf("Low-contrast theme files found: %s", strings.Join(lowContrast, ", "))
	}
	diagnostics = append(diagnostics, themeDiag)

	return diagnostics
}
