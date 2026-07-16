# Contributing to PomoGo

Thank you for your interest in contributing to PomoGo! This guide will help you get started.

## Philosophy

PomoGo follows a **one-idea-at-a-time** approach. We ship complete phases, not half-finished features. See [04-roadmap.md](04-roadmap.md) and [PIPELINE-V2.md](../PIPELINE-V2.md) for the full plan.

## Before You Start

- Read [00-product-vision.md](00-product-vision.md) to understand the why
- Read [02-architecture.md](02-architecture.md) to understand the structure
- Check [PIPELINE-V2.md](../PIPELINE-V2.md) to see what's being worked on

## Development Setup

### Prerequisites
- Go 1.26+ (or check [go.mod](../go.mod))
- Arch Linux / Omarchy (for testing integrations)
- `notify-send` (libnotify)

### Build & Test
```bash
git clone https://github.com/Ibnu-Afdel/pomogo
cd pomogo

# Build
go build ./...

# Test
go test ./... -v

# Lint (requires golangci-lint)
golangci-lint run

# Run
./pomogo
```

## Workflow

We use [PIPELINE-V2.md](../PIPELINE-V2.md) as the single source of truth for what to build.

### Starting Work

1. Check [PIPELINE-V2.md](../PIPELINE-V2.md) — find the first unchecked task in the active phase
2. **Work on one task ID at a time** (e.g., `P1.4`)
3. Create a feature branch: `git checkout -b P1.4-ui-timer`
4. Implement the task
5. Run the **Verify** command from the task
6. Commit with message: `P1.4: UI timer screen`
7. Mark the task done in `PIPELINE-V2.md`
8. Push and open a PR

### Code Style

- **Format:** `gofmt` (enforced by CI)
- **Lint:** `golangci-lint run`
- **Imports:** organize with `goimports`
- **Tests:** table-driven for pure logic, mocks for I/O
- **Docs:** comment exported functions and types

### Testing Guidelines

- **Pure logic** (`timer`, `config`, `stats`): ≥90% coverage, table-driven tests
- **I/O** (`statefile`, `notify`): use temp files or mocks
- **TUI** (`ui`): manual testing against quality bar (LazyGit/btop/Yazi)
- **Integrations:** tested on-device (Hyprland/Waybar/tmux)

Example table-driven test:
```go
func TestSessionTransitions(t *testing.T) {
  tests := []struct {
    name    string
    initial SessionState
    event   Event
    want    SessionState
  }{
    {"Start from Idle", StateIdle, EventStart, StateWork},
    {"Skip Work", StateWork, EventSkip, StateShortBreak},
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      s := &Session{State: tt.initial}
      got := s.Handle(tt.event)
      if got != tt.want {
        t.Errorf("got %v, want %v", got, tt.want)
      }
    })
  }
}
```

## Architecture Principles

1. **Pure state machine** — `internal/timer` has no I/O, only pure functions
2. **Dependency injection** — `Clock` interface injected, not `time.Now()` called directly
3. **Atomic writes** — state file writes via temp file + rename
4. **No global state** — all config injected into models
5. **Minimal dependencies** — only Charm stack, TOML, sqlite (Phase 2+)

## Documentation

- **Code comments:** explain the "why", not the "what"
- **Task docs:** update [02-architecture.md](02-architecture.md) as you change schema
- **User docs:** contributor-friendly docs live in `docs/`
- **Inline examples:** code snippets in `docs/` for major features

## Quality Bar

Your code should feel like:
- **LazyGit** — clean, minimal, scannable
- **btop** — high quality, zero clutter
- **Yazi** — responsive, keyboard-driven

If a UX component wouldn't make it into those projects, it doesn't belong here.

## Common Tasks

### Add a new config field

1. Update `internal/config/config.go` struct + TOML tags
2. Update [02-architecture.md](02-architecture.md) schema section
3. Update `.golangci.yml` if adding a linter
4. Write tests in `internal/config/config_test.go`
5. Update `README.md` and example config
6. Commit: `config: add <field_name>`

### Add a new subcommand

1. Add handler in `cmd/pomogo/main.go`
2. Add help text in `handleHelp()`
3. Write tests in `cmd/pomogo/main_test.go`
4. Update `contrib/completions/` (shell completion)
5. Commit: `cli: add <subcommand>`

### Add a new theme

1. Add to `internal/theme/theme.go`
2. Test visually in TUI
3. Update [03-ui.md](03-ui.md) with color codes
4. Commit: `theme: add <theme_name>`

## Commit Messages

Format: `P<phase>.<task>: <what>`

Examples:
- `P1.4: UI timer screen`
- `P2.1: SQLite schema + migrations`
- `P3.2: Waybar module integration`

## Opening a PR

1. Ensure CI passes (`go build`, `go vet`, `golangci-lint`, `go test`)
2. Update [PIPELINE-V2.md](../PIPELINE-V2.md) with task status
3. Reference the task ID in the PR title: `P1.4: UI timer screen`
4. Link the `PIPELINE-V2.md` task in the PR description
5. Wait for code review

## Questions?

- Open an issue for discussion
- Check existing closed issues for context
- Read the phase docs in `docs/`

## Code of Conduct

Be respectful, inclusive, and kind. This is a collaborative project.

---

**Happy coding! Let's build a beautiful focus companion.** 🍅
