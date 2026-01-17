package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/olamilekan000/etcd-tui/internal/config"
	"github.com/olamilekan000/etcd-tui/internal/tui/constants"
	"github.com/olamilekan000/etcd-tui/internal/tui/model"
)

var (
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "etcd-tui",
		Short: "A terminal UI for etcd",
		Long: `etcd-tui is a terminal user interface for interacting with etcd.

Configuration can be provided via:
  - Config file: ~/.etcd-tui/config.json (or path specified with --config)
  - Environment variables: ETCDCTL_ENDPOINTS, ETCDCTL_CACERT, etc.`,
		Run: runTUI,
	}

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: ~/.etcd-tui/config.json)")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(constants.Version)
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) {
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
