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
| `s` | Start / Pause / Resume timer |
| `n` | Skip current segment |
| `t` | Edit task description (supports autocomplete) |
| `p` | Edit project category (supports autocomplete) |
| `T` | Cycle theme live |
| `L` | Cycle layout live |
| `e` | Cycle ambient background effects (none / stars / snow / rain) |
| `v` | Cycle task verb labels (Focusing, Building, Fixing...) |
| `S` | Toggle Zen / Screenshot mode |
| `tab` | Toggle statistics screen |
| `r` | Reset session |
| `?` | Show help menu |
| `q` | Quit |

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

Product direction lives in [`docs/00-product-vision.md`](docs/00-product-vision.md), with current research notes in [`docs/07-product-research.md`](docs/07-product-research.md).

Run tests and linters:
```sh
go test ./...
golangci-lint run
```
