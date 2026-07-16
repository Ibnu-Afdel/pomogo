# PRD: PomoGo — Product Requirements Document

## Executive Summary

PomoGo is a keyboard-first, native Pomodoro timer for Linux developers. It ships as a single static binary, runs no daemon, integrates with existing Linux tooling (tmux, Waybar, Hyprland, Ghostty), and tracks session history for focus insights.

**MVP Target (Phase 1):** A beautiful timer with notifications, restore on crash, zero required configuration.

## User Personas

### Primary: Terminal-First Developer on Omarchy/Arch Linux
- Uses Neovim, tmux, TUI apps daily
- Wayland/Hyprland setup
- Keyboard-first workflows (no mouse)
- Wants focus tracking without distraction

### Secondary: Hardcore Focus Worker
- Needs integration with existing status bars and multiplexers
- Values statistics and streaks
- Uses custom shell prompts and status lines
- Wants plugins for Neovim, shell hooks, etc.

## Core Features

### Phase 1: MVP Timer
- **Start/Pause/Resume/Skip** via keyboard (s, space, n, r)
- **Centered, minimal UI** with large timer display
- **Notifications** on session transitions (via notify-send)
- **Config file** (TOML, `~/.config/pomogo/config.toml`)
- **Themes** (Tokyo Night default, Catppuccin, Gruvbox)
- **Restore on crash** — re-launch to resume mid-session
- **State file** (`$XDG_RUNTIME_DIR/pomogo/state.json`) for integrations

### Phase 2: Productivity
- **SQLite history** (`~/.local/share/pomogo/pomogo.db`)
- **Current task** labeling (set with `t` key)
- **Session notes** (optional prompt on session end)
- **Stats screen** (`Tab` in TUI, or `pomogo stats --week`)
- **Daily/weekly/monthly summaries** with completion rates and streaks

### Phase 3: Omarchy Integration
- **Waybar module** (`interval: 1`, `pomogo status --format waybar`)
- **tmux integration** (status-right snippet using `pomogo status --format tmux`)
- **Lock detection** (pause on screen lock via loginctl)
- **Suspend detection** (wall-clock jump → auto-pause)
- **Terminal title updates** (OSC 2 countdown in Ghostty/Alacritty)
- **Clipboard copy** (OSC 52 stats line)
- **Desktop entry** (Walker/wofi launcher)
- **D-Bus notifications** with action buttons under Mako

### Phase 4: Power User
- **Projects** table and picker (e.g., `Backend`, `Frontend`, `Research`)
- **Profiles** in config (e.g., `[profiles.backend]` with custom durations/theme)
- **Smart launch** (`pomogo start backend` auto-applies profile/project)
- **Notification sounds** (profile-level, disable by default)

### Phase 5: Ecosystem (Optional, Demand-Driven)
- `pomogo doctor` — checks environment setup
- `pomogo export --json|--csv|--obsidian` — data export
- Shell hooks — `on_work_start`, `on_work_end`, etc.
- Neovim plugin (separate repo)
- Starship prompt integration
- GitHub correlation report

## Technical Constraints

- **Single static binary** — no CGo, no dynamic linking
- **No daemon** — TUI writes state; integrations read it
- **Minimal deps** — only Go stdlib, Charm stack (Tea/Bubbles/Lip Gloss), TOML, sqlite
- **Keyboard only** — no mouse support required
- **Wayland-first, X11 tolerated** — no Xlib dependency
- **Performance:** startup < 50 ms, idle CPU ≈ 0%

## Out of Scope (Phase 5+)

- Cloud sync, accounts, or login
- Electron/web UI
- Mobile companion apps
- AI/ML integration
- Custom TUI themes at runtime (config only)

## Success Metrics

- Phase 1: A stranger runs `pomogo` and completes a focus cycle with notifications
- Phase 2: `pomogo stats --week` truthfully represents a week of focus
- Phase 3: Waybar, tmux, and Hyprland notifications work without extra setup
- Phase 4: Power users can `pomogo start backend` and it Just Works
- Phase 5: Ecosystem features enable infinite extensibility via hooks and exports
