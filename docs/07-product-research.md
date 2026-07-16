# Product Research: Deep-Work Companion Direction

This note freezes the product direction before the next implementation pass. PomoGo is not
"a Pomodoro app with themes"; it is a lightweight terminal companion for developers who want
focus, flow, and a screen they are happy to keep open.

## Core Promise

> PomoGo is a beautiful, lightweight terminal deep-work companion for developers. Pomodoro is
> the internal engine, not the product.

The user should think in modes:

- **Quick Focus:** start a 25 minute focus session. When work ends, the 5 minute break continues
  naturally as part of the classic cycle. This mode is for low-friction work.
- **Deep Focus:** start a 1-4 hour block. PomoGo classifies the block behind the scenes into
  work and break segments, but the main display is the whole block countdown. The user should
  feel like they are in a deep-work block, not babysitting Pomodoro intervals.

## What Must Feel True

- The app is calm enough to leave open all day.
- The screen is visually distinctive enough to screenshot.
- The main timer always answers the user's mental model: "how much focus block is left?"
- The implementation remains tiny: no daemon, no telemetry, no network, no Electron, no busy loops.
- The aesthetics come from typography, spacing, palettes, layouts, and restrained motion, not clutter.

## Aesthetic Rules

Keep the current layout variants, but add more only when each one has a different job:

- **Classic:** balanced default, readable in a normal terminal.
- **Minimal:** split-pane / corner-terminal mode.
- **Centered:** screenshot-first, large timer, quiet metadata.
- **Compact:** very small terminal mode.
- **Retro:** characterful but still readable.
- **Next variants to explore:** dashboard, monolith, glass, terminal-rice, focus-card, tinybar.

Every new layout must pass these checks:

- Works at 80x24 and 120x32.
- Has a zen/screenshot variant.
- Does not duplicate an existing layout with only different borders.
- Does not add visual noise around the timer.
- Uses stable dimensions so timer ticks do not shift the frame.

Theme expansion should follow published palettes where possible. Candidate theme families:

- Night Owl
- One Dark
- Ayu Mirage
- Monokai Pro
- Solarized Dark
- GitHub Dark
- Oxocarbon
- Palenight
- Iceberg
- Hyper

## Research Notes

The current stack remains the right default.

- **Bubble Tea** is still the best fit for the application shell because it is designed for
  full-window terminal apps, uses the Elm-style update/view model, and includes high-quality
  keyboard handling and rendering.
- **Lip Gloss** should stay the layout/styling layer. Do not hand-roll ANSI styling except in
  small pure helpers.
- **Bubbles** should be used for focused inputs and controls, not for the main timer render if
  custom rendering is cleaner.
- **termenv** is already in the Charm ecosystem and useful through Lip Gloss for color profile
  detection and graceful color degradation. Prefer it over custom hex-to-ANSI maps.
- **go-runewidth / uniseg** are good candidates for project icons, emoji, and wide glyph
  alignment. Use one consistent width helper before adding more icon-heavy layouts.
- **VHS** is the right external tool for README demos and regression-friendly visual captures.
- **Freeze** is useful for static terminal screenshots, but it should remain a contributor tool,
  not a runtime dependency.

Avoid switching to a different TUI framework right now. Libraries like `tview` are strong for
widget-heavy dashboards, but PomoGo's best surface is a custom-rendered focus scene. A framework
switch would create churn without improving the core experience.

## Tooling Decisions

Runtime dependencies should stay conservative:

- Keep Bubble Tea, Lip Gloss, Bubbles, TOML, and modernc SQLite.
- Add a width/segmentation helper only when project icons or new layouts prove current alignment
  is insufficient.
- Do not add table/dashboard libraries until stats becomes too complex for pure render helpers.
- Do not add markdown rendering to the TUI. Docs are outside the app; the app should stay focused.

Development tools worth using:

- `contrib/bench.sh` for performance gates.
- `vhs` for demo GIFs and visual marketing.
- Golden render tests for every layout/theme-sensitive change.
- `go test ./...` as the minimum correctness gate.
- `goreleaser release --snapshot --clean` before packaging changes.

## Product Gaps Before More Features

The next implementation pass should focus on quality, not breadth:

1. Make the mode model exact: Quick Focus cycles naturally; Deep Focus classifies internally.
2. Fix statefile block identity so restore can reconnect to an active deep block.
3. Finish config semantics for `[quick_focus]` and `[deep_focus]`.
4. Add more layouts/themes only after a visual target sheet exists.
5. Reduce UI/model complexity before adding more visual variants.
6. Produce real screenshots/GIFs and judge the product by how it looks in a terminal, not by
   feature count.

## Sources Checked

- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Bubbles: https://github.com/charmbracelet/bubbles
- termenv: https://github.com/muesli/termenv
- go-runewidth: https://github.com/mattn/go-runewidth
- uniseg: https://github.com/rivo/uniseg
- VHS: https://github.com/charmbracelet/vhs
- Freeze: https://github.com/charmbracelet/freeze
- tview: https://github.com/rivo/tview
