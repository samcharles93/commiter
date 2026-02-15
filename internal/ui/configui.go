package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/config"
)

// ConfigModel is the TUI model for configuration
type ConfigModel struct {
	config           *config.Config
	state            string
	list             list.Model
	hookList         list.Model
	textInput        textinput.Model
	err              error
	editingField     string
	editingHookType  string
	editingHookIndex int
	errorReturnState string
}

const (
	configStateMenu     = "menu"
	configStateEdit     = "edit"
	configStateHookList = "hook-list"
	configStateHookEdit = "hook-edit"
	configStateSaved    = "saved"
	configStateError    = "error"
)

type configMenuItem struct {
	name        string
	description string
	value       string
}

func (i configMenuItem) Title() string       { return i.name }
func (i configMenuItem) Description() string { return fmt.Sprintf("%s: %s", i.description, i.value) }
func (i configMenuItem) FilterValue() string { return i.name }

type hookCommandItem struct {
	command string
	isAdd   bool
}

func (i hookCommandItem) Title() string {
	if i.isAdd {
		return "+ Add Hook"
	}
	return i.command
}

func (i hookCommandItem) Description() string {
	if i.isAdd {
		return "Create a new hook command"
	}
	return "Hook command"
}

func (i hookCommandItem) FilterValue() string {
	return i.command
}

// NewConfigModel creates a new configuration TUI model
func NewConfigModel(cfg *config.Config) ConfigModel {
	items := configItems(cfg)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedFileStyle
	delegate.Styles.SelectedDesc = SubtleStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "Configuration"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	hookList := list.New([]list.Item{}, delegate, 0, 0)
	hookList.Title = "Hooks"
	hookList.SetShowStatusBar(false)
	hookList.SetFilteringEnabled(false)
	hookList.Styles.Title = TitleStyle

	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 500
	ti.Width = 60

	return ConfigModel{
		config:           cfg,
		state:            configStateMenu,
		list:             l,
		hookList:         hookList,
		textInput:        ti,
		editingHookIndex: -1,
	}
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case configStateMenu:
			switch msg.String() {
			case "enter":
				selected := m.list.SelectedItem()
				if selected != nil {
					item := selected.(configMenuItem)
					switch item.name {
					case "Save":
						if err := config.Save(m.config); err != nil {
							m.setError(configStateMenu, err)
							return m, nil
						}
						m.state = configStateSaved
						return m, tea.Quit

					case "Confirm Quit":
						current := m.config.GetConfirmQuit()
						newValue := !current
						m.config.ConfirmQuit = &newValue
						m.updateListItems()
						return m, nil

					case "Pre-Commit Hooks":
						m.startHookList("pre")
						return m, nil

					case "Post-Commit Hooks":
						m.startHookList("post")
						return m, nil

					default:
						m.editingField = item.name
						m.state = configStateEdit
						m.textInput.SetValue(m.getFieldValue(item.name))
						m.textInput.Focus()
						return m, textinput.Blink
					}
				}
			case "q", "esc":
				return m, tea.Quit
			}

		case configStateEdit:
			switch msg.String() {
			case "enter":
				if err := m.setFieldValue(m.editingField, m.textInput.Value()); err != nil {
					m.setError(configStateEdit, err)
					return m, nil
				}
				m.updateListItems()
				m.state = configStateMenu
				m.textInput.Blur()
				return m, nil

			case "esc":
				m.state = configStateMenu
				m.textInput.Blur()
				return m, nil
			}

		case configStateHookList:
			switch msg.String() {
			case "enter":
				selected := m.hookList.SelectedItem()
				if selected == nil {
					return m, nil
				}

				item := selected.(hookCommandItem)
				if item.isAdd {
					m.startHookEdit(-1, "")
					return m, textinput.Blink
				}

				m.startHookEdit(m.hookList.Index(), item.command)
				return m, textinput.Blink

			case "a":
				m.startHookEdit(-1, "")
				return m, textinput.Blink

			case "d":
				idx := m.hookList.Index()
				hooks := m.currentHooks()
				if idx >= 0 && idx < len(hooks) {
					hooks = append(hooks[:idx], hooks[idx+1:]...)
					m.setCurrentHooks(hooks)
					m.reloadHookListItems()
					m.updateListItems()
				}
				return m, nil

			case "q", "esc":
				m.state = configStateMenu
				return m, nil
			}

		case configStateHookEdit:
			switch msg.String() {
			case "enter":
				value := strings.TrimSpace(m.textInput.Value())
				if value == "" {
					m.setError(configStateHookEdit, fmt.Errorf("hook command cannot be empty"))
					return m, nil
				}

				hooks := m.currentHooks()
				if m.editingHookIndex < 0 {
					hooks = append(hooks, value)
				} else if m.editingHookIndex < len(hooks) {
					hooks[m.editingHookIndex] = value
				} else {
					m.setError(configStateHookList, fmt.Errorf("invalid hook index %d", m.editingHookIndex))
					return m, nil
				}

				m.setCurrentHooks(hooks)
				m.reloadHookListItems()
				m.updateListItems()
				m.editingHookIndex = -1
				m.state = configStateHookList
				m.textInput.Blur()
				return m, nil

			case "esc":
				m.editingHookIndex = -1
				m.state = configStateHookList
				m.textInput.Blur()
				return m, nil
			}

		case configStateError:
			switch msg.String() {
			case "enter", "esc", "q":
				if m.errorReturnState == "" {
					return m, tea.Quit
				}
				m.state = m.errorReturnState
				m.err = nil
				return m, nil
			}

		case configStateSaved:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		appH, appV := AppStyle.GetFrameSize()
		h, v := BoxStyle.GetFrameSize()
		width := msg.Width - appH - h
		height := msg.Height - appV - v - 5
		if width < 0 {
			width = 0
		}
		if height < 0 {
			height = 0
		}
		m.list.SetSize(width, height)
		m.hookList.SetSize(width, height)

		inputWidth := width - 8
		if inputWidth < 20 {
			inputWidth = 20
		}
		m.textInput.Width = inputWidth
	}

	switch m.state {
	case configStateMenu:
		m.list, cmd = m.list.Update(msg)
	case configStateEdit, configStateHookEdit:
		m.textInput, cmd = m.textInput.Update(msg)
	case configStateHookList:
		m.hookList, cmd = m.hookList.Update(msg)
	}

	return m, cmd
}

