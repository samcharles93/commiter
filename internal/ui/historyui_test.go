package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/git"
)

func TestHistoryDetailMarkdownRenderAndResize(t *testing.T) {
	m := NewHistoryModel([]git.CommitInfo{
		{
			Hash:    "0123456789abcdef",
			Author:  "Test User",
			Date:    "today",
			Subject: "feat: summary",
			Body:    "- item one\n- item two",
		},
	})

	m.state = historyStateDetail
	m.commitDetail = &git.CommitInfo{
		Hash:    "0123456789abcdef",
		Author:  "Test User",
		Date:    "today",
		Subject: "feat: summary",
		Body:    "- item one\n- item two",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("history view panicked on resize/render: %v", r)
		}
	}()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model := updated.(HistoryModel)
	view := model.View()
	if !strings.Contains(view, "Commit Details") {
		t.Fatalf("expected commit detail header in view, got:\n%s", view)
	}
}
