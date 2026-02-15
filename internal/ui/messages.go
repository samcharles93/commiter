package ui

// Tea Messages for async operations

type GenerateMsg struct {
	Message string
	Err     error
}

type SummaryMsg struct {
	Summary string
	Err     error
}

type CommitSuccessMsg struct{}

type CommitErrorMsg struct {
	Err error
}
