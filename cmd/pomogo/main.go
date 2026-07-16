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
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/ui"
	"github.com/charmbracelet/lipgloss"
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
	case "themes":
		handleThemes()
	case "recap":
		handleRecap()
	case "status":
		handleStatus()
	case "completion":
		handleCompletion()
	case "projects":
		handleProjects()
	case "start":
		handleStart()
	case "doctor":
		handleDoctor()
	case "export":
		handleExport()
	case "report":
		handleReport()
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
	fmt.Println("  themes             List all available color themes with swatches")
	fmt.Println("  recap              Show summary of the last completed block")
	fmt.Println("  status             Show current session status")
	fmt.Println("  completion         Generate shell completion scripts")
	fmt.Println("  projects           Manage focus projects")
	fmt.Println("  start [profile|prj] Start timer with a specific profile or project")
	fmt.Println("  doctor             Check system dependencies and configuration health")
	fmt.Println("  export             Export focus session history to JSON or CSV format")
	fmt.Println("  report             Generate a weekly focus report in Markdown")
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
		fmt.Println("Weekly Focus Activity (mins):")
		fmt.Println("-----------------------------")
		for _, wd := range s.WeekDays {
			barLen := wd.Minutes / 10
			if barLen > 15 {
				barLen = 15
			}
			bar := ""
			if barLen > 0 {
				bar = strings.Repeat("█", barLen)
			} else {
				bar = "░"
			}
			fmt.Printf("%s  %-15s %dm\n", wd.Date.Format("Mon"), bar, wd.Minutes)
		}
		fmt.Println()
		fmt.Printf("Total Week Completed: %d sessions (%d mins)\n", s.WeekCount, s.WeekMinutes)
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
	fmt.Printf("Today Completed:    %d sessions (%d mins focused)\n", s.TodayCount, s.TodayMinutes)
	fmt.Printf("Current Streak:     %d days\n", s.CurrentStreak)
	fmt.Printf("Best Streak:        %d days\n", s.BestStreak)
	fmt.Printf("Monthly Completed:  %d sessions (Rate: %.0f%%)\n", s.MonthCount, s.CompletionRate*100)
	fmt.Printf("Lifetime Completed: %d sessions (%dh %dm focused)\n", s.LifetimeSessions, s.LifetimeMinutes/60, s.LifetimeMinutes%60)
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

func handleThemes() {
	_ = theme.LoadExternalThemes()
	themes := theme.List()
	fmt.Println("Available Color Themes:")
	fmt.Println("-----------------------")
	for _, name := range themes {
		t := theme.Get(name)
		swatchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Work.String()))
		swatch := swatchStyle.Render("■ Work")

		breakStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Break.String()))
		brk := breakStyle.Render("■ Break")

		accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent.String()))
		acc := accentStyle.Render("■ Accent")

		fmt.Printf("  %-16s %s  %s  %s  - %s\n", t.Name, swatch, brk, acc, t.Description)
	}
}

