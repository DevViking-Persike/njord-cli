package jira

// Issue is the minimal Jira issue representation the app cares about.
type Issue struct {
	Key      string
	Summary  string
	Status   string
	Type     string // Task, Story, Epic, Bug, Sub-task
	Assignee string
	EpicKey  string // parent epic key, empty if none or if the issue IS an epic
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
