# PomoGo — Development Pipeline

> **PomoGo — Native Focus Companion for Linux Developers.**
> This file is the single source of truth for *what to build, in what order, and when it counts as done*.
> Vision and philosophy live in `plan.md` and `docs/00-vision.md`. This file is execution only.

---

## How to use this file

**For humans:** work top to bottom. Never start a phase before the previous phase's exit gate is checked. Update checkboxes as you go.

**For AI agents:** follow these rules on every session:

1. Read this file first. Find the first unchecked task in the active phase.
2. Work on **one task ID at a time** (e.g. `P1.4`). Do not batch tasks.
3. A task is done only when its **Verify** command/check passes. Run it.
4. When done: check the box, commit with message `P1.4: <short description>`.
5. Never skip ahead across a phase gate. If a task is blocked, mark it `⚠` with a note and move to the next unblocked task in the same phase.
6. If a decision is not covered here or in `docs/`, prefer the **Design Rules** below over inventing new behavior.

Status legend: `[ ]` todo · `[x]` done · `⚠` blocked (add a note inline)

---

## Locked decisions (do not relitigate)

| Area | Decision | Why |
|---|---|---|
| Language | Go 1.24+ | Single static binary, fast startup |
| TUI | Bubble Tea **v2** + Bubbles v2 + Lip Gloss | Current major version of the Charm stack |
| Config | TOML at `~/.config/pomogo/config.toml` | plan.md Rule 5 |
| Storage | SQLite via `modernc.org/sqlite` | **CGo-free** — keeps the single-binary rule intact |
| CLI | stdlib `flag` + manual subcommand dispatch | No cobra; opinionated, tiny, zero deps |
| Process model | **No daemon.** TUI process writes a state file; integrations read it | plan.md Rule 7 |
| State file | JSON at `$XDG_RUNTIME_DIR/pomogo/state.json` (fallback `~/.cache/pomogo/`) | Waybar/tmux/Neovim poll this — designed in Phase 1, exploited in Phase 3 |
| Paths | XDG only: `~/.config/pomogo`, `~/.local/share/pomogo`, `~/.cache/pomogo` | plan.md Rule 6 |
| Notifications | `notify-send` (libnotify) shell-out, degrade silently if absent | Works with Mako/Hyprland out of the box |
| Target | Wayland-first Linux; X11 tolerated, never designed for | plan.md Rule 9 |

### Design Rules (from plan.md — apply to every task)

1. Keyboard first. 2. No pop-up windows — native notifications. 3. Everything works from the terminal. 4. Sane defaults, zero required config. 5. TOML config. 6. XDG paths. 7. Single binary, no daemon. 8. Near-zero idle CPU (tick only when visible/running; use timer deadlines, not busy loops). 9. Wayland-first. 10. Quality bar: LazyGit / btop / Yazi.

---

## Target repository layout

```
pomogo/
├── cmd/pomogo/main.go        # entrypoint + subcommand dispatch
├── internal/
│   ├── timer/                # session state machine (pure, no I/O)
│   ├── ui/                   # Bubble Tea models, views, keymaps
│   ├── theme/                # Lip Gloss styles, theme registry
│   ├── config/               # TOML load/validate/defaults
│   ├── statefile/            # runtime state.json read/write
│   ├── store/                # SQLite persistence (Phase 2)
│   ├── notify/               # notify-send wrapper
│   └── integrations/         # waybar, tmux, hypr (Phase 3)
├── docs/                     # 00-vision … 05-contributing
├── contrib/                  # waybar snippet, tmux snippet, .desktop, completions
├── .github/workflows/ci.yml
├── .golangci.yml
├── PIPELINE.md               # ← this file
└── README.md
```

---

# Phase 0 — Foundation

**Goal:** a repo where every later task has rails: build, lint, test, CI, docs skeleton.
**Entry:** nothing exists but `plan.md`. **Exit gate below.**

- [x] **P0.1 — Init repo & module**
  - Do: `git init`; `go mod init github.com/<owner>/pomogo`; add Go `.gitignore`; first commit includes `plan.md` + `PIPELINE.md`.
  - Deliverable: versioned repo with module.
  - Verify: `git log --oneline` shows initial commit; `go mod tidy` exits 0.

