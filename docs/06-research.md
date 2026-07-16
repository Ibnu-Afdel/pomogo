# Research Pass Findings

Research notes on Bubble Tea, Lip Gloss, text input components, and layout/design inspiration from terminal user interfaces.

## 1. Bubble Tea v1.x Tick Patterns
*   **Mechanism**: Repeating tick events in Bubble Tea must be scheduled using `tea.Cmd`. We return a tick command during model initialization (`Init`) and whenever the model handles a tick message (`tickMsg`), we return the next tick command to schedule the subsequent tick.
*   **Utility**: `tea.Tick(duration, func(time.Time) tea.Msg)` is standard and simple. In v1.x, recurring events can also be driven by `tea.Every(duration, fn)`, but for a 1-second focus timer tick, `tea.Tick` is reliable, lightweight, and standard.
*   **Idle CPU Control**: Ticks must only be scheduled when the timer is active (`m.session.IsRunning && !m.session.IsPaused`). When the session pauses or completes, we stop returning the tick command. This ensures zero CPU consumption during idle states.

## 2. Lip Gloss Place & Whitespace Options
*   **Library Version**: The codebase is using `github.com/charmbracelet/lipgloss v1.1.0`.
*   **API Usage**: Lip Gloss v1.x supports adding whitespace styling to layout boxes. Specifically, we can chain properties on `lipgloss.Style`:
    *   `WithWhitespaceChars(string)`: Defines the character sequence used to fill background/empty space (e.g. `.` or ` ` or particle glyphs).
    *   `WithWhitespaceForeground(lipgloss.Color)`: Styles the color of these background characters.
*   **Particle compositor**: In D3.2, rather than drawing particles directly, we can composite a particle field by setting background characters on the Lip Gloss styles, or by using a custom compositor that merges the particle matrix with the content box.

## 3. Bubbles Progress & TextInput
*   **Status**: `github.com/charmbracelet/bubbles v1.0.0` is imported.
*   **TextInput**: The textinput component maintains focused state and handles cursor movements. It is updated in our loop with `m.textInput, cmd = m.textInput.Update(msg)`.
*   **Progress**: We compute our progress percentage manually as a float64 `[0.0, 1.0]` and render it into a progress bar character string. This gives us full visual control over filled vs. empty segments without standard widget overhead.

## 4. Reference TUI Visual Layouts
*   **LazyGit**:
    *   *Lessons*: Clean panels, keyboard-driven navigation with visible and consistent key bindings on the bottom row. Quiet uppercase labels.
    *   *Application*: We will replace the static help text with a unified `KeyMap` struct that drives both the input handlers and the help overlays dynamically.
*   **btop**:
    *   *Lessons*: Rich, vibrant theme palettes, big clock indicators, rounded borders, high text contrast.
    *   *Application*: Keep the big ASCII-art characters for the time display, and use theme-specific accent colors for borders to guide user focus.
*   **glow**:
    *   *Lessons*: Excellent margin management and layout centering using Lip Gloss padding/margins.
    *   *Application*: Use whitespace margins as design elements to prevent terminal text clutter.

## Decisions Affected

*   **Ambient Particle Engine Substrate**: Confirmed. Using Lip Gloss styling APIs (`WithWhitespaceChars` and `WithWhitespaceForeground`) combined with a deterministic layout compositor is the ideal way to implement ambient particle fields (stars, snow, rain) without spawning heavy drawing goroutines or causing high CPU consumption.
*   **Decoupled Display State**: We will create a `render.DisplayState` transfer object that layouts will consume, making them completely independent of the Bubble Tea model.

## Baseline Metrics (D0.2)

Recorded on 2026-07-15 using `contrib/bench.sh`:

| Metric | Target / Budget | Pre-Refactor Baseline (D0.2) |
|---|---|---|
| **Binary Size** | (none, reference size) | 12.81 MB (13,439,786 bytes) |
| **Cold-Start Time** | < 50 ms | 4 ms |
| **Idle RSS** | < 15 MB | 14.26 MB (14,608 KB) |
| **Idle CPU %** | < 1 % | 0.0 % |

## Final Metrics (D5.6)

Recorded on 2026-07-16 using `contrib/bench.sh`:

| Metric | Budget | Baseline (D0.2) | Final (D5.6) | Status |
|---|---|---|---|---|
| **Binary Size** | ≤ 14.73 MB | 12.81 MB | 13.10 MB | Pass |
| **Cold-Start Time** | < 50 ms | 4 ms | 4 ms | Pass |
| **Idle RSS** | < 15 MB | 14.26 MB | 13.25 MB | Pass |
| **Idle CPU %** | < 1 % | 0.0 % | 0.0 % | Pass |

