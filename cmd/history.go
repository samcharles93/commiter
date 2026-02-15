package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/samcharles93/commiter/internal/git"
	"github.com/samcharles93/commiter/internal/ui"
)

var (
	historyLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Browse commit history",
	Long:  `Browse and search recent commit history with diffs.`,
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 50, "Number of commits to show")
}

func runHistory(cmd *cobra.Command, args []string) error {
	// Get commit history
	commits, err := git.GetCommitHistory(historyLimit)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No commits found.")
		return nil
	}

	// Create and run history TUI
	m := ui.NewHistoryModel(commits)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running history browser: %w", err)
	}

	return nil
}
