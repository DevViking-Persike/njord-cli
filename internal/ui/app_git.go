package ui

import (
	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	projectapp "github.com/DevViking-Persike/njord-cli/internal/app/project"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/githubclient"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	cloneui "github.com/DevViking-Persike/njord-cli/internal/ui/clone"
	githubui "github.com/DevViking-Persike/njord-cli/internal/ui/github"
	gitlabui "github.com/DevViking-Persike/njord-cli/internal/ui/gitlab"
	projectui "github.com/DevViking-Persike/njord-cli/internal/ui/project"
	"github.com/DevViking-Persike/njord-cli/internal/ui/settings"
	tea "github.com/charmbracelet/bubbletea"
)

// Handlers das telas do "universo Git": hub GitLab + listas (GitLab/GitHub/Local)
// e ações correspondentes. Separado de app.go pra manter cada arquivo ≤ 300
// linhas (regra 1).

func (m AppModel) updateGitLabHub(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.gitlabHub, cmd = m.gitlabHub.Update(msg)

	if m.gitlabHub.GoBack() {
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
			m.refreshGrid()
		}
		m.screen = ScreenGrid
		return m, nil
	}
	if sel := m.gitlabHub.Selected(); sel != nil {
		m.gitlabHub.ClearSelection()
		return m.openHubSelection(*sel)
	}
	return m, cmd
}

func (m AppModel) openHubSelection(kind gitlabui.HubKind) (tea.Model, tea.Cmd) {
	switch kind {
	case gitlabui.HubGitLab:
		if m.config.GitLab.Token == "" {
			m.settings = settings.NewModel(m.config, m.configPath)
			m.settings.SetSize(m.width, m.height)
			m.screen = ScreenSettings
			return m, m.settings.Init()
		}
		m.ensureGitLabClient()
		m.gitlabScreen = gitlabui.NewModel(m.config, m.configPath, m.gitlabClient, githubapp.FilterGitLab)
		m.gitlabScreen.SetSize(m.width, m.height)
		m.screen = ScreenGitLab
		return m, m.gitlabScreen.Init()
	case gitlabui.HubGitHub:
		m.githubList = githubui.NewModel(m.config)
		m.githubList.SetSize(m.width, m.height)
		m.screen = ScreenGitHubList
		return m, m.githubList.Init()
	case gitlabui.HubLocal:
		m.localList = gitlabui.NewLocalModel(m.config)
		m.localList.SetSize(m.width, m.height)
		m.screen = ScreenLocalList
		return m, m.localList.Init()
	case gitlabui.HubClone:
		m.ensureGitLabClient()
		m.ensureGitHubClient()
		m.cloneGroups = cloneui.NewGroupsModel(m.gitlabClient)
		m.cloneGroups.SetSize(m.width, m.height)
		m.screen = ScreenCloneGroups
		return m, m.cloneGroups.Init()
	}
	return m, nil
}

func (m *AppModel) ensureGitLabClient() {
	if m.gitlabClient != nil || m.config.GitLab.Token == "" {
		return
	}
	if client, err := gitlabclient.NewClient(m.config.GitLab.Token, m.config.GitLab.GitLabURL()); err == nil {
		m.gitlabClient = client
	}
}

func (m *AppModel) ensureGitHubClient() {
	if m.githubClient != nil || m.config.GitHub.Token == "" {
		return
	}
	if client, err := githubclient.NewClient(m.config.GitHub.Token); err == nil {
		m.githubClient = client
	}
}

func (m AppModel) updateCloneGroups(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.cloneGroups, cmd = m.cloneGroups.Update(msg)

	if m.cloneGroups.GoBack() {
		m.gitlabHub = gitlabui.NewHubModel(m.config)
		m.gitlabHub.SetSize(m.width, m.height)
		m.screen = ScreenGitLabHub
		return m, m.gitlabHub.Init()
	}
	if scope := m.cloneGroups.Selected(); scope != nil {
		m.cloneGroups.ClearSelection()
		m.cloneScreen = cloneui.NewModel(m.gitlabClient, m.githubClient, *scope)
		m.cloneScreen.SetSize(m.width, m.height)
		m.screen = ScreenClone
		return m, m.cloneScreen.Init()
	}
	return m, cmd
}

