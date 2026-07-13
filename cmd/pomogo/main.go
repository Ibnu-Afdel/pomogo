package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/ui"
)

// Version is injected at build time via -ldflags "-X main.Version=..."
var Version = "dev"

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
	fmt.Printf("pomogo version %s\n", Version)
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
	fmt.Println("  help               Show this help message")
	fmt.Println()
	fmt.Println("Run 'pomogo' to start the timer.")
	fmt.Println()
	fmt.Println("Configuration file: ~/.config/pomogo/config.toml")
	fmt.Println("State file: $XDG_RUNTIME_DIR/pomogo/state.json")
}

func init() {
	flag.Parse()
}
