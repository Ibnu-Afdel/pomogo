# Architecture: PomoGo — Technical Design

## High-Level Overview

PomoGo is a **state machine wrapped in a TUI**, with integrations consuming a state file.

```
┌─────────────────────────────────────────────────────┐
│                    PomoGo (TUI)                    │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │   Bubble Tea v2 Model                       │   │
│  │   ├─ Session State Machine (pure)           │   │
│  │   ├─ Keybinds (s, space, n, r, q, ?)       │   │
│  │   └─ Lip Gloss Rendering                    │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  On every state change:                             │
│  ├─ Write state.json                               │
│  ├─ Send notify-send                               │
│  └─ Update config (persist task/project)           │
│                                                     │
└─────────────────────────────────────────────────────┘
         ↓ (writes JSON + shell-outs)
┌─────────────────────────────────────────────────────┐
│         Integration Layer (Phase 3+)                │
│                                                     │
│  pomogo status --format waybar  → Waybar           │
│  pomogo status --format tmux    → tmux status-right│
│  pomogo stats --week            → shell/prompt     │
│  notify-send (custom)           → Mako/dunst       │
│  loginctl (D-Bus)               → lock detection   │
│                                                     │
└─────────────────────────────────────────────────────┘
         ↓ (reads from)
┌─────────────────────────────────────────────────────┐
│     Persistent Storage                              │
│                                                     │
│  state.json (runtime, Phase 1)                      │
│  └─ $XDG_RUNTIME_DIR/pomogo/state.json              │
│                                                     │
│  config.toml (config, Phase 1)                      │
│  └─ ~/.config/pomogo/config.toml                    │
│                                                     │
│  pomogo.db (history, Phase 2+)                      │
│  └─ ~/.local/share/pomogo/pomogo.db                 │
│                                                     │
└─────────────────────────────────────────────────────┘
```

## Database Schema (pomogo.db)

Located at `~/.local/share/pomogo/pomogo.db`.

PomoGo uses a standard CGo-free SQLite database. The schema is versioned using a migrations table:

### `schema_migrations` Table
- `version` INTEGER PRIMARY KEY (tracks applied migration versions)

### `sessions` Table
- `id` INTEGER PRIMARY KEY AUTOINCREMENT
- `type` TEXT NOT NULL (type of session: e.g., "work")
- `task` TEXT (optional active task description)
- `note` TEXT (optional user notes recorded upon completion)
- `started_at` DATETIME NOT NULL (start time in local/UTC format)
- `ended_at` DATETIME (completion/skip time)
- `completed` INTEGER NOT NULL (1 for completed focus, 0 for skipped)
- `duration_secs` INTEGER NOT NULL (duration in seconds)

## Folder Structure

```
pomogo/
├── cmd/pomogo/main.go              # Entrypoint: subcommand dispatch
├── internal/
│   ├── timer/                      # State machine (pure Go, no I/O)
│   │   └── timer.go
│   ├── ui/                         # Bubble Tea TUI
│   │   ├── model.go                # BubbleTea Model
│   │   ├── view.go                 # View/Render logic
│   │   └── keybinds.go             # Keyboard handlers
│   ├── theme/                      # Lip Gloss theming
│   │   └── theme.go
│   ├── config/                     # TOML config
│   │   └── config.go
│   ├── statefile/                  # state.json I/O
│   │   └── statefile.go
│   ├── store/                      # SQLite (Phase 2)
│   │   └── store.go
│   ├── notify/                     # notify-send shell-out
│   │   └── notify.go
│   └── integrations/               # (Phase 3)
│       ├── waybar.go
│       ├── tmux.go
│       └── status.go
├── docs/                           # Full docs tree
├── contrib/                        # Scripts, configs, desktop files
│   ├── waybar/
│   ├── tmux/
│   ├── completions/
│   ├── pomogo.desktop
│   └── aur/
├── .github/workflows/ci.yml
├── .golangci.yml
├── go.mod / go.sum
├── PIPELINE-V2.md
└── README.md
```

