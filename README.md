# PomoGo

**A beautiful terminal deep-work companion for developers.**

PomoGo is a fast, keyboard-driven, single-binary TUI for Quick Focus sessions and Deep Focus blocks. Pomodoro timing runs behind the scenes; the product is a calm terminal companion you can leave open while you work.

<p align="center">
  <img src="assets/running.png" alt="PomoGo TUI running" width="60%">
</p>

---

## Key Features

*   **Quick Focus Mode**: Start a classic 25 minute focus session and continue into the short break cycle naturally.
*   **Deep Focus Mode**: Start a 1-4 hour block; PomoGo classifies work and break segments behind the scenes while showing the whole block countdown.
*   **10 Built-In Themes**: High-contrast palettes including Tokyo Night, Catppuccin, Gruvbox, Everforest, Rose Pine, Dracula, Kanagawa, Nord, and Carbon. Exposes a live swatch viewer: `pomogo themes`.
*   **5 Responsive Layouts**: Choose from `classic`, `minimal`, `centered`, `compact`, and `retro` views. Switch layouts instantly using the `L` key.
*   **Ambient Effects Engine**: Deterministic particle animations like falling snow, rain, or blinking stars that render behind layouts. Toggle using the `e` key.
*   **Screenshot Zen Mode**: Press `S` to instantly hide all hints, status messages, and dots for a clean screenshot-friendly aesthetic.
*   **Git & Tmux Integration**: Automatically detects your current active Git branch and TMUX session name to display in the status region without executing process calls.
*   **CLI Recap & Stats**: Check stats at a glance with `pomogo stats` or review your last session with `pomogo recap`.

---

## Installation

```sh
go install github.com/Ibnu-Afdel/pomogo/cmd/pomogo@latest
```

Or download a pre-built binary from the [Releases](https://github.com/Ibnu-Afdel/pomogo/releases) page.

---

## Usage

```sh
pomogo              # Start the TUI
pomogo config init  # Initialize default configuration
pomogo stats        # View focus stats (use --week or --month)
pomogo recap        # Show recap of the last completed block
pomogo themes       # List all themes with colored CLI swatches
pomogo projects     # Manage focus projects
```

### Keyboard Shortcuts (TUI)

| Key | Action |
|---|---|
| `s` | Start the current focus session |
| `space` | Pause / resume a running timer |
| `n` | Skip current segment |
| `t` | Edit task description (supports autocomplete) |
| `p` | Edit project category (supports autocomplete) |
| `d` | Choose a Deep Focus duration before starting |
| `tab` | Toggle statistics screen |
| `y` | Copy stats summary to clipboard |
| `T` | Cycle theme live |
| `L` | Cycle layout live |
| `a` | Choose sound profile |
| `S` | Toggle Zen / Screenshot mode |
| `e` | Cycle ambient background effects (none / stars / snow / rain) |
| `v` | Cycle task verb labels (Focusing, Building, Fixing...) |
| `r` | Reset session |
| `esc` | Back / close overlay |
| `?` | Show help menu |
| `q`, `ctrl+c` | Quit |

### Overlay Shortcuts

| Screen | Key | Action |
|---|---|---|
| Help | `?`, `esc` | Close help |
| Restore prompt | `y` | Restore the saved session |
| Restore prompt | `n`, `esc` | Discard the saved session |
| Deep Focus duration picker | `up`, `k`, `shift+tab` | Move selection up |
| Deep Focus duration picker | `down`, `j`, `tab` | Move selection down |
| Deep Focus duration picker | `1`-`4` | Jump to a 1-4 hour preset |
| Deep Focus duration picker | `enter` | Select duration |
| Deep Focus duration picker | `esc` | Cancel picker |
| Custom duration input | `enter` | Apply typed duration |
| Custom duration input | `esc` | Return to duration picker |
| Sound picker | `up`, `k`, `shift+tab` | Move selection up |
| Sound picker | `down`, `j`, `tab` | Move selection down |
| Sound picker | `space` | Preview selected sound |
| Sound picker | `enter` | Apply selected sound profile |
| Sound picker | `esc` | Cancel picker |
| Any picker / prompt | `q`, `ctrl+c` | Quit |

---

## Configuration

To create a configuration file at `~/.config/pomogo/config.toml`, run:
```sh
pomogo config init
```

### Example Config:
```toml
# Main Durations
work_duration = 25
short_break_duration = 5
long_break_duration = 15
sessions_before_long_break = 4

# UI Aesthetics
theme = "tokyo-night"
layout = "classic"
effects = "none"

# Developer Options
show_git = true
show_tmux = false
```

---

## Performance Targets

*   **RSS Memory**: < 15 MB
*   **CPU Overhead**: < 1.0% (with particle animations active)
*   **Cold Start Time**: < 50 ms
*   **Zero Daemon Dependencies**: SQLite DB & JSON statefile based.

---

## Development

Run tests and linters:
```sh
go test ./...
golangci-lint run
```
