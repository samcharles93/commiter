package hooks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunSuccessSequence(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "hooks.log")
	err := Run(context.Background(), RunOptions{
		Phase: PhasePreCommit,
		Commands: []string{
			appendLineCommand(outPath, "first"),
			appendLineCommand(outPath, "second"),
		},
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(data) != "first\nsecond\n" {
		t.Fatalf("unexpected output sequence: %q", string(data))
	}
}

func TestRunFailureShortCircuits(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "hooks.log")
	err := Run(context.Background(), RunOptions{
		Phase: PhasePreCommit,
		Commands: []string{
			appendLineCommand(outPath, "first"),
			failingCommand(),
			appendLineCommand(outPath, "third"),
		},
		Timeout: 2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected hook failure, got nil")
	}

	var hookErr *HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected HookError, got %T", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(data) != "first\n" {
		t.Fatalf("expected short-circuit after first command, got: %q", string(data))
	}
}

func TestRunTimeout(t *testing.T) {
	t.Parallel()

	err := Run(context.Background(), RunOptions{
		Phase:    PhasePreCommit,
		Commands: []string{sleepCommand()},
		Timeout:  100 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	var hookErr *HookError
	if !errors.As(err, &hookErr) {
		t.Fatalf("expected HookError, got %T", err)
	}
	if !hookErr.TimedOut {
		t.Fatalf("expected timeout hook error, got: %v", hookErr)
	}
}

func TestRunSetsEnvironment(t *testing.T) {
	t.Parallel()

	outPath := filepath.Join(t.TempDir(), "env.log")
	err := Run(context.Background(), RunOptions{
		Phase:         PhasePostCommit,
		Commands:      []string{printEnvCommand(outPath)},
		Timeout:       2 * time.Second,
		CommitMessage: "feat: add env test",
		IsAmend:       true,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read env output: %v", err)
	}

	got := strings.TrimSpace(string(data))
	expected := "post_commit|feat: add env test|true"
	if got != expected {
		t.Fatalf("unexpected env output: got %q, want %q", got, expected)
	}
}

func appendLineCommand(path, line string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`echo %s>>"%s"`, line, path)
	}
	return fmt.Sprintf(`printf "%s\n" >> %q`, line, path)
}

func failingCommand() string {
	if runtime.GOOS == "windows" {
		return `echo hook failed 1>&2 && exit /b 1`
	}
	return `echo "hook failed" >&2; exit 1`
}

func sleepCommand() string {
	if runtime.GOOS == "windows" {
		return `ping -n 3 127.0.0.1 >NUL`
	}
	return `sleep 1`
}

func printEnvCommand(path string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`echo %%COMMITER_HOOK_PHASE%%^|%%COMMITER_COMMIT_MESSAGE%%^|%%COMMITER_IS_AMEND%%>"%s"`, path)
	}
	return fmt.Sprintf(`printf "%%s|%%s|%%s" "$COMMITER_HOOK_PHASE" "$COMMITER_COMMIT_MESSAGE" "$COMMITER_IS_AMEND" > %q`, path)
}
