# PomoGo v2 — Deep Work Companion Pipeline

> **Transformation: from "a Pomodoro timer" to "a beautiful terminal Deep Work companion."**
> This file is the single source of truth for the v2 transformation: what to build, in what
> order, and when each task counts as done. Product vision lives in `docs/00-product-vision.md`.
> The completed v1 pipeline was removed from the repo to keep the v2 direction unambiguous.

---

## How to use this file

**For humans:** work top to bottom. Never start a phase before the previous phase's exit gate passes.

**For AI agents:** on every session:

1. Read this file first. Find the first unchecked task in the active phase.
2. Work on **one task ID at a time** (e.g. `D1.3`). Do not batch tasks.
3. A task is done only when its **Verify** check passes. Run it.
4. When done: check the box, commit with message `D1.3: <short description>`.
5. Never skip ahead across a phase gate. If blocked, mark the task `⚠` with an inline note and take the next unblocked task in the same phase.
6. Decisions not covered here: prefer the **Locked Decisions** and **Design Rules** below over inventing behavior. If still ambiguous, choose the option with less code.

Status legend: `[ ]` todo · `[x]` done · `⚠` blocked (note inline)

---

## Audit of the current codebase (July 2026)

Read this before touching anything. The prompt-level assumption "inspect before rewriting"
has been done; these are the facts the plan is built on.

### What exists and is GOOD — keep, do not rewrite

| Component | File | Assessment |
|---|---|---|
| Timer state machine | `internal/timer/timer.go` | Pure, no I/O, injected `Clock`, deadline-based (`EndsAt`), well-tested. **Extend, don't replace.** |
| SQLite store | `internal/store/store.go` | CGo-free (`modernc.org/sqlite`), versioned migrations (currently v3), sessions + projects + notes. |
| Config | `internal/config/config.go` | TOML, XDG paths, defaults-first, profiles with overrides, hooks. |
| Crash recovery | `internal/statefile/` + `internal/restore/` | Atomic JSON state file in `$XDG_RUNTIME_DIR/pomogo/state.json`; PID-staleness + expiry checks; restore prompt on launch. |
| Integrations | `internal/integrations/` | `pomogo status --format waybar|tmux|json` reads the state file. No daemon. Doctor diagnostics. |
| Notifications | `internal/notify/` | `notify-send` + `canberra-gtk-play` shell-outs, degrade silently, DBus action buttons. |
| Stats engine | `internal/stats/stats.go` | Pure aggregation: today/week/month, streaks, completion rate. |
| CLI | `cmd/pomogo/main.go` | stdlib `flag` + manual dispatch: stats, history, status, export, report, doctor, projects, start, completion. |

### What exists and must CHANGE

| Component | Problem | v2 answer |
|---|---|---|
| `internal/ui/model.go` (1,288 lines) | Monolith: key handling, rendering, DB writes, hooks, OSC sequences all in one file. Rendering is hardcoded to one centered-box layout. | Split into `model/update/keymap` + new `render` package with a display-snapshot boundary (Phase D2). |
| `internal/theme/theme.go` | Only 3 themes; 8 color roles; a hand-rolled hex→ANSI256 map that duplicates what lipgloss/termenv already do automatically. | Expanded palette struct, 10 built-ins, random/daily selection, external TOML themes. Delete the ANSI256 map. |
| `internal/timer` cycle logic | `advancePhase()` hardcodes work→short→…→long forever. No concept of a finite block. | Segment-plan layer on top (Phase D1). The Session mechanics (pause/resume/tick/deadlines) are reused as-is. |
| Phase transitions | Completing a phase drops to idle; user presses `s` for the next phase. | Deep Focus must auto-advance (a 2-hour block cannot ask for a keypress every 25 min). Quick Focus keeps confirm-to-continue, configurable. |
| Statefile schema | No notion of mode or block. | Additive v2 fields; existing fields keep meaning so waybar/tmux configs don't break. |
| DB schema | Sessions have no mode/block linkage. | Migrations 4+: `mode`, `block_id` columns; new `blocks` table. |

