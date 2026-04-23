package jiraclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, "me@example.com", "token")
	if err != nil {
		t.Fatal(err)
	}
	return c, srv
}

func TestValidateCreate(t *testing.T) {
	cases := []struct {
		name    string
		in      CreateIssueInput
		wantErr bool
	}{
		{"ok task", CreateIssueInput{ProjectKey: "GAP", Summary: "x", Type: "Task"}, false},
		{"subtask sem parent", CreateIssueInput{ProjectKey: "GAP", Summary: "x", Type: "Subtask"}, true},
		{"subtask com parent", CreateIssueInput{ProjectKey: "GAP", Summary: "x", Type: "Subtask", ParentKey: "GAP-1"}, false},
		{"sem project", CreateIssueInput{Summary: "x", Type: "Task"}, true},
		{"sem summary", CreateIssueInput{ProjectKey: "GAP", Type: "Task"}, true},
		{"sem type", CreateIssueInput{ProjectKey: "GAP", Summary: "x"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreate(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestIsSubtaskType(t *testing.T) {
	cases := map[string]bool{
		"Subtask":  true,
		"subtask":  true,
		"Sub-task": true,
		"SUB-TASK": true,
		"Task":     false,
		"":         false,
	}
	for in, want := range cases {
		if got := isSubtaskType(in); got != want {
			t.Errorf("isSubtaskType(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestBuildCreateFields_Task(t *testing.T) {
	fields := buildCreateFields(CreateIssueInput{
		ProjectKey: "GAP", Summary: "titulo", Type: "Task",
		Description: "desc", AssigneeAccount: "acc-1",
	})
	if fields["summary"] != "titulo" {
		t.Fatalf("summary mapping: %v", fields["summary"])
	}
	if _, has := fields["parent"]; has {
		t.Fatalf("Task não deve ter parent: %+v", fields)
	}
	assignee := fields["assignee"].(map[string]string)
	if assignee["accountId"] != "acc-1" {
		t.Fatalf("assignee = %+v", assignee)
	}
}

func TestBuildCreateFields_SubtaskWithParent(t *testing.T) {
	fields := buildCreateFields(CreateIssueInput{
		ProjectKey: "GAP", Summary: "s", Type: "Subtask", ParentKey: "GAP-42",
	})
	parent, ok := fields["parent"].(map[string]string)
	if !ok || parent["key"] != "GAP-42" {
		t.Fatalf("parent = %+v", fields["parent"])
	}
}

func TestCreateIssue_Success(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/issue" {
			t.Fatalf("unexpected req: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		fields := body["fields"].(map[string]any)
		if fields["summary"] != "oi" {
			t.Fatalf("summary = %v", fields["summary"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"key":"GAP-99"}`))
	})
	issue, err := c.CreateIssue(CreateIssueInput{ProjectKey: "GAP", Summary: "oi", Type: "Task"})
	if err != nil {
		t.Fatal(err)
	}
	if issue.Key != "GAP-99" {
		t.Fatalf("Key = %q", issue.Key)
	}
}

func TestCreateIssue_HTTPError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"errorMessages":["bad"]}`))
	})
	if _, err := c.CreateIssue(CreateIssueInput{ProjectKey: "GAP", Summary: "x", Type: "Task"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateIssue_SkipsEmptyFields(t *testing.T) {
	err := (&Client{}).UpdateIssue("GAP-1", UpdateIssueInput{})
	if err == nil || !strings.Contains(err.Error(), "nada pra atualizar") {
		t.Fatalf("expected 'nada pra atualizar' err, got %v", err)
	}
}

func TestUpdateIssue_SendsPut(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/rest/api/3/issue/GAP-1" {
			t.Fatalf("unexpected req: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	})
	if err := c.UpdateIssue("GAP-1", UpdateIssueInput{Summary: "novo"}); err != nil {
		t.Fatal(err)
	}
}

func TestListTransitions(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/GAP-1/transitions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"transitions":[{"id":"11","name":"Em desenvolvimento","to":{"name":"Em desenvolvimento","statusCategory":{"key":"indeterminate"}}}]}`))
	})
	ts, err := c.ListTransitions("GAP-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 1 || ts[0].ID != "11" || ts[0].StatusCat != "indeterminate" {
		t.Fatalf("got %+v", ts)
	}
}

func TestTransitionIssue(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/rest/api/3/issue/GAP-1/transitions" {
			t.Fatalf("unexpected req")
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["transition"].(map[string]any)["id"] != "11" {
			t.Fatalf("payload = %v", body)
		}
		w.WriteHeader(204)
	})
	if err := c.TransitionIssue("GAP-1", "11"); err != nil {
		t.Fatal(err)
	}
}

func TestADFDescription(t *testing.T) {
	doc := adfDescription("hello")
	if doc["type"] != "doc" {
		t.Fatalf("type = %v", doc["type"])
	}
}
