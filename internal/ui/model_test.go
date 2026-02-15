package ui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/config"
	"github.com/samcharles93/commiter/internal/git"
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

func TestCommitSuccessWithRemainingChangesShowsContinuePrompt(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{})
	m.commitMsg = "feat: test"

	updated, cmd := m.Update(CommitSuccessMsg{
		HasRemainingChanges: true,
		RemainingFiles: []git.ChangedFile{
			{Path: "main.go", Status: "modified"},
		},
	})

	if cmd != nil {
		t.Fatalf("expected no command when entering continue prompt state")
	}

	model := updated.(Model)
	if model.state != StateContinueConfirm {
		t.Fatalf("expected state %q, got %q", StateContinueConfirm, model.state)
	}
	if len(model.files) != 1 || model.files[0].Path != "main.go" {
		t.Fatalf("expected remaining files to be stored in model")
	}
}

func TestContinueConfirmEnterStartsGeneratingWhenDiffExists(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{})
	m.state = StateContinueConfirm
	m.diff = "diff --git a/main.go b/main.go"
	m.history = []llm.Message{{Role: "assistant", Content: "old"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected generate command when continuing with staged diff")
	}

	model := updated.(Model)
	if model.state != StateGenerating {
		t.Fatalf("expected state %q, got %q", StateGenerating, model.state)
	}
	if len(model.history) != 0 {
		t.Fatalf("expected history to reset for next commit flow")
	}
}
