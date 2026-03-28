package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mrf/dcc/internal/app"
	"github.com/mrf/dcc/internal/config"
)

// version is set via -ldflags at build time.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	showConfigPath := flag.Bool("config-path", false, "print resolved config file path and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	if *showConfigPath {
		fmt.Println(config.Path())
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create model
	model := app.NewModel(cfg)

	// Create and run program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
