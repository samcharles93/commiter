package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	notInGitRepoMessage = "run this tool inside a git repository"
)

// EnsureGitRepository checks if the current directory is inside a git repository
func EnsureGitRepository() error {
	out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").CombinedOutput()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("git is not installed or not in PATH")
		}

		output := strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(output), "not a git repository") {
			return fmt.Errorf(notInGitRepoMessage)
		}

		if output == "" {
			return fmt.Errorf("failed to check git repository: %w", err)
		}
		return fmt.Errorf("failed to check git repository: %s", output)
	}

	if strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf(notInGitRepoMessage)
	}

	return nil
}

// GetStagedDiff returns the diff of staged changes
func GetStagedDiff() ([]byte, error) {
	if err := EnsureGitRepository(); err != nil {
		return nil, err
	}

	diff, err := exec.Command("git", "diff", "--cached").CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(diff))
		if output == "" {
			return nil, fmt.Errorf("failed to read staged changes: %w", err)
		}
		return nil, fmt.Errorf("failed to read staged changes: %s", output)
	}

	return diff, nil
}

// GetFileDiff returns the diff for a specific file
func GetFileDiff(path string) ([]byte, error) {
	if err := EnsureGitRepository(); err != nil {
		return nil, err
	}

	diff, err := exec.Command("git", "diff", "--", path).CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(diff))
		if output == "" {
			return nil, fmt.Errorf("failed to read file diff: %w", err)
		}
		return nil, fmt.Errorf("failed to read file diff: %s", output)
	}

	return diff, nil
}

// ListUnstagedChanges returns a list of files with unstaged changes
func ListUnstagedChanges() ([]ChangedFile, error) {
	if err := EnsureGitRepository(); err != nil {
		return nil, err
	}

	out, err := exec.Command("git", "status", "--porcelain").CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if output == "" {
			return nil, fmt.Errorf("failed to read unstaged changes: %w", err)
		}
		return nil, fmt.Errorf("failed to read unstaged changes: %s", output)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\r\n"), "\n")
	changes := make([]ChangedFile, 0, len(lines))
	seen := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || len(line) < 3 {
			continue
		}

		statusCode := line[:2]
		rawPath := strings.TrimSpace(line[3:])
		if rawPath == "" {
			continue
		}

		switch {
		case statusCode == "??":
			if _, exists := seen[rawPath]; exists {
				continue
			}
			seen[rawPath] = struct{}{}
			changes = append(changes, ChangedFile{Path: rawPath, Status: "untracked"})
		case statusCode[1] != ' ':
			path := rawPath
			if strings.Contains(rawPath, " -> ") {
				parts := strings.SplitN(rawPath, " -> ", 2)
				path = strings.TrimSpace(parts[1])
			}
			if _, exists := seen[path]; exists {
				continue
			}
			seen[path] = struct{}{}
			changes = append(changes, ChangedFile{Path: path, Status: statusLabel(statusCode[1])})
		}
	}

	return changes, nil
}

// StageFiles stages the specified files
func StageFiles(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	args := append([]string{"add", "--"}, paths...)
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		output := strings.TrimSpace(string(out))
		if output == "" {
			return fmt.Errorf("failed to stage selected files: %w", err)
		}
		return fmt.Errorf("failed to stage selected files: %s", output)
	}
	return nil
}

// Commit creates a new commit with the given message
func Commit(message string) error {
	out, err := exec.Command("git", "commit", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

// CanAmend checks if there is a commit to amend
func CanAmend() bool {
	out, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// GetLastCommit returns information about the last commit
func GetLastCommit() (*CommitInfo, error) {
	// Get commit hash
	hashOut, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}
	hash := strings.TrimSpace(string(hashOut))

	// Get commit message and details
	logOut, err := exec.Command("git", "log", "-1", "--pretty=format:%an%n%ad%n%s%n%b", "HEAD").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit details: %w", err)
	}

	lines := strings.Split(string(logOut), "\n")
	info := &CommitInfo{Hash: hash}
	if len(lines) > 0 {
		info.Author = lines[0]
	}
	if len(lines) > 1 {
		info.Date = lines[1]
	}
	if len(lines) > 2 {
		info.Subject = lines[2]
	}
	if len(lines) > 3 {
		info.Body = strings.Join(lines[3:], "\n")
	}

	return info, nil
}

// AmendCommit amends the last commit with a new message
func AmendCommit(message string) error {
	out, err := exec.Command("git", "commit", "--amend", "-m", message).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

// GetCommitHistory returns the last n commits
func GetCommitHistory(limit int) ([]CommitInfo, error) {
	if err := EnsureGitRepository(); err != nil {
		return nil, err
	}

	args := []string{"log", fmt.Sprintf("--max-count=%d", limit), "--pretty=format:%H%x00%an%x00%ad%x00%s"}
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	commits := make([]CommitInfo, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "\x00")
		if len(parts) < 4 {
			continue
		}

		commits = append(commits, CommitInfo{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    parts[2],
			Subject: parts[3],
		})
	}

	return commits, nil
}

// GetCommitDetails returns detailed information and diff for a commit
func GetCommitDetails(hash string) (*CommitInfo, string, error) {
	if err := EnsureGitRepository(); err != nil {
		return nil, "", err
	}

	// Get commit details
	logOut, err := exec.Command("git", "log", "-1", "--pretty=format:%H%x00%an%x00%ad%x00%s%x00%b", hash).CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get commit details: %w", err)
	}

	parts := strings.Split(string(logOut), "\x00")
	info := &CommitInfo{}
	if len(parts) > 0 {
		info.Hash = parts[0]
	}
	if len(parts) > 1 {
		info.Author = parts[1]
	}
	if len(parts) > 2 {
		info.Date = parts[2]
	}
	if len(parts) > 3 {
		info.Subject = parts[3]
	}
	if len(parts) > 4 {
		info.Body = parts[4]
	}

	// Get commit diff
	diffOut, err := exec.Command("git", "show", "--pretty=format:", hash).CombinedOutput()
	if err != nil {
		return info, "", fmt.Errorf("failed to get commit diff: %w", err)
	}

	return info, string(diffOut), nil
}

func statusLabel(code byte) string {
	switch code {
	case 'M':
		return "modified"
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case 'U':
		return "unmerged"
	default:
		return "changed"
	}
}
