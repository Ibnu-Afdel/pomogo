package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/integrations"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/stats"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/ui"
)

// Build-time variables injected via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		handleDefault()
		return
	}

	switch os.Args[1] {
	case "version":
		handleVersion()
	case "config":
		handleConfig()
	case "stats":
		handleStats()
	case "history":
		handleHistory()
	case "status":
		handleStatus()
	case "completion":
		handleCompletion()
	case "projects":
		handleProjects()
	case "start":
		handleStart()
	case "help", "-h", "--help":
		handleHelp()
	default:
		handleDefault()
	}
}

func handleDefault() {
	// Load config (use defaults if file doesn't exist)
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// Launch TUI
	model := ui.NewModel(cfg)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func handleVersion() {
	if Version == "dev" {
		fmt.Printf("pomogo %s\n", Version)
	} else {
		fmt.Printf("pomogo %s (commit %s, built %s)\n", Version, Commit, Date)
	}
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: pomogo config [init]")
		fmt.Println()
		fmt.Println("Subcommands:")
		fmt.Println("  init    Create a default config file")
		return
	}

	switch os.Args[2] {
	case "init":
		if err := config.WriteDefault(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config file created at: %s\n", config.ConfigFilePath())
	default:
		fmt.Fprintf(os.Stderr, "Unknown config subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func handleHelp() {
	fmt.Println("Usage: pomogo [COMMAND]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  version            Show version information")
	fmt.Println("  config init        Create a default config file")
	fmt.Println("  stats              Show focus statistics")
	fmt.Println("  history            Show detailed session history")
	fmt.Println("  status             Show current session status")
	fmt.Println("  completion         Generate shell completion scripts")
	fmt.Println("  projects           Manage focus projects")
	fmt.Println("  start [profile|prj] Start timer with a specific profile or project")
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Run 'pomogo' to start the timer.")
	fmt.Println()
	fmt.Println("Configuration file: ~/.config/pomogo/config.toml")
	fmt.Println("State file: $XDG_RUNTIME_DIR/pomogo/state.json")
}

func handleStats() {
	statsCmd := flag.NewFlagSet("stats", flag.ExitOnError)
	weekFlag := statsCmd.Bool("week", false, "Show weekly activity")
	monthFlag := statsCmd.Bool("month", false, "Show monthly summary")

	if err := statsCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	st, err := store.New(config.DBFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	sessions, err := st.GetSessions(start, now.Add(24*time.Hour))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}

	s := stats.Calculate(sessions, now)

	if *weekFlag {
		fmt.Println("Weekly Focus Activity:")
		fmt.Println("----------------------")
		for _, wd := range s.WeekDays {
			bar := ""
			if wd.Count > 0 {
				bar = strings.Repeat("█", wd.Count)
			} else {
				bar = "░"
			}
			fmt.Printf("%s  %-10s %d\n", wd.Date.Format("Mon"), bar, wd.Count)
		}
		fmt.Println()
		fmt.Printf("Total Week Completed: %d sessions\n", s.WeekCount)
		return
	}

	if *monthFlag {
		fmt.Println("Monthly Focus Summary:")
		fmt.Println("----------------------")
		fmt.Printf("This Month Completed: %d sessions\n", s.MonthCount)
		fmt.Printf("Focus Completion Rate: %.0f%%\n", s.CompletionRate*100)
		return
	}

	fmt.Println("PomoGo Focus Statistics:")
	fmt.Println("------------------------")
	fmt.Printf("Today Completed:   %d sessions (%d mins focused)\n", s.TodayCount, s.TodayMinutes)
	fmt.Printf("Current Streak:    %d days\n", s.CurrentStreak)
	fmt.Printf("Best Streak:       %d days\n", s.BestStreak)
	fmt.Printf("Monthly Completed: %d sessions (Rate: %.0f%%)\n", s.MonthCount, s.CompletionRate*100)
}

func handleHistory() {
	st, err := store.New(config.DBFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	sessions, err := st.GetSessions(start, now.Add(24*time.Hour))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Recent Work Sessions:")
	fmt.Println("---------------------")
	count := 0
	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		if s.Type != "work" {
			continue
		}
		status := "completed"
		if !s.Completed {
			status = "skipped"
		}
		
		task := s.Task
		if task == "" {
			task = "[no task]"
		}
		
		timeStr := s.StartedAt.Local().Format("2006-01-02 15:04")
		fmt.Printf("%s  %-15s (%s)", timeStr, task, status)
		if s.Note != "" {
			fmt.Printf(" - %s", s.Note)
		}
		fmt.Println()
		count++
		if count >= 20 { // show last 20 work sessions
			break
		}
	}
	if count == 0 {
		fmt.Println("No work sessions recorded yet.")
	}
}

func handleStatus() {
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	formatFlag := statusCmd.String("format", "default", "Output format: default, waybar, tmux, json")

	if err := statusCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	manager, err := statefile.NewManager()
	if err != nil {
		out, _ := integrations.FormatStatus(nil, *formatFlag)
		fmt.Println(out)
		return
	}

	state, err := manager.Read()
	if err != nil {
		out, _ := integrations.FormatStatus(nil, *formatFlag)
		fmt.Println(out)
		return
	}

	out, err := integrations.FormatStatus(state, *formatFlag)
	if err != nil {
		out, _ = integrations.FormatStatus(nil, *formatFlag)
	}
	fmt.Println(out)
}

func handleCompletion() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: pomogo completion [bash|zsh|fish]")
		return
	}

	shell := os.Args[2]
	switch shell {
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "bash":
		fmt.Print(bashCompletionScript)
	case "fish":
		fmt.Print(fishCompletionScript)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s. Supported: bash, zsh, fish\n", shell)
		os.Exit(1)
	}
}

const zshCompletionScript = `#compdef pomogo

_pomogo() {
    local -a subcmds
    subcmds=(
        'version:Show version information'
        'config:Manage configuration'
        'stats:Show focus statistics'
        'history:Show detailed session history'
        'status:Show current session status'
        'completion:Generate shell completions'
        'help:Show help message'
    )
    
    _arguments \
        '1: :->cmds' \
        '*:: :->args'

    case "$state" in
        cmds)
            _describe -t commands 'pomogo commands' subcmds
            ;;
        args)
            case "$words[1]" in
                config)
                    _values 'config subcommands' 'init:Create a default config file'
                    ;;
                stats)
                    _arguments '--week[Show weekly activity]' '--month[Show monthly summary]'
                    ;;
                completion)
                    _values 'shells' 'zsh:Zsh completion' 'bash:Bash completion' 'fish:Fish completion'
                    ;;
                status)
                    _arguments '--format[Output format]:format:(default waybar tmux json)'
                    ;;
            esac
            ;;
    esac
}

compdef _pomogo pomogo
`

const bashCompletionScript = `_pomogo_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="version config stats history completion help status"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi

    case "${prev}" in
        config)
            COMPREPLY=( $(compgen -W "init" -- ${cur}) )
            return 0
            ;;
        stats)
            COMPREPLY=( $(compgen -W "--week --month" -- ${cur}) )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- ${cur}) )
            return 0
            ;;
        status)
            COMPREPLY=( $(compgen -W "--format" -- ${cur}) )
            return 0
            ;;
    esac
}
complete -F _pomogo_completion pomogo
`

const fishCompletionScript = `complete -c pomogo -f
complete -c pomogo -n "not __fish_seen_subcommand_from version config stats history completion help status" -a "version config stats history completion help status"
complete -c pomogo -n "__fish_seen_subcommand_from config" -a "init"
complete -c pomogo -n "__fish_seen_subcommand_from stats" -l week -d "Show weekly activity"
complete -c pomogo -n "__fish_seen_subcommand_from stats" -l month -d "Show monthly summary"
complete -c pomogo -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
complete -c pomogo -n "__fish_seen_subcommand_from status" -l format -d "Output format (default, waybar, tmux, json)"
`

func init() {
	flag.Parse()
}

func handleProjects() {
	st, err := store.New(config.DBFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open store: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if len(os.Args) < 3 {
		listProjects(st)
		return
	}

	switch os.Args[2] {
	case "list":
		listProjects(st)
	case "add":
		if len(os.Args) < 4 {
			fmt.Println("Usage: pomogo projects add <name> [color]")
			os.Exit(1)
		}
		name := os.Args[3]
		color := ""
		if len(os.Args) >= 5 {
			color = os.Args[4]
		}
		p := &store.Project{Name: name, Color: color}
		if err := st.CreateProject(p); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project '%s' added successfully.\n", name)
	case "archive":
		if len(os.Args) < 4 {
			fmt.Println("Usage: pomogo projects archive <name>")
			os.Exit(1)
		}
		name := os.Args[3]
		if err := st.ArchiveProject(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Project '%s' archived successfully.\n", name)
	default:
		fmt.Fprintf(os.Stderr, "Unknown projects subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func listProjects(st *store.Store) {
	projects, err := st.GetProjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load projects: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return
	}

	var active []string
	var archived []string
	for _, p := range projects {
		colorSuffix := ""
		if p.Color != "" {
			colorSuffix = fmt.Sprintf(" (%s)", p.Color)
		}
		if p.Archived {
			archived = append(archived, p.Name+colorSuffix)
		} else {
			active = append(active, p.Name+colorSuffix)
		}
	}

	fmt.Println("Active Projects:")
	if len(active) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, a := range active {
			fmt.Printf("  - %s\n", a)
		}
	}

	if len(archived) > 0 {
		fmt.Println("\nArchived Projects:")
		for _, a := range archived {
			fmt.Printf("  - %s (archived)\n", a)
		}
	}
}

func handleStart() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: pomogo start [profile|project]")
		os.Exit(1)
	}
	target := os.Args[2]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	project := ""
	// 1. Check if profile matches
	if cfg.Profiles != nil {
		if _, exists := cfg.Profiles[target]; exists {
			cfg, project = cfg.ResolveProfile(target)
		}
	}

	// 2. If it wasn't resolved as a profile, check if it matches a project in the database
	if project == "" {
		st, err := store.New(config.DBFilePath())
		if err == nil {
			if p, err := st.GetProjectByName(target); err == nil && p != nil {
				project = p.Name
			}
			st.Close()
		}
	}

	// If it matched neither a profile nor a project, we treat the target as the active project name
	if project == "" {
		project = target
	}

	// Launch TUI with resolved config and project
	model := ui.NewModel(cfg)
	model.SetProjectByName(project)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
