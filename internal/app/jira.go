package app

import (
	"fmt"

	"github.com/DevViking-Persike/njord-cli/internal/jira"
)

// JiraGateway is the minimum surface the app needs from a Jira client.
// Keeping it as a package-local interface lets us mock it in tests without
// coupling to the concrete HTTP client in internal/jira.
type JiraGateway interface {
	CurrentUser() (jira.User, error)
	SearchIssues(jql string) (jira.SearchResult, error)
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
func (s *JiraService) ListMyOpenIssues() ([]jira.Issue, error) {
	const jql = `assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC`
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing my open issues: %w", err)
	}
	return res.Issues, nil
}

// ListMyOpenEpics returns epics the user is assigned to that are still open.
func (s *JiraService) ListMyOpenEpics() ([]jira.Issue, error) {
	const jql = `issuetype = Epic AND assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC`
	res, err := s.gw.SearchIssues(jql)
	if err != nil {
		return nil, fmt.Errorf("listing my open epics: %w", err)
	}
	return res.Issues, nil
}

// ListEpicChildren returns every child issue of the given epic. Empty result
// is valid (empty epic). Returns error if epicKey is empty.
func (s *JiraService) ListEpicChildren(epicKey string) ([]jira.Issue, error) {
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

// CheckConnection verifies credentials by hitting /myself.
func (s *JiraService) CheckConnection() (jira.User, error) {
	u, err := s.gw.CurrentUser()
	if err != nil {
		return jira.User{}, fmt.Errorf("checking jira connection: %w", err)
	}
	return u, nil
}
