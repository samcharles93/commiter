package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/samcharles93/commiter/internal/config"
)

func TestInteractiveCommitPreHookFailureBlocksCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := initRepoForModelHookTests(t)
	writeFileForModelHookTest(t, filepath.Join(repoDir, "tracked.txt"), "updated\n")
	runGitForModelHookTest(t, repoDir, "add", "tracked.txt")

	cfg := &config.Config{
		PreCommitHooks:     []string{failingHookCommandForModelTest()},
		HookTimeoutSeconds: 5,
	}
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", cfg, false)
	m.commitMsg = "feat: blocked by pre hook"

	withWorkingDirForModelHookTest(t, repoDir, func() {
		msg := m.commitChanges()()
		commitErr, ok := msg.(CommitErrorMsg)
		if !ok {
			t.Fatalf("expected CommitErrorMsg, got %T", msg)
		}
		if !strings.Contains(commitErr.Err.Error(), "pre-commit hook failed") {
			t.Fatalf("expected pre-hook failure error, got %v", commitErr.Err)
		}
	})

	if count := commitCountForModelHookTest(t, repoDir); count != 1 {
		t.Fatalf("expected pre-hook to block commit, got commit count %d", count)
	}
}

func TestInteractiveAmendPreHookFailureBlocksAmend(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := initRepoForModelHookTests(t)
	originalSubject := strings.TrimSpace(runGitForModelHookTest(t, repoDir, "log", "-1", "--pretty=%s"))

	cfg := &config.Config{
		PreCommitHooks:     []string{failingHookCommandForModelTest()},
		HookTimeoutSeconds: 5,
	}
	m := NewModel(stubProvider{}, nil, "", "openai", "gpt-4o", cfg, false)
	m.commitMsg = "feat: amended message"
	m.isAmending = true

	withWorkingDirForModelHookTest(t, repoDir, func() {
		msg := m.commitChanges()()
		commitErr, ok := msg.(CommitErrorMsg)
		if !ok {
			t.Fatalf("expected CommitErrorMsg, got %T", msg)
		}
		if !strings.Contains(commitErr.Err.Error(), "pre-commit hook failed") {
			t.Fatalf("expected pre-hook failure error, got %v", commitErr.Err)
		}
	})

	if count := commitCountForModelHookTest(t, repoDir); count != 1 {
		t.Fatalf("expected amend to be blocked, got commit count %d", count)
	}

	subject := strings.TrimSpace(runGitForModelHookTest(t, repoDir, "log", "-1", "--pretty=%s"))
	if subject != originalSubject {
		t.Fatalf("expected commit subject unchanged, got %q (want %q)", subject, originalSubject)
	}
}

func initRepoForModelHookTests(t *testing.T) string {
	t.Helper()

	repoDir := t.TempDir()
	runGitForModelHookTest(t, repoDir, "init")
	runGitForModelHookTest(t, repoDir, "config", "user.email", "test@example.com")
	runGitForModelHookTest(t, repoDir, "config", "user.name", "Test User")

	writeFileForModelHookTest(t, filepath.Join(repoDir, "tracked.txt"), "initial\n")
	runGitForModelHookTest(t, repoDir, "add", "tracked.txt")
	runGitForModelHookTest(t, repoDir, "commit", "-m", "chore: init")

	return repoDir
}

func writeFileForModelHookTest(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func runGitForModelHookTest(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func withWorkingDirForModelHookTest(t *testing.T, dir string, fn func()) {
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

func commitCountForModelHookTest(t *testing.T, dir string) int {
	t.Helper()
	out := strings.TrimSpace(runGitForModelHookTest(t, dir, "rev-list", "--count", "HEAD"))
	n, err := strconv.Atoi(out)
	if err != nil {
		t.Fatalf("parse commit count %q: %v", out, err)
	}
	return n
}

func failingHookCommandForModelTest() string {
	if runtime.GOOS == "windows" {
		return `echo hook failed 1>&2 && exit /b 1`
	}
	return `echo "hook failed" >&2; exit 1`
}