func handleRecap() {
	st, err := store.New(config.DBFilePath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	b, err := st.GetLastBlock()
	if err != nil {
		fmt.Println("No completed blocks found.")
		return
	}

	modeStr := "Quick Focus"
	if b.Mode == "deep" {
		modeStr = "Deep Focus"
	}

	hrs := b.PlannedSecs / 3600
	mins := (b.PlannedSecs % 3600) / 60
	secs := b.PlannedSecs % 60
	var timeStr string
	if hrs > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hrs, mins)
	} else if mins > 0 {
		timeStr = fmt.Sprintf("%dm %ds", mins, secs)
	} else {
		timeStr = fmt.Sprintf("%ds", secs)
	}

	statusStr := "Completed"
	if !b.Completed {
		statusStr = "Abandoned"
	}

	fmt.Println("Last Focus Cycle Recap:")
	fmt.Println("-----------------------")
	fmt.Printf("  Mode:            %s\n", modeStr)
	fmt.Printf("  Status:          %s\n", statusStr)
	fmt.Printf("  Started At:      %s\n", b.StartedAt.Local().Format("2006-01-02 15:04"))
	fmt.Printf("  Planned Time:    %s\n", timeStr)
	fmt.Printf("  Pauses Taken:    %d\n", b.Pauses)
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
			fmt.Println("Usage: pomogo projects add <name> [--icon <icon>] [color]")
			os.Exit(1)
		}
		name := os.Args[3]
		icon := ""
		color := ""

		args := os.Args[4:]
		for i := 0; i < len(args); i++ {
			if args[i] == "--icon" && i+1 < len(args) {
				icon = args[i+1]
				i++
			} else if color == "" {
				color = args[i]
			}
		}

		p := &store.Project{Name: name, Color: color, Icon: icon}
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

		displayName := p.Name
		if p.Icon != "" {
			displayName = p.Icon + " " + p.Name
		}

		if p.Archived {
			archived = append(archived, displayName+colorSuffix)
		} else {
			active = append(active, displayName+colorSuffix)
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
	soundEvent := ""
	// 1. Check if profile matches
	if cfg.Profiles != nil {
		if _, exists := cfg.Profiles[target]; exists {
			cfg, project, soundEvent = cfg.ResolveProfile(target)
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
	if soundEvent != "" {
		model.SetCustomSoundEvent(soundEvent)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func handleDoctor() {
	fmt.Println("Checking PomoGo system health...")
	fmt.Println()
	diags := integrations.RunDoctor()
	allPassed := true
	for _, d := range diags {
		status := "✔"
		if !d.Passed {
			status = "✘"
			allPassed = false
		}
		fmt.Printf("[%s] %-40s : %s\n", status, d.Name, d.Message)
	}
	fmt.Println()
	if allPassed {
		fmt.Println("All systems normal! PomoGo is fully operational.")
	} else {
		fmt.Println("Some warnings/errors were detected. Please check the reports above.")
	}
}

func handleExport() {
	exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
	format := exportCmd.String("format", "json", "Output format: json or csv")
	startStr := exportCmd.String("start", "", "Start date (YYYY-MM-DD)")
	endStr := exportCmd.String("end", "", "End date (YYYY-MM-DD)")

	if err := exportCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Defaults: export last 365 days
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now().Add(24 * time.Hour)

	if *startStr != "" {
		if t, err := time.Parse("2006-01-02", *startStr); err == nil {
			start = t
		} else {
			fmt.Println("Invalid start date format. Use YYYY-MM-DD.")
			os.Exit(1)
		}
	}
	if *endStr != "" {
		if t, err := time.Parse("2006-01-02", *endStr); err == nil {
			end = t
		} else {
			fmt.Println("Invalid end date format. Use YYYY-MM-DD.")
			os.Exit(1)
		}
	}

	st, err := store.New(config.DBFilePath())
	if err != nil {
		log.Fatalf("Error: failed to open store: %v", err)
	}
	defer st.Close()

	output, err := st.ExportSessions(*format, start, end)
	if err != nil {
		log.Fatalf("Error: export failed: %v", err)
	}

	fmt.Print(output)
}

func handleReport() {
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	startStr := reportCmd.String("start", "", "Start date (YYYY-MM-DD)")
	endStr := reportCmd.String("end", "", "End date (YYYY-MM-DD)")

	if err := reportCmd.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Defaults: report last 7 days
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	if *startStr != "" {
		if t, err := time.Parse("2006-01-02", *startStr); err == nil {
			start = t
		} else {
			fmt.Println("Invalid start date format. Use YYYY-MM-DD.")
			os.Exit(1)
		}
	}
	if *endStr != "" {
		if t, err := time.Parse("2006-01-02", *endStr); err == nil {
			end = t
		} else {
			fmt.Println("Invalid end date format. Use YYYY-MM-DD.")
			os.Exit(1)
		}
	}

	st, err := store.New(config.DBFilePath())
	if err != nil {
		log.Fatalf("Error: failed to open store: %v", err)
	}
	defer st.Close()

	report, err := st.GenerateMarkdownReport(start, end)
	if err != nil {
		log.Fatalf("Error: report generation failed: %v", err)
	}

	fmt.Print(report)
}
