package git

// ChangedFile represents a file with changes in the git repository
type ChangedFile struct {
	Path     string
	Status   string
	Selected bool
}

// Title implements list.Item interface
func (f ChangedFile) Title() string { return f.Path }

// Description implements list.Item interface
func (f ChangedFile) Description() string { return f.Status }

// FilterValue implements list.Item interface
func (f ChangedFile) FilterValue() string { return f.Path }

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash    string
	Date    string
	Author  string
	Subject string
	Body    string
}
