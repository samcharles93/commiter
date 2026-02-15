package config

// GetDefaultTemplates returns the default commit message templates
func GetDefaultTemplates() []CommitTemplate {
	return []CommitTemplate{
		{
			Key:    "conventional",
			Name:   "Conventional Commits",
			Types:  []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
			Format: "{type}: {subject}",
			Prompt: `Generate a commit message following Conventional Commits format.
Start with a type prefix (feat/fix/docs/style/refactor/test/chore) followed by a colon and space.
Use "feat" for new features, "fix" for bug fixes, "docs" for documentation, etc.
Keep the subject line under 50 characters when possible.
If the change touches multiple files or concerns, include a body after a blank line with 1-3 concise bullet points summarizing impact.
Example: "feat: add user authentication"`,
		},
		{
			Key:    "simple",
			Name:   "Simple",
			Format: "{subject}",
			Prompt: `Generate a simple, concise commit message without any prefix.
Focus on what changed and why.
Keep the subject under 50 characters when possible.
If the change spans multiple files or concerns, include a short body after a blank line with 1-3 bullet points.
Example: "improve error handling in API client"`,
		},
	}
}
