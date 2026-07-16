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
	Theme   string `toml:"theme"`
	Layout  string `toml:"layout"`
	Effects string `toml:"effects"`

	// Notifications
	NotificationsEnabled bool `toml:"notifications_enabled"`
	SoundEnabled         bool `toml:"sound_enabled"`

	// Notes
	PromptForNotes bool `toml:"prompt_for_notes"`

	// Lock & Suspend
	PauseOnLock          bool `toml:"pause_on_lock"`
	PauseOnSuspend       bool `toml:"pause_on_suspend"`
	TerminalTitleEnabled bool `toml:"terminal_title_enabled"`
	ShowGit              bool `toml:"show_git"`
	ShowTmux             bool `toml:"show_tmux"`

	// Profiles
	Profiles map[string]Profile `toml:"profiles"`

	// Mode configurations
	QuickFocus QuickFocusConfig `toml:"quick_focus"`
	DeepFocus  DeepFocusConfig  `toml:"deep_focus"`

	// Hooks
	OnWorkStart  string `toml:"on_work_start"`
	OnWorkEnd    string `toml:"on_work_end"`
	OnBreakStart string `toml:"on_break_start"`
	OnBreakEnd   string `toml:"on_break_end"`
}

type QuickFocusConfig struct {
	WorkDuration            *int  `toml:"work_duration"`
	ShortBreakDuration      *int  `toml:"short_break_duration"`
	LongBreakDuration       *int  `toml:"long_break_duration"`
	SessionsBeforeLongBreak *int  `toml:"sessions_before_long_break"`
	AutoAdvance             *bool `toml:"auto_advance"`
}

type DeepFocusConfig struct {
	WorkDuration       *int `toml:"work_duration"`
	ShortBreakDuration *int `toml:"short_break_duration"`
	DefaultDuration    *int `toml:"default_duration"`
}

// Profile represents a set of overrides for custom focus states.
type Profile struct {
	WorkDuration            *int    `toml:"work_duration"`
	ShortBreakDuration      *int    `toml:"short_break_duration"`
	LongBreakDuration       *int    `toml:"long_break_duration"`
	SessionsBeforeLongBreak *int    `toml:"sessions_before_long_break"`
	Theme                   *string `toml:"theme"`
	Layout                  *string `toml:"layout"`
	Project                 *string `toml:"project"`
	SoundEvent              *string `toml:"sound_event"`
}

