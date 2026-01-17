package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/olamilekan000/etcd-tui/internal/config"
	"github.com/olamilekan000/etcd-tui/internal/tui/model"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file (default: ~/.etcd-tui/config.json)")
	flag.StringVar(&configPath, "c", "", "Path to config file (shorthand)")
	flag.Parse()

	if configPath != "" {
		config.SetConfigPath(configPath)
	}

	p := tea.NewProgram(
		model.New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
