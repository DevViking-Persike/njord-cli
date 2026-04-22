package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jira"
)

type fakeJiraGW struct {
	user       jira.User
	userErr    error
	searchRes  jira.SearchResult
	searchErr  error
	lastJQL    string
	callCount  int
}

func (f *fakeJiraGW) CurrentUser() (jira.User, error) {
	f.callCount++
	return f.user, f.userErr
}

func (f *fakeJiraGW) SearchIssues(jql string) (jira.SearchResult, error) {
	f.lastJQL = jql
	f.callCount++
	return f.searchRes, f.searchErr
}

func TestListMyOpenIssues_UsesExpectedJQL(t *testing.T) {
	gw := &fakeJiraGW{searchRes: jira.SearchResult{Issues: []jira.Issue{{Key: "P-1"}}}}
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
	gw := &fakeJiraGW{searchRes: jira.SearchResult{Issues: []jira.Issue{{Key: "C-1"}, {Key: "C-2"}}}}
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
	gw := &fakeJiraGW{user: jira.User{DisplayName: "V"}}
	svc := NewJiraService(gw)
	u, err := svc.CheckConnection()
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "V" {
		t.Errorf("user = %+v", u)
	}
}

func TestCheckConnection_Error(t *testing.T) {
	gw := &fakeJiraGW{userErr: errors.New("401")}
	svc := NewJiraService(gw)
	if _, err := svc.CheckConnection(); err == nil {
		t.Fatal("expected error")
	}
}
