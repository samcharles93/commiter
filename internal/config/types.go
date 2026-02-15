package config

// Config represents the application configuration
type Config struct {
	APIKey      string            `json:"api_key"`
	Model       string            `json:"model"`
	BaseURL     string            `json:"base_url"`
	Provider    string            `json:"provider"`
	ConfirmQuit *bool             `json:"confirm_quit,omitempty"`
	Templates   []CommitTemplate  `json:"templates,omitempty"`
}

// CommitTemplate represents a commit message template
type CommitTemplate struct {
	Name   string   `json:"name"`
	Types  []string `json:"types,omitempty"`
	Format string   `json:"format"`
	Prompt string   `json:"prompt"`
}
