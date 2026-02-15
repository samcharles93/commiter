package components

// ConfirmationDialog represents a confirmation dialog
type ConfirmationDialog struct {
	Title   string
	Message string
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(title, message string) ConfirmationDialog {
	return ConfirmationDialog{
		Title:   title,
		Message: message,
	}
}
