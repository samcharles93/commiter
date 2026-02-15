package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitRepositoryOutsideRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	tempDir := t.TempDir()
	withWorkingDir(t, tempDir, func() {
		err := EnsureGitRepository()
		if err == nil {
			t.Fatal("expected error outside git repository")
		}
		if !strings.Contains(err.Error(), notInGitRepoMessage) {
			t.Fatalf("expected %q in error, got %q", notInGitRepoMessage, err.Error())
		}
	})
}

func TestGetStagedDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")

	file := filepath.Join(repoDir, "hello.txt")
	if err := os.WriteFile(file, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	runGit(t, repoDir, "add", "hello.txt")

	withWorkingDir(t, repoDir, func() {
		diff, err := GetStagedDiff()
		if err != nil {
			t.Fatalf("GetStagedDiff returned error: %v", err)
		}
		if !strings.Contains(string(diff), "hello.txt") {
			t.Fatalf("expected staged diff to contain file name, got:\n%s", string(diff))
		}
	})
}

func TestListUnstagedChanges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	repoDir := t.TempDir()
	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "test@example.com")
	runGit(t, repoDir, "config", "user.name", "Test User")

	tracked := filepath.Join(repoDir, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write tracked file: %v", err)
	}
	runGit(t, repoDir, "add", "tracked.txt")
	runGit(t, repoDir, "commit", "-m", "feat: add tracked file")

	if err := os.WriteFile(tracked, []byte("hello world\n"), 0o644); err != nil {
		t.Fatalf("modify tracked file: %v", err)
	}

	untracked := filepath.Join(repoDir, "new.txt")
	if err := os.WriteFile(untracked, []byte("new\n"), 0o644); err != nil {
		t.Fatalf("write untracked file: %v", err)
	}

	withWorkingDir(t, repoDir, func() {
		status := runGit(t, repoDir, "status", "--porcelain")
		t.Logf("Git status output:\n%s", status)

		changes, err := ListUnstagedChanges()
		if err != nil {
			t.Fatalf("ListUnstagedChanges returned error: %v", err)
		}

		statusByPath := map[string]string{}
		for _, c := range changes {
			statusByPath[c.Path] = c.Status
		}

		if got := statusByPath["tracked.txt"]; got != "modified" {
			t.Fatalf("expected tracked.txt status to be modified, got %q (full map: %#v)", got, statusByPath)
		}
		if got := statusByPath["new.txt"]; got != "untracked" {
			t.Fatalf("expected new.txt status to be untracked, got %q (full map: %#v)", got, statusByPath)
		}
	})
}

func withWorkingDir(t *testing.T, dir string, fn func()) {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change to %s: %v", dir, err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
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
