package jira

import (
	"errors"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeLoader struct {
	projects []jiraclient.Project
	err      error
}

func (f *fakeLoader) ListSpaces() ([]jiraclient.Project, error) {
	return f.projects, f.err
}

func TestJiraSpaces_InitShowsLoading(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{})
	if !strings.Contains(m.View(), "Carregando") {
		t.Error("expected loading state in initial view")
	}
}

func TestJiraSpaces_LoadedShowsProjects(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{projects: []jiraclient.Project{
		{Key: "GAP", Name: "Squad GAP"},
		{Key: "BILL", Name: "Squad Billing"},
	}})
	m.SetSize(80, 40)

	m, _ = m.Update(spacesLoadedMsg{projects: []jiraclient.Project{
		{Key: "GAP", Name: "Squad GAP"},
		{Key: "BILL", Name: "Squad Billing"},
	}})

	view := m.View()
	if !strings.Contains(view, "Squad GAP") || !strings.Contains(view, "Squad Billing") {
		t.Errorf("view missing projects:\n%s", view)
	}
}

func TestJiraSpaces_LoadError(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{err: errors.New("401 unauthorized")})
	m, _ = m.Update(spacesLoadedMsg{err: errors.New("401 unauthorized")})

	view := m.View()
	if !strings.Contains(view, "401 unauthorized") {
		t.Errorf("error should be visible, got:\n%s", view)
	}
}

func TestJiraSpaces_EscTriggersGoBack(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{})
	m, _ = m.Update(spacesLoadedMsg{projects: nil})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.GoBack() {
		t.Error("esc should trigger goBack")
	}
}

func TestJiraSpaces_EnterSelectsProject(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{})
	m.SetSize(80, 40)
	m, _ = m.Update(spacesLoadedMsg{projects: []jiraclient.Project{
		{Key: "GAP", Name: "Squad GAP"},
	}})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	sel := m.Selected()
	if sel == nil || sel.Key != "GAP" {
		t.Errorf("expected GAP selected, got %+v", sel)
	}

	m.ClearSelection()
	if m.Selected() != nil {
		t.Error("ClearSelection did not reset selection")
	}
}

func TestJiraSpaces_EmptyList(t *testing.T) {
	m := NewSpacesModel(&fakeLoader{})
	m, _ = m.Update(spacesLoadedMsg{projects: nil})
	if !strings.Contains(m.View(), "Nenhum projeto encontrado") {
		t.Errorf("expected empty-state message, got:\n%s", m.View())
	}
}
