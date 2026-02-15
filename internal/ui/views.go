package ui

import (
	"fmt"
	"strings"

	"github.com/samcharles93/commiter/internal/git"
)

func (m Model) renderTemplateSelection() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🔍 Commiter") + "\n")
	b.WriteString(SubtleStyle.Render("Select a commit message template") + "\n\n")
	b.WriteString(m.templateList.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓: navigate • enter: select • ?: help • q: quit"))
	return b.String()
}

func (m Model) renderFileSelection() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🔍 Commiter") + "\n")
	b.WriteString(SubtleStyle.Render("No staged changes detected") + "\n\n")
	b.WriteString(m.fileList.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓: navigate • space: toggle • enter: stage • a: stage all • u: unselect all • d: diff • q: quit"))
	return b.String()
}

func (m Model) renderGenerating() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🤖 Commiter") + "\n")
	b.WriteString(m.spinner.View() + " Generating commit message with " + m.modelName + "...\n")
	return b.String()
}

func (m Model) renderReview() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("📝 Proposed Commit Message") + "\n")
	b.WriteString(CommitMsgStyle.Render(m.commitMsg) + "\n\n")
	b.WriteString(SubtleStyle.Render("Provider: "+m.providerName+" | Model: "+m.modelName) + "\n\n")

	amendOption := ""
	if git.CanAmend() {
		amendOption = " • [a] amend"
	}

	b.WriteString(HelpStyle.Render("[y] accept • [n] regenerate • [r] refine • [s] summary • [d] diff" + amendOption + " • [?] help • [q] quit"))
	return b.String()
}

func (m Model) renderRefining() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("✏️  Refine Commit Message") + "\n")
	b.WriteString(SubtleStyle.Render("Current message:") + "\n")
	b.WriteString(BoxStyle.Render(m.commitMsg) + "\n")
	b.WriteString(SubtleStyle.Render("How should it be improved?") + "\n\n")
	b.WriteString(m.textarea.View() + "\n\n")
	b.WriteString(HelpStyle.Render("ctrl+s: submit • esc: cancel"))
	return b.String()
}

func (m Model) renderSummary() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("📊 Change Summary") + "\n")

	if m.summary == "" {
		b.WriteString(m.spinner.View() + " Analyzing changes...\n")
	} else {
		b.WriteString(BoxStyle.Render(m.summary) + "\n\n")
		b.WriteString(HelpStyle.Render("Press any key to return"))
	}

	return b.String()
}

func (m Model) renderCommitting() string {
	var b strings.Builder
	action := "Committing"
	if m.isAmending {
		action = "Amending commit"
	}
	b.WriteString(TitleStyle.Render("💾 "+action) + "\n")
	b.WriteString(m.spinner.View() + " " + action + "...\n")
	return b.String()
}

func (m Model) renderSuccess() string {
	var b strings.Builder
	b.WriteString("\n")
	action := "Committed"
	if m.isAmending {
		action = "Amended"
	}
	b.WriteString(SuccessStyle.Render("✅ Success!") + "\n\n")
	b.WriteString(BoxStyle.Render(action+" with message:\n\n"+m.commitMsg) + "\n")
	b.WriteString(SubtleStyle.Render("Returning to terminal...") + "\n")
	return b.String()
}

func (m Model) renderContinueConfirm() string {
	var b strings.Builder
	b.WriteString("\n")
	action := "Committed"
	if m.isAmending {
		action = "Amended"
	}
	b.WriteString(SuccessStyle.Render("✅ Success!") + "\n\n")
	b.WriteString(BoxStyle.Render(action+" with message:\n\n"+m.commitMsg) + "\n")
	b.WriteString(SubtleStyle.Render("More changes detected. Continue committing?") + "\n")
	b.WriteString(HelpStyle.Render("[enter] continue • [n/esc] exit"))
	return b.String()
}

func (m Model) renderError() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("❌ Error") + "\n\n")
	b.WriteString(ErrorBoxStyle.Render(m.err.Error()) + "\n")
	return b.String()
}

func (m Model) renderDiffPreview() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("📄 Diff Preview") + "\n")
	b.WriteString(m.diffViewer.View() + "\n")
	b.WriteString(HelpStyle.Render("↑↓/jk: scroll • pgup/pgdn: page • g/G: top/bottom • q/esc: exit"))
	return b.String()
}

func (m Model) renderAmendConfirm() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("🔄 Amend Last Commit") + "\n\n")
	b.WriteString(SubtleStyle.Render("Current commit:") + "\n")
	if m.lastCommit != nil {
		currentMsg := fmt.Sprintf("%s\n\n%s", m.lastCommit.Subject, m.lastCommit.Body)
		b.WriteString(BoxStyle.Render(strings.TrimSpace(currentMsg)) + "\n\n")
	}
	b.WriteString(SubtleStyle.Render("New commit message:") + "\n")
	b.WriteString(CommitMsgStyle.Render(m.commitMsg) + "\n\n")
	b.WriteString(HelpStyle.Render("[y] amend commit • [n/esc] cancel"))
	return b.String()
}

func (m Model) renderQuitConfirm() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render("❓ Quit") + "\n\n")
	b.WriteString(BoxStyle.Render("Are you sure you want to quit?\nAny unsaved work will be lost.") + "\n\n")
	b.WriteString(HelpStyle.Render("[y] quit • [n/esc] cancel"))
	return b.String()
}

func (m Model) renderHelp() string {
	help := `
┌─ Keyboard Shortcuts ─────────────────────────┐
│                                              │
│  Template Selection                          │
│    ↑↓        Navigate templates              │
│    Enter     Select template                 │
│                                              │
│  File Selection                              │
│    ↑↓        Navigate files                  │
│    Space     Toggle file selection           │
│    Enter     Stage selected files            │
│    a         Stage all files                 │
│    u         Unselect all files              │
│    d         Preview file diff               │
│                                              │
│  Commit Review                               │
│    y         Accept commit message           │
│    n         Regenerate (different option)   │
│    r         Refine with feedback            │
│    s         Show change summary             │
│    a         Amend last commit               │
│    d         Preview full diff               │
│                                              │
│  Diff Preview                                │
│    ↑↓/jk     Scroll up/down                  │
│    PgUp/Dn   Page up/down                    │
│    g/G       Jump to top/bottom              │
│    q/Esc     Exit diff view                  │
│                                              │
│  Refining                                    │
│    Ctrl+S    Submit feedback                 │
│    Esc       Cancel refinement               │
│                                              │
│  General                                     │
│    ?         Toggle this help                │
│    q         Quit (with confirmation)        │
│    Ctrl+C    Force quit (always)             │
│                                              │
└──────────────────────────────────────────────┘

Press ? to close this help
`
	return BoxStyle.Render(help)
}
