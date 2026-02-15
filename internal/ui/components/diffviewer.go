package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	diffStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4"))
)

// DiffViewer is a component for viewing diffs with syntax highlighting
type DiffViewer struct {
	viewport viewport.Model
	content  string
	ready    bool
}

// NewDiffViewer creates a new diff viewer
func NewDiffViewer() DiffViewer {
	vp := viewport.New(80, 20)
	vp.Style = diffStyle
	return DiffViewer{
		viewport: vp,
	}
}

// SetContent sets the content to display
func (d *DiffViewer) SetContent(content string) {
	d.content = content
	// Apply basic syntax highlighting (can be enhanced with Chroma later)
	highlighted := highlightDiff(content)
	d.viewport.SetContent(highlighted)
}

// SetSize sets the viewport size
func (d *DiffViewer) SetSize(width, height int) {
	d.viewport.Width = width - 4
	d.viewport.Height = height - 4
	d.ready = true
	if d.content != "" {
		highlighted := highlightDiff(d.content)
		d.viewport.SetContent(highlighted)
	}
}

// Update updates the diff viewer
func (d DiffViewer) Update(msg tea.Msg) (DiffViewer, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			d.viewport.GotoTop()
		case "G":
			d.viewport.GotoBottom()
		case "k", "up":
			d.viewport.ScrollUp(1)
		case "j", "down":
			d.viewport.ScrollDown(1)
		case "pgup":
			d.viewport.HalfPageUp()
		case "pgdown":
			d.viewport.HalfPageDown()
		}
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return d, cmd
}

// View renders the diff viewer
func (d DiffViewer) View() string {
	if !d.ready {
		return "Loading..."
	}
	return d.viewport.View()
}

// highlightDiff applies basic syntax highlighting to diff content
func highlightDiff(content string) string {
	var b strings.Builder
	lines := strings.Split(content, "\n")

	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n")
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			b.WriteString(addedStyle.Render(line))
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			b.WriteString(removedStyle.Render(line))
		} else if strings.HasPrefix(line, "@@") {
			b.WriteString(headerStyle.Render(line))
		} else if strings.HasPrefix(line, "diff --git") || strings.HasPrefix(line, "index") ||
			strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			b.WriteString(metaStyle.Render(line))
		} else {
			b.WriteString(line)
		}
	}

	return b.String()
}
