package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = ".commiter.json"
)

// Load loads the configuration from disk
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	paths := []string{
		ConfigFileName,
		filepath.Join(home, ConfigFileName),
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

			return &cfg, nil
		}
	}

	// Return default config
	defaultConfirmQuit := true
	return &Config{
		ConfirmQuit: &defaultConfirmQuit,
		Templates:   GetDefaultTemplates(),
	}, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	path := filepath.Join(home, ConfigFileName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfirmQuit safely returns the ConfirmQuit value
func (c *Config) GetConfirmQuit() bool {
	if c.ConfirmQuit == nil {
		return true
	}
	return *c.ConfirmQuit
}
