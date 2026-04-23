package jira

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
)

type fakeJiraGW struct {
	user        jiraclient.User
	userErr     error
	searchRes   jiraclient.SearchResult
	searchErr   error
	projects    []jiraclient.Project
	projectsErr error
	lastJQL     string
	callCount   int

	// mutation hooks (usados só nos testes que mexem em CRUD)
	createdInput  jiraclient.CreateIssueInput
	createResult  jiraclient.Issue
	createErr     error
	updateKey     string
	updateInput   jiraclient.UpdateIssueInput
	updateErr     error
	transitions   []jiraclient.Transition
	transErr      error
	transitionKey string
	transitionID  string
}

func (f *fakeJiraGW) CreateIssue(in jiraclient.CreateIssueInput) (jiraclient.Issue, error) {
	f.createdInput = in
	return f.createResult, f.createErr
}
func (f *fakeJiraGW) UpdateIssue(key string, in jiraclient.UpdateIssueInput) error {
	f.updateKey = key
	f.updateInput = in
	return f.updateErr
}
func (f *fakeJiraGW) ListTransitions(string) ([]jiraclient.Transition, error) {
	return f.transitions, f.transErr
}
func (f *fakeJiraGW) TransitionIssue(key, id string) error {
	f.transitionKey = key
	f.transitionID = id
	return nil
}

func (f *fakeJiraGW) CurrentUser() (jiraclient.User, error) {
	f.callCount++
	return f.user, f.userErr
}

func (f *fakeJiraGW) SearchIssues(jql string) (jiraclient.SearchResult, error) {
	f.lastJQL = jql
	f.callCount++
	return f.searchRes, f.searchErr
}

func (f *fakeJiraGW) ListProjects() ([]jiraclient.Project, error) {
	f.callCount++
	return f.projects, f.projectsErr
}

func TestListMyOpenIssues_UsesExpectedJQL(t *testing.T) {
	gw := &fakeJiraGW{searchRes: jiraclient.SearchResult{Issues: []jiraclient.Issue{{Key: "P-1"}}}}
	svc := NewJiraService(gw)

	got, err := svc.ListMyOpenIssues()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Key != "P-1" {
		t.Errorf("got = %+v", got)
	}
	wantFragments := []string{"assignee = currentUser()", "statusCategory != Done"}
	for _, frag := range wantFragments {
		if !strings.Contains(gw.lastJQL, frag) {
			t.Errorf("JQL missing %q: %s", frag, gw.lastJQL)
		}
	}
}

