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

func init() {
	flag.Parse()
}
