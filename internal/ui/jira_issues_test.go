package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeIssuesLoader struct {
	lastKey string
	issues  []jiraclient.Issue
	err     error
}

func (f *fakeIssuesLoader) ListMyIssuesInProject(key string) ([]jiraclient.Issue, error) {
	f.lastKey = key
	return f.issues, f.err
}

func TestJiraIssues_Init_CallsWithProjectKey(t *testing.T) {
	loader := &fakeIssuesLoader{}
	m := NewJiraIssuesModel(loader, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	cmd := m.Init()
	cmd() // fire to trigger loader
	if loader.lastKey != "GAP" {
		t.Errorf("loader called with key %q, want GAP", loader.lastKey)
	}
}

func TestJiraIssues_GroupsByStatus(t *testing.T) {
	m := NewJiraIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(jiraIssuesLoadedMsg{issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "Task A", Status: "Desenvolvimento em 2.2", Type: "Task"},
		{Key: "GAP-2", Summary: "Task B", Status: "Desenvolvimento em 2.1", Type: "Story"},
		{Key: "GAP-3", Summary: "Task C", Status: "Desenvolvimento em 2.2", Type: "Bug"},
	}})

	view := m.View()
	for _, want := range []string{"Squad GAP — Minhas issues", "Desenvolvimento em 2.2", "Desenvolvimento em 2.1", "(2)", "(1)", "GAP-1", "GAP-2", "GAP-3"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestJiraIssues_LoadError(t *testing.T) {
	m := NewJiraIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m, _ = m.Update(jiraIssuesLoadedMsg{err: errors.New("401 bad token")})
	if !strings.Contains(m.View(), "401 bad token") {
		t.Errorf("error should be visible, got:\n%s", m.View())
	}
}

func TestJiraIssues_EmptyState(t *testing.T) {
	m := NewJiraIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "X", Name: "X"})
	m, _ = m.Update(jiraIssuesLoadedMsg{issues: nil})
	if !strings.Contains(m.View(), "Nenhuma issue") {
		t.Errorf("expected empty-state message, got:\n%s", m.View())
	}
}

func TestJiraIssues_EscGoesBack(t *testing.T) {
	m := NewJiraIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "X"})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.GoBack() {
		t.Error("esc should trigger goBack")
	}
}

func TestFormatIssueLine_TruncatesLongSummary(t *testing.T) {
	iss := jiraclient.Issue{Key: "X-1", Type: "Task", Summary: strings.Repeat("a", 200)}
	line := formatIssueLine(iss)
	if !strings.Contains(line, "...") {
		t.Error("expected truncation ellipsis in long summary")
	}
	if !strings.Contains(line, "X-1") || !strings.Contains(line, "[Task]") {
		t.Errorf("missing key or type, got %q", line)
	}
}