### Stale decision corrected

The old v1 pipeline locked "Bubble Tea **v2**". The shipped code is `bubbletea v1.3.10` /
`bubbles v1.0.0` / `lipgloss v1.1.0`. **v2 migration is explicitly OUT of scope for this
transformation** — do not mix a framework major-version migration with a product redesign.
It becomes a candidate task after D5.

---

## Locked decisions (do not relitigate)

| Area | Decision | Why |
|---|---|---|
| Stack | Go 1.26, Bubble Tea v1.3.x, Bubbles v1, Lip Gloss v1, `modernc.org/sqlite` | What ships today; stable; migration ≠ redesign |
| Product frame | Two modes: **Quick Focus** (classic 25/5) and **Deep Focus** (1–4 h block, Pomodoro internally, only the block timer shown) | `docs/00-product-vision.md` |
| Engine shape | New `internal/session` package: a `Block` is an ordered list of `Segment`s executed by the existing deadline mechanics. Pomodoro becomes an implementation detail. | Reuses proven timing code; finite & infinite plans share one engine |
| Deep Focus countdown | Wall-clock remaining of the whole block **including breaks** (`02:13:45 remaining`) | That is the promise of the mode |
| Deep Focus advance | Auto-advances between segments, no keypress. Notifications still fire on work↔break. | A block must run unattended |
| Quick Focus advance | Keeps current behavior (idle + press `s`), new config `auto_advance` can change it | Backward compatible |
| Segment splitting | Fill the block with `work,break,work,break,…`; **last segment is always work**; trailing partial work segment allowed (min 10 min, else fold into previous work); no trailing break | Ending a deep block on a break is absurd |
| Theme/Layout independence | Themes = colors only. Layouts = arrangement only. Any theme × any layout. Both support `random` (per launch) and `daily` (date-seeded). | Combinatorial variety, zero extra maintenance |
| Render boundary | UI update logic produces a `DisplayState` snapshot struct; layouts are pure functions `func(DisplayState, *Theme, Frame) string` | Layouts/testable without a live timer |
| Screenshot mode | Key `S` (capital) toggles: hides hints/controls/status, keeps project/task/mode/timer/progress, centers | Flagship feature |
| Ambient effects | Optional, default **off**, driven by the existing 1 s tick — **no extra goroutines, no extra tickers**. Rendered only into whitespace around the content box. | Performance rule 8 |
| External themes | `~/.config/pomogo/themes/*.toml`, same field names as built-ins; loaded at startup; name collision → user file wins | Extensible without Go changes |
| Statefile | Schema v2 is additive. `state`, `session_type`, `ends_at`, `remaining_secs`, `pid` keep their exact v1 meaning (current segment). New: `mode`, `block_ends_at`, `block_remaining_secs`, `segment_index`, `segment_count`, `version` | waybar/tmux/starship consumers must not break |
| DB | Additive migrations only (4, 5, …). Never edit migrations 1–3. | Existing user data |
| Scope exclusions | ❌ Kanban, calendar sync, notes editor, AI coach, teams, cloud accounts, auth, complex graphs, Spotify, daemon, telemetry | `docs/00-product-vision.md` "What I would NOT add" |

### Design rules (apply to every task)

1. Keyboard first. 2. Native notifications, no popups. 3. Terminal is the whole product.
4. Sane defaults, zero required config. 5. TOML. 6. XDG paths. 7. Single binary, no daemon.
8. **Near-zero idle CPU: one 1 s tick while running, none while idle; no busy loops; no goroutines beyond the existing hook-runner and DBus listener.** 9. Wayland-first Linux.
10. Quality bar: LazyGit / btop / Yazi. 11. Whitespace is a feature — when in doubt, show less.
12. Performance budget: **< 15 MB RSS, < 1 % CPU, < 50 ms cold start** (measured in D0.3, enforced at every phase gate).

