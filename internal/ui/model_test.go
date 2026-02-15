package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/config"
	"github.com/samcharles93/commiter/internal/llm"
)

type stubProvider struct{}

func (stubProvider) GenerateMessage(context.Context, string, []llm.Message, *config.CommitTemplate) (string, error) {
	return "test commit", nil
}

func (stubProvider) SummarizeChanges(context.Context, string) (string, error) {
	return "summary", nil
}

func TestModelWindowResizeWithStagedDiffDoesNotPanic(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "diff --git a/file b/file", "openai", "gpt-4o", &config.Config{})

	runResize := func(msg tea.WindowSizeMsg) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Update panicked on window resize: %v", r)
			}
		}()

		_, _ = m.Update(msg)
	}

	runResize(tea.WindowSizeMsg{Width: 120, Height: 40})
	runResize(tea.WindowSizeMsg{Width: 1, Height: 1})
}