func (m ConfigModel) View() string {
	var content string
	switch m.state {
	case configStateMenu:
		content = m.renderMenu()
	case configStateEdit:
		content = m.renderEdit()
	case configStateHookList:
		content = m.renderHookList()
	case configStateHookEdit:
		content = m.renderHookEdit()
	case configStateSaved:
		content = m.renderSaved()
	case configStateError:
		content = m.renderError()
	}
	return AppStyle.Render(content)
}

func (m ConfigModel) renderMenu() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("⚙️  Configuration") + "\n\n")
	b.WriteString(m.list.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓: navigate • enter: edit/toggle/save • q: quit"))
	return b.String()
}

func (m ConfigModel) renderEdit() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("✏️  Edit "+m.editingField) + "\n\n")
	if m.editingField == "Default Template" {
		b.WriteString(SubtleStyle.Render("Leave blank to prompt on startup.") + "\n")
		b.WriteString(SubtleStyle.Render("Accepted values: "+formatTemplateOptions(m.config)) + "\n\n")
	}
	b.WriteString(SubtleStyle.Render("Current value:") + "\n")
	b.WriteString(BoxStyle.Render(m.getFieldValue(m.editingField)) + "\n\n")
	b.WriteString(SubtleStyle.Render("New value:") + "\n")
	b.WriteString(m.textInput.View() + "\n\n")
	b.WriteString(HelpStyle.Render("enter: save • esc: cancel"))
	return b.String()
}

func (m ConfigModel) renderHookList() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🪝 "+m.currentHookLabel()+" Hooks") + "\n\n")
	b.WriteString(m.hookList.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓: navigate • enter: edit/add • a: add • d: delete • esc: back"))
	return b.String()
}

func (m ConfigModel) renderHookEdit() string {
	var b strings.Builder
	action := "Add"
	if m.editingHookIndex >= 0 {
		action = "Edit"
	}

	b.WriteString(TitleStyle.Render("✏️  "+action+" "+m.currentHookLabel()+" Hook") + "\n\n")
	b.WriteString(SubtleStyle.Render("Hook command:") + "\n")
	b.WriteString(m.textInput.View() + "\n\n")
	b.WriteString(HelpStyle.Render("enter: save • esc: cancel"))
	return b.String()
}

func (m ConfigModel) renderSaved() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(SuccessStyle.Render("✅ Configuration Saved!") + "\n\n")
	b.WriteString(BoxStyle.Render("Your settings have been saved.") + "\n")
	return b.String()
}

func (m ConfigModel) renderError() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(ErrorBoxStyle.Render(m.err.Error()) + "\n")
	b.WriteString(HelpStyle.Render("enter/esc: go back"))
	return b.String()
}

func (m *ConfigModel) getFieldValue(field string) string {
	switch field {
	case "Provider":
		if m.config.Provider == "" {
			return "deepseek"
		}
		return m.config.Provider
	case "API Key":
		return m.config.APIKey
	case "Model":
		return m.config.Model
	case "Default Template":
		return m.config.DefaultTemplate
	case "Base URL":
		return m.config.BaseURL
	case "Hook Timeout (sec)":
		return strconv.Itoa(m.config.GetHookTimeoutSeconds())
	default:
		return ""
	}
}

