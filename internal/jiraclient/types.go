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

// Transition é uma opção de mudança de status disponível pra uma issue; o ID
// é o que o POST /transitions consome.
type Transition struct {
	ID        string
	Name      string // display (ex.: "Em desenvolvimento")
	ToStatus  string
	StatusCat string // "new" | "indeterminate" | "done"
}

// CreateIssueInput reúne o mínimo pra criar uma issue via POST /issue.
// ParentKey é obrigatório quando Type == "Subtask" (ou outro subtipo).
type CreateIssueInput struct {
	ProjectKey      string
	Summary         string
	Description     string
	Type            string // "Task" | "Bug" | "Story" | "Subtask"
	ParentKey       string // usado só se Type for subtask
	AssigneeAccount string // accountId do assignee (usamos o usuário atual)
}

// UpdateIssueInput atualiza campos editáveis via PUT /issue/{key}. Campos
// vazios são ignorados no payload.
type UpdateIssueInput struct {
	Summary     string
	Description string
}
