package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/git"
	"github.com/samcharles93/commiter/internal/ui/components"
)

// HistoryModel is the TUI model for browsing commit history
type HistoryModel struct {
	state        string
	commits      []git.CommitInfo
	list         list.Model
	diffViewer   components.DiffViewer
	selectedHash string
	commitDetail *git.CommitInfo
	commitDiff   string
	showDiff     bool
	filterInput  textinput.Model
	filtering    bool
	err          error
}

const (
	historyStateList   = "list"
	historyStateDetail = "detail"
	historyStateFilter = "filter"
	historyStateError  = "error"
)

type commitItem struct {
	git.CommitInfo
}

func (c commitItem) Title() string {
	return fmt.Sprintf("%s - %s", c.Hash[:8], c.Subject)
}

func (c commitItem) Description() string {
	return fmt.Sprintf("%s by %s", c.Date, c.Author)
}

func (c commitItem) FilterValue() string {
	return c.Subject + " " + c.Author + " " + c.Hash
}

// NewHistoryModel creates a new history browser model
func NewHistoryModel(commits []git.CommitInfo) HistoryModel {
	items := make([]list.Item, len(commits))
	for i, c := range commits {
		items[i] = commitItem{c}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedFileStyle
	delegate.Styles.SelectedDesc = SubtleStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "Commit History"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	filterInput := textinput.New()
	filterInput.Placeholder = "Search commits..."
	filterInput.CharLimit = 100
	filterInput.Width = 60

	return HistoryModel{
		state:       historyStateList,
		commits:     commits,
		list:        l,
		diffViewer:  components.NewDiffViewer(),
		filterInput: filterInput,
	}
}

func (m HistoryModel) Init() tea.Cmd {
	return nil
}

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case historyStateList:
			switch msg.String() {
			case "enter", " ":
				// View commit details
				selected := m.list.SelectedItem()
				if selected != nil {
					commit := selected.(commitItem)
					m.selectedHash = commit.Hash

					// Fetch commit details
					detail, diff, err := git.GetCommitDetails(commit.Hash)
					if err != nil {
						m.state = historyStateError
						m.err = err
						return m, nil
					}

					m.commitDetail = detail
					m.commitDiff = diff
					m.diffViewer.SetContent(diff)
					m.showDiff = true
					m.state = historyStateDetail
					return m, nil
				}
			case "/":
				// Start filtering
				m.filtering = true
				m.filterInput.Focus()
				m.state = historyStateFilter
				return m, textinput.Blink
			case "q", "esc":
				return m, tea.Quit
			}

		case historyStateDetail:
			switch msg.String() {
			case "d":
				// Toggle diff view
				m.showDiff = !m.showDiff
				return m, nil
			case "y":
				// Copy hash to clipboard (would need clipboard library)
				// For now, just show a message
				// TODO: Implement clipboard copy
				return m, nil
			case "q", "esc":
				m.state = historyStateList
				return m, nil
			}

		case historyStateFilter:
			switch msg.String() {
			case "enter":
				// Apply filter
				filter := strings.TrimSpace(m.filterInput.Value())
				if filter != "" {
					m.filterCommits(filter)
				}
				m.filtering = false
				m.filterInput.Blur()
				m.state = historyStateList
				return m, nil
			case "esc":
				m.filtering = false
				m.filterInput.Blur()
				m.state = historyStateList
				return m, nil
			}

		case historyStateError:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := BoxStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
		m.diffViewer.SetSize(msg.Width, msg.Height-10)
	}

	// Update child components
	switch m.state {
	case historyStateList:
		m.list, cmd = m.list.Update(msg)
	case historyStateDetail:
		if m.showDiff {
			m.diffViewer, cmd = m.diffViewer.Update(msg)
		}
	case historyStateFilter:
		m.filterInput, cmd = m.filterInput.Update(msg)
	}

	return m, cmd
}

func (m HistoryModel) View() string {
	switch m.state {
	case historyStateList:
		return m.renderList()
	case historyStateDetail:
		return m.renderDetail()
	case historyStateFilter:
		return m.renderFilter()
	case historyStateError:
		return m.renderError()
	}
	return ""
}

func (m HistoryModel) renderList() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("📜 Commit History") + "\n\n")
	b.WriteString(m.list.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓: navigate • enter/space: view details • /: filter • q: quit"))
	return b.String()
}

func (m HistoryModel) renderDetail() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("📝 Commit Details") + "\n\n")

	if m.commitDetail != nil {
		b.WriteString(SubtleStyle.Render("Hash: ") + m.commitDetail.Hash + "\n")
		b.WriteString(SubtleStyle.Render("Author: ") + m.commitDetail.Author + "\n")
		b.WriteString(SubtleStyle.Render("Date: ") + m.commitDetail.Date + "\n\n")
		b.WriteString(BoxStyle.Render(m.commitDetail.Subject+"\n\n"+m.commitDetail.Body) + "\n\n")
	}

	if m.showDiff {
		b.WriteString(SubtleStyle.Render("Diff:") + "\n")
		b.WriteString(m.diffViewer.View() + "\n")
	}

	diffToggle := "[d] show diff"
	if m.showDiff {
		diffToggle = "[d] hide diff"
	}

	b.WriteString(HelpStyle.Render(diffToggle + " • [y] copy hash • q/esc: back"))
	return b.String()
}

func (m HistoryModel) renderFilter() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🔍 Filter Commits") + "\n\n")
	b.WriteString(SubtleStyle.Render("Enter search term:") + "\n\n")
	b.WriteString(m.filterInput.View() + "\n\n")
	b.WriteString(HelpStyle.Render("enter: apply filter • esc: cancel"))
	return b.String()
}

func (m HistoryModel) renderError() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(ErrorBoxStyle.Render(m.err.Error()) + "\n")
	return b.String()
}

func (m *HistoryModel) filterCommits(filter string) {
	filter = strings.ToLower(filter)
	var filtered []list.Item

	for _, commit := range m.commits {
		if strings.Contains(strings.ToLower(commit.Subject), filter) ||
			strings.Contains(strings.ToLower(commit.Author), filter) ||
			strings.Contains(strings.ToLower(commit.Hash), filter) {
			filtered = append(filtered, commitItem{commit})
		}
	}

	m.list.SetItems(filtered)
}
