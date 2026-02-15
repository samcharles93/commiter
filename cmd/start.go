package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/samcharles93/commiter/internal/config"
	"github.com/samcharles93/commiter/internal/git"
	"github.com/samcharles93/commiter/internal/llm"
	"github.com/samcharles93/commiter/internal/ui"
)

const (
	DefaultDeepSeekAPIURL = "https://api.deepseek.com/v1/chat/completions"
	DefaultOpenAIAPIURL   = "https://api.openai.com/v1/chat/completions"
)

func runStart(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Warning: error loading config: %v\n", err)
		cfg = &config.Config{}
	}

	// Determine provider
	providerName := "deepseek"
	if cfg.Provider != "" {
		providerName = cfg.Provider
	}
	if providerFlag != "" {
		providerName = providerFlag
	}

	// Get API key, model, and baseURL based on provider
	var apiKey, model, baseURL string
	switch strings.ToLower(providerName) {
	case "openai":
		apiKey = os.Getenv("OPENAI_API_KEY")
		model = "gpt-4o"
		baseURL = DefaultOpenAIAPIURL
	case "deepseek":
		apiKey = os.Getenv("DEEPSEEK_API_KEY")
		model = "deepseek-chat"
		baseURL = DefaultDeepSeekAPIURL
	default:
		return fmt.Errorf("unknown provider %q", providerName)
	}

	// Override with config values
	if cfg.APIKey != "" {
		apiKey = cfg.APIKey
	}
	if cfg.Model != "" {
		model = cfg.Model
	}
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}

	// Override with flags
	if modelFlag != "" {
		model = modelFlag
	}

	// Check if bypass mode
	if bypassMode {
		return runBypassMode(args, apiKey, model, baseURL, providerName)
	}

	// Interactive mode
	return runInteractiveMode(apiKey, model, baseURL, providerName, cfg)
}

func runBypassMode(files []string, apiKey, model, baseURL, providerName string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified for bypass mode")
	}

	// Stage files
	if err := git.StageFiles(files); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Get diff
	diff, err := git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get staged diff: %w", err)
	}

	if len(diff) == 0 {
		return fmt.Errorf("no changes to commit")
	}

	// Generate or use custom message
	var message string
	if customMessage != "" {
		message = customMessage
	} else {
		if apiKey == "" {
			return fmt.Errorf("API key for %s not found", providerName)
		}

		provider := llm.NewGenericProvider(apiKey, model, baseURL)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		message, err = provider.GenerateMessage(ctx, string(diff), nil, nil)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}
	}

	// Commit
	if err := git.Commit(message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Print success
	fmt.Println("✓ Committed:")
	fmt.Println(message)

	return nil
}

func runInteractiveMode(apiKey, model, baseURL, providerName string, cfg *config.Config) error {
	// Get staged diff
	diff, err := git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	var files []git.ChangedFile
	if len(diff) == 0 {
		// No staged changes, list unstaged
		unstaged, err := git.ListUnstagedChanges()
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}
		if len(unstaged) == 0 {
			fmt.Println("No changes detected.")
			return nil
		}
		files = unstaged
	}

	if apiKey == "" {
		return fmt.Errorf("API key for %s not found", providerName)
	}

	// Create provider and model
	provider := llm.NewGenericProvider(apiKey, model, baseURL)
	m := ui.NewModel(provider, files, string(diff), providerName, model, cfg)

	// Run TUI
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
}
