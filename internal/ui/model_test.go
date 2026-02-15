package ui

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	m := NewModel(stubProvider{}, nil, "diff --git a/file b/file", "openai", "gpt-4o", &config.Config{}, false)

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
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{}, false)
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
	if model.hookWarning != "" {
		t.Fatalf("expected no hook warning, got %q", model.hookWarning)
	}
}

func TestCommitSuccessStoresHookWarning(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{}, false)
	m.commitMsg = "feat: test"

	updated, _ := m.Update(CommitSuccessMsg{
		HookWarning: "### Post-commit hook failed\n\n```text\nboom\n```",
	})

	model := updated.(Model)
	if model.state != StateSuccess {
		t.Fatalf("expected state %q, got %q", StateSuccess, model.state)
	}
	if model.hookWarning == "" {
		t.Fatal("expected hook warning to be stored in model")
	}
}

func TestContinueConfirmEnterStartsGeneratingWhenDiffExists(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{}, false)
	m.state = StateContinueConfirm
	m.diff = "diff --git a/main.go b/main.go"
	m.history = []llm.Message{{Role: "assistant", Content: "old"}}
	m.hookWarning = "warning"

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
	if model.hookWarning != "" {
		t.Fatalf("expected hook warning to be reset, got %q", model.hookWarning)
	}
}

func TestModelWindowResizeWithMarkdownDoesNotPanic(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{}, false)
	m.summary = "### Summary\n\n- one\n- two"
	m.state = StateSummary

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Update panicked on markdown resize: %v", r)
		}
	}()

	_, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
}

func TestRenderSuccessIncludesHookWarning(t *testing.T) {
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", &config.Config{}, false)
	m.state = StateSuccess
	m.commitMsg = "feat: test"
	m.hookWarning = "### Post-commit hook failed\n\n```text\nboom\n```"

	view := m.View()
	if !strings.Contains(view, "Hook warning:") {
		t.Fatalf("expected hook warning section in success view, got:\n%s", view)
	}
}

func TestTemplateSelectionPersistsDefaultAndGeneratesWithStagedDiff(t *testing.T) {
	home := t.TempDir()
	setHomeForModelTest(t, home)

	cfg := &config.Config{
		Templates: config.GetDefaultTemplates(),
	}
	m := NewModel(stubProvider{}, nil, "diff --git a/file b/file", "openai", "gpt-4o", cfg, false)
	if m.state != StateTemplateSelection {
		t.Fatalf("expected state %q, got %q", StateTemplateSelection, m.state)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected generate command after selecting template")
	}

	model := updated.(Model)
	if model.state != StateGenerating {
		t.Fatalf("expected state %q, got %q", StateGenerating, model.state)
	}
	if model.template == nil {
		t.Fatal("expected selected template to be stored")
	}
	if cfg.DefaultTemplate == "" {
		t.Fatal("expected default template to be persisted in config")
	}

	data, err := os.ReadFile(filepath.Join(home, config.ConfigFileName))
	if err != nil {
		t.Fatalf("read saved config: %v", err)
	}

	var saved map[string]any
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("parse saved config: %v", err)
	}
	if saved["default_template"] != cfg.DefaultTemplate {
		t.Fatalf("expected saved default_template %q, got %v", cfg.DefaultTemplate, saved["default_template"])
	}
}

func TestNewModelUsesConfiguredDefaultTemplate(t *testing.T) {
	cfg := &config.Config{
		Templates:       config.GetDefaultTemplates(),
		DefaultTemplate: "Simple",
	}

	withDiff := NewModel(stubProvider{}, nil, "diff --git a/file b/file", "openai", "gpt-4o", cfg, false)
	if withDiff.state != StateGenerating {
		t.Fatalf("expected staged-diff flow to start in %q, got %q", StateGenerating, withDiff.state)
	}
	if withDiff.template == nil || withDiff.template.Name != "Simple" {
		t.Fatalf("expected default template %q to be used, got %+v", "Simple", withDiff.template)
	}

	files := []git.ChangedFile{{Path: "main.go", Status: "modified"}}
	noDiff := NewModel(stubProvider{}, files, "", "openai", "gpt-4o", cfg, false)
	if noDiff.state != StateFileSelection {
		t.Fatalf("expected unstaged flow to skip template prompt and start in %q, got %q", StateFileSelection, noDiff.state)
	}
}

func setHomeForModelTest(t *testing.T, home string) {
	t.Helper()

	originalHome, hadHome := os.LookupEnv("HOME")
	originalUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("set HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", home); err != nil {
		t.Fatalf("set USERPROFILE: %v", err)
	}

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})
}
