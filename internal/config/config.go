package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	ConfigFileName            = ".commiter.json"
	DefaultHookTimeoutSeconds = 30
)

// Load loads the configuration from disk
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	homePath := filepath.Join(home, ConfigFileName)
	paths := []string{
		ConfigFileName,
		homePath,
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg Config
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}

			// Auto-upgrade: set default values for new fields
			if cfg.ConfirmQuit == nil {
				defaultConfirmQuit := true
				cfg.ConfirmQuit = &defaultConfirmQuit
			}
			if cfg.Templates == nil {
				cfg.Templates = GetDefaultTemplates()
			}
			if cfg.PreCommitHooks == nil {
				cfg.PreCommitHooks = []string{}
			}
			if cfg.PostCommitHooks == nil {
				cfg.PostCommitHooks = []string{}
			}
			if cfg.HookTimeoutSeconds <= 0 {
				cfg.HookTimeoutSeconds = DefaultHookTimeoutSeconds
			}

			cfg.sourcePath = normalizeConfigPath(path)
			return &cfg, nil
		}
	}

	// Return default config
	defaultConfirmQuit := true
	return &Config{
		ConfirmQuit:        &defaultConfirmQuit,
		Templates:          GetDefaultTemplates(),
		PreCommitHooks:     []string{},
		PostCommitHooks:    []string{},
		HookTimeoutSeconds: DefaultHookTimeoutSeconds,
		sourcePath:         homePath,
	}, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	path, err := cfg.configPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	cfg.sourcePath = normalizeConfigPath(path)
	return nil
}

// GetConfirmQuit safely returns the ConfirmQuit value
func (c *Config) GetConfirmQuit() bool {
	if c.ConfirmQuit == nil {
		return true
	}
	return *c.ConfirmQuit
}

// GetHookTimeoutSeconds safely returns hook timeout in seconds.
func (c *Config) GetHookTimeoutSeconds() int {
	if c == nil || c.HookTimeoutSeconds <= 0 {
		return DefaultHookTimeoutSeconds
	}
	return c.HookTimeoutSeconds
}

// FindTemplate returns a template by name.
func (c *Config) FindTemplate(name string) *CommitTemplate {
	if c == nil {
		return nil
	}

	key := normalizeTemplateValue(name)
	if key == "" {
		return nil
	}
	if key == "default" && len(c.Templates) > 0 {
		return &c.Templates[0]
	}

	for i := range c.Templates {
		if c.Templates[i].ConfigValue() == key {
			return &c.Templates[i]
		}
	}
	for i := range c.Templates {
		if normalizeTemplateValue(c.Templates[i].DisplayName()) == key {
			return &c.Templates[i]
		}
	}
	return nil
}

// ResolveDefaultTemplate returns the configured default template, if valid.
func (c *Config) ResolveDefaultTemplate() *CommitTemplate {
	if c == nil {
		return nil
	}
	return c.FindTemplate(c.DefaultTemplate)
}

// ConfigValue returns the stable config value for this template.
func (t CommitTemplate) ConfigValue() string {
	if key := normalizeTemplateValue(t.Key); key != "" {
		return key
	}
	return normalizeTemplateValue(t.Name)
}

// DisplayName returns the template label shown in the UI.
func (t CommitTemplate) DisplayName() string {
	if strings.TrimSpace(t.Name) != "" {
		return strings.TrimSpace(t.Name)
	}
	return t.ConfigValue()
}

func (c *Config) configPath() (string, error) {
	if c != nil && c.sourcePath != "" {
		return normalizeConfigPath(c.sourcePath), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ConfigFileName), nil
}

func normalizeConfigPath(path string) string {
	if path == "" {
		return ""
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}

func normalizeTemplateValue(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range v {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash && b.Len() > 0:
			b.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(b.String(), "-")
}
