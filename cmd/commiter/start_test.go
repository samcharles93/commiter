package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/samcharles93/commiter/internal/config"
)

func TestRunBypassModePreHookFailureBlocksCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := initRepoForBypassTests(t)
	writeFile(t, filepath.Join(repoDir, "a.txt"), "hello\n")

	cfg := &config.Config{
		PreCommitHooks:     []string{failingHookCommand()},
		HookTimeoutSeconds: 5,
	}

	prevCustomMessage := customMessage
	customMessage = "feat: blocked by hook"
	t.Cleanup(func() {
		customMessage = prevCustomMessage
	})

	withWorkingDir(t, repoDir, func() {
		err := runBypassMode([]string{"a.txt"}, "", "", "", "openai", cfg, false)
		if err == nil {
			t.Fatal("expected pre-hook failure, got nil")
		}
		if !strings.Contains(err.Error(), "pre-commit hook failed") {
			t.Fatalf("expected pre-hook failure error, got: %v", err)
		}
	})

	if commits := commitCount(t, repoDir); commits != 1 {
		t.Fatalf("expected no new commit after pre-hook failure, got commit count %d", commits)
	}
}

func TestRunBypassModePostHookFailureWarnsOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := initRepoForBypassTests(t)
	writeFile(t, filepath.Join(repoDir, "b.txt"), "hello\n")

	cfg := &config.Config{
		PostCommitHooks:    []string{failingHookCommand()},
		HookTimeoutSeconds: 5,
	}

	prevCustomMessage := customMessage
	customMessage = "feat: post hook warning"
	t.Cleanup(func() {
		customMessage = prevCustomMessage
	})

	stderrOutput := captureStderr(t, func() {
		withWorkingDir(t, repoDir, func() {
			err := runBypassMode([]string{"b.txt"}, "", "", "", "openai", cfg, false)
			if err != nil {
				t.Fatalf("expected commit success, got error: %v", err)
			}
		})
	})

	if commits := commitCount(t, repoDir); commits != 2 {
		t.Fatalf("expected commit to succeed despite post-hook failure, got commit count %d", commits)
	}
	if !strings.Contains(stderrOutput, "post-commit hook failed") {
		t.Fatalf("expected post-hook warning in stderr, got: %q", stderrOutput)
	}
}

func TestRunBypassModeNoHooksSkipsConfiguredHooks(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := initRepoForBypassTests(t)
	writeFile(t, filepath.Join(repoDir, "c.txt"), "hello\n")

	cfg := &config.Config{
		PreCommitHooks:     []string{failingHookCommand()},
		PostCommitHooks:    []string{failingHookCommand()},
		HookTimeoutSeconds: 5,
	}

	prevCustomMessage := customMessage
	customMessage = "feat: skip hooks"
	t.Cleanup(func() {
		customMessage = prevCustomMessage
	})

	withWorkingDir(t, repoDir, func() {
		if err := runBypassMode([]string{"c.txt"}, "", "", "", "openai", cfg, true); err != nil {
			t.Fatalf("expected commit success with hooks disabled, got: %v", err)
		}
	})

	if commits := commitCount(t, repoDir); commits != 2 {
		t.Fatalf("expected commit success when hooks are disabled, got commit count %d", commits)
	}
}

func initRepoForBypassTests(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	writeFile(t, filepath.Join(repoDir, ".init"), "seed\n")
	runGit(t, repoDir, "add", ".init")
	runGit(t, repoDir, "commit", "-m", "chore: init")

	return repoDir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(original); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	fn()
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func commitCount(t *testing.T, dir string) int {
	t.Helper()
	out := runGit(t, dir, "rev-list", "--count", "HEAD")
	out = strings.TrimSpace(out)
	if out == "" {
		return 0
	}
	count, err := strconv.Atoi(out)
	if err != nil {
		t.Fatalf("parse commit count %q: %v", out, err)
	}
	return count
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = old
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stderr output: %v", err)
	}
	return string(data)
}

func failingHookCommand() string {
	if runtime.GOOS == "windows" {
		return `echo hook failed 1>&2 && exit /b 1`
	}
	return `echo "hook failed" >&2; exit 1`
}