- [x] **P0.2 — Walking skeleton binary**
  - Do: `cmd/pomogo/main.go` with subcommand dispatch (`pomogo`, `pomogo version`, `pomogo help`). `pomogo` prints a placeholder; `version` prints version injected via `-ldflags`.
  - Deliverable: buildable binary with CLI surface stub.
  - Verify: `go build ./... && ./pomogo version` prints a version string.

- [x] **P0.3 — Lint + format + CI**
  - Do: add `.golangci.yml` (start from charmbracelet/bubbletea-app-template's config); GitHub Actions workflow running `go build`, `go vet`, `golangci-lint`, `go test ./...` on push.
  - Deliverable: CI green on main.
  - Verify: `golangci-lint run` exits 0 locally; Actions run passes.

- [x] **P0.4 — Docs skeleton**
  - Do: create `docs/00-vision.md` (distilled from plan.md), `01-prd.md`, `02-architecture.md` (state machine + state-file design from this file), `03-ui.md` (UI philosophy + quality bar), `04-roadmap.md` (points here), `05-contributing.md`. Stubs are fine except 00 and 02, which must be real.
  - Deliverable: `docs/` tree.
  - Verify: files exist; 00-vision and 02-architecture each ≥ 1 page of substance.

- [x] **P0.5 — Release tooling**
  - Do: GoReleaser config (single linux/amd64 + linux/arm64 binary, tarball, checksums). No publishing yet.
  - Deliverable: `.goreleaser.yml`.
  - Verify: `goreleaser release --snapshot --clean` produces binaries in `dist/`.

**⛔ Phase 0 exit gate:** CI passes on a clean clone (`go build`, lint, test); `pomogo version` works; docs skeleton committed.

---

# Phase 1 — MVP: A Beautiful Timer

**Goal:** `pomogo` opens a gorgeous full-screen timer you'd screenshot next to btop. Ship-quality, tiny scope.
**Entry:** Phase 0 gate passed.

- [x] **P1.1 — Session state machine (pure Go, no UI)**
  - Do: `internal/timer`: states `Idle → Work → ShortBreak/LongBreak → …`, events `start/pause/resume/skip/complete/reset`, long break every N work sessions. Pure functions + struct; time injected via clock interface. Table-driven tests for every transition.
  - Deliverable: `internal/timer` with ≥90% coverage of transitions.
  - Verify: `go test ./internal/timer/ -cover` passes, coverage ≥ 90%.

- [x] **P1.2 — Config system**
  - Do: `internal/config`: TOML load from XDG path, full defaults when file missing (Rule 4), validation with friendly errors. Fields: work/short/long durations, sessions-before-long-break, theme name, notifications on/off, sound on/off. `pomogo config --init` writes a commented default file.
  - Deliverable: config package + `config --init` subcommand.
  - Verify: `go test ./internal/config/`; delete config file → app still runs; `pomogo config --init` writes valid TOML.

- [x] **P1.3 — Theme system**
  - Do: `internal/theme`: theme struct (colors, accents, timer digit style), 3 built-in themes (default = Tokyo Night-ish to match Omarchy, plus Catppuccin, Gruvbox). Selected via config. Adaptive to terminal background.
  - Deliverable: theme registry.
  - Verify: `go test ./internal/theme/`; switching theme in config visibly changes UI.

- [x] **P1.4 — TUI: timer screen**
  - Do: Bubble Tea v2 model wrapping `internal/timer`. Centered layout (Rule: whitespace is a feature), large timer display, session type label, progress indication, minimal keybind hint line. Keys: `s` start, `space` pause/resume, `n` skip, `r` reset, `?` help overlay, `q` quit. Resizes cleanly. Tick only while running (Rule 8).
  - Deliverable: working interactive timer.
  - Verify: manual run through a full work→break→work cycle; resize terminal mid-session; `top` shows ~0% CPU while paused.

- [x] **P1.5 — Notifications**
  - Do: `internal/notify`: shell out to `notify-send` with app name, urgency, session-appropriate message on every transition. Missing binary → silent no-op. Respect config toggle.
  - Deliverable: notifications on session transitions.
  - Verify: run under Mako/dunst → notification appears on work-session end; `PATH=/dev/null pomogo` does not crash.

- [x] **P1.6 — Runtime state file**
  - Do: `internal/statefile`: on every state change, atomically write `state.json` (`state`, `session_type`, `ends_at`, `paused`, `remaining_secs`, `pid`) to `$XDG_RUNTIME_DIR/pomogo/`. Remove/mark stale on clean exit. This is the Phase 3 integration contract — document the schema in `docs/02-architecture.md`.
  - Deliverable: state file with documented schema.
  - Verify: `go test ./internal/statefile/`; `cat $XDG_RUNTIME_DIR/pomogo/state.json` mid-session shows live data.

- [x] **P1.7 — Crash/close restore**
  - Do: on startup, if state file exists with a live unexpired session (and pid dead), offer one-key resume.
  - Deliverable: accidental-close recovery.
  - Verify: `kill -9` the app mid-session → relaunch → resume prompt restores correct remaining time.

- [x] **P1.8 — Polish pass + v0.1.0**
  - Do: review every screen against the quality bar (LazyGit/btop/Yazi); fix spacing, colors, help overlay; write README with screenshot/GIF (vhs); tag `v0.1.0`, GoReleaser release.
  - Deliverable: **`pomogo` — that's it.** Released v0.1.0.
  - Verify: fresh Arch container/VM: download binary → runs with zero config.

**⛔ Phase 1 exit gate:** a stranger can `pomogo` → full pomodoro cycle with notifications, no config, no docs reading. CI green, v0.1.0 tagged.

---

# Phase 2 — Productivity: It Tracks Focus

**Goal:** sessions persist; the tool shows you your focus, not just a countdown.
**Entry:** Phase 1 gate passed.

- [ ] **P2.1 — SQLite store**
  - Do: `internal/store` on `modernc.org/sqlite`, DB at `~/.local/share/pomogo/pomogo.db`. Schema v1: `sessions(id, type, task, note, started_at, ended_at, completed, duration_secs)` + `schema_migrations`. Embedded migrations, forward-only.
  - Deliverable: store package with migration runner.
  - Verify: `go test ./internal/store/` (temp-dir DB); schema survives open→migrate→reopen.

- [ ] **P2.2 — Session recording**
  - Do: TUI writes every finished/skipped work session to the store (breaks configurable). Completion flag distinguishes finished vs abandoned.
  - Deliverable: automatic history.
  - Verify: complete 2 sessions, skip 1 → `sqlite3 …/pomogo.db 'select count(*) from sessions'` matches.

- [ ] **P2.3 — Current task + session notes**
  - Do: `t` sets the current task (text input, shown on timer screen, saved with session); on session end, optional one-line note prompt (skippable with Enter, config-off-able).
  - Deliverable: task + note capture.
  - Verify: manual: set task, finish session, check both stored in DB.

- [ ] **P2.4 — Stats engine**
  - Do: pure `internal/stats` over store queries: today/week/month totals, completion rate, current & best daily streak.
  - Deliverable: stats package.
  - Verify: `go test ./internal/stats/` with seeded fixture DB (streak edge cases: gaps, today-empty, timezone midnight).

- [ ] **P2.5 — Stats TUI + CLI**
  - Do: `Tab` toggles a stats view: today summary, week bar graph (Lip Gloss bars), streak display, monthly summary. Also non-interactive `pomogo stats [--week|--month]` printing the same to stdout (Rule 3).
  - Deliverable: stats screen + subcommand.
  - Verify: manual review against quality bar; `pomogo stats --week` renders sane output in a plain pipe (`pomogo stats | cat`).

- [ ] **P2.6 — v0.2.0**
  - Do: docs update (02-architecture: schema; README: stats screenshots), tag and release.
  - Verify: upgrade from v0.1.0 data dir works (no DB → creates; existing config untouched).

**⛔ Phase 2 exit gate:** week of dogfooding recorded correctly; `pomogo stats` truthful; v0.2.0 tagged.

---

# Phase 3 — Omarchy Integration: Feels Like It Shipped With the OS

**Goal:** every surface a keyboard-first Hyprland user looks at can show PomoGo state — all via the state file, still no daemon.
**Entry:** Phase 2 gate passed. Everything here reads `state.json` through one command: **`pomogo status`**.

- [ ] **P3.1 — `pomogo status` command**
  - Do: reads state file, output formats: default human line, `--format waybar` (JSON: `text`, `class`, `tooltip`), `--format tmux` (plain `🍅 13:22`), `--format json` (raw). Exits 0 with empty/idle output when no session — never errors in a status bar.
  - Deliverable: the universal integration endpoint.
  - Verify: `go test` on formatter; `pomogo status --format waybar | jq .` mid-session and with no session.

- [ ] **P3.2 — Waybar module**
  - Do: `contrib/waybar/`: custom module config (`interval: 1`, `exec: pomogo status --format waybar`, `return-type: json`) + CSS classes per state (work/break/paused/idle) + install instructions in docs.
  - Deliverable: copy-paste Waybar integration → `🍅 18:24` in the bar.
  - Verify: on a Waybar setup: module shows live countdown; color changes work→break.

- [ ] **P3.3 — tmux integration**
  - Do: `contrib/tmux/pomogo.tmux` snippet for `status-right` using `#(pomogo status --format tmux)`; docs section.
  - Deliverable: `🍅 13:22` in tmux status line.
  - Verify: countdown ticks in tmux status-right.

- [ ] **P3.4 — Idle / lock / suspend awareness**
  - Do: detect suspend by wall-clock jump across ticks (monotonic vs wall time delta) → auto-pause + notify on resume. Lock detection via `loginctl`/D-Bus `org.freedesktop.login1` session `LockedHint` (poll on tick, cheap). Config: `pause_on_lock`, `pause_on_suspend`.
  - Deliverable: timer doesn't lie after lid-close or lock.
  - Verify: manual: lock screen 1 min mid-session → timer paused; suspend/resume → correct remaining time.

- [ ] **P3.5 — Terminal title + clipboard**
  - Do: OSC 2 terminal title updates (`🍅 18:24 · work`, throttled to 1/s, off-able) — works in Ghostty/Alacritty. `y` copies current stats line via OSC 52 (Wayland-safe, no wl-copy dependency).
  - Deliverable: title + yank support.
  - Verify: title visibly counts down in Ghostty; `y` then paste yields stats line.

- [ ] **P3.6 — Desktop entry + launcher**
  - Do: `contrib/pomogo.desktop` (launches in `$TERMINAL`), shows up in Walker/wofi. Shell completions (`pomogo completion zsh|bash|fish`) into `contrib/completions/`.
  - Deliverable: launcher + completions.
  - Verify: entry appears in Walker and launches; `pomogo <TAB>` completes subcommands in zsh.

- [ ] **P3.7 — Mako notification actions**
  - Do: upgrade notify to D-Bus (`org.freedesktop.Notifications`, e.g. godbus) to support actions: "Start break" / "Skip" / "+5 min" buttons on transition notifications. Fallback to notify-send if D-Bus unavailable. **Constraint:** actions only work while the TUI runs (no daemon) — the TUI process listens for the action signal.
  - Deliverable: actionable notifications under Mako.
  - Verify: session ends → notification with buttons → clicking "Start break" starts break in TUI.

- [ ] **P3.8 — v0.3.0 + AUR**
  - Do: docs `03-ui.md`/`02-architecture.md` updates; PKGBUILD in `contrib/aur/`; publish `pomogo-bin` (or `pomogo`) to AUR; tag v0.3.0.
  - Verify: `makepkg -si` on clean Arch installs a working binary with completions + desktop entry.

**⛔ Phase 3 exit gate:** on an Omarchy machine: Waybar shows the timer, tmux shows the timer, lock pauses it, notifications have buttons, installed from AUR. Feels stock.

---

# Phase 4 — Power User: Projects & Profiles

**Goal:** `pomogo start backend` just works.
**Entry:** Phase 3 gate passed.

- [ ] **P4.1 — Projects**
  - Do: `projects` table + FK on sessions; `p` picker in TUI (Bubbles list, create-on-type); `pomogo projects` CLI (list/add/archive). Stats become filterable by project.
  - Verify: store tests; sessions attribute to the picked project; `pomogo stats --project backend` filters.

- [ ] **P4.2 — Profiles in config**
  - Do: `[profiles.<name>]` TOML tables overriding durations, theme, notification sound, auto-start behavior. Profile can pin a default project.
  - Deliverable: profile-aware config with validation.
  - Verify: config tests: profile overrides defaults, unknown keys → friendly error.

- [ ] **P4.3 — `pomogo start [profile|project]`**
  - Do: `pomogo start backend` → resolves profile and/or project, launches TUI already running. `pomogo start` with no args uses defaults. Fuzzy-ish matching, helpful error listing known names.
  - Deliverable: the plan.md headline command.
  - Verify: `pomogo start backend` opens mid-countdown with the right durations/theme/project.

- [ ] **P4.4 — Notification sounds**
  - Do: per-profile sound via `notify-send` sound hints / `canberra-gstreamer` or paplay shell-out; silent when unavailable. Off by default (opinionated).
  - Verify: enable in profile → sound on transition; disable → silence; missing player → no crash.

- [ ] **P4.5 — v0.4.0**
  - Do: docs (config reference gets a full page: every key, default, example), tag, release, AUR bump.
  - Verify: config file from v0.3.0 still loads unmodified.

**⛔ Phase 4 exit gate:** daily driver with ≥2 profiles and ≥2 projects for a week without touching config mid-week; v0.4.0 shipped.

---

# Phase 5 — Ecosystem (optional, demand-driven)

**Goal:** open the data and the state to everything else. Each task independent — cherry-pick by user demand; order below is suggested.
**Entry:** Phase 4 gate passed.

- [ ] **P5.1 — `pomogo doctor`** — checks notify-send, D-Bus, sqlite db integrity, waybar/tmux/mako presence, state-file writability; prints the ✔/✘ report from plan.md. Verify: run on a stripped container → correct ✘s, helpful hints.
- [ ] **P5.2 — Exports** — `pomogo export --json|--csv [range]`; `pomogo report --markdown --week` generates a Markdown focus report. Verify: golden-file tests; output round-trips into sqlite/pandas.
- [ ] **P5.3 — Hooks (the real plugin system)** — config-declared shell hooks: `on_work_start`, `on_work_end`, `on_break_start`, `on_break_end` receiving state as env vars/JSON stdin. This *is* the plugin system — no Go plugin loading. Verify: hook script writes a log line on each transition.
- [ ] **P5.4 — Neovim plugin** — separate repo `pomogo.nvim`: statusline component reading `pomogo status --format json`, `:PomoGo` commands driving via `pomogo` CLI. Verify: lualine shows live countdown.
- [ ] **P5.5 — Shell prompt + widgets** — Starship custom-command snippet + zsh prompt segment in `contrib/`. Verify: prompt shows 🍅 during session, nothing when idle.
- [ ] **P5.6 — GitHub correlation report** — `pomogo report --github <user>`: overlays contribution calendar (via `gh api`) with focus history. Read-only, gh optional. Verify: renders both series for a seeded week.
- [ ] **P5.7 — Obsidian export** — daily-note-formatted Markdown (`pomogo export --obsidian --vault <path>`), appends focus log section. Verify: idempotent append (running twice doesn't duplicate).
- [ ] **P5.8 — v1.0.0** — API/CLI/state-file/config formats declared stable; SemVer commitment documented; announcement post. Verify: docs freeze; all gates green.

**⛔ Phase 5 / v1.0 gate:** state-file schema, CLI surface, and config format frozen and documented; `pomogo doctor` all-green on Omarchy.

---

## Cross-cutting conventions

- **Testing:** pure logic (`timer`, `stats`, `config`, `statefile`, formatters) gets table-driven unit tests in the same task — never deferred. TUI verified manually against the quality bar; integrations verified on-device.
- **Commits:** one task ID per commit, message `P<phase>.<n>: <what>`. Keeps history mappable to this file.
- **Dependencies:** adding any dependency beyond the Charm stack, TOML, sqlite, and godbus requires a note in `docs/02-architecture.md` justifying it.
- **Performance budget:** startup < 50 ms perceived; idle CPU ~0%; state-file writes only on transitions/pauses, not per tick (`ends_at` makes per-second writes unnecessary).
- **Releases:** every phase ends in a tagged release + AUR bump (from Phase 3 on). Release = GoReleaser + updated README + docs.
- **Docs debt rule:** a phase gate cannot pass with `02-architecture.md` out of date.

## Progress at a glance

| Phase | Name | Tasks | Status |
|---|---|---|---|
| 0 | Foundation | 5 | complete |
| 1 | MVP timer | 8 | complete |
| 2 | Productivity | 6 | not started |
| 3 | Omarchy integration | 8 | not started |
| 4 | Power user | 5 | not started |
| 5 | Ecosystem | 8 | not started |

*Update this table when a phase starts/completes. Detailed truth is always the checkboxes above.*
