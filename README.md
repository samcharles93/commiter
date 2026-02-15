# Commiter

Fast AI-powered git commits in a TUI.

## Features

### Features
- **Fast AI commit messages**: generate contextual commit messages using git diff
- **Interactive TUI**: Beautiful terminal user interface built with Bubble Tea
- **Multi-Select Files**: Select multiple files to stage with spacebar
- **Diff Preview**: View syntax-highlighted diffs before committing
- **Commit Templates**: Conventional Commits and custom templates
- **Commit History Browser**: Browse and search recent commits with diffs
- **Amend Support**: Option to amend the last commit
- **Configurable Commit Hooks**: Run custom pre/post commit commands
- **Quit Confirmation**: Optional confirmation before quitting
- **Bypass Mode**: Quick commit without interactive prompts
- **Markdown Rendering**: Rich markdown output for summaries, history details, and hook diagnostics

### 🛠️ Commands

#### Default (Interactive Mode)
```bash
commiter
```
Launch the interactive TUI to stage files and generate commit messages.

#### Config
```bash
commiter config
```
Interactive configuration editor:
- Provider (DeepSeek, OpenAI, custom)
- API Key
- Model name
- Base URL
- Pre-commit hooks
- Post-commit hooks
- Hook timeout (per command)
- Confirm quit setting

#### History
```bash
commiter history
```
Browse commit history with:
- Commit details view
- Diff preview
- Search/filter
- Copy commit hash

```bash
commiter history -n 100  # Show last 100 commits
```

#### Bypass Mode
```bash
# Stage and commit file(s) immediately
commiter file.txt -y

# With custom message
commiter file1.go file2.go -y -m "fix: bug in handler"

# Multiple files
commiter src/*.go -y

# Skip configured hooks for this run
commiter src/*.go -y --no-hooks
```

### ⌨️ Keyboard Shortcuts

#### Template Selection
- `↑↓` - Navigate templates
- `Enter` - Select template

#### File Selection
- `↑↓` - Navigate files
- `Space` - Toggle file selection
- `Enter` - Stage selected files
- `a` - Stage all files
- `u` - Unselect all files
- `d` - Preview file diff
- `q` - Quit (with confirmation)

#### Commit Review
- `y` - Accept commit message
- `n` - Regenerate (different option)
- `r` - Refine with feedback
- `s` - Show change summary
- `a` - Amend last commit (if available)
- `d` - Preview full diff
- `?` - Toggle help
- `q` - Quit (with confirmation)

#### Diff Preview
- `↑↓/jk` - Scroll up/down
- `PgUp/PgDn` - Page up/down
- `g/G` - Jump to top/bottom
- `q/Esc` - Exit diff view

#### History Browser
- `↑↓` - Navigate commits
- `Enter/Space` - View commit details
- `/` - Filter commits
- `d` - Toggle diff view
- `y` - Copy commit hash
- `q` - Quit

### 📋 Configuration

Configuration is stored in `~/.commiter.json`:

```json
{
  "provider": "deepseek",
  "api_key": "your-api-key",
  "model": "deepseek-chat",
  "base_url": "https://api.deepseek.com/v1/chat/completions",
  "pre_commit_hooks": [
    "go test ./..."
  ],
  "post_commit_hooks": [
    "echo \"commit created\""
  ],
  "hook_timeout_seconds": 30,
  "confirm_quit": true,
  "templates": [
    {
      "name": "Conventional Commits",
      "types": ["feat", "fix", "docs", "style", "refactor", "test", "chore"],
      "format": "{type}: {subject}",
      "prompt": "Generate a commit message following Conventional Commits format..."
    }
  ]
}
```

### 🔧 Setup

1. **Build**:
   ```bash
   go build -o commiter ./cmd/commiter
   ```

2. **Configure**:
   ```bash
   ./commiter config
   ```
   Or set environment variables:
   ```bash
   export DEEPSEEK_API_KEY="your-key-here"
   # or
   export OPENAI_API_KEY="your-key-here"
   ```

3. **Use**:
   ```bash
   cd your-git-repo
   ./commiter
   ```

### 🚀 Usage Examples

**Basic workflow:**
```bash
# Stage files and generate commit message
commiter

# Stage specific files
# Use space to select, enter to stage
# Review and accept generated message
```

**Quick commit:**
```bash
# Bypass interactive mode
commiter main.go -y

# With custom message
commiter main.go -m "fix: handle nil pointer" -y
```

**Browse history:**
```bash
# View recent commits
commiter history

# Search for commits
# Press '/' and type search term

# View commit details
# Press Enter on a commit
```

**Configure settings:**
```bash
# Open config editor
commiter config

# Change provider, API key, model, etc.
# Select "Save" to persist
```

### 🎨 Commit Templates

Default templates included:
- **Conventional Commits**: `feat:`, `fix:`, `docs:`, etc.
- **Simple**: Plain commit messages

Templates guide the LLM to generate consistent commit messages following your preferred format.

### 📦 Architecture

```
commiter/
├── main.go                    # Entry point
├── cmd/
│   ├── root.go               # Cobra CLI root
│   ├── start.go              # Default TUI flow
│   ├── config.go             # Config command
│   └── history.go            # History command
├── internal/
│   ├── git/
│   │   ├── operations.go     # Git commands
│   │   └── types.go          # Git types
│   ├── llm/
│   │   ├── provider.go       # LLM provider
│   │   └── types.go          # LLM types
│   ├── config/
│   │   ├── config.go         # Config management
│   │   ├── templates.go      # Template system
│   │   └── types.go          # Config types
│   └── ui/
│       ├── model.go          # Main TUI model
│       ├── states.go         # State constants
│       ├── views.go          # View rendering
│       ├── styles.go         # Lipgloss styles
│       ├── messages.go       # Bubbletea messages
│       ├── configui.go       # Config TUI
│       ├── historyui.go      # History TUI
│       └── components/
│           ├── filelist.go   # Multi-select list
│           ├── diffviewer.go # Diff viewer
│           └── confirmation.go # Confirmation dialog
```

### 🔄 State Machine

The TUI uses a state machine with the following states:
- `template-selection` - Choose commit message template
- `file-selection` - Select files to stage
- `diff-preview` - View file/staged diff
- `generating` - LLM generating commit message
- `review` - Review commit message
- `refining` - Refine message with feedback
- `summary` - View change summary
- `amend-confirm` - Confirm amending last commit
- `quit-confirm` - Confirm quitting
- `continue-confirm` - Continue with remaining changes
- `committing` - Creating commit
- `success` - Commit succeeded
- `error` - Error occurred

### 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### 📝 Development

**Prerequisites:**
- Go 1.22+
- Git
- API key for DeepSeek or OpenAI

**Dependencies:**
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/spf13/cobra` - CLI framework

### 🐛 Troubleshooting

**"API key not found":**
- Set `DEEPSEEK_API_KEY` or `OPENAI_API_KEY` environment variable
- Or configure via `commiter config`

**"Not a git repository":**
- Run commiter inside a git repository
- Initialize with `git init` if needed

**Diff preview not showing:**
- Ensure you have git installed
- Check file has changes: `git status`

### 🎯 Roadmap

- [x] Multi-select files
- [x] Diff preview with syntax highlighting
- [x] Commit templates
- [x] History browser
- [x] Amend support
- [x] Quit confirmation
- [x] Bypass mode
- [x] Config command
- [ ] Clipboard support for hash copying
- [ ] Template editor in config UI
- [ ] Syntax highlighting with Chroma

### 📄 License

MIT

### 🤝 Contributing

Contributions welcome! Please open an issue or PR.

---

Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea)
