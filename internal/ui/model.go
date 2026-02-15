package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/config"
	"github.com/samcharles93/commiter/internal/git"
	"github.com/samcharles93/commiter/internal/llm"
	"github.com/samcharles93/commiter/internal/ui/components"
)

const successExitDelay = 1200 * time.Millisecond

// Model is the main TUI model
type Model struct {
	state         string
	provider      llm.Provider
	diff          string
	commitMsg     string
	summary       string
	history       []llm.Message
	err           error
	spinner       spinner.Model
	fileList      list.Model
	textarea      textarea.Model
	templateList  list.Model
	files         []git.ChangedFile
	templates     []config.CommitTemplate
	template      *config.CommitTemplate
	showHelp      bool
	providerName  string
	modelName     string
	confirmQuit   bool
	isAmending    bool
	previousState string
	diffViewer    components.DiffViewer
	lastCommit    *git.CommitInfo
}

// NewModel creates a new TUI model
func NewModel(provider llm.Provider, files []git.ChangedFile, diff string, providerName, modelName string, cfg *config.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SelectedFileStyle

	delegate := components.NewFileItemDelegate()
	fileList := list.New([]list.Item{}, delegate, 0, 0)
	fileList.Title = "Select Files to Stage"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(true)
	fileList.Styles.Title = TitleStyle

	var initialState string

	if len(diff) == 0 {
		items := make([]list.Item, len(files))
		for i, f := range files {
			items[i] = f
		}

		fileList.SetItems(items)

		// Check if we should show template selection first
		if len(cfg.Templates) > 0 {
			initialState = StateTemplateSelection
		} else {
			initialState = StateFileSelection
		}
	} else {
		initialState = StateGenerating
	}

	ta := textarea.New()
	ta.Placeholder = "Describe how to improve the commit message..."
	ta.Focus()

	// Setup template list if templates are configured
	var templateList list.Model
	var templates []config.CommitTemplate
	if len(cfg.Templates) > 0 {
		templates = cfg.Templates
		items := make([]list.Item, len(templates))
		for i, t := range templates {
			items[i] = templateItem{CommitTemplate: t}
		}

		delegate := list.NewDefaultDelegate()
		delegate.Styles.SelectedTitle = SelectedFileStyle
		delegate.Styles.SelectedDesc = SubtleStyle

		templateList = list.New(items, delegate, 0, 0)
		templateList.Title = "Select Commit Message Template"
		templateList.SetShowStatusBar(false)
		templateList.SetFilteringEnabled(false)
		templateList.Styles.Title = TitleStyle
	}

	m := Model{
		state:        initialState,
		provider:     provider,
		diff:         diff,
		spinner:      s,
		fileList:     fileList,
		textarea:     ta,
		templateList: templateList,
		files:        files,
		templates:    templates,
		providerName: providerName,
		modelName:    modelName,
		confirmQuit:  cfg.GetConfirmQuit(),
		diffViewer:   components.NewDiffViewer(),
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.state == StateGenerating {
		return tea.Batch(m.spinner.Tick, m.generateCommitMsg())
	}
	return m.spinner.Tick
}

func (m Model) generateCommitMsg() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		msg, err := m.provider.GenerateMessage(ctx, m.diff, m.history, m.template)
		return GenerateMsg{Message: msg, Err: err}
	}
}

func (m Model) generateSummary() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		summary, err := m.provider.SummarizeChanges(ctx, m.diff)
		return SummaryMsg{Summary: summary, Err: err}
	}
}

func (m Model) commitChanges() tea.Cmd {
	return func() tea.Msg {
		var err error
		if m.isAmending {
			err = git.AmendCommit(m.commitMsg)
		} else {
			err = git.Commit(m.commitMsg)
		}

		if err != nil {
			return CommitErrorMsg{Err: err}
		}

		remainingDiff, remainingFiles, err := m.collectRemainingChanges()
		if err != nil {
			// Commit already succeeded; skip follow-up prompt if we cannot inspect remaining changes.
			return CommitSuccessMsg{}
		}

		return CommitSuccessMsg{
			HasRemainingChanges: len(remainingDiff) > 0 || len(remainingFiles) > 0,
			RemainingDiff:       remainingDiff,
			RemainingFiles:      remainingFiles,
		}
	}
}

func (m Model) quitAfterSuccess() tea.Cmd {
	return tea.Tick(successExitDelay, func(time.Time) tea.Msg {
		return AutoQuitMsg{}
	})
}

