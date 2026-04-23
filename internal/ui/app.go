package ui

import (
	"fmt"
	"time"

	jiraapp "github.com/DevViking-Persike/njord-cli/internal/app/jira"
	projectapp "github.com/DevViking-Persike/njord-cli/internal/app/project"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	dockerpkg "github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/githubclient"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	cloneui "github.com/DevViking-Persike/njord-cli/internal/ui/clone"
	dockerui "github.com/DevViking-Persike/njord-cli/internal/ui/docker"
	githubui "github.com/DevViking-Persike/njord-cli/internal/ui/github"
	gitlabui "github.com/DevViking-Persike/njord-cli/internal/ui/gitlab"
	"github.com/DevViking-Persike/njord-cli/internal/ui/grid"
	jiraui "github.com/DevViking-Persike/njord-cli/internal/ui/jira"
	projectui "github.com/DevViking-Persike/njord-cli/internal/ui/project"
	"github.com/DevViking-Persike/njord-cli/internal/ui/settings"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	"github.com/DevViking-Persike/njord-cli/internal/ui/stack"
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
	pushes []grid.RecentPushAlias
	err    error
}

// Async message for pending MRs
type pendingMRsMsg struct {
	mrs []grid.PendingMRAlias
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
	ScreenGitLabHub
	ScreenGitLab
	ScreenGitLabActions
	ScreenGitHubList
	ScreenGitHubActions
	ScreenLocalList
	ScreenCloneGroups
	ScreenClone
	ScreenJiraSpaces
	ScreenJiraActions
	ScreenJiraIssues
	ScreenJiraCreate
	ScreenJiraEdit
)

// Result is the final output from the TUI.
type Result struct {
	// Command is the shell command to eval (e.g. cd "/path" && code .)
	Command string
}

type AppModel struct {
	screen Screen
	config *config.Config
	docker *dockerpkg.Client
	result *Result
	width  int
	height int

	// Sub-models
	grid          grid.Model
	projects      projectui.Model
	dockerScreen  dockerui.Model
	dockerActions dockerui.ActionsModel
	addProject    projectui.AddModel
	addStack      stack.AddModel
	settings      settings.Model
	gitlabHub     gitlabui.HubModel
	gitlabScreen  gitlabui.Model
	gitlabActions gitlabui.ActionsModel
	githubList    githubui.Model
	githubActions githubui.ActionsModel
	localList     gitlabui.LocalModel
	cloneGroups   cloneui.GroupsModel
	cloneScreen   cloneui.Model
	jiraSpaces    jiraui.SpacesModel
	jiraActions   jiraui.SpaceActionsModel
	jiraIssues    jiraui.IssuesModel
	jiraCreate    jiraui.CreateModel
	jiraEdit      jiraui.EditModel

	// GitLab client (lazy init)
	gitlabClient *gitlabclient.Client

	// GitHub client (lazy init)
	githubClient *githubclient.Client

	// Jira client (lazy init)
	jiraClient *jiraclient.Client

	// Context for transitions
	selectedCatID string
	selectedStack *config.DockerStack
	configPath    string
	quitting      bool
}

func NewApp(cfg *config.Config, dockerClient *dockerpkg.Client, configPath string) AppModel {
	grid := grid.NewModel(cfg)
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
		m.gitlabHub.SetSize(msg.Width, msg.Height)
		m.gitlabScreen.SetSize(msg.Width, msg.Height)
		m.gitlabActions.SetSize(msg.Width, msg.Height)
		m.githubList.SetSize(msg.Width, msg.Height)
		m.githubActions.SetSize(msg.Width, msg.Height)
		m.localList.SetSize(msg.Width, msg.Height)
		m.cloneGroups.SetSize(msg.Width, msg.Height)
		m.cloneScreen.SetSize(msg.Width, msg.Height)
		m.jiraSpaces.SetSize(msg.Width, msg.Height)
		m.jiraActions.SetSize(msg.Width, msg.Height)
		m.jiraIssues.SetSize(msg.Width, msg.Height)
		m.jiraCreate.SetSize(msg.Width, msg.Height)
		m.jiraEdit.SetSize(msg.Width, msg.Height)
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
			m.grid.SetPushError(msg.err.Error())
		}
		return m, nil

	case pendingMRsMsg:
		m.grid.SetPendingMRs(msg.mrs)
		if msg.err != nil {
			m.grid.SetMRsError(msg.err.Error())
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
	case ScreenGitLabHub:
		return m.updateGitLabHub(msg)
	case ScreenGitLab:
		return m.updateGitLab(msg)
	case ScreenGitLabActions:
		return m.updateGitLabActions(msg)
	case ScreenGitHubList:
		return m.updateGitHubList(msg)
	case ScreenGitHubActions:
		return m.updateGitHubActions(msg)
	case ScreenLocalList:
		return m.updateLocalList(msg)
	case ScreenCloneGroups:
		return m.updateCloneGroups(msg)
	case ScreenClone:
		return m.updateClone(msg)
	case ScreenJiraSpaces:
		return m.updateJiraSpaces(msg)
	case ScreenJiraActions:
		return m.updateJiraActions(msg)
	case ScreenJiraIssues:
		return m.updateJiraIssues(msg)
	case ScreenJiraCreate:
		return m.updateJiraCreate(msg)
	case ScreenJiraEdit:
		return m.updateJiraEdit(msg)
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
	case ScreenGitLabHub:
		content = m.gitlabHub.View()
	case ScreenGitLab:
		content = m.gitlabScreen.View()
	case ScreenGitLabActions:
		content = m.gitlabActions.View()
	case ScreenGitHubList:
		content = m.githubList.View()
	case ScreenGitHubActions:
		content = m.githubActions.View()
	case ScreenLocalList:
		content = m.localList.View()
	case ScreenCloneGroups:
		content = m.cloneGroups.View()
	case ScreenClone:
		content = m.cloneScreen.View()
	case ScreenJiraSpaces:
		content = m.jiraSpaces.View()
	case ScreenJiraActions:
		content = m.jiraActions.View()
	case ScreenJiraIssues:
		content = m.jiraIssues.View()
	case ScreenJiraCreate:
		content = m.jiraCreate.View()
	case ScreenJiraEdit:
		content = m.jiraEdit.View()
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
	pushes := m.grid.RecentPushes()
	mrs := m.grid.PendingMRs()
	pushErr := m.grid.PushError()
	mrsErr := m.grid.MRsError()

	m.grid = grid.NewModel(m.config)
	m.grid.SetSize(m.width, m.height)
	m.grid.SetRecentPushes(pushes)
	m.grid.SetPendingMRs(mrs)
	m.grid.SetPushError(pushErr)
	m.grid.SetMRsError(mrsErr)
}

