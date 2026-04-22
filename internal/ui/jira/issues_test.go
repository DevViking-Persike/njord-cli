package jira

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

func (f *fakeIssuesLoader) ListProjectBacklog(key string) ([]jiraclient.Issue, error) {
	f.lastKey = key
	return f.issues, f.err
}

func TestJiraIssues_Init_CallsWithProjectKey(t *testing.T) {
	loader := &fakeIssuesLoader{}
	m := NewIssuesModel(loader, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	cmd := m.Init()
	cmd() // fire to trigger loader
	if loader.lastKey != "GAP" {
		t.Errorf("loader called with key %q, want GAP", loader.lastKey)
	}
}

func TestJiraIssues_GroupsByStatus(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "Task A", Status: "Desenvolvimento em 2.2", Type: "Task"},
		{Key: "GAP-2", Summary: "Task B", Status: "Desenvolvimento em 2.1", Type: "Story"},
		{Key: "GAP-3", Summary: "Task C", Status: "Desenvolvimento em 2.2", Type: "Bug"},
	}})

	view := m.View()
	for _, want := range []string{"Squad GAP — Backlog", "Desenvolvimento em 2.2", "Desenvolvimento em 2.1", "(2)", "(1)", "GAP-1", "GAP-2", "GAP-3"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestJiraIssues_SearchFiltersLive(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "Fix login", Status: "Em dev", Type: "Task"},
		{Key: "GAP-42", Summary: "Retry caching", Status: "Em dev", Type: "Task"},
		{Key: "GAP-7", Summary: "Add logout", Status: "Pronto", Type: "Story"},
	}})

	// Type "login" — só GAP-1 deve aparecer
	for _, ch := range "login" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	view := m.View()
	if !strings.Contains(view, "GAP-1") {
		t.Errorf("expected GAP-1 after search=login, got:\n%s", view)
	}
	if strings.Contains(view, "GAP-42") || strings.Contains(view, "GAP-7") {
		t.Errorf("expected only GAP-1, got:\n%s", view)
	}
	if !strings.Contains(view, "(1 de 3)") {
		t.Errorf("expected counter '(1 de 3)', got:\n%s", view)
	}
}

func TestJiraIssues_SearchByKey(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "A", Status: "s", Type: "Task"},
		{Key: "GAP-42", Summary: "B", Status: "s", Type: "Task"},
	}})
	for _, ch := range "42" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	view := m.View()
	if !strings.Contains(view, "GAP-42") || strings.Contains(view, "GAP-1\n") {
		t.Errorf("search by key 42 should keep only GAP-42, got:\n%s", view)
	}
}

func TestJiraIssues_BackspaceRemovesChar(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{{Key: "GAP-1", Summary: "foo", Status: "s", Type: "T"}}})
	for _, ch := range "gap" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.search != "ga" {
		t.Errorf("search after backspace = %q, want %q", m.search, "ga")
	}
}

func TestJiraIssues_EscClearsSearchThenGoesBack(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{{Key: "GAP-1", Summary: "x", Status: "s", Type: "T"}}})
	for _, ch := range "abc" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	// First esc clears search
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.search != "" {
		t.Errorf("first esc should clear search, got %q", m.search)
	}
	if m.GoBack() {
		t.Error("first esc should NOT go back yet")
	}
	// Second esc goes back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.GoBack() {
		t.Error("second esc should go back")
	}
}

func TestJiraIssues_LoadError(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m, _ = m.Update(issuesLoadedMsg{err: errors.New("401 bad token")})
	if !strings.Contains(m.View(), "401 bad token") {
		t.Errorf("error should be visible, got:\n%s", m.View())
	}
}

func TestJiraIssues_EmptyBacklog(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "X", Name: "X"})
	m, _ = m.Update(issuesLoadedMsg{issues: nil})
	if !strings.Contains(m.View(), "Backlog vazio") {
		t.Errorf("expected empty-backlog message, got:\n%s", m.View())
	}
}

func TestJiraIssues_SearchWithNoResults(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "X", Name: "X"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{issues: []jiraclient.Issue{{Key: "X-1", Summary: "foo", Status: "s", Type: "T"}}})
	for _, ch := range "zzz" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	if !strings.Contains(m.View(), "Nenhuma issue corresponde") {
		t.Errorf("expected no-results message, got:\n%s", m.View())
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
