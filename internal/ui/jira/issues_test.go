package jira

import (
	"errors"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeIssuesLoader struct {
	backlogKey  string
	mineKey     string
	backlog     []jiraclient.Issue
	mine        []jiraclient.Issue
	backlogErr  error
	mineErr     error
}

func (f *fakeIssuesLoader) ListProjectBacklog(key string) ([]jiraclient.Issue, error) {
	f.backlogKey = key
	return f.backlog, f.backlogErr
}

func (f *fakeIssuesLoader) ListMyProjectIssues(key string) ([]jiraclient.Issue, error) {
	f.mineKey = key
	return f.mine, f.mineErr
}

func TestJiraIssues_InitLoadsBacklog(t *testing.T) {
	loader := &fakeIssuesLoader{}
	m := NewIssuesModel(loader, jiraclient.Project{Key: "GAP", Name: "Squad GAP"})
	m.Init()()
	if loader.backlogKey != "GAP" || loader.mineKey != "" {
		t.Errorf("initial load should hit backlog only; backlog=%q mine=%q", loader.backlogKey, loader.mineKey)
	}
}

func TestJiraIssues_RightArrowSwitchesToMine(t *testing.T) {
	loader := &fakeIssuesLoader{}
	m := NewIssuesModel(loader, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	// First deliver backlog
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: []jiraclient.Issue{{Key: "GAP-1", Status: "Em dev", StatusCategory: "indeterminate", Type: "Task"}}})
	// Press right
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.mode != modeMine {
		t.Errorf("mode = %v, want modeMine", m.mode)
	}
	// Fire returned cmd — should call ListMyProjectIssues
	if cmd == nil {
		t.Fatal("expected load command after mode switch")
	}
	cmd()
	if loader.mineKey != "GAP" {
		t.Errorf("mine loader should have been called with GAP, got %q", loader.mineKey)
	}
}

func TestJiraIssues_LeftArrowGoesBacklog(t *testing.T) {
	loader := &fakeIssuesLoader{}
	m := NewIssuesModel(loader, jiraclient.Project{Key: "GAP"})
	m.mode = modeMine
	m.loading = false
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.mode != modeBacklog {
		t.Errorf("mode should be modeBacklog after left")
	}
	if cmd == nil {
		t.Fatal("expected load command")
	}
}

func TestJiraIssues_TabCyclesStatusFilter(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: nil})

	wantCycle := []string{"indeterminate", "done", "new", ""}
	for _, want := range wantCycle {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		if m.statusFilter != want {
			t.Errorf("expected statusFilter=%q after tab, got %q", want, m.statusFilter)
		}
	}
}

func TestJiraIssues_StatusFilterHidesOtherCategories(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "A", Status: "Em dev", StatusCategory: "indeterminate", Type: "Task"},
		{Key: "GAP-2", Summary: "B", Status: "Concluído", StatusCategory: "done", Type: "Task"},
	}})

	// Cycle once → "indeterminate"
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	view := m.View()
	if !strings.Contains(view, "GAP-1") {
		t.Errorf("expected GAP-1 when filter=indeterminate, got:\n%s", view)
	}
	if strings.Contains(view, "GAP-2") {
		t.Errorf("GAP-2 should be hidden when filter=indeterminate, got:\n%s", view)
	}
}

func TestJiraIssues_StaleResponseIgnored(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.mode = modeMine
	// Delivery of backlog while mode is mine — ignore.
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: []jiraclient.Issue{{Key: "STALE-1"}}})
	if m.issues != nil && len(m.issues) > 0 {
		t.Errorf("stale response from mode %v should be ignored in mode %v", modeBacklog, m.mode)
	}
}

func TestJiraIssues_SearchFiltersLive(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: []jiraclient.Issue{
		{Key: "GAP-1", Summary: "Fix login", Status: "Em dev", Type: "Task"},
		{Key: "GAP-42", Summary: "Retry caching", Status: "Em dev", Type: "Task"},
	}})
	for _, ch := range "login" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	view := m.View()
	if !strings.Contains(view, "GAP-1") || strings.Contains(view, "GAP-42") {
		t.Errorf("search=login should keep only GAP-1, got:\n%s", view)
	}
}

func TestJiraIssues_EscClearsSearchThenGoesBack(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m.SetSize(120, 40)
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, issues: []jiraclient.Issue{{Key: "GAP-1", Status: "s", Type: "T"}}})
	for _, ch := range "abc" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	// First esc clears
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.search != "" || m.GoBack() {
		t.Errorf("first esc should clear search not go back; search=%q goBack=%v", m.search, m.GoBack())
	}
	// Second esc goes back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.GoBack() {
		t.Error("second esc should go back")
	}
}

func TestJiraIssues_LoadError(t *testing.T) {
	m := NewIssuesModel(&fakeIssuesLoader{}, jiraclient.Project{Key: "GAP"})
	m, _ = m.Update(issuesLoadedMsg{mode: modeBacklog, err: errors.New("401")})
	if !strings.Contains(m.View(), "401") {
		t.Errorf("error not visible, got:\n%s", m.View())
	}
}

func TestNextInCycle(t *testing.T) {
	cycle := []string{"a", "b", "c"}
	tests := []struct{ in, want string }{
		{"", "a"},
		{"a", "b"},
		{"b", "c"},
		{"c", "a"},
		{"unknown", "a"},
	}
	for _, tt := range tests {
		if got := nextInCycle(cycle, tt.in); got != tt.want {
			t.Errorf("nextInCycle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatIssueLine_TruncatesLongSummary(t *testing.T) {
	iss := jiraclient.Issue{Key: "X-1", Type: "Task", Summary: strings.Repeat("a", 200)}
	line := formatIssueLine(iss)
	if !strings.Contains(line, "...") {
		t.Error("expected truncation ellipsis in long summary")
	}
}
