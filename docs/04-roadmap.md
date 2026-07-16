# Roadmap: Phases & Milestones

This is the long-term vision for PomoGo. See [PIPELINE-V2.md](../PIPELINE-V2.md) for detailed task-level breakdown.

## Phase 1: MVP Timer (v0.1.0)

**Goal:** A beautiful, professional-grade timer you'd screenshot next to btop.

**Deliverables:**
- Full-screen Bubble Tea UI with centered timer
- Session transitions (Work → ShortBreak → LongBreak → Work)
- Keyboard controls (start, pause, resume, skip, reset)
- Native notifications on every transition
- Theme system (Tokyo Night, Catppuccin, Gruvbox)
- TOML config at `~/.config/pomogo/config.toml` (with sane defaults)
- Runtime state file at `$XDG_RUNTIME_DIR/pomogo/state.json`
- Restore on accidental close (mid-session recovery)

**Quality Bar:** Ship-quality timer, zero required setup.

**Status:** In progress (P0.2 complete, P1.1-P1.8 next)

---

## Phase 2: Productivity — It Tracks Focus (v0.2.0)

**Goal:** Sessions persist; the tool shows you your focus, not just a countdown.

**Deliverables:**
- SQLite database at `~/.local/share/pomogo/pomogo.db`
- Automatic session recording (completed vs abandoned)
- Current task labeling (set with `t` key)
- Optional session notes prompt on completion
- Stats engine: today/week/month totals, completion rates, streaks
- Stats screen in TUI (`Tab` key)
- CLI command: `pomogo stats [--week|--month]`
- Weekly bar graphs showing focus per day

**Quality Bar:** A week of truthful, scannable focus data.

**Timeline:** After Phase 1 passes.

---

## Phase 3: Omarchy Integration (v0.3.0)

**Goal:** Every surface a keyboard-first Hyprland user sees shows PomoGo state.

**Integration Points:**
1. **Waybar module** — `🍅 18:24` in the status bar (JSON output, CSS classes)
2. **tmux status-right** — live countdown via `pomogo status --format tmux`
3. **Lock-aware pausing** — auto-pause on screen lock (loginctl D-Bus detection)
4. **Suspend detection** — wall-clock jump → auto-pause + notify on resume
5. **Terminal title** — OSC 2 countdown in Ghostty/Alacritty
6. **Clipboard support** — `y` key copies stats line via OSC 52
7. **Desktop entry + launcher** — shows in Walker/wofi
8. **Shell completions** — zsh, bash, fish auto-complete
9. **Mako notification actions** — "Start break", "Skip", "+5 min" buttons
10. **AUR release** — install via `pacman -S pomogo` / `pomogo-bin`

**Deliverables:**
- `pomogo status` — universal integration endpoint
- `contrib/waybar/` — module config + CSS
- `contrib/tmux/` — tmux snippet
- `contrib/pomogo.desktop` — desktop entry
- `contrib/completions/` — shell completions
- D-Bus listener for Mako action buttons
- Idle/lock/suspend awareness
- Full AUR package

**Quality Bar:** Feels like it shipped with Omarchy.

**Timeline:** After Phase 2 passes.

---

## Phase 4: Power User — Projects & Profiles (v0.4.0)

**Goal:** `pomogo start backend` just works. The tool adapts to your workflow.

**Deliverables:**
- **Projects:** simple picker (e.g., Backend, Frontend, Research)
- **Profiles:** config-level overrides (durations, theme, sounds, auto-start)
- **Smart launch:** `pomogo start [profile|project]` with fuzzy matching
- **Per-profile sounds** — enable/disable notification audio
- **Stats filtering** — `pomogo stats --project backend`
- **Config reference documentation** — every field, default, example

**Example Config:**
```toml
[profiles.coding]
work_duration = 50
short_break_duration = 5
theme = "tokyo-night"
default_project = "pomogo"

[profiles.studying]
work_duration = 40
short_break_duration = 10
notification_sound_enabled = true
```

**Quality Bar:** Daily driver with ≥2 profiles and ≥2 projects for a week.

**Timeline:** After Phase 3 passes.

---

## Phase 5: Ecosystem (v1.0.0 & beyond)

**Goal:** Open the data and state to everything else.

**Optional, Demand-Driven Tasks:**

1. **pomogo doctor** — environment health check (notify-send, D-Bus, sqlite, waybar, etc.)
2. **Exports** — JSON, CSV, Markdown reports; time-range filtering
3. **Shell hooks** — `on_work_start`, `on_work_end`, etc. with env/JSON input
4. **Neovim plugin** — separate repo (pomogo.nvim), lualine integration
5. **Shell prompt integration** — Starship custom command, zsh prompt segment
6. **GitHub correlation** — overlay focus history with contribution graph
7. **Obsidian export** — daily-note appends with focus log section
8. **v1.0.0 freeze** — API/CLI/config/state-file formats declared stable, SemVer commitment

**Quality Bar:** State-file schema frozen; `pomogo doctor` all-green on Omarchy.

**Timeline:** User-driven, cherry-pick as needed.

---

## Success Checkpoints

- **Phase 1:** Stranger can `pomogo` → full cycle with notifications, no config.
- **Phase 2:** `pomogo stats --week` is truthful and scannable.
- **Phase 3:** Waybar/tmux/Hyprland all show live timer; lock detection works.
- **Phase 4:** Power users run `pomogo start backend` and it just works.
- **Phase 5:** Ecosystem enables infinite extensions via hooks and exports.

---

## Release Cadence

- **v0.1.0** — Phase 1 completion, GitHub + AUR
- **v0.2.0** — Phase 2 completion, GitHub + AUR
- **v0.3.0** — Phase 3 completion, GitHub + AUR
- **v0.4.0** — Phase 4 completion, GitHub + AUR
- **v1.0.0** — Phase 5 completion, freeze API/config/state formats

Each release ships complete, usable software — not beta or incomplete feature sets.

---

## Out of Scope

- Cloud sync, accounts, or login
- Electron/web UI (PomoGo is terminal-only by design)
- Mobile apps (focus on desktop Linux first)
- AI/ML integration (data belongs to the user)
- GUI preferences dialog (TOML + `pomogo config --init`)
- Custom TUI themes at runtime (config only, not in-app editing)
