package config

// GetDefaultTemplates returns the default commit message templates
func GetDefaultTemplates() []CommitTemplate {
	return []CommitTemplate{
		{
			Name:   "Conventional Commits",
			Types:  []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
			Format: "{type}: {subject}",
			Prompt: `Generate a commit message following Conventional Commits format.
Start with a type prefix (feat/fix/docs/style/refactor/test/chore) followed by a colon and space.
Use "feat" for new features, "fix" for bug fixes, "docs" for documentation, etc.
Keep the subject line under 50 characters.
Example: "feat: add user authentication"`,
		},
		{
			Name:   "Simple",
			Format: "{subject}",
			Prompt: `Generate a simple, concise commit message without any prefix.
Focus on what changed and why.
Keep it under 50 characters if possible.
Example: "improve error handling in API client"`,
		},
	}
}
