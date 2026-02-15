package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/samcharles93/commiter/internal/config"
	"github.com/samcharles93/commiter/internal/ui"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure commiter settings",
	Long:  `Interactive configuration for commiter. Edit provider, API key, model, and other settings.`,
	RunE:  runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	// Load current config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: error loading config: %v\n", err)
		cfg = &config.Config{}
	}

	// Create and run config TUI
	m := ui.NewConfigModel(cfg)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running config TUI: %w", err)
	}

	return nil
}
