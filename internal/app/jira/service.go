package jira

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
)

// JiraGateway is the minimum surface the app needs from a Jira client.
// Keeping it as a package-local interface lets us mock it in tests without
// coupling to the concrete HTTP client in internal/jira.
type JiraGateway interface {
	CurrentUser() (jiraclient.User, error)
	SearchIssues(jql string) (jiraclient.SearchResult, error)
	ListProjects() ([]jiraclient.Project, error)
}

// JiraService composes use cases for tasks, stories and epics.
type JiraService struct {
	gw JiraGateway
}

// NewJiraService wires a gateway to the service.
func NewJiraService(gw JiraGateway) *JiraService {
	return &JiraService{gw: gw}
}

// ListMyOpenIssues returns issues assigned to the authenticated user that are
// not Done yet, ordered by last update desc. Includes tasks, stories, bugs.
func (s *JiraService) ListMyOpenIssues() ([]jiraclient.Issue, error) {
	const jql = `assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC`
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing my open issues: %w", err)
	}
	return res.Issues, nil
}

// ListMyOpenEpics returns epics the user is assigned to that are still open.
func (s *JiraService) ListMyOpenEpics() ([]jiraclient.Issue, error) {
	const jql = `issuetype = Epic AND assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC`
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing my open epics: %w", err)
	}
	return res.Issues, nil
}

// ListEpicChildren returns every child issue of the given epic. Empty result
// is valid (empty epic). Returns error if epicKey is empty.
func (s *JiraService) ListEpicChildren(epicKey string) ([]jiraclient.Issue, error) {
	if epicKey == "" {
		return nil, fmt.Errorf("listing epic children: epicKey is required")
	}
	jql := fmt.Sprintf(`parent = %q ORDER BY rank ASC`, epicKey)
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing epic children: %w", err)
	}
	return res.Issues, nil
}

// ListProjectBacklog returns every non-Done issue in a project, ordered by
// status then last update desc. Used for browsing + search at the space level.
// Empty projectKey is an error.
func (s *JiraService) ListProjectBacklog(projectKey string) ([]jiraclient.Issue, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("listing project backlog: projectKey is required")
	}
	jql := fmt.Sprintf(`project = %q AND statusCategory != Done ORDER BY status ASC, updated DESC`, projectKey)
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing project backlog: %w", err)
	}
	return res.Issues, nil
}

// ListMyProjectIssues returns all issues (any status) assigned to the current
// user in a project. "Meu histórico" — inclui concluídas para consulta.
func (s *JiraService) ListMyProjectIssues(projectKey string) ([]jiraclient.Issue, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("listing my project issues: projectKey is required")
	}
	jql := fmt.Sprintf(`project = %q AND assignee = currentUser() ORDER BY status ASC, updated DESC`, projectKey)
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing my project issues: %w", err)
	}
	return res.Issues, nil
}

// FilterByStatusCategory keeps only issues whose StatusCategory matches the
// given Jira category key ("new", "indeterminate", "done"). Empty category
// returns the input unchanged.
func FilterByStatusCategory(issues []jiraclient.Issue, category string) []jiraclient.Issue {
	if category == "" {
		return issues
	}
	out := make([]jiraclient.Issue, 0, len(issues))
	for _, iss := range issues {
		if iss.StatusCategory == category {
			out = append(out, iss)
		}
	}
	return out
}

// FilterIssues returns the subset of issues whose Key or Summary contains the
// query (case-insensitive). Empty query returns the full slice unchanged.
func FilterIssues(issues []jiraclient.Issue, query string) []jiraclient.Issue {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return issues
	}
	out := make([]jiraclient.Issue, 0, len(issues))
	for _, iss := range issues {
		if strings.Contains(strings.ToLower(iss.Key), q) ||
			strings.Contains(strings.ToLower(iss.Summary), q) {
			out = append(out, iss)
		}
	}
	return out
}

// GroupedByStatus groups issues by status name, preserving first-seen order.
// Returns the ordered status list and the grouping map.
func GroupedByStatus(issues []jiraclient.Issue) (statuses []string, byStatus map[string][]jiraclient.Issue) {
	byStatus = make(map[string][]jiraclient.Issue)
	for _, iss := range issues {
		status := iss.Status
		if status == "" {
			status = "Sem status"
		}
		if _, seen := byStatus[status]; !seen {
			statuses = append(statuses, status)
		}
		byStatus[status] = append(byStatus[status], iss)
	}
	return statuses, byStatus
}

// ListSpaces returns all Jira projects (aka espaços) visible to the user,
// sorted by name ascending. Returns empty slice when there is none.
func (s *JiraService) ListSpaces() ([]jiraclient.Project, error) {
	projects, err := s.gw.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("listing jira spaces: %w", err)
	}
	return projects, nil
}

// CheckConnection verifies credentials by hitting /myself.
func (s *JiraService) CheckConnection() (jiraclient.User, error) {
	u, err := s.gw.CurrentUser()
	if err != nil {
		return jiraclient.User{}, fmt.Errorf("checking jira connection: %w", err)
	}
	return u, nil
}