func (m *ConfigModel) setFieldValue(field, value string) error {
	value = strings.TrimSpace(value)

	switch field {
	case "Provider":
		m.config.Provider = value
		return nil
	case "API Key":
		m.config.APIKey = value
		return nil
	case "Model":
		m.config.Model = value
		return nil
	case "Default Template":
		if value == "" {
			m.config.DefaultTemplate = ""
			return nil
		}
		tmpl := m.config.FindTemplate(value)
		if tmpl == nil {
			return fmt.Errorf("unknown template %q (valid: %s)", value, formatTemplateOptions(m.config))
		}
		m.config.DefaultTemplate = tmpl.ConfigValue()
		return nil
	case "Base URL":
		m.config.BaseURL = value
		return nil
	case "Hook Timeout (sec)":
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds <= 0 {
			return fmt.Errorf("hook timeout must be a positive integer")
		}
		m.config.HookTimeoutSeconds = seconds
		return nil
	default:
		return fmt.Errorf("unknown field %q", field)
	}
}

func (m *ConfigModel) updateListItems() {
	m.list.SetItems(configItems(m.config))
}

func (m *ConfigModel) setError(returnState string, err error) {
	m.err = err
	m.errorReturnState = returnState
	m.state = configStateError
}

func (m *ConfigModel) startHookList(hookType string) {
	m.editingHookType = hookType
	m.reloadHookListItems()
	m.state = configStateHookList
}

func (m *ConfigModel) startHookEdit(index int, value string) {
	m.editingHookIndex = index
	m.textInput.SetValue(value)
	m.textInput.Focus()
	m.state = configStateHookEdit
}

func (m *ConfigModel) reloadHookListItems() {
	hooks := m.currentHooks()
	items := make([]list.Item, 0, len(hooks)+1)
	for _, command := range hooks {
		items = append(items, hookCommandItem{command: command})
	}
	items = append(items, hookCommandItem{isAdd: true})
	m.hookList.SetItems(items)
	m.hookList.Title = m.currentHookLabel() + " Hooks"
}

func (m *ConfigModel) currentHooks() []string {
	switch m.editingHookType {
	case "post":
		return append([]string(nil), m.config.PostCommitHooks...)
	default:
		return append([]string(nil), m.config.PreCommitHooks...)
	}
}

func (m *ConfigModel) setCurrentHooks(hooks []string) {
	switch m.editingHookType {
	case "post":
		m.config.PostCommitHooks = append([]string(nil), hooks...)
	default:
		m.config.PreCommitHooks = append([]string(nil), hooks...)
	}
}

func (m *ConfigModel) currentHookLabel() string {
	if m.editingHookType == "post" {
		return "Post-Commit"
	}
	return "Pre-Commit"
}

func configItems(cfg *config.Config) []list.Item {
	return []list.Item{
		configMenuItem{"Provider", "LLM provider", cfg.Provider},
		configMenuItem{"API Key", "API key for the provider", maskAPIKey(cfg.APIKey)},
		configMenuItem{"Model", "Model name", cfg.Model},
		configMenuItem{"Default Template", "Template used on startup", formatDefaultTemplateValue(cfg)},
		configMenuItem{"Base URL", "API base URL", cfg.BaseURL},
		configMenuItem{"Confirm Quit", "Ask before quitting", fmt.Sprintf("%t", cfg.GetConfirmQuit())},
		configMenuItem{"Pre-Commit Hooks", "Commands to run before commit", formatHookSummary(cfg.PreCommitHooks)},
		configMenuItem{"Post-Commit Hooks", "Commands to run after commit", formatHookSummary(cfg.PostCommitHooks)},
		configMenuItem{"Hook Timeout (sec)", "Per-command timeout", strconv.Itoa(cfg.GetHookTimeoutSeconds())},
		configMenuItem{"Save", "Save configuration", ""},
	}
}

func formatDefaultTemplateValue(cfg *config.Config) string {
	if cfg == nil {
		return "(prompt on startup)"
	}
	if cfg.DefaultTemplate == "" {
		return "(prompt on startup)"
	}
	tmpl := cfg.ResolveDefaultTemplate()
	if tmpl == nil {
		return cfg.DefaultTemplate + " (missing)"
	}
	return fmt.Sprintf("%s = %s", tmpl.ConfigValue(), tmpl.DisplayName())
}

func formatTemplateOptions(cfg *config.Config) string {
	if cfg == nil || len(cfg.Templates) == 0 {
		return "default"
	}
	options := make([]string, 0, len(cfg.Templates)+1)
	options = append(options, "default")
	for _, tmpl := range cfg.Templates {
		options = append(options, fmt.Sprintf("%s=%s", tmpl.ConfigValue(), tmpl.DisplayName()))
	}
	return strings.Join(options, ", ")
}

func formatHookSummary(hooks []string) string {
	if len(hooks) == 0 {
		return "(none)"
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return fmt.Sprintf("%d hooks", len(hooks))
}

func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