## State Machine Design

### States

```
Idle
  ↓ (start) →
Work (25 min, configurable)
  ↓ (complete or skip) →
ShortBreak (5 min, configurable)
  ├─ (return to Work after N short breaks) →
  └─ LongBreak (15 min, after every N=4 work sessions)
  ↓ (complete) →
Work (next session)
```

### Events

- **`start`** → Idle → Work
- **`pause`** → Work/Break → Paused
- **`resume`** → Paused → Work/Break
- **`skip`** → Work/Break → next state immediately
- **`complete`** → Work/Break → next state (mark as completed)
- **`reset`** → any → Idle

### Transition Rules

1. **Timer fires** → advance state (play notification)
2. **User pauses** → freeze remaining time (wall-clock independent)
3. **Resume after suspend** → wall-clock jump detection; auto-pause + notify
4. **Lock detected** → pause + notify if configured
5. **App closes** → state file remains; relaunch resumes

## State File Schema (state.json)

Located at `$XDG_RUNTIME_DIR/pomogo/state.json` (fallback `~/.cache/pomogo/state.json`).

Atomic writes via temp file + rename.

```json
{
  "state": "work",
  "session_type": "work",
  "session_count": 3,
  "started_at": 1783945800,
  "ends_at": 1783947300,
  "paused": false,
  "remaining_secs": 600,
  "total_secs": 1500,
  "pid": 12345,
  "updated_at": 1783946700
}
```

Fields are intentionally minimal in Phase 1. Future versions may add task,
project, theme, and schema version fields; integrations must ignore fields they
do not understand.

### Contract for Integrations

Integrations MUST:
1. Read this file atomically (copy + parse, or lock)
2. Ignore missing fields (forward/backward compat)
3. Assume missing file = idle state
4. Poll at 1 Hz or use inotify (cheap)
5. Never write to this file

Example Waybar module:
```json
{ "text": "🍅 18:24", "class": "pomogo-work", "tooltip": "Work session ends at 15:55" }
```

## Config File Schema (config.toml)

Located at `~/.config/pomogo/config.toml`.

Sane defaults built in; file is optional.

```toml
# Durations (minutes)
work_duration = 25
short_break_duration = 5
long_break_duration = 15
sessions_before_long_break = 4

# Theme: tokyo-night, catppuccin, gruvbox
theme = "tokyo-night"

# Notifications
notifications_enabled = true
sound_enabled = false

# Projects (Phase 4)
[projects]
backend = { color = "blue" }
frontend = { color = "purple" }

# Profiles (Phase 4)
[profiles.coding]
work_duration = 50
short_break_duration = 5
theme = "tokyo-night"

[profiles.studying]
work_duration = 40
short_break_duration = 10
theme = "catppuccin"
sound_enabled = true
```

## Package Interfaces

### `timer` (Internal, Pure)

```go
type Session struct {
  State SessionState
  RemainingTime time.Duration
  TotalTime time.Duration
  ...
}

type SessionState int
const (
  StateIdle SessionState = iota
  StateWork
  StateShortBreak
  StateLongBreak
)

type Clock interface {
  Now() time.Time
}

func (s *Session) Tick(clock Clock)
func (s *Session) Start(clock Clock, duration time.Duration) error
func (s *Session) Pause()
func (s *Session) Resume()
func (s *Session) Skip() SessionState
func (s *Session) IsRunning() bool
```

### `config` (Internal, I/O)

```go
type Profile struct {
  WorkDuration            *int
  ShortBreakDuration      *int
  LongBreakDuration       *int
  SessionsBeforeLongBreak *int
  Theme                   *string
  Project                 *string
  SoundEvent              *string
}

type Config struct {
  WorkDuration            int
  ShortBreakDuration      int
  LongBreakDuration       int
  SessionsBeforeLongBreak int
  Theme                   string
  Profiles                map[string]Profile
  ...
}

func Load() (*Config, error)
func (c *Config) ResolveProfile(name string) (*Config, string, string)
```