// Default returns a Config with sensible defaults (no file required).
func Default() *Config {
	return &Config{
		WorkDuration:            25,
		ShortBreakDuration:      5,
		LongBreakDuration:       15,
		SessionsBeforeLongBreak: 4,
		Theme:                   "tokyo-night",
		Layout:                  "classic",
		Effects:                 "none",
		NotificationsEnabled:    true,
		SoundEnabled:            true,
		PromptForNotes:          true,
		PauseOnLock:             true,
		PauseOnSuspend:          true,
		TerminalTitleEnabled:    true,
		ShowGit:                 true,
		ShowTmux:                false,
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

	validLayouts := map[string]bool{
		"classic": true, "minimal": true, "centered": true,
		"compact": true, "retro": true, "random": true, "daily": true,
		"": true,
	}
	if !validLayouts[c.Layout] {
		return fmt.Errorf("unknown layout: %q", c.Layout)
	}

	validEffects := map[string]bool{
		"none": true, "stars": true, "snow": true, "rain": true, "random": true, "": true,
	}
	if !validEffects[c.Effects] {
		return fmt.Errorf("unknown effects: %q", c.Effects)
	}

	if c.QuickFocus.WorkDuration != nil && *c.QuickFocus.WorkDuration <= 0 {
		return errors.New("quick_focus.work_duration must be positive")
	}
	if c.QuickFocus.ShortBreakDuration != nil && *c.QuickFocus.ShortBreakDuration <= 0 {
		return errors.New("quick_focus.short_break_duration must be positive")
	}
	if c.QuickFocus.LongBreakDuration != nil && *c.QuickFocus.LongBreakDuration <= 0 {
		return errors.New("quick_focus.long_break_duration must be positive")
	}
	if c.QuickFocus.SessionsBeforeLongBreak != nil && *c.QuickFocus.SessionsBeforeLongBreak <= 0 {
		return errors.New("quick_focus.sessions_before_long_break must be positive")
	}
	if c.DeepFocus.WorkDuration != nil && *c.DeepFocus.WorkDuration <= 0 {
		return errors.New("deep_focus.work_duration must be positive")
	}
	if c.DeepFocus.ShortBreakDuration != nil && *c.DeepFocus.ShortBreakDuration <= 0 {
		return errors.New("deep_focus.short_break_duration must be positive")
	}
	if c.DeepFocus.DefaultDuration != nil && *c.DeepFocus.DefaultDuration <= 0 {
		return errors.New("deep_focus.default_duration must be positive")
	}

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

# Theme: "tokyo-night", "catppuccin", "gruvbox", "rose-pine", etc. (default: tokyo-night)
theme = "tokyo-night"

# Layout: "classic", "minimal", "centered", "compact", "retro" (default: classic)
layout = "classic"

# Effects: "none", "stars", "snow", "rain", "random" (default: none)
effects = "none"

[quick_focus]
# Continue automatically from focus to break and back again (default: true)
auto_advance = true

[deep_focus]
# Default Deep Focus block duration in minutes (default: 60)
default_duration = 60
# Internal segment durations for long blocks
work_duration = 25
short_break_duration = 5


# Enable notifications on session transitions (default: true)
notifications_enabled = true

# Enable sound on session transitions via canberra-gtk-play (default: true)
sound_enabled = true

# Prompt for a brief note when completing a work session (default: true)
prompt_for_notes = true

# Pause the timer automatically when the system locks (default: true)
pause_on_lock = true

# Pause the timer automatically when the system suspends (default: true)
pause_on_suspend = true

# Show the timer countdown in the terminal window title (default: true)
terminal_title_enabled = true

# Show current Git branch in layouts if repo is found (default: true)
show_git = true

# Show current tmux session in layouts if inside TMUX (default: false)
show_tmux = false
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResolveProfile resolves config settings by applying profile overrides.
// It returns a copy of the Config with overridden settings, a default project name, and a custom sound event.
func (c *Config) ResolveProfile(name string) (*Config, string, string) {
	resolved := *c
	project := ""
	soundEvent := ""

	if c.Profiles == nil {
		return &resolved, project, soundEvent
	}

	p, exists := c.Profiles[name]
	if !exists {
		return &resolved, project, soundEvent
	}

	if p.WorkDuration != nil {
		resolved.WorkDuration = *p.WorkDuration
	}
	if p.ShortBreakDuration != nil {
		resolved.ShortBreakDuration = *p.ShortBreakDuration
	}
	if p.LongBreakDuration != nil {
		resolved.LongBreakDuration = *p.LongBreakDuration
	}
	if p.SessionsBeforeLongBreak != nil {
		resolved.SessionsBeforeLongBreak = *p.SessionsBeforeLongBreak
	}
	if p.Theme != nil {
		resolved.Theme = *p.Theme
	}
	if p.Project != nil {
		project = *p.Project
	}
	if p.SoundEvent != nil {
		soundEvent = *p.SoundEvent
	}

	return &resolved, project, soundEvent
}

func (c *Config) QuickFocusWorkDurationAsDuration() time.Duration {
	if c.QuickFocus.WorkDuration != nil {
		return time.Duration(*c.QuickFocus.WorkDuration) * time.Minute
	}
	return c.WorkDurationAsDuration()
}

func (c *Config) QuickFocusShortBreakDurationAsDuration() time.Duration {
	if c.QuickFocus.ShortBreakDuration != nil {
		return time.Duration(*c.QuickFocus.ShortBreakDuration) * time.Minute
	}
	return c.ShortBreakDurationAsDuration()
}

func (c *Config) QuickFocusLongBreakDurationAsDuration() time.Duration {
	if c.QuickFocus.LongBreakDuration != nil {
		return time.Duration(*c.QuickFocus.LongBreakDuration) * time.Minute
	}
	return c.LongBreakDurationAsDuration()
}

func (c *Config) QuickFocusSessionsBeforeLongBreak() int {
	if c.QuickFocus.SessionsBeforeLongBreak != nil {
		return *c.QuickFocus.SessionsBeforeLongBreak
	}
	return c.SessionsBeforeLongBreak
}

func (c *Config) QuickFocusAutoAdvance() bool {
	if c.QuickFocus.AutoAdvance != nil {
		return *c.QuickFocus.AutoAdvance
	}
	return true
}

func (c *Config) DeepFocusWorkDurationAsDuration() time.Duration {
	if c.DeepFocus.WorkDuration != nil {
		return time.Duration(*c.DeepFocus.WorkDuration) * time.Minute
	}
	return c.WorkDurationAsDuration()
}

func (c *Config) DeepFocusShortBreakDurationAsDuration() time.Duration {
	if c.DeepFocus.ShortBreakDuration != nil {
		return time.Duration(*c.DeepFocus.ShortBreakDuration) * time.Minute
	}
	return c.ShortBreakDurationAsDuration()
}

func (c *Config) DeepFocusDefaultDurationAsDuration() time.Duration {
	if c.DeepFocus.DefaultDuration != nil {
		return time.Duration(*c.DeepFocus.DefaultDuration) * time.Minute
	}
	return time.Hour
}
