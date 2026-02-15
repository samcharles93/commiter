package config

// Config represents the application configuration
type Config struct {
	APIKey             string           `json:"api_key"`
	Model              string           `json:"model"`
	BaseURL            string           `json:"base_url"`
	Provider           string           `json:"provider"`
	DefaultTemplate    string           `json:"default_template,omitempty"`
	ConfirmQuit        *bool            `json:"confirm_quit,omitempty"`
	Templates          []CommitTemplate `json:"templates,omitempty"`
	PreCommitHooks     []string         `json:"pre_commit_hooks,omitempty"`
	PostCommitHooks    []string         `json:"post_commit_hooks,omitempty"`
	HookTimeoutSeconds int              `json:"hook_timeout_seconds,omitempty"`
	sourcePath         string           `json:"-"`
}

// CommitTemplate represents a commit message template
type CommitTemplate struct {
	Key    string   `json:"key,omitempty"`
	Name   string   `json:"name"`
	Types  []string `json:"types,omitempty"`
	Format string   `json:"format"`
	Prompt string   `json:"prompt"`
}