### `statefile` (Internal, I/O)

```go
type StateSnapshot struct {
  State string
  EndsAt time.Time
  RemainingSeconds int
  ...
}

func Write(path string, snapshot *StateSnapshot) error
func Read(path string) (*StateSnapshot, error)
```

### `notify` (Internal, Shell-out)

```go
func Notify(title, body string, urgency string) error
// Shells out to notify-send if available; no error if absent.
```

### `store` (Internal, I/O)

```go
type Session struct {
  ID           int64
  Type         string
  Task         string
  Note         string
  StartedAt    time.Time
  EndedAt      time.Time
  Completed    bool
  DurationSecs int
  ProjectID    *int64
  ProjectName  string
}

type Project struct {
  ID       int64
  Name     string
  Color    string
  Archived bool
}

func New(dbPath string) (*Store, error)
func (s *Store) Close() error
func (s *Store) SaveSession(sess *Session) error
func (s *Store) GetSessions(start, end time.Time) ([]*Session, error)
func (s *Store) CreateProject(p *Project) error
func (s *Store) GetProjectByName(name string) (*Project, error)
func (s *Store) GetProjects() ([]*Project, error)
func (s *Store) ArchiveProject(name string) error
```

### `stats` (Internal, Pure)

```go
type DayStats struct {
  Date  time.Time
  Count int
}

type Stats struct {
  TodayCount     int
  TodayMinutes   int
  WeekCount      int
  MonthCount     int
  CurrentStreak  int
  BestStreak     int
  CompletionRate float64
  WeekDays       [7]DayStats
}

func Calculate(sessions []*store.Session, now time.Time) *Stats
```

### `integrations` (Internal, I/O)

```go
type WaybarOutput struct {
  Text    string
  Class   string
  Tooltip string
}

func FormatStatus(state *statefile.State, format string) (string, error)
func IsSessionLocked() (bool, error)
```



## Performance Budget

- **Startup:** < 50 ms (measured via `time ./pomogo`)
- **Idle CPU:** ≈ 0% (use `times.AfterFunc()` or `context.WithTimeout()`, never busy-loop)
- **State file writes:** Only on transitions/pauses, not per tick
- **Memory:** < 10 MB resident

## Testing Strategy

- **Pure logic** (`timer`, `config`, `stats`): table-driven unit tests, ≥90% coverage
- **I/O** (`statefile`, `notify`): integration tests with temp files / mocks
- **TUI** (`ui`): manual visual testing against quality bar (LazyGit/btop/Yazi)
- **Integrations:** tested on-device under Hyprland/Waybar/tmux

## Dependencies

**Core (Phase 1):**
- `github.com/charmbracelet/bubbletea` v2
- `github.com/charmbracelet/bubbles` v2
- `github.com/charmbracelet/lipgloss` (included in Tea)
- `github.com/pelletier/go-toml/v2` (config)

**Phase 2:**
- `modernc.org/sqlite` (CGo-free SQLite)

**Phase 3:**
- `github.com/godbus/dbus/v5` (D-Bus for lock detection + Mako actions)

**Explicitly NOT included:**
- Cobra (too opinionated; use stdlib flag + manual dispatch)
- Viper (too complex; direct TOML load)
- gRPC, protobuf (no RPC layer needed)

## Versioning & Compatibility

- **Semantic Versioning:** `major.minor.patch`
- **Phase gates:** each phase ships a tagged release (`v0.1.0`, `v0.2.0`, etc.)
- **State file schema:** versioned field in JSON; migrations handled in statefile pkg
- **Config schema:** new fields have defaults; old files work unmodified
- **Database schema:** embedded migrations via `embed` package; forward-only

## Next Steps

- Phase 1: Implement timer state machine with full test coverage
- Phase 2: Add SQLite schema + session recording
- Phase 3: Implement status formatters for integrations
