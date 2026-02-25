package ui

import (
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Screen int

const (
	ScreenGrid Screen = iota
	ScreenProjects
	ScreenDocker
	ScreenDockerActions
	ScreenAddProject
	ScreenAddStack
	ScreenSettings
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
	return m.grid.Init()
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
	}

	help := theme.HelpStyle.Render("  ↑↓←→ navigate  enter select  esc back  q quit")
	return lipgloss.JoinVertical(lipgloss.Left, content, "", help)
}

// GetResult returns the result after the TUI exits.
func (m AppModel) GetResult() *Result {
	return m.result
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
		path := m.config.ResolveProjectPath(*proj)

		if path == "@rdp" {
			m.result = &Result{Command: "gnome-terminal -- bash -c 'echo \"Connecting to RDP...\"; sleep 1' &"}
		} else {
			editor := m.config.Settings.Editor
			m.result = &Result{Command: "cd " + shellQuote(path) + " && " + editor + " ."}
		}
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
			m.grid = NewGridModel(m.config)
			m.grid.SetSize(m.width, m.height)
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
			m.grid = NewGridModel(m.config)
			m.grid.SetSize(m.width, m.height)
		}
		m.screen = ScreenGrid
		return m, nil
	}

	return m, cmd
}

func shellQuote(s string) string {
	return "\"" + s + "\""
}
