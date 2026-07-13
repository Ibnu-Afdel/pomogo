// Package config provides TOML-based configuration for PomoGo.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration.
type Config struct {
	// Durations (in minutes, converted to time.Duration internally)
	WorkDuration            int `toml:"work_duration"`
	ShortBreakDuration      int `toml:"short_break_duration"`
	LongBreakDuration       int `toml:"long_break_duration"`
	SessionsBeforeLongBreak int `toml:"sessions_before_long_break"`

	// Display
	Theme string `toml:"theme"`

	// Notifications
	NotificationsEnabled bool `toml:"notifications_enabled"`
	SoundEnabled         bool `toml:"sound_enabled"`
}

// Default returns a Config with sensible defaults (no file required).
func Default() *Config {
	return &Config{
		WorkDuration:            25,
		ShortBreakDuration:      5,
		LongBreakDuration:       15,
		SessionsBeforeLongBreak: 4,
		Theme:                   "tokyo-night",
		NotificationsEnabled:    true,
		SoundEnabled:            true,
	}
}

// XDGConfigDir returns the XDG config directory path for PomoGo.
func XDGConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configHome, "pomogo")
}

// ConfigFilePath returns the full path to the config file.
func ConfigFilePath() string {
	return filepath.Join(XDGConfigDir(), "config.toml")
}

// XDGDataDir returns the XDG data directory path for PomoGo.
func XDGDataDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(os.Getenv("HOME"), ".local/share")
	}
	return filepath.Join(dataHome, "pomogo")
}

// DBFilePath returns the path to the sqlite database.
func DBFilePath() string {
	return filepath.Join(XDGDataDir(), "pomogo.db")
}

// Load loads the configuration from the config file.
// If the file does not exist, returns the default configuration.
// If the file exists but is invalid, returns an error.
func Load() (*Config, error) {
	path := ConfigFilePath()

	// If file doesn't exist, return defaults
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse TOML
	cfg := Default() // Start with defaults, then override with file values
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	if c.WorkDuration <= 0 {
		return errors.New("work_duration must be positive")
	}
	if c.ShortBreakDuration <= 0 {
		return errors.New("short_break_duration must be positive")
	}
	if c.LongBreakDuration <= 0 {
		return errors.New("long_break_duration must be positive")
	}
	if c.SessionsBeforeLongBreak <= 0 {
		return errors.New("sessions_before_long_break must be positive")
	}
	if c.Theme == "" {
		return errors.New("theme must be set")
	}
	// Valid themes are checked elsewhere; we allow any string here for extensibility
	return nil
}

// WorkDurationAsDuration returns the work duration as time.Duration.
func (c *Config) WorkDurationAsDuration() time.Duration {
	return time.Duration(c.WorkDuration) * time.Minute
}

// ShortBreakDurationAsDuration returns the short break duration as time.Duration.
func (c *Config) ShortBreakDurationAsDuration() time.Duration {
	return time.Duration(c.ShortBreakDuration) * time.Minute
}

// LongBreakDurationAsDuration returns the long break duration as time.Duration.
func (c *Config) LongBreakDurationAsDuration() time.Duration {
	return time.Duration(c.LongBreakDuration) * time.Minute
}

// WriteDefault writes a commented default configuration file to the config directory.
func WriteDefault() error {
	dir := XDGConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	path := ConfigFilePath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists at %s", path)
	}

	// Create a default config with comments
	content := `# PomoGo Configuration
# This file configures the Pomodoro timer behavior.

# Work session duration in minutes (default: 25)
work_duration = 25

# Short break duration in minutes (default: 5)
short_break_duration = 5

# Long break duration in minutes (default: 15)
long_break_duration = 15

# Number of work sessions before a long break (default: 4)
sessions_before_long_break = 4

# Theme: "tokyo-night", "catppuccin", or "gruvbox" (default: tokyo-night)
theme = "tokyo-night"

# Enable notifications on session transitions (default: true)
notifications_enabled = true

# Enable sound on session transitions via canberra-gtk-play (default: true)
sound_enabled = true
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