---

## Target architecture

```
pomogo/
├── cmd/pomogo/main.go            # dispatch (unchanged shape)
├── internal/
│   ├── timer/                    # UNCHANGED pure segment mechanics (Session keeps
│   │                             #   pause/resume/tick/deadline logic)
│   ├── session/                  # NEW  Block/Segment plan: builds segment lists for
│   │                             #   Quick & Deep modes, tracks block progress,
│   │                             #   drives timer.Session, emits SegmentEnded/BlockEnded
│   ├── ui/
│   │   ├── model.go              # Bubble Tea model: state + routing only (slim)
│   │   ├── update.go             # message handling, key dispatch
│   │   ├── keymap.go             # single source of key bindings (feeds help view)
│   │   ├── display.go            # builds DisplayState snapshot from model
│   │   └── screens/              # input, restore-prompt, stats, recap screens
│   ├── render/                   # NEW  pure rendering
│   │   ├── layout.go             # Layout interface + registry + random/daily pick
│   │   ├── classic.go minimal.go centered.go compact.go retro.go
│   │   ├── bigclock.go           # big digits (moved from model.go), H:MM:SS capable
│   │   ├── widgets.go            # progress bar, session dots, labels
│   │   └── ambient.go            # particle fields composited into Place() whitespace
│   ├── theme/                    # expanded palette, 10 built-ins, loader for user TOML,
│   │                             #   random/daily; ANSI256 map DELETED
│   ├── config/                   # + [quick_focus] [deep_focus] sections, theme/layout/
│   │                             #   effects keys, legacy key aliases
│   ├── statefile/                # schema v2 (additive)
│   ├── restore/                  # block-aware restore
│   ├── store/                    # + migrations 4/5: blocks table, mode/block_id cols
│   ├── stats/                    # + deep-work hours, block recaps (pure funcs)
│   ├── devinfo/                  # NEW  git branch (read .git/HEAD, no exec), tmux session
│   ├── notify/  integrations/    # unchanged, minor additions
├── docs/                         # updated per phase
└── contrib/                      # unchanged consumers, verified against statefile v2
```

**Data flow:** `config → session.Block → timer.Session mechanics → ui.Model → DisplayState
→ render.Layout(theme, frame) → screen`. Storage and statefile writes happen only in
`ui` update paths (as today), never in render.

---

# Phase D0 — Groundwork & verification rails

**Goal:** measurements and seams in place so the redesign is safe.
**Exit gate:** `go test ./...` green · benchmark baseline recorded · `DisplayState` exists and the current single layout renders through it.

- [ ] **D0.1 — Research pass (recorded, time-boxed)**
  - Do: verify current versions/idioms for bubbletea v1.x tick patterns, lipgloss `Place` whitespace options (`WithWhitespaceChars`/`WithWhitespaceForeground` — planned substrate for ambient effects), bubbles progress/textinput, and skim 3 reference TUIs for layout/typography ideas (lazygit, btop, glow). Record findings + links in `docs/06-research.md`. Max half a day; no code.
  - Verify: `docs/06-research.md` exists with a "decisions affected" section (confirm or amend the ambient-effects substrate decision).
- [ ] **D0.2 — Baseline metrics**
  - Do: script `contrib/bench.sh`: binary size, cold-start time (`hyperfine 'pomogo version'`), RSS of idle TUI after 10 s, CPU % over 60 s running (via `pidstat`). Record numbers in `docs/06-research.md`.
  - Verify: script runs; baseline table committed.
- [ ] **D0.3 — Extract `DisplayState` seam**
  - Do: define `ui.DisplayState` (mode label, project, task, phase kind, segment remaining, block remaining, progress 0–1, segment index/count, paused/running/idle, status message, hints visibility, theme name, layout name). `renderTimerScreen` becomes `render.Classic(ds, theme, frame)` fed by a `model.displayState()` builder. Pixel-identical output.
  - Verify: golden-string test for the classic render at 80×24; existing tests green.
