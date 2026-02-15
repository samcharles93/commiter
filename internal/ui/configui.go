package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/samcharles93/commiter/internal/config"
)

// ConfigModel is the TUI model for configuration
type ConfigModel struct {
	config       *config.Config
	state        string
	list         list.Model
	textInput    textinput.Model
	err          error
	editingField string
}

const (
	configStateMenu   = "menu"
	configStateEdit   = "edit"
	configStateSaved  = "saved"
	configStateError  = "error"
)

type configMenuItem struct {
	name        string
	description string
	value       string
}

func (i configMenuItem) Title() string       { return i.name }
func (i configMenuItem) Description() string { return fmt.Sprintf("%s: %s", i.description, i.value) }
func (i configMenuItem) FilterValue() string { return i.name }

// NewConfigModel creates a new configuration TUI model
func NewConfigModel(cfg *config.Config) ConfigModel {
	items := []list.Item{
		configMenuItem{"Provider", "LLM provider", cfg.Provider},
		configMenuItem{"API Key", "API key for the provider", maskAPIKey(cfg.APIKey)},
		configMenuItem{"Model", "Model name", cfg.Model},
		configMenuItem{"Base URL", "API base URL", cfg.BaseURL},
		configMenuItem{"Confirm Quit", "Ask before quitting", fmt.Sprintf("%t", cfg.GetConfirmQuit())},
		configMenuItem{"Save", "Save configuration", ""},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = SelectedFileStyle
	delegate.Styles.SelectedDesc = SubtleStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "Configuration"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle

	ti := textinput.New()
	ti.Placeholder = "Enter value..."
	ti.CharLimit = 200
	ti.Width = 60

	return ConfigModel{
		config: cfg,
		state:  configStateMenu,
		list:   l,
		textInput: ti,
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
						// Save configuration
						if err := config.Save(m.config); err != nil {
							m.state = configStateError
							m.err = err
							return m, nil
						}
						m.state = configStateSaved
						return m, tea.Quit
					case "Confirm Quit":
						// Toggle confirm quit
						current := m.config.GetConfirmQuit()
						newValue := !current
						m.config.ConfirmQuit = &newValue
						m.updateListItems()
						return m, nil
					default:
						// Edit field
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
				// Save edited value
				m.setFieldValue(m.editingField, m.textInput.Value())
				m.updateListItems()
				m.state = configStateMenu
				m.textInput.Blur()
				return m, nil
			case "esc":
				m.state = configStateMenu
				m.textInput.Blur()
				return m, nil
			}

		case configStateSaved, configStateError:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		h, v := BoxStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)
	}

	// Update child components
	switch m.state {
	case configStateMenu:
		m.list, cmd = m.list.Update(msg)
	case configStateEdit:
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m ConfigModel) View() string {
	switch m.state {
	case configStateMenu:
		return m.renderMenu()
	case configStateEdit:
		return m.renderEdit()
	case configStateSaved:
		return m.renderSaved()
	case configStateError:
		return m.renderError()
	}
	return ""
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
	b.WriteString(SubtleStyle.Render("Current value:") + "\n")
	b.WriteString(BoxStyle.Render(m.getFieldValue(m.editingField)) + "\n\n")
	b.WriteString(SubtleStyle.Render("New value:") + "\n")
	b.WriteString(m.textInput.View() + "\n\n")
	b.WriteString(HelpStyle.Render("enter: save • esc: cancel"))
	return b.String()
}

func (m ConfigModel) renderSaved() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(SuccessStyle.Render("✅ Configuration Saved!") + "\n\n")
	b.WriteString(BoxStyle.Render("Your settings have been saved to ~/.commiter.json") + "\n")
	return b.String()
}

func (m ConfigModel) renderError() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(ErrorBoxStyle.Render(m.err.Error()) + "\n")
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
	case "Base URL":
		return m.config.BaseURL
	default:
		return ""
	}
}

func (m *ConfigModel) setFieldValue(field, value string) {
	switch field {
	case "Provider":
		m.config.Provider = value
	case "API Key":
		m.config.APIKey = value
	case "Model":
		m.config.Model = value
	case "Base URL":
		m.config.BaseURL = value
	}
}

func (m *ConfigModel) updateListItems() {
	items := []list.Item{
		configMenuItem{"Provider", "LLM provider", m.config.Provider},
		configMenuItem{"API Key", "API key for the provider", maskAPIKey(m.config.APIKey)},
		configMenuItem{"Model", "Model name", m.config.Model},
		configMenuItem{"Base URL", "API base URL", m.config.BaseURL},
		configMenuItem{"Confirm Quit", "Ask before quitting", fmt.Sprintf("%t", m.config.GetConfirmQuit())},
		configMenuItem{"Save", "Save configuration", ""},
	}
	m.list.SetItems(items)
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