func (m Model) collectRemainingChanges() (string, []git.ChangedFile, error) {
	stagedDiff, err := git.GetStagedDiff()
	if err != nil {
		return "", nil, err
	}
	if len(stagedDiff) > 0 {
		return string(stagedDiff), nil, nil
	}

	unstaged, err := git.ListUnstagedChanges()
	if err != nil {
		return "", nil, err
	}

	return "", unstaged, nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle space key for file selection BEFORE other processing
		if m.state == StateFileSelection && msg.String() == " " {
			idx := m.fileList.Index()
			if idx >= 0 && idx < len(m.files) {
				m.files[idx].Selected = !m.files[idx].Selected
				// Update the list items
				items := make([]list.Item, len(m.files))
				for i, f := range m.files {
					items[i] = f
				}
				m.fileList.SetItems(items)
			}
			return m, nil
		}

		if msg.String() == "?" && m.state != StateRefining {
			m.showHelp = !m.showHelp
			return m, nil
		}

		switch m.state {
		case StateTemplateSelection:
			switch msg.String() {
			case "enter":
				selected := m.templateList.SelectedItem()
				if selected != nil {
					tmpl := selected.(templateItem)
					m.template = &tmpl.CommitTemplate
					m.state = StateFileSelection
					return m, nil
				}
			case "q", "esc":
				if m.confirmQuit {
					m.previousState = m.state
					m.state = StateQuitConfirm
					return m, nil
				}
				return m, tea.Quit
			}

		case StateFileSelection:
			switch msg.String() {
			case "u":
				// Unselect all
				for i := range m.files {
					m.files[i].Selected = false
				}
				items := make([]list.Item, len(m.files))
				for i, f := range m.files {
					items[i] = f
				}
				m.fileList.SetItems(items)
				return m, nil
			case "enter":
				// Stage selected files, or current if none selected
				var paths []string
				hasSelected := false
				for _, f := range m.files {
					if f.Selected {
						paths = append(paths, f.Path)
						hasSelected = true
					}
				}

				if !hasSelected {
					// No files selected, stage current
					selected := m.fileList.SelectedItem()
					if selected != nil {
						file := selected.(git.ChangedFile)
						paths = []string{file.Path}
					}
				}

				if len(paths) > 0 {
					if err := git.StageFiles(paths); err != nil {
						m.state = StateError
						m.err = err
						return m, nil
					}

					diff, err := git.GetStagedDiff()
					if err != nil {
						m.state = StateError
						m.err = err
						return m, nil
					}

					m.diff = string(diff)
					m.state = StateGenerating
					return m, tea.Batch(m.spinner.Tick, m.generateCommitMsg())
				}
			case "a":
				paths := make([]string, len(m.files))
				for i, f := range m.files {
					paths[i] = f.Path
				}
				if err := git.StageFiles(paths); err != nil {
					m.state = StateError
					m.err = err
					return m, nil
				}

				diff, err := git.GetStagedDiff()
				if err != nil {
					m.state = StateError
					m.err = err
					return m, nil
				}

				m.diff = string(diff)
				m.state = StateGenerating
				return m, tea.Batch(m.spinner.Tick, m.generateCommitMsg())
			case "d":
				// Show diff preview for selected file
				selected := m.fileList.SelectedItem()
				if selected != nil {
					file := selected.(git.ChangedFile)
					diff, err := git.GetFileDiff(file.Path)
					if err != nil {
						m.state = StateError
						m.err = err
						return m, nil
					}
					m.diffViewer.SetContent(string(diff))
					m.previousState = m.state
					m.state = StateDiffPreview
					return m, nil
				}
			case "q", "esc":
				if m.confirmQuit {
					m.previousState = m.state
					m.state = StateQuitConfirm
					return m, nil
				}
				return m, tea.Quit
			}

		case StateReview:
			switch msg.String() {
			case "y":
				m.state = StateCommitting
				return m, tea.Batch(m.spinner.Tick, m.commitChanges())
			case "n":
				m.history = append(m.history, llm.Message{Role: "assistant", Content: m.commitMsg})
				m.history = append(m.history, llm.Message{Role: "user", Content: "Give me a different option."})
				m.state = StateGenerating
				return m, tea.Batch(m.spinner.Tick, m.generateCommitMsg())
			case "r":
				m.state = StateRefining
				m.textarea.Reset()
				m.textarea.Focus()
				return m, textarea.Blink
			case "s":
				m.state = StateSummary
				return m, tea.Batch(m.spinner.Tick, m.generateSummary())
			case "a":
				// Amend option
				if git.CanAmend() {
					lastCommit, err := git.GetLastCommit()
					if err != nil {
						m.state = StateError
						m.err = err
						return m, nil
					}
					m.lastCommit = lastCommit
					m.previousState = m.state
					m.state = StateAmendConfirm
					return m, nil
				}
			case "d":
				// Show full staged diff
				m.diffViewer.SetContent(m.diff)
				m.previousState = m.state
				m.state = StateDiffPreview
				return m, nil
			case "q", "esc":
				if m.confirmQuit {
					m.previousState = m.state
					m.state = StateQuitConfirm
					return m, nil
				}
				return m, tea.Quit
			}

		case StateRefining:
			switch msg.String() {
			case "esc":
				m.state = StateReview
				return m, nil
			case "ctrl+s":
				feedback := strings.TrimSpace(m.textarea.Value())
				if feedback == "" {
					m.state = StateReview
					return m, nil
				}
				m.history = append(m.history, llm.Message{Role: "assistant", Content: m.commitMsg})
				m.history = append(m.history, llm.Message{Role: "user", Content: feedback})
				m.state = StateGenerating
				return m, tea.Batch(m.spinner.Tick, m.generateCommitMsg())
			}

		case StateSummary:
			if msg.String() != "?" {
				m.state = StateReview
				return m, nil
			}

		case StateDiffPreview:
			switch msg.String() {
			case "q", "esc":
				m.state = m.previousState
				return m, nil
			}

		case StateAmendConfirm:
			switch msg.String() {
			case "y":
				m.isAmending = true
				m.state = StateCommitting
				return m, tea.Batch(m.spinner.Tick, m.commitChanges())
			case "n", "esc":
				m.isAmending = false
				m.state = m.previousState
				return m, nil
			}

		case StateQuitConfirm:
			switch msg.String() {
			case "y":
				return m, tea.Quit
			case "n", "esc":
				m.state = m.previousState
				return m, nil
			}

		case StateContinueConfirm:
			switch msg.String() {
			case "enter":
				m.isAmending = false
				m.lastCommit = nil
				m.summary = ""
				m.history = nil

				if len(m.diff) > 0 {
					m.state = StateGenerating
					return m, tea.Batch(m.spinner.Tick, m.generateCommitMsg())
				}

				if len(m.files) == 0 {
					return m, tea.Quit
				}

				items := make([]list.Item, len(m.files))
				for i := range m.files {
					m.files[i].Selected = false
					items[i] = m.files[i]
				}
				m.fileList.SetItems(items)

				if len(m.templates) > 0 && m.template == nil {
					m.state = StateTemplateSelection
				} else {
					m.state = StateFileSelection
				}
				return m, nil

			case "n", "esc", "q":
				return m, tea.Quit
			}

		case StateSuccess, StateError:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := BoxStyle.GetFrameSize()
		fileListWidth := msg.Width - h
		fileListHeight := msg.Height - v - 5
		if fileListWidth < 0 {
			fileListWidth = 0
		}
		if fileListHeight < 0 {
			fileListHeight = 0
		}

		m.fileList.SetSize(fileListWidth, fileListHeight)
		if len(m.templates) > 0 {
			m.templateList.SetSize(fileListWidth, fileListHeight)
		}
		textareaWidth := msg.Width - h - 4
		if textareaWidth < 0 {
			textareaWidth = 0
		}
		m.textarea.SetWidth(textareaWidth)
		m.textarea.SetHeight(8)
		m.diffViewer.SetSize(msg.Width, msg.Height-4)

	case GenerateMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
			return m, nil
		}
		m.commitMsg = msg.Message
		m.state = StateReview
		return m, nil

	case SummaryMsg:
		if msg.Err != nil {
			m.state = StateError
			m.err = msg.Err
			return m, nil
		}
		m.summary = msg.Summary
		return m, nil

	case CommitSuccessMsg:
		if msg.HasRemainingChanges {
			m.diff = msg.RemainingDiff
			m.files = msg.RemainingFiles
			m.state = StateContinueConfirm
			return m, nil
		}

		m.state = StateSuccess
		return m, m.quitAfterSuccess()

	case CommitErrorMsg:
		m.state = StateError
		m.err = msg.Err
		return m, nil

	case AutoQuitMsg:
		if m.state == StateSuccess {
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update child components
	switch m.state {
	case StateFileSelection:
		m.fileList, cmd = m.fileList.Update(msg)
		cmds = append(cmds, cmd)
	case StateTemplateSelection:
		m.templateList, cmd = m.templateList.Update(msg)
		cmds = append(cmds, cmd)
	case StateRefining:
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	case StateDiffPreview:
		m.diffViewer, cmd = m.diffViewer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.state {
	case StateTemplateSelection:
		return m.renderTemplateSelection()
	case StateFileSelection:
		return m.renderFileSelection()
	case StateGenerating:
		return m.renderGenerating()
	case StateReview:
		return m.renderReview()
	case StateRefining:
		return m.renderRefining()
	case StateSummary:
		return m.renderSummary()
	case StateCommitting:
		return m.renderCommitting()
	case StateSuccess:
		return m.renderSuccess()
	case StateContinueConfirm:
		return m.renderContinueConfirm()
	case StateError:
		return m.renderError()
	case StateDiffPreview:
		return m.renderDiffPreview()
	case StateAmendConfirm:
		return m.renderAmendConfirm()
	case StateQuitConfirm:
		return m.renderQuitConfirm()
	}

	return ""
}

// templateItem implements list.Item for commit templates
type templateItem struct {
	config.CommitTemplate
}

func (t templateItem) Title() string       { return t.Name }
func (t templateItem) Description() string { return t.Format }
func (t templateItem) FilterValue() string { return t.Name }
