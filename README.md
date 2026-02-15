# Commiter

`commiter` is a simple, AI-powered git commit tool for your terminal. It helps you stage files, see what changed, and writes the commit messages for you so you don't have to.

### What it does

- **AI Messages**: Generates commit messages based on your actual changes.
- **Interactive UI**: A clean terminal interface to pick files and preview diffs.
- **Templates**: Supports Conventional Commits out of the box.
- **History**: Browse and search through your previous commits.
- **Hooks**: Run your tests or linters automatically before you commit.

### How to use it

#### Installation

```bash
go install github.com/samcharles93/commiter/cmd/commiter@latest
```

#### Basic Usage

Just run `commiter` in any git repo:

```bash
commiter
```

- Use `Space` to select files you want to stage.
- Press `Enter` to generate a message.
- Review it, and if it looks good, hit `y` to commit.

#### Configuration

To set up your AI provider (OpenAI, DeepSeek, etc.) or add hooks:

```bash
commiter config
```

#### History

To look back at what you've done:

```bash
commiter history
```

### Shortcuts

- `Space` - Select/unselect a file
- `Enter` - Stage selected files
- `d` - Show diff for a file
- `y` - Accept and commit
- `r` - Give feedback to refine the message
- `q` - Quit
