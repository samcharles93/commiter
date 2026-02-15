package hooks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	defaultCommandTimeout = 30 * time.Second
	outputLimitBytes      = 8 * 1024
)

// Phase identifies the hook phase.
type Phase string

const (
	PhasePreCommit  Phase = "pre_commit"
	PhasePostCommit Phase = "post_commit"
)

// RunOptions configures hook execution.
type RunOptions struct {
	Phase         Phase
	Commands      []string
	Timeout       time.Duration
	CommitMessage string
	IsAmend       bool
}

// HookError carries details about a failed hook command.
type HookError struct {
	Phase    Phase
	Command  string
	Index    int
	Output   string
	Timeout  time.Duration
	TimedOut bool
	Err      error
}

func (e *HookError) Error() string {
	phase := strings.ReplaceAll(string(e.Phase), "_", "-")
	if e.TimedOut {
		if e.Output != "" {
			return fmt.Sprintf("%s hook %d timed out after %s: %q\n%s", phase, e.Index+1, e.Timeout, e.Command, e.Output)
		}
		return fmt.Sprintf("%s hook %d timed out after %s: %q", phase, e.Index+1, e.Timeout, e.Command)
	}

	if e.Output != "" {
		return fmt.Sprintf("%s hook %d failed: %q\n%s", phase, e.Index+1, e.Command, e.Output)
	}

	return fmt.Sprintf("%s hook %d failed: %q (%v)", phase, e.Index+1, e.Command, e.Err)
}

// Unwrap returns the underlying execution error.
func (e *HookError) Unwrap() error {
	return e.Err
}

// Run executes configured hook commands sequentially.
func Run(ctx context.Context, opts RunOptions) error {
	commands := normalizeCommands(opts.Commands)
	if len(commands) == 0 {
		return nil
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}

	for i, command := range commands {
		if err := runCommand(ctx, opts, i, command, timeout); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(ctx context.Context, opts RunOptions, index int, command string, timeout time.Duration) error {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	name, args := shellCommand(command)
	cmd := exec.CommandContext(runCtx, name, args...)
	cmd.Env = append(os.Environ(),
		"COMMITER_HOOK_PHASE="+string(opts.Phase),
		"COMMITER_COMMIT_MESSAGE="+opts.CommitMessage,
		fmt.Sprintf("COMMITER_IS_AMEND=%t", opts.IsAmend),
	)

	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	trimmedOutput := trimOutput(string(output))
	timedOut := errors.Is(runCtx.Err(), context.DeadlineExceeded)
	return &HookError{
		Phase:    opts.Phase,
		Command:  command,
		Index:    index,
		Output:   trimmedOutput,
		Timeout:  timeout,
		TimedOut: timedOut,
		Err:      err,
	}
}

func shellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

func normalizeCommands(commands []string) []string {
	normalized := make([]string, 0, len(commands))
	for _, command := range commands {
		trimmed := strings.TrimSpace(command)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return normalized
}

func trimOutput(output string) string {
	trimmed := strings.TrimSpace(output)
	if len(trimmed) <= outputLimitBytes {
		return trimmed
	}

	return trimmed[:outputLimitBytes] + "\n...[output truncated]"
}
