package components

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/samcharles93/commiter/internal/git"
)

var (
	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	selectedTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	normalTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFF"))

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// FileItemDelegate is a custom delegate for file items with checkbox support
type FileItemDelegate struct{}

// NewFileItemDelegate creates a new file item delegate
func NewFileItemDelegate() FileItemDelegate {
	return FileItemDelegate{}
}

func (d FileItemDelegate) Height() int                             { return 1 }
func (d FileItemDelegate) Spacing() int                            { return 0 }
func (d FileItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d FileItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	file, ok := item.(git.ChangedFile)
	if !ok {
		return
	}

	checkbox := "[ ]"
	if file.Selected {
		checkbox = "[✓]"
	}
	checkbox = checkboxStyle.Render(checkbox)

	title := file.Path
	desc := file.Status

	titleStyle := normalTitleStyle
	if index == m.Index() {
		titleStyle = selectedTitleStyle
	}

	str := fmt.Sprintf("%s %s %s",
		checkbox,
		titleStyle.Render(title),
		descStyle.Render("("+desc+")"),
	)

	fmt.Fprint(w, str)
}
