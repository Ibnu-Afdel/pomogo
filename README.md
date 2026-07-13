# PomoGo

**A keyboard-first Pomodoro timer for Linux terminals.**

Single static binary. No daemon. No config required. Integrates with Waybar, tmux, and
Neovim via a runtime state file — coming in Phase 3.

<!-- demo GIF: run `vhs contrib/demo.tape` to regenerate -->

## Install

**Download** the latest binary from the [Releases](https://github.com/Ibnu-Afdel/pomogo/releases) page,
or build from source:

```sh
go install github.com/Ibnu-Afdel/pomogo/cmd/pomogo@latest
```

Or clone and build:

```sh
git clone https://github.com/Ibnu-Afdel/pomogo
cd pomogo
go build -o pomogo ./cmd/pomogo
./pomogo
```

## Usage

```sh
pomogo            # open the timer
pomogo version    # show version
pomogo config init  # write a commented default config file
pomogo help       # show command list
```

### Keybindings

| Key | Action |
|---|---|
| `s` | Start the queued work or break session |
| `space` | Pause / resume |
| `n` | Skip to next phase |
| `r` | Reset and clear runtime state |
| `?` | Toggle help overlay |
| `q` / `ctrl+c` | Quit (state saved if a session is running) |

When a running session is interrupted, PomoGo leaves a state file behind. On the next
launch it offers a one-key restore prompt.

## Configuration

PomoGo works with zero config. To create an editable default file:

```sh
pomogo config init
```

Config path: `~/.config/pomogo/config.toml`

```toml
work_duration = 25               # minutes
short_break_duration = 5
long_break_duration = 15
sessions_before_long_break = 4
theme = "tokyo-night"            # tokyo-night | catppuccin | gruvbox
notifications_enabled = true
sound_enabled = false
```

## Themes

Three built-in themes selectable via `theme` in the config:

| Name | Style |
|---|---|
| `tokyo-night` | Dark, neon accents (default) |
| `catppuccin` | Light pastels |
| `gruvbox` | Warm earth tones |

## Runtime State File

While a session is active, PomoGo writes:

```
$XDG_RUNTIME_DIR/pomogo/state.json   (fallback: ~/.cache/pomogo/state.json)
```

The file is written atomically on every state change and removed on clean exit. It is the
integration contract for Waybar, tmux, and Neovim widgets (Phase 3). Schema documented in
`docs/02-architecture.md`.

## Notifications

Session transitions send a native desktop notification via `notify-send`. Works with Mako
and dunst out of the box. If `notify-send` is absent the app runs silently.

## Generating the Demo GIF

Requires [VHS](https://github.com/charmbracelet/vhs):

```sh
vhs contrib/demo.tape
```

## Development

```sh
go test ./...
go build -o pomogo ./cmd/pomogo
golangci-lint run
```

Phase roadmap and task tracking: see `PIPELINE.md`.