- [ ] **D0.4 — Split `ui/model.go`**
  - Do: mechanical split into `model.go` / `update.go` / `keymap.go` / `display.go` / `screens/*.go` + `render/bigclock.go`, `render/widgets.go`. No behavior change. Introduce a `KeyMap` struct that the help overlay renders from (kills the hand-maintained help text drift).
  - Verify: `go build` clean; golden test from D0.3 still passes; no file in `internal/ui` > 400 lines.

---

# Phase D1 — Engine: modes & blocks (the core change)

**Goal:** Deep Focus works end-to-end in the existing UI. Pomodoro becomes an implementation detail.
**Exit gate:** start a 1 h Deep Focus block → auto-advancing 25/5 segments → block countdown displayed → sessions recorded with `block_id` → quit/restart mid-block restores correctly.

- [ ] **D1.1 — `internal/session`: Segment plans**
  - Do: `Segment{Kind SegmentKind /* Work|ShortBreak|LongBreak */, Duration time.Duration}`.
    `BuildDeepPlan(total, work, brk time.Duration) []Segment` implementing the locked splitting rule.
    `Block{Mode, Segments, Index, PlannedTotal}` with `Remaining(segRemaining time.Duration) time.Duration`, `Advance() (Segment, bool)`, `Progress()`. Quick mode = `Block` with a cyclic generator (no finite list). Pure package, table-driven tests incl. edge cases: total < work, exact fits, trailing 7 min.
  - Verify: `go test ./internal/session/` — includes property check: sum of work+break segments == planned total (± the fold rule), last segment is Work.
- [ ] **D1.2 — Wire Block to the timer mechanics**
  - Do: `session.Runner` owns a `timer.Session` configured per-segment (feed each segment's duration as the work duration; the runner, not `advancePhase`, decides what's next). On `Tick()==true`: record segment, `Advance()`, auto-start next segment when `block.AutoAdvance`. Emits typed events (`SegmentEnded`, `BlockEnded`) the UI consumes. Quick mode reproduces today's behavior exactly through the same runner.
  - Verify: unit tests with fake clock: 1 h block runs 25/5/25(/5)/… to completion with zero keypresses; quick mode still idles between phases.
- [ ] **D1.3 — Mode selection UI**
  - Do: idle screen gains a mode picker: `[q] Quick Focus  [d] Deep Focus`. Deep Focus opens a duration select (1h / 2h / 3h / 4h / custom input, reusing the existing textinput screen). Selection stored on the model; `s` starts the block.
  - Verify: manual: pick 2 h, see `Deep Focus · 01:59:59 remaining`; golden test for the picker screen.
