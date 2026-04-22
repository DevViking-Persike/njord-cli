package jira

import (
	"fmt"

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

// ListMyIssuesInProject returns the authenticated user's issues in a given
// project, ordered by status then updated desc. Caller groups by status.
// Empty projectKey is an error.
func (s *JiraService) ListMyIssuesInProject(projectKey string) ([]jiraclient.Issue, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("listing project issues: projectKey is required")
	}
	jql := fmt.Sprintf(`project = %q AND assignee = currentUser() ORDER BY status ASC, updated DESC`, projectKey)
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing project issues: %w", err)
	}
	return res.Issues, nil
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
