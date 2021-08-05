package persistence

type ListOptions struct {
	Limit      int
	Offset     int
	Sort       string
	ShowHidden bool
	Search     string
}