// --- Screen transition handlers ---

func (m AppModel) updateGrid(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.grid, cmd = m.grid.Update(msg)

	// Check for grid selection
	if sel := m.grid.Selected(); sel != nil {
		m.grid.ClearSelection()

		switch sel.Type {
		case grid.ItemCategory:
			if sel.CatID == "*" {
				// "Todos" category - show all projects
				m.projects = projectui.NewModel(m.config, sel.CatID, m.config.AllProjects())
				m.projects.SetSize(m.width, m.height)
			} else {
				for _, cat := range m.config.Categories {
					if cat.ID == sel.CatID {
						m.projects = projectui.NewModel(m.config, sel.CatID, cat.Projects)
						m.projects.SetSize(m.width, m.height)
						break
					}
				}
			}
			m.selectedCatID = sel.CatID
			m.screen = ScreenProjects
			return m, m.projects.Init()

		case grid.ItemDocker:
			m.dockerScreen = dockerui.NewModel(m.config, m.docker, m.configPath)
			m.dockerScreen.SetSize(m.width, m.height)
			m.screen = ScreenDocker
			return m, m.dockerScreen.Init()

		case grid.ItemGitLab:
			m.gitlabHub = gitlabui.NewHubModel(m.config)
			m.gitlabHub.SetSize(m.width, m.height)
			m.screen = ScreenGitLabHub
			return m, m.gitlabHub.Init()

		case grid.ItemAdd:
			m.addProject = projectui.NewAddModel(m.config, m.configPath)
			m.addProject.SetSize(m.width, m.height)
			m.screen = ScreenAddProject
			return m, m.addProject.Init()

		case grid.ItemSettings:
			m.settings = settings.NewModel(m.config, m.configPath)
			m.settings.SetSize(m.width, m.height)
			m.screen = ScreenSettings
			return m, m.settings.Init()

		case grid.ItemJira:
			if m.jiraClient == nil {
				client, err := jiraclient.NewClient(m.config.Jira.URL, m.config.Jira.Email, m.config.Jira.Token)
				if err != nil {
					return m, nil
				}
				m.jiraClient = client
			}
			svc := jiraapp.NewJiraService(m.jiraClient)
			m.jiraSpaces = jiraui.NewSpacesModel(svc)
			m.jiraSpaces.SetSize(m.width, m.height)
			m.screen = ScreenJiraSpaces
			return m, m.jiraSpaces.Init()
		}
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

func (m AppModel) updateDocker(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dockerScreen, cmd = m.dockerScreen.Update(msg)

	if m.dockerScreen.GoBack() {
		m.screen = ScreenGrid
		return m, nil
	}

	if stack := m.dockerScreen.SelectedStack(); stack != nil {
		m.selectedStack = stack
		m.dockerActions = dockerui.NewActionsModel(m.config, m.docker, *stack)
		m.dockerActions.SetSize(m.width, m.height)
		m.dockerScreen.ClearSelection()
		m.screen = ScreenDockerActions
		return m, m.dockerActions.Init()
	}

	if m.dockerScreen.WantsAddStack() {
		m.dockerScreen.ClearAddStack()
		m.addStack = stack.NewAddModel(m.config, m.configPath)
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
		m.dockerScreen = dockerui.NewModel(m.config, m.docker, m.configPath)
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
		m.dockerScreen = dockerui.NewModel(m.config, m.docker, m.configPath)
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

		var result []grid.RecentPushAlias
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

			result = append(result, grid.RecentPushAlias{
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

		var result []grid.PendingMRAlias
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

			result = append(result, grid.PendingMRAlias{
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
