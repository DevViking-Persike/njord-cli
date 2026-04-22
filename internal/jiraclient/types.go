package jiraclient

// Issue is the minimal Jira issue representation the app cares about.
type Issue struct {
	Key            string
	Summary        string
	Status         string
	StatusCategory string // "new" (to-do), "indeterminate" (in progress), "done"
	Type           string // Task, Story, Epic, Bug, Sub-task
	Assignee       string
	EpicKey        string // parent epic key, empty if none or if the issue IS an epic
}

// User is the authenticated account.
type User struct {
	AccountID    string
	DisplayName  string
	EmailAddress string
}

// SearchResult is one page of a JQL search.
type SearchResult struct {
	Issues     []Issue
	Total      int
	IsLast     bool
	NextCursor string
}

// Project is a Jira project (aka "espaço" for the user — Squad GAP, Squad Billing).
type Project struct {
	Key  string // short identifier, used in issue keys (e.g. GAP-123)
	Name string // display name (e.g. "Squad GAP")
	ID   string
}