func (m AppModel) updateClone(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.cloneScreen, cmd = m.cloneScreen.Update(msg)

	if m.cloneScreen.GoBack() {
		// Um esc volta pro scope picker, não pro hub.
		m.screen = ScreenCloneGroups
		return m, nil
	}
	if sel := m.cloneScreen.Selected(); sel != nil {
		m.cloneScreen.ClearSelection()
		m.addProject = projectui.NewAddModelWithURL(m.config, m.configPath, sel.CloneSSH)
		m.addProject.SetSize(m.width, m.height)
		m.screen = ScreenAddProject
		return m, m.addProject.Init()
	}
	return m, cmd
}

func (m AppModel) updateGitLab(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.gitlabScreen, cmd = m.gitlabScreen.Update(msg)

	if m.gitlabScreen.GoBack() {
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
			m.refreshGrid()
		}
		m.gitlabHub = gitlabui.NewHubModel(m.config)
		m.gitlabHub.SetSize(m.width, m.height)
		m.screen = ScreenGitLabHub
		return m, m.gitlabHub.Init()
	}

	if sel := m.gitlabScreen.Selected(); sel != nil {
		m.gitlabScreen.ClearSelection()
		if m.gitlabClient != nil {
			m.gitlabActions = gitlabui.NewActionsModel(m.gitlabClient, sel.GitLabPath, sel.Alias, m.config.GitLab.GitLabURL())
			m.gitlabActions.SetSize(m.width, m.height)
			m.screen = ScreenGitLabActions
			return m, m.gitlabActions.Init()
		}
	}

	return m, cmd
}

func (m AppModel) updateGitLabActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.gitlabActions, cmd = m.gitlabActions.Update(msg)

	if m.gitlabActions.GoBack() {
		m.gitlabScreen = gitlabui.NewModel(m.config, m.configPath, m.gitlabClient, githubapp.FilterGitLab)
		m.gitlabScreen.SetSize(m.width, m.height)
		m.screen = ScreenGitLab
		return m, m.gitlabScreen.Init()
	}

	return m, cmd
}

func (m AppModel) updateGitHubList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.githubList, cmd = m.githubList.Update(msg)

	if m.githubList.GoBack() {
		m.gitlabHub = gitlabui.NewHubModel(m.config)
		m.gitlabHub.SetSize(m.width, m.height)
		m.screen = ScreenGitLabHub
		return m, m.gitlabHub.Init()
	}
	if sel := m.githubList.Selected(); sel != nil {
		m.githubList.ClearSelection()
		m.githubActions = githubui.NewActionsModel(m.config, m.configPath, *sel)
		m.githubActions.SetSize(m.width, m.height)
		m.screen = ScreenGitHubActions
		return m, m.githubActions.Init()
	}
	return m, cmd
}

func (m AppModel) updateGitHubActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.githubActions, cmd = m.githubActions.Update(msg)

	if command := m.githubActions.Command(); command != "" {
		m.githubActions.ClearCommand()
		m.result = &Result{Command: command}
		m.quitting = true
		return m, tea.Quit
	}

	if m.githubActions.GoBack() {
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
		}
		m.githubList = githubui.NewModel(m.config)
		m.githubList.SetSize(m.width, m.height)
		m.screen = ScreenGitHubList
		return m, m.githubList.Init()
	}

	return m, cmd
}

func (m AppModel) updateLocalList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.localList, cmd = m.localList.Update(msg)

	if m.localList.GoBack() {
		m.gitlabHub = gitlabui.NewHubModel(m.config)
		m.gitlabHub.SetSize(m.width, m.height)
		m.screen = ScreenGitLabHub
		return m, m.gitlabHub.Init()
	}
	if proj := m.localList.Selected(); proj != nil {
		command, err := projectapp.BuildProjectCommand(m.config, *proj)
		if err != nil {
			return m, nil
		}
		m.result = &Result{Command: command}
		m.quitting = true
		return m, tea.Quit
	}
	return m, cmd
}
