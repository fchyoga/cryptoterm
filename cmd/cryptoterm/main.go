package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fchyoga/cryptoterm/internal/config"
	"github.com/fchyoga/cryptoterm/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	Version = "1.1.0"
)

func main() {
	showVer := flag.Bool("v", false, "Show version")
	showVerLong := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVer || *showVerLong {
		fmt.Printf("cryptoterm v%s\n", Version)
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	appModel := ui.NewUIModel(cfg)
	p := tea.NewProgram(
		appModel,
		tea.WithAltScreen(),       // Use alternate terminal screen buffer
		tea.WithMouseCellMotion(), // Support mouse scrolling
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Terminal runtime error: %v\n", err)
		os.Exit(1)
	}
}
