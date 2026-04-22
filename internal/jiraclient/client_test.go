package jiraclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		in  string
		n   int
		out string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello…"},
		{"", 3, ""},
	}
	for _, tt := range tests {
		if got := truncate(tt.in, tt.n); got != tt.out {
			t.Errorf("truncate(%q,%d) = %q, want %q", tt.in, tt.n, got, tt.out)
		}
	}
}

func TestResolveEpicKey(t *testing.T) {
	tests := []struct {
		name string
		flds rawIssueFlds
		want string
	}{
		{
			name: "EpicLink takes precedence",
			flds: rawIssueFlds{EpicLink: "ABC-1", Parent: &rawParent{Key: "XYZ-9"}},
			want: "ABC-1",
		},
		{
			name: "Parent is Epic",
			flds: rawIssueFlds{Parent: &rawParent{Key: "EPIC-5", Fields: rawParentFlds{IssueType: rawNamedField{Name: "Epic"}}}},
			want: "EPIC-5",
		},
		{
			name: "Parent is Story (ignored)",
			flds: rawIssueFlds{Parent: &rawParent{Key: "STORY-1", Fields: rawParentFlds{IssueType: rawNamedField{Name: "Story"}}}},
			want: "",
		},
		{
			name: "no parent, no epic link",
			flds: rawIssueFlds{},
			want: "",
		},
		{
			name: "Parent type is epic lowercase",
			flds: rawIssueFlds{Parent: &rawParent{Key: "E-7", Fields: rawParentFlds{IssueType: rawNamedField{Name: "epic"}}}},
			want: "E-7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveEpicKey(tt.flds); got != tt.want {
				t.Errorf("resolveEpicKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseIssue_FullFields(t *testing.T) {
	status := rawStatusField{Name: "In Progress"}
	status.Category.Key = "indeterminate"
	r := rawIssue{
		Key: "PROJ-1",
		Fields: rawIssueFlds{
			Summary:   "Fix login",
			Status:    status,
			IssueType: rawNamedField{Name: "Task"},
			Assignee:  &rawAssigneeFld{DisplayName: "Victor"},
			EpicLink:  "EPIC-1",
		},
	}
	got := parseIssue(r)
	want := Issue{Key: "PROJ-1", Summary: "Fix login", Status: "In Progress", StatusCategory: "indeterminate", Type: "Task", Assignee: "Victor", EpicKey: "EPIC-1"}
	if got != want {
		t.Errorf("parseIssue() = %+v, want %+v", got, want)
	}
}

func TestParseIssue_NoAssignee(t *testing.T) {
	r := rawIssue{
		Key: "PROJ-2",
		Fields: rawIssueFlds{
			Summary:   "Unassigned",
			Status:    rawStatusField{Name: "To Do"},
			IssueType: rawNamedField{Name: "Story"},
		},
	}
	got := parseIssue(r)
	if got.Assignee != "" {
		t.Errorf("Assignee = %q, want empty", got.Assignee)
	}
}

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name, baseURL, email, token string
		wantErr                     bool
	}{
		{"ok", "https://x.atlassian.net", "a@b", "t", false},
		{"missing baseURL", "", "a@b", "t", true},
		{"missing email", "https://x", "", "t", true},
		{"missing token", "https://x", "a@b", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.baseURL, tt.email, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	c, err := NewClient("https://x.atlassian.net/", "a@b", "t")
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != "https://x.atlassian.net" {
		t.Errorf("baseURL = %q, want trimmed", c.baseURL)
	}
}

func TestCurrentUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Basic ") {
			t.Errorf("missing basic auth header, got %q", got)
		}
		if r.URL.Path != "/rest/api/3/myself" {
			t.Errorf("path = %q, want /rest/api/3/myself", r.URL.Path)
		}
		fmt.Fprint(w, `{"accountId":"1","displayName":"V","emailAddress":"v@a"}`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	u, err := c.CurrentUser()
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "V" || u.EmailAddress != "v@a" || u.AccountID != "1" {
		t.Errorf("user = %+v", u)
	}
}

func TestCurrentUser_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errorMessages":["Unauthorized"]}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	_, err := c.CurrentUser()
	if err == nil {
		t.Fatal("expected error on HTTP 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got %v", err)
	}
}

func TestSearchIssues_ParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("jql") != "assignee = currentUser()" {
			t.Errorf("jql = %q", r.URL.Query().Get("jql"))
		}
		fmt.Fprint(w, `{
			"total": 2, "isLast": true,
			"issues": [
				{"key":"P-1","fields":{"summary":"a","status":{"name":"To Do"},"issuetype":{"name":"Task"}}},
				{"key":"P-2","fields":{"summary":"b","status":{"name":"Done"},"issuetype":{"name":"Bug"},"assignee":{"displayName":"X"}}}
			]
		}`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	res, err := c.SearchIssues("assignee = currentUser()")
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 2 || len(res.Issues) != 2 {
		t.Fatalf("res = %+v", res)
	}
	if res.Issues[0].Key != "P-1" || res.Issues[1].Assignee != "X" {
		t.Errorf("issues = %+v", res.Issues)
	}
	if !res.IsLast {
		t.Error("expected IsLast=true")
	}
}

func TestSearchIssues_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad jql", http.StatusBadRequest)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	_, err := c.SearchIssues("invalid jql")
	if err == nil {
		t.Fatal("expected error on HTTP 400")
	}
}

func TestListProjects_ReturnsRecent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project/recent" {
			t.Errorf("path = %q, want /rest/api/3/project/recent", r.URL.Path)
		}
		if r.URL.Query().Get("maxResults") != "50" {
			t.Errorf("maxResults = %q, want 50", r.URL.Query().Get("maxResults"))
		}
		fmt.Fprint(w, `[
			{"id":"1","key":"GAP","name":"Squad GAP"},
			{"id":"2","key":"BILL","name":"Squad Billing"},
			{"id":"3","key":"SPAVT","name":"Suporte Tecnologia"}
		]`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	projects, err := c.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 3 {
		t.Fatalf("len = %d, want 3", len(projects))
	}
	if projects[0].Key != "GAP" || projects[2].Name != "Suporte Tecnologia" {
		t.Errorf("projects = %+v", projects)
	}
}

func TestListProjects_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	projects, err := c.ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 0 {
		t.Errorf("expected empty slice, got %d items", len(projects))
	}
}

func TestListProjects_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	if _, err := c.ListProjects(); err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchIssues_DecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not-json`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, "v@a", "t", WithHTTPClient(srv.Client()))
	_, err := c.SearchIssues("x")
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("expected decode error, got %v", err)
	}
}
