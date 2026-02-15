package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndSaveUseLoadedConfigPath(t *testing.T) {
	workDir := t.TempDir()
	homeDir := t.TempDir()
	setHomeForConfigTest(t, homeDir)

	localPath := filepath.Join(workDir, ConfigFileName)
	if err := os.WriteFile(localPath, []byte(`{"provider":"openai"}`), 0o600); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(workDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	cfg.DefaultTemplate = "Simple"
	if err := Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	localData, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("read local config: %v", err)
	}

	var localCfg map[string]any
	if err := json.Unmarshal(localData, &localCfg); err != nil {
		t.Fatalf("parse local config: %v", err)
	}
	if localCfg["default_template"] != "Simple" {
		t.Fatalf("expected local default_template to be saved, got %v", localCfg["default_template"])
	}

	homePath := filepath.Join(homeDir, ConfigFileName)
	if _, err := os.Stat(homePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no home config write, stat error: %v", err)
	}
}

func TestSaveDefaultsToHomePathWhenNoSourcePath(t *testing.T) {
	homeDir := t.TempDir()
	setHomeForConfigTest(t, homeDir)

	cfg := &Config{
		Provider: "openai",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	homePath := filepath.Join(homeDir, ConfigFileName)
	if _, err := os.Stat(homePath); err != nil {
		t.Fatalf("expected config file at %s: %v", homePath, err)
	}
}

func TestFindTemplateSupportsValueAliases(t *testing.T) {
	cfg := &Config{
		Templates: GetDefaultTemplates(),
	}

	tests := []struct {
		input    string
		expected string
	}{
		{input: "default", expected: "conventional"},
		{input: "conventional", expected: "conventional"},
		{input: "Conventional Commits", expected: "conventional"},
		{input: "simple", expected: "simple"},
		{input: "Simple", expected: "simple"},
	}

	for _, tc := range tests {
		tmpl := cfg.FindTemplate(tc.input)
		if tmpl == nil {
			t.Fatalf("expected template for input %q", tc.input)
		}
		if tmpl.ConfigValue() != tc.expected {
			t.Fatalf("expected input %q to resolve to %q, got %q", tc.input, tc.expected, tmpl.ConfigValue())
		}
	}
}

func setHomeForConfigTest(t *testing.T, home string) {
	t.Helper()

	originalHome, hadHome := os.LookupEnv("HOME")
	originalUserProfile, hadUserProfile := os.LookupEnv("USERPROFILE")

	if err := os.Setenv("HOME", home); err != nil {
		t.Fatalf("set HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", home); err != nil {
		t.Fatalf("set USERPROFILE: %v", err)
	}

	t.Cleanup(func() {
		if hadHome {
			_ = os.Setenv("HOME", originalHome)
		} else {
			_ = os.Unsetenv("HOME")
		}
		if hadUserProfile {
			_ = os.Setenv("USERPROFILE", originalUserProfile)
		} else {
			_ = os.Unsetenv("USERPROFILE")
		}
	})
}
