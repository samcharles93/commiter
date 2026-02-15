package ui

import "github.com/samcharles93/commiter/internal/git"

// Tea Messages for async operations

type GenerateMsg struct {
	Message string
	Err     error
}

type SummaryMsg struct {
	Summary string
	Err     error
}

type CommitSuccessMsg struct {
	HasRemainingChanges bool
	RemainingDiff       string
	RemainingFiles      []git.ChangedFile
}

type CommitErrorMsg struct {
	Err error
}

type AutoQuitMsg struct{}
