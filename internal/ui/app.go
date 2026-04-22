package ui

import (
	"fmt"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/app/jira"
	"github.com/DevViking-Persike/njord-cli/internal/app/project"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Periodic refresh tick
type gitlabRefreshTickMsg struct{}

// Async message for GitLab client initialization
type gitlabInitMsg struct {
	client *gitlabclient.Client
	cfg    *config.Config
	err    error
}

// Async message for recent pushes
type recentPushesMsg struct {
	pushes []RecentPushAlias
	err    error
}

// Async message for pending MRs
type pendingMRsMsg struct {
	mrs []PendingMRAlias
	err error
}

type Screen int

const (
	ScreenGrid Screen = iota
	ScreenProjects
	ScreenDocker
	ScreenDockerActions
	ScreenAddProject
	ScreenAddStack
	ScreenSettings
	ScreenGitLab
	ScreenGitLabActions
	ScreenJiraSpaces
	ScreenJiraIssues
)

// Result is the final output from the TUI.
type Result struct {
	// Command is the shell command to eval (e.g. cd "/path" && code .)
	Command string
}

type AppModel struct {
	screen Screen
	config *config.Config
	docker *docker.Client
	result *Result
	width  int
	height int

	// Sub-models
	grid          GridModel
	projects      ProjectsModel
	dockerScreen  DockerModel
	dockerActions DockerActionsModel
	addProject    AddProjectModel
	addStack      AddStackModel
	settings      SettingsModel
	gitlabScreen  GitLabModel
	gitlabActions GitLabActionsModel
	jiraSpaces    JiraSpacesModel
	jiraIssues    JiraIssuesModel

	// GitLab client (lazy init)
	gitlabClient *gitlabclient.Client

	// Jira client (lazy init)
	jiraClient *jiraclient.Client

	// Context for transitions
	selectedCatID string
	selectedStack *config.DockerStack
	configPath    string
	quitting      bool
}

func NewApp(cfg *config.Config, dockerClient *docker.Client, configPath string) AppModel {
	grid := NewGridModel(cfg)
	return AppModel{
		screen:     ScreenGrid,
		config:     cfg,
		docker:     dockerClient,
		grid:       grid,
		configPath: configPath,
	}
}

func (m AppModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.grid.Init()}
	if m.config.GitLab.Token != "" {
		cmds = append(cmds, m.initGitLabFetches())
	}
	return tea.Batch(cmds...)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.grid.SetSize(msg.Width, msg.Height)
		m.projects.SetSize(msg.Width, msg.Height)
		m.dockerScreen.SetSize(msg.Width, msg.Height)
		m.dockerActions.SetSize(msg.Width, msg.Height)
		m.addProject.SetSize(msg.Width, msg.Height)
		m.addStack.SetSize(msg.Width, msg.Height)
		m.settings.SetSize(msg.Width, msg.Height)
		m.gitlabScreen.SetSize(msg.Width, msg.Height)
		m.gitlabActions.SetSize(msg.Width, msg.Height)
		m.jiraSpaces.SetSize(msg.Width, msg.Height)
		m.jiraIssues.SetSize(msg.Width, msg.Height)
		return m, nil

	case gitlabInitMsg:
		if msg.err != nil {
			return m, nil
		}
		m.gitlabClient = msg.client
		return m, tea.Batch(
			fetchRecentPushes(msg.client, msg.cfg),
			fetchPendingMRs(msg.client, msg.cfg),
			scheduleRefresh(),
		)

	case gitlabRefreshTickMsg:
		if m.gitlabClient != nil {
			return m, tea.Batch(
				fetchRecentPushes(m.gitlabClient, m.config),
				fetchPendingMRs(m.gitlabClient, m.config),
				scheduleRefresh(),
			)
		}
		return m, nil

	case recentPushesMsg:
		m.grid.SetRecentPushes(msg.pushes)
		if msg.err != nil {
			m.grid.pushError = msg.err.Error()
		}
		return m, nil

	case pendingMRsMsg:
		m.grid.SetPendingMRs(msg.mrs)
		if msg.err != nil {
			m.grid.mrsError = msg.err.Error()
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.screen {
	case ScreenGrid:
		return m.updateGrid(msg)
	case ScreenProjects:
		return m.updateProjects(msg)
	case ScreenDocker:
		return m.updateDocker(msg)
	case ScreenDockerActions:
		return m.updateDockerActions(msg)
	case ScreenAddProject:
		return m.updateAddProject(msg)
	case ScreenAddStack:
		return m.updateAddStack(msg)
	case ScreenSettings:
		return m.updateSettings(msg)
	case ScreenGitLab:
		return m.updateGitLab(msg)
	case ScreenGitLabActions:
		return m.updateGitLabActions(msg)
	case ScreenJiraSpaces:
		return m.updateJiraSpaces(msg)
	case ScreenJiraIssues:
		return m.updateJiraIssues(msg)
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.quitting {
		return ""
	}

	var content string
	switch m.screen {
	case ScreenGrid:
		content = m.grid.View()
	case ScreenProjects:
		content = m.projects.View()
	case ScreenDocker:
		content = m.dockerScreen.View()
	case ScreenDockerActions:
		content = m.dockerActions.View()
	case ScreenAddProject:
		content = m.addProject.View()
	case ScreenAddStack:
		content = m.addStack.View()
	case ScreenSettings:
		content = m.settings.View()
	case ScreenGitLab:
		content = m.gitlabScreen.View()
	case ScreenGitLabActions:
		content = m.gitlabActions.View()
	case ScreenJiraSpaces:
		content = m.jiraSpaces.View()
	case ScreenJiraIssues:
		content = m.jiraIssues.View()
	}

	help := theme.HelpStyle.Render("  ↑↓←→ navigate  enter select  esc back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "", help)
}

// GetResult returns the result after the TUI exits.
func (m AppModel) GetResult() *Result {
	return m.result
}

// refreshGrid recria o grid com a config atualizada, preservando dados do header.
func (m *AppModel) refreshGrid() {
	pushes := m.grid.recentPushes
	mrs := m.grid.pendingMRs
	pushErr := m.grid.pushError
	mrsErr := m.grid.mrsError

	m.grid = NewGridModel(m.config)
	m.grid.SetSize(m.width, m.height)
	m.grid.recentPushes = pushes
	m.grid.pendingMRs = mrs
	m.grid.pushError = pushErr
	m.grid.mrsError = mrsErr
}

// --- Screen transition handlers ---

func (m AppModel) updateGrid(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.grid, cmd = m.grid.Update(msg)

	// Check for grid selection
	if sel := m.grid.Selected(); sel != nil {
		m.grid.ClearSelection()

		switch sel.Type {
		case GridItemCategory:
			if sel.CatID == "*" {
				// "Todos" category - show all projects
				m.projects = NewProjectsModel(m.config, sel.CatID, m.config.AllProjects())
				m.projects.SetSize(m.width, m.height)
			} else {
				for _, cat := range m.config.Categories {
					if cat.ID == sel.CatID {
						m.projects = NewProjectsModel(m.config, sel.CatID, cat.Projects)
						m.projects.SetSize(m.width, m.height)
						break
					}
				}
			}
			m.selectedCatID = sel.CatID
			m.screen = ScreenProjects
			return m, m.projects.Init()

		case GridItemDocker:
			m.dockerScreen = NewDockerModel(m.config, m.docker, m.configPath)
			m.dockerScreen.SetSize(m.width, m.height)
			m.screen = ScreenDocker
			return m, m.dockerScreen.Init()

		case GridItemGitLab:
			// Check if token is configured
			if m.config.GitLab.Token == "" {
				// Redirect to settings to configure token
				m.settings = NewSettingsModel(m.config, m.configPath)
				m.settings.SetSize(m.width, m.height)
				m.screen = ScreenSettings
				return m, m.settings.Init()
			}
			// Lazy init gitlab client
			if m.gitlabClient == nil {
				client, err := gitlabclient.NewClient(m.config.GitLab.Token, m.config.GitLab.GitLabURL())
				if err == nil {
					m.gitlabClient = client
				}
			}
			m.gitlabScreen = NewGitLabModel(m.config, m.configPath, m.gitlabClient)
			m.gitlabScreen.SetSize(m.width, m.height)
			m.screen = ScreenGitLab
			return m, m.gitlabScreen.Init()

		case GridItemAdd:
			m.addProject = NewAddProjectModel(m.config, m.configPath)
			m.addProject.SetSize(m.width, m.height)
			m.screen = ScreenAddProject
			return m, m.addProject.Init()

		case GridItemSettings:
			m.settings = NewSettingsModel(m.config, m.configPath)
			m.settings.SetSize(m.width, m.height)
			m.screen = ScreenSettings
			return m, m.settings.Init()

		case GridItemJira:
			if m.jiraClient == nil {
				client, err := jiraclient.NewClient(m.config.Jira.URL, m.config.Jira.Email, m.config.Jira.Token)
				if err != nil {
					return m, nil
				}
				m.jiraClient = client
			}
			svc := jira.NewJiraService(m.jiraClient)
			m.jiraSpaces = NewJiraSpacesModel(svc)
			m.jiraSpaces.SetSize(m.width, m.height)
			m.screen = ScreenJiraSpaces
			return m, m.jiraSpaces.Init()
		}
	}

	return m, cmd
}

func (m AppModel) updateJiraSpaces(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraSpaces, cmd = m.jiraSpaces.Update(msg)

	if m.jiraSpaces.GoBack() {
		m.screen = ScreenGrid
		return m, nil
	}
	if sel := m.jiraSpaces.Selected(); sel != nil {
		m.jiraSpaces.ClearSelection()
		svc := jira.NewJiraService(m.jiraClient)
		m.jiraIssues = NewJiraIssuesModel(svc, *sel)
		m.jiraIssues.SetSize(m.width, m.height)
		m.screen = ScreenJiraIssues
		return m, m.jiraIssues.Init()
	}
	return m, cmd
}

func (m AppModel) updateJiraIssues(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.jiraIssues, cmd = m.jiraIssues.Update(msg)

	if m.jiraIssues.GoBack() {
		m.screen = ScreenJiraSpaces
		return m, nil
	}
	return m, cmd
}

func (m AppModel) updateProjects(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.projects, cmd = m.projects.Update(msg)

	if m.projects.GoBack() {
		m.screen = ScreenGrid
		return m, nil
	}

	if proj := m.projects.Selected(); proj != nil {
		command, err := project.BuildProjectCommand(m.config, *proj)
		if err != nil {
			return m, nil
		}
		m.result = &Result{Command: command}
		m.quitting = true
		return m, tea.Quit
	}

	return m, cmd
}

func (m AppModel) updateDocker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dockerScreen, cmd = m.dockerScreen.Update(msg)

	if m.dockerScreen.GoBack() {
		m.screen = ScreenGrid
		return m, nil
	}

	if stack := m.dockerScreen.SelectedStack(); stack != nil {
		m.selectedStack = stack
		m.dockerActions = NewDockerActionsModel(m.config, m.docker, *stack)
		m.dockerActions.SetSize(m.width, m.height)
		m.dockerScreen.ClearSelection()
		m.screen = ScreenDockerActions
		return m, m.dockerActions.Init()
	}

	if m.dockerScreen.WantsAddStack() {
		m.dockerScreen.ClearAddStack()
		m.addStack = NewAddStackModel(m.config, m.configPath)
		m.addStack.SetSize(m.width, m.height)
		m.screen = ScreenAddStack
		return m, m.addStack.Init()
	}

	return m, cmd
}

func (m AppModel) updateDockerActions(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dockerActions, cmd = m.dockerActions.Update(msg)

	if m.dockerActions.GoBack() {
		m.dockerScreen = NewDockerModel(m.config, m.docker, m.configPath)
		m.dockerScreen.SetSize(m.width, m.height)
		m.screen = ScreenDocker
		return m, m.dockerScreen.Init()
	}

	return m, cmd
}

func (m AppModel) updateAddProject(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.addProject, cmd = m.addProject.Update(msg)

	if m.addProject.GoBack() {
		// Reload config in case it was modified
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
			m.refreshGrid()
		}
		m.screen = ScreenGrid
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateAddStack(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.addStack, cmd = m.addStack.Update(msg)

	if m.addStack.GoBack() {
		// Reload config
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
		}
		m.dockerScreen = NewDockerModel(m.config, m.docker, m.configPath)
		m.dockerScreen.SetSize(m.width, m.height)
		m.screen = ScreenDocker
		return m, m.dockerScreen.Init()
	}

	return m, cmd
}

func (m AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.settings, cmd = m.settings.Update(msg)

	if m.settings.GoBack() {
		// Reload config in case it was modified
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
			m.refreshGrid()
		}
		m.screen = ScreenGrid
		return m, nil
	}

	return m, cmd
}

func (m AppModel) updateGitLab(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.gitlabScreen, cmd = m.gitlabScreen.Update(msg)

	if m.gitlabScreen.GoBack() {
		// Reload config in case gitlab_path was modified
		if updated, err := config.Load(m.configPath); err == nil {
			m.config = updated
			m.refreshGrid()
		}
		m.screen = ScreenGrid
		return m, nil
	}

	if sel := m.gitlabScreen.Selected(); sel != nil {
		m.gitlabScreen.ClearSelection()
		if m.gitlabClient != nil {
			m.gitlabActions = NewGitLabActionsModel(m.gitlabClient, sel.project.GitLabPath, sel.project.Alias, m.config.GitLab.GitLabURL())
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
		m.gitlabScreen = NewGitLabModel(m.config, m.configPath, m.gitlabClient)
		m.gitlabScreen.SetSize(m.width, m.height)
		m.screen = ScreenGitLab
		return m, m.gitlabScreen.Init()
	}

	return m, cmd
}

func (m AppModel) initGitLabFetches() tea.Cmd {
	token := m.config.GitLab.Token
	url := m.config.GitLab.GitLabURL()
	cfg := m.config

	return func() tea.Msg {
		client, err := gitlabclient.NewClient(token, url)
		if err != nil {
			return gitlabInitMsg{err: err}
		}
		return gitlabInitMsg{client: client, cfg: cfg}
	}
}

func fetchRecentPushes(client *gitlabclient.Client, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		pushes, err := client.ListRecentPushes(3)
		if err != nil {
			return recentPushesMsg{err: err}
		}

		pathToAlias := cfg.PathToAliasMap()

		// Filter to last 6 hours only
		sixHoursAgo := time.Now().Add(-6 * time.Hour)

		var result []RecentPushAlias
		for _, push := range pushes {
			if push.CreatedAt.Before(sixHoursAgo) {
				continue
			}

			projectPath, err := client.ResolveProjectPath(push.ProjectID)
			if err != nil {
				continue
			}
			alias := projectPath
			if a, ok := pathToAlias[projectPath]; ok {
				alias = a
			}

			// Fetch approval status for this project
			approvalIcon := ""
			approval, _ := client.GetProjectLatestMRApproval(projectPath)
			if approval != nil {
				if approval.Approved {
					approvalIcon = "✓"
				} else {
					approvalIcon = fmt.Sprintf("⏳ %d/%d", approval.ApprovalsGiven, approval.ApprovalsRequired)
					if approval.RuleName != "" {
						approvalIcon += " " + approval.RuleName
					}
				}
			}

			result = append(result, RecentPushAlias{
				Alias:    alias,
				Ago:      shared.TimeAgo(push.CreatedAt),
				Approval: approvalIcon,
			})
		}

		if len(result) > 6 {
			result = result[:6]
		}

		return recentPushesMsg{pushes: result}
	}
}

func fetchPendingMRs(client *gitlabclient.Client, cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		mrs, err := client.ListMyOpenMRs()
		if err != nil {
			return pendingMRsMsg{err: err}
		}

		pathToAlias := cfg.PathToAliasMap()

		var result []PendingMRAlias
		for _, mr := range mrs {
			projectPath, err := client.ResolveProjectPath(mr.ProjectID)
			if err != nil {
				continue
			}

			alias := projectPath
			if a, found := pathToAlias[projectPath]; found {
				alias = a
			}

			approval := client.GetMRApproval(projectPath, mr.IID, mr.Title)
			if approval != nil && approval.Approved {
				continue // já aprovado, não é pendente
			}

			approvalIcon := ""
			if approval != nil && approval.ApprovalsRequired > 0 {
				approvalIcon = fmt.Sprintf("⏳ %d/%d", approval.ApprovalsGiven, approval.ApprovalsRequired)
			}

			result = append(result, PendingMRAlias{
				Alias:    alias,
				IID:      mr.IID,
				Title:    mr.Title,
				Ago:      shared.TimeAgo(mr.CreatedAt),
				Approval: approvalIcon,
			})
		}

		if len(result) > 5 {
			result = result[:5]
		}

		return pendingMRsMsg{mrs: result}
	}
}

func scheduleRefresh() tea.Cmd {
	return tea.Tick(2*time.Minute, func(time.Time) tea.Msg {
		return gitlabRefreshTickMsg{}
	})
}