func TestListMyOpenIssues_PropagatesError(t *testing.T) {
	gw := &fakeJiraGW{searchErr: errors.New("boom")}
	svc := NewJiraService(gw)
	if _, err := svc.ListMyOpenIssues(); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestListMyOpenEpics_JQLFiltersEpic(t *testing.T) {
	gw := &fakeJiraGW{}
	svc := NewJiraService(gw)
	_, _ = svc.ListMyOpenEpics()
	if !strings.Contains(gw.lastJQL, "issuetype = Epic") {
		t.Errorf("JQL missing issuetype = Epic: %s", gw.lastJQL)
	}
}

func TestListEpicChildren_RequiresKey(t *testing.T) {
	svc := NewJiraService(&fakeJiraGW{})
	_, err := svc.ListEpicChildren("")
	if err == nil {
		t.Fatal("expected error for empty epicKey")
	}
}

func TestListEpicChildren_QueriesByParent(t *testing.T) {
	gw := &fakeJiraGW{searchRes: jiraclient.SearchResult{Issues: []jiraclient.Issue{{Key: "C-1"}, {Key: "C-2"}}}}
	svc := NewJiraService(gw)

	children, err := svc.ListEpicChildren("EPIC-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 2 {
		t.Errorf("got %d children, want 2", len(children))
	}
	if !strings.Contains(gw.lastJQL, `parent = "EPIC-1"`) {
		t.Errorf("JQL missing parent filter: %s", gw.lastJQL)
	}
}

func TestListEpicChildren_PropagatesError(t *testing.T) {
	gw := &fakeJiraGW{searchErr: errors.New("net")}
	svc := NewJiraService(gw)
	if _, err := svc.ListEpicChildren("E-1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCheckConnection_Success(t *testing.T) {
	gw := &fakeJiraGW{user: jiraclient.User{DisplayName: "V"}}
	svc := NewJiraService(gw)
	u, err := svc.CheckConnection()
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "V" {
		t.Errorf("user = %+v", u)
	}
}

func TestListSpaces_ReturnsAll(t *testing.T) {
	gw := &fakeJiraGW{projects: []jiraclient.Project{
		{Key: "GAP", Name: "Squad GAP"},
		{Key: "BILL", Name: "Squad Billing"},
	}}
	svc := NewJiraService(gw)
	got, err := svc.ListSpaces()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Key != "GAP" || got[1].Name != "Squad Billing" {
		t.Errorf("got = %+v", got)
	}
}

func TestListSpaces_Error(t *testing.T) {
	gw := &fakeJiraGW{projectsErr: errors.New("boom")}
	svc := NewJiraService(gw)
	if _, err := svc.ListSpaces(); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestListProjectBacklog_RequiresKey(t *testing.T) {
	svc := NewJiraService(&fakeJiraGW{})
	if _, err := svc.ListProjectBacklog(""); err == nil {
		t.Error("expected error on empty projectKey")
	}
}

func TestListProjectBacklog_JQL(t *testing.T) {
	gw := &fakeJiraGW{searchRes: jiraclient.SearchResult{Issues: []jiraclient.Issue{{Key: "GAP-1"}}}}
	svc := NewJiraService(gw)
	got, err := svc.ListProjectBacklog("GAP")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Key != "GAP-1" {
		t.Errorf("issues = %+v", got)
	}
	wantFragments := []string{`project = "GAP"`, "statusCategory != Done", "ORDER BY status"}
	absentFragments := []string{"assignee = currentUser()"}
	for _, frag := range wantFragments {
		if !strings.Contains(gw.lastJQL, frag) {
			t.Errorf("JQL missing %q: %s", frag, gw.lastJQL)
		}
	}
	for _, frag := range absentFragments {
		if strings.Contains(gw.lastJQL, frag) {
			t.Errorf("JQL should NOT contain %q (backlog is everyone's): %s", frag, gw.lastJQL)
		}
	}
}

func TestListProjectBacklog_PropagatesError(t *testing.T) {
	gw := &fakeJiraGW{searchErr: errors.New("boom")}
	svc := NewJiraService(gw)
	if _, err := svc.ListProjectBacklog("GAP"); err == nil {
		t.Error("expected error")
	}
}

func TestFilterIssues(t *testing.T) {
	issues := []jiraclient.Issue{
		{Key: "GAP-1", Summary: "Fix login bug"},
		{Key: "GAP-42", Summary: "Add caching"},
		{Key: "BILL-7", Summary: "Retry on 429"},
	}

	tests := []struct {
		query    string
		wantKeys []string
	}{
		{"", []string{"GAP-1", "GAP-42", "BILL-7"}},
		{"   ", []string{"GAP-1", "GAP-42", "BILL-7"}},
		{"gap", []string{"GAP-1", "GAP-42"}},
		{"GAP-42", []string{"GAP-42"}},
		{"login", []string{"GAP-1"}},
		{"CACH", []string{"GAP-42"}},
		{"nonexistent", nil},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := FilterIssues(issues, tt.query)
			var keys []string
			for _, iss := range got {
				keys = append(keys, iss.Key)
			}
			if !reflect.DeepEqual(keys, tt.wantKeys) {
				t.Errorf("FilterIssues(%q) = %v, want %v", tt.query, keys, tt.wantKeys)
			}
		})
	}
}

func TestGroupedByStatus(t *testing.T) {
	issues := []jiraclient.Issue{
		{Key: "A-1", Status: "Desenvolvimento em 2.2"},
		{Key: "A-2", Status: "Desenvolvimento em 2.1"},
		{Key: "A-3", Status: "Desenvolvimento em 2.2"},
		{Key: "A-4", Status: ""},
	}
	order, grouped := GroupedByStatus(issues)
	if len(order) != 3 {
		t.Fatalf("order = %v, want 3 groups", order)
	}
	if order[0] != "Desenvolvimento em 2.2" {
		t.Errorf("first group = %q, want first-seen order", order[0])
	}
	if len(grouped["Desenvolvimento em 2.2"]) != 2 {
		t.Errorf("expected 2 issues in 'Desenvolvimento em 2.2', got %d", len(grouped["Desenvolvimento em 2.2"]))
	}
	if len(grouped["Sem status"]) != 1 {
		t.Errorf("expected 1 issue in 'Sem status' (empty status fallback)")
	}
}

func TestCheckConnection_Error(t *testing.T) {
	gw := &fakeJiraGW{userErr: errors.New("401")}
	svc := NewJiraService(gw)
	if _, err := svc.CheckConnection(); err == nil {
		t.Fatal("expected error")
	}
}
