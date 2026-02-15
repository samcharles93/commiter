package main

import (
	"github.com/spf13/cobra"
)

var (
	// Flags
	modelFlag     string
	providerFlag  string
	bypassMode    bool
	customMessage string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "commiter",
	Short: "AI-powered git commit message generator",
	Long: `Commiter is a TUI application that generates intelligent commit messages
using LLMs (DeepSeek, OpenAI, etc.). It analyzes your git changes and creates
conventional, meaningful commit messages.`,
	SilenceUsage: true,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "The model to use (e.g., gpt-5, deepseek-chat)")
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "The provider to use (deepseek or openai)")
	rootCmd.PersistentFlags().BoolVarP(&bypassMode, "bypass", "y", false, "Bypass interactive mode and commit immediately")
	rootCmd.PersistentFlags().StringVarP(&customMessage, "message", "m", "", "Use custom commit message (skips LLM generation)")

	// Set the run function
	rootCmd.RunE = runStart

	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(historyCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