- [ ] **D1.4 — Deep Focus display semantics**
  - Do: `DisplayState` block fields populated; classic layout shows big block countdown (H:MM:SS — extend `bigclock.go`), phase shown subtly (`focus · break in 12m` style, muted), progress bar = block progress. Session dots hidden in deep mode (that's the point). Break segments swap the accent color as today.
  - Verify: golden tests for deep-mode work + break frames.
- [ ] **D1.5 — Persistence: blocks & modes**
  - Do: migration 4: `CREATE TABLE blocks(id, mode TEXT, planned_secs INT, started_at, ended_at, completed INT, pauses INT DEFAULT 0)`; migration 5: `ALTER TABLE sessions ADD COLUMN mode TEXT`, `ADD COLUMN block_id INTEGER REFERENCES blocks(id)`. Store methods: `CreateBlock`, `FinishBlock`, `IncrementBlockPauses`. Runner records work segments with `block_id`. Note prompts: **only at block end** in deep mode (never between segments).
  - Verify: `go test ./internal/store/`; migrating a copy of a real v3 DB preserves all rows.
- [ ] **D1.6 — Statefile v2 + restore**
  - Do: add locked v2 fields; bump `version:2`. `restore` reconstructs mode, block deadline, segment index (block remaining is authoritative; segment remaining derived). v1 files (no `version`) restore as quick mode.
  - Verify: `go test ./internal/statefile/ ./internal/restore/`; kill -9 mid-block → relaunch → `y` → countdown continues within 2 s of true remaining. `pomogo status --format waybar` output unchanged for quick mode.
- [ ] **D1.7 — Config: mode sections**
  - Do: add `[quick_focus] work/short_break/long_break/sessions_before_long_break/auto_advance` and `[deep_focus] work/break/default_duration` sections. Legacy top-level keys still parse as quick_focus aliases (warn in `doctor`, don't break). `pomogo config init` writes the new commented format. CLI: `pomogo start --deep 2h` and profile field `mode`.
  - Verify: config tests: old file parses identically; new file round-trips; `pomogo doctor` flags legacy keys.

---

# Phase D2 — Personality: themes & layouts

**Goal:** any of 10 themes × 5 layouts, random/daily selection, external theme files.
**Exit gate:** `theme = "random"` + `layout = "random"` give a different pairing across launches; every combination renders without panic at 60×18 and 200×50 (scripted sweep test).

- [ ] **D2.1 — Expanded Theme struct**
  - Do: grow palette to ~12 roles: `Work, Break, LongBreak, Idle, Accent, Text, Muted, Subtle, Border, ProgressFill, ProgressTrack, Ambient`. Migrate 3 existing themes. **Delete** the `ANSI256()` map — lipgloss/termenv degrade automatically. Fix `catppuccin` to Mocha (dark — what devs expect); keep Latte as `catppuccin-latte`.
  - Verify: theme tests; classic golden updated once; `TERM=xterm-256color` and truecolor both render (manual spot check).
- [ ] **D2.2 — Ten built-in themes**
  - Do: add `rose-pine`, `tokyo-night` (exists), `catppuccin` (mocha), `catppuccin-latte`, `gruvbox` (exists), `everforest`, `nord`, `dracula`, `kanagawa`, `carbon`. Palettes from each project's published spec — no invented colors. Cite source hex tables in a comment.
  - Verify: `pomogo themes` CLI lists all with color swatches; sweep test renders each.
- [ ] **D2.3 — Random & daily selection**
  - Do: `theme = "random"` → seeded from time+pid at launch; `"daily"` → seeded from `YYYY-MM-DD` (theme-of-the-day). Same for `layout`. Statusline shows resolved name so screenshots are attributable. Key `T` cycles theme live, `L` cycles layout live (not persisted).
  - Verify: unit test daily determinism; manual: three launches with random ≠ constant.
- [ ] **D2.4 — Layout system**
  - Do: `render.Layout` = `func(ds DisplayState, th *theme.Theme, f Frame) string` + registry.
    Implement: **classic** (current box), **minimal** (timer + task + thin bar, no border), **centered** (huge clock, everything else whisper-quiet), **compact** (≤ 8 rows, for splits/corner terminals), **retro** (double-line border, `▒` textures). Each declares min width/height; model picks nearest fitting layout when the terminal is too small instead of the "terminal too small" wall (that message only below the global minimum, 40×10).
  - Verify: golden test per layout (both modes, running+paused+idle); sweep test across sizes.
- [ ] **D2.5 — External themes**
  - Do: load `~/.config/pomogo/themes/*.toml` (fields = Theme struct, kebab-case) at startup; register by filename stem; user wins collisions; malformed file → warning in `doctor`, skipped at runtime. Ship `contrib/themes/example.toml`.
  - Verify: drop example into config dir → `pomogo themes` lists it → `theme = "example"` renders.
- [ ] **D2.6 — Typography & polish pass**
  - Do: one deliberate pass over all layouts: spacing rhythm, label casing (quiet uppercase for labels, sentence case for values), muted hint bar, consistent glyph set (`━ ─ ● ○ ✦`). Update README screenshots (retake with `contrib/demo.tape`).
  - Verify: side-by-side before/after screenshots in the PR; goldens updated once, intentionally.

---

# Phase D3 — Flagship: screenshot mode & ambient effects

**Goal:** the two shareability features.
**Exit gate:** `S` produces a clean frame in every theme×layout; effects on = CPU still < 1 % (measured).

- [ ] **D3.1 — Screenshot mode**
  - Do: `S` toggles `ds.Zen`: layouts drop hints, status line, keybind bar, mode picker残; keep project/task/mode label/timer/progress. Auto-exits on any other key. Works in every layout (each layout implements its zen variant — usually just omitting sections).
  - Verify: golden zen frame per layout; manual screenshot session.
- [ ] **D3.2 — Ambient effects engine**
  - Do: `render/ambient.go`: deterministic particle field (`stars`, `snow`, `rain`) computed from `(seed, tickCount, width, height)` — **pure function, no state, no goroutine, advances only on the existing 1 s tick**. Composite: render field as background lines, overlay the content box at its `Place` offset (replace `lipgloss.Place` with a small compositor that knows the box rect). Particles use `theme.Ambient` (very muted). Config `effects = "none"|"stars"|"snow"|"rain"|"random"`, default `none`; key `e` cycles live.
  - Verify: unit test determinism (same inputs → same frame); `contrib/bench.sh` with effects on: CPU delta < 0.3 %; no allocation regression > 20 % in a render benchmark.
- [ ] **D3.3 — Session recap**
  - Do: on `BlockEnded` (deep) or long-break completion (quick), show recap screen: total focused time, segments completed, breaks, pauses (from `blocks.pauses`), streak line. One accent line of celebration, no confetti spam. `enter` dismisses → idle. Recap also printable: `pomogo recap` shows the last block.
  - Verify: fake-clock integration test drives a block to completion → recap DisplayState correct; golden frame.

---

# Phase D4 — Statistics & identity details

**Goal:** stats worthy of the rest; small identity touches.
**Exit gate:** stats screen + CLI show deep-work hours; all pure-function tested.

- [ ] **D4.1 — Deep-work aware stats**
  - Do: extend `stats.Calculate`: today/week/lifetime **hours** (not just session counts), blocks completed, per-day week bars scaled by minutes (cap bar width — current code renders one `█` per session, unbounded). Lifetime line: `482h · 917 sessions · 48 day streak`.
  - Verify: table-driven tests; `pomogo stats` output reviewed.
- [ ] **D4.2 — Stats screen redesign**
  - Do: re-render stats screen through the render package with theme roles: Today / Week (ASCII bars) / Lifetime / Recent sessions. Respect zen mode (hides it).
  - Verify: golden test; fits 80×24.
- [ ] **D4.3 — Focus level heuristic**
  - Do: per-block score 1–10 from simple arithmetic: start at 10, −1 per pause, −1 per skipped break, −2 per abandoned segment. Shown on recap + stats (`Deep ■■■■■■■ 7/10`). Pure function, documented formula, no cleverness.
  - Verify: unit tests over the formula table.
- [ ] **D4.4 — Project icons**
  - Do: optional `icon` column on projects (migration 6); settable via `pomogo projects add <name> --icon 🐹` and shown as `🐹 Fawz` in layouts. Free-text (any emoji/nerd-font glyph), width-measured with lipgloss so alignment survives.
  - Verify: store test; layout golden with icon set.
- [ ] **D4.5 — Session titles**
  - Do: work segments display a verb label instead of "Work": from task keywords (`build/fix/debug/read/write/learn/design/research` → "Building", …) with manual override key `v` cycling the list. Fallback: "Focusing". Pure mapping + tests.
  - Verify: unit tests; golden shows `Building` for task "build auth".

---

# Phase D5 — Developer features, docs & release

**Goal:** the dev-identity layer and shipping.
**Exit gate:** README rewrite merged; release artifacts build; benchmarks within budget.

- [ ] **D5.1 — Git awareness (`internal/devinfo`)**
  - Do: at launch and on project change, read `.git/HEAD` upward from cwd (**file read only, never exec git**): branch name (or short SHA when detached). Layouts show it muted under project (` feature/timer`). Config `show_git = true` default; absent repo → absent line.
  - Verify: unit tests with fixture `.git` dirs (branch, detached, none); zero exec calls (grep).
- [ ] **D5.2 — tmux presence**
  - Do: if `$TMUX` set, read session name (`tmux display-message -p '#S'` once at startup — single exec, cached). Optional line in classic/retro layouts; `show_tmux = false` default.
  - Verify: manual inside/outside tmux; no exec when `$TMUX` empty.
- [ ] **D5.3 — Statefile consumers refresh**
  - Do: update `contrib/waybar`, `contrib/tmux`, `contrib/starship`, `contrib/zsh` snippets to prefer v2 fields (`block_remaining_secs`) with v1 fallback; `pomogo status` formats gain deep-mode output (` 1:53:21`).
  - Verify: `integrations` tests cover v1 and v2 statefiles for every format.
- [ ] **D5.4 — Docs rewrite**
  - Do: README rewritten around the new identity — "A beautiful terminal deep-work companion for developers", not "Pomodoro timer": hero GIF (vhs `demo.tape`), theme gallery grid, layout gallery, screenshot-mode callout, honest performance numbers from `bench.sh`. Update `docs/02-architecture.md` to the target diagram above; `docs/03-ui.md` documents DisplayState/layout/theme contracts for contributors.
  - Verify: reviewed rendering on GitHub; all links/images resolve.
- [ ] **D5.5 — Packaging**
  - Do: GoReleaser config (GitHub Releases, linux amd64/arm64, darwin arm64), AUR `PKGBUILD` in `contrib/aur/`, Homebrew formula template, Scoop manifest. CI job builds release artifacts on tag.
  - Verify: `goreleaser release --snapshot --clean` succeeds; PKGBUILD `makepkg -si` in a clean chroot (or documented manual check).
- [ ] **D5.6 — Final performance gate**
  - Do: run `contrib/bench.sh`, compare to D0.2 baseline. Budget: RSS < 15 MB, CPU < 1 % (effects on), cold start < 50 ms, binary ≤ baseline + 15 %.
  - Verify: numbers table committed to `docs/06-research.md`; any breach fixed before tagging `v2.0.0`.

---

## Post-v2 candidates (not scheduled — do not start)

- Bubble Tea v2 / Bubbles v2 / Lip Gloss v2 migration (isolated PR, goldens make it safe)
- Plugin-style layout files (layouts stay Go-only until someone actually asks)
- Spotify / MPRIS now-playing line (explicitly deprioritized in `docs/00-product-vision.md`)

## Risk register

| Risk | Mitigation |
|---|---|
| Redesign breaks existing users' muscle memory / configs | Legacy config keys alias to `[quick_focus]`; default mode remains Quick; keymap unchanged except new keys (`d`, `S`, `T`, `L`, `e`, `v`) |
| Statefile change breaks waybar/tmux users silently | v2 is additive; D1.6 + D5.3 test both schema versions in every output format |
| Golden tests churn on every visual tweak | Goldens updated only in tasks that declare it (D2.1, D2.6); other tasks must not touch them |
| Ambient effects sneak in CPU cost | Effects derive from the existing tick; D3.2 and D5.6 measure; default off |
| `session` runner duplicates `timer` logic | Runner owns *sequencing* only; all timing stays in `timer.Session` — if you're writing deadline math in `session/`, stop |
| Scope creep toward productivity suite | Locked exclusions table; anything not in this file needs a new locked decision first |
