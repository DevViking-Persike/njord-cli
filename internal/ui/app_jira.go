package ui

import (
	jiraapp "github.com/DevViking-Persike/njord-cli/internal/app/jira"
	jiraui "github.com/DevViking-Persike/njord-cli/internal/ui/jira"
	tea "github.com/charmbracelet/bubbletea"
)

// Handlers das telas do universo Jira: spaces → actions → (issues | create | edit).
// Separado de app.go pra manter cada arquivo ≤ 300 linhas (regra 1).

func (m AppModel) updateJiraSpaces(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraSpaces, cmd = m.jiraSpaces.Update(msg)

	if m.jiraSpaces.GoBack() {
		m.screen = ScreenGrid
		return m, nil
	}
	if sel := m.jiraSpaces.Selected(); sel != nil {
		m.jiraSpaces.ClearSelection()
		m.jiraActions = jiraui.NewSpaceActionsModel(*sel)
		m.jiraActions.SetSize(m.width, m.height)
		m.screen = ScreenJiraActions
		return m, m.jiraActions.Init()
	}
	return m, cmd
}

func (m AppModel) updateJiraActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraActions, cmd = m.jiraActions.Update(msg)

	if m.jiraActions.GoBack() {
		m.screen = ScreenJiraSpaces
		return m, nil
	}
	if sel := m.jiraActions.Selected(); sel != nil {
		m.jiraActions.ClearSelection()
		proj := m.jiraActions.Project()
		svc := jiraapp.NewJiraService(m.jiraClient)
		switch *sel {
		case jiraui.ActionList:
			m.jiraIssues = jiraui.NewIssuesModel(svc, proj)
			m.jiraIssues.SetSize(m.width, m.height)
			m.screen = ScreenJiraIssues
			return m, m.jiraIssues.Init()
		case jiraui.ActionCreate:
			m.jiraCreate = jiraui.NewCreateModel(svc, proj)
			m.jiraCreate.SetSize(m.width, m.height)
			m.screen = ScreenJiraCreate
			return m, m.jiraCreate.Init()
		case jiraui.ActionEdit:
			m.jiraEdit = jiraui.NewEditModel(svc, proj)
			m.jiraEdit.SetSize(m.width, m.height)
			m.screen = ScreenJiraEdit
			return m, m.jiraEdit.Init()
		}
	}
	return m, cmd
}

func (m AppModel) updateJiraIssues(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraIssues, cmd = m.jiraIssues.Update(msg)

	if m.jiraIssues.GoBack() {
		m.screen = ScreenJiraActions
		return m, nil
	}
	return m, cmd
}

func (m AppModel) updateJiraCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraCreate, cmd = m.jiraCreate.Update(msg)

	if m.jiraCreate.GoBack() {
		m.screen = ScreenJiraActions
		return m, nil
	}
	return m, cmd
}

func (m AppModel) updateJiraEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraEdit, cmd = m.jiraEdit.Update(msg)

	if m.jiraEdit.GoBack() {
		m.screen = ScreenJiraActions
		return m, nil
	}
	return m, cmd
}
