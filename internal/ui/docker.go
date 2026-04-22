package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stackStatusMsg struct {
	statuses map[int]docker.StackStatus
}

type tickRefreshMsg struct{}

type DockerModel struct {
	cfg        *config.Config
	docker     *docker.Client
	configPath string
	stacks     []config.DockerStack
	statuses   map[int]docker.StackStatus
	available  bool
	cursor     int
	goBack     bool
	selected   *config.DockerStack
	addStack   bool
	width      int
	height     int
}

func NewDockerModel(cfg *config.Config, dockerClient *docker.Client, configPath string) DockerModel {
	return DockerModel{
		cfg:        cfg,
		docker:     dockerClient,
		configPath: configPath,
		stacks:     cfg.DockerStacks,
		statuses:   make(map[int]docker.StackStatus),
		available:  dockerClient != nil,
	}
}

func (m DockerModel) Init() tea.Cmd {
	return m.refreshStatuses()
}

func (m DockerModel) Update(msg tea.Msg) (DockerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case stackStatusMsg:
		m.statuses = msg.statuses
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return tickRefreshMsg{}
		})

	case tickRefreshMsg:
		return m, m.refreshStatuses()

	case tea.KeyMsg:
		totalItems := len(m.stacks) + 1

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < totalItems-1 {
				m.cursor++
			}
		case "esc", "q":
			m.goBack = true
			return m, nil
		case "enter":
			if m.cursor < len(m.stacks) {
				stack := m.stacks[m.cursor]
				m.selected = &stack
			} else {
				m.addStack = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m DockerModel) View() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(theme.DockerBlue).Render("  Docker Stacks")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	headerLine := fmt.Sprintf("  %-18s │ %-30s │ %s", "Stack", "Descrição", "Status")
	b.WriteString(theme.DimStyle.Render(headerLine) + "\n")
	b.WriteString(theme.DimStyle.Render("  "+strings.Repeat("─", 18)+"─┼─"+strings.Repeat("─", 30)+"─┼─"+strings.Repeat("─", 20)) + "\n")

	for i, stack := range m.stacks {
		selected := i == m.cursor
		status := m.statuses[i]

		var statusStr string
		switch status.Symbol {
		case "●":
			statusStr = theme.StatusRunning.Render("● " + status.Label)
		case "◐":
			statusStr = theme.StatusPartial.Render("◐ " + status.Label)
		case "!":
			statusStr = theme.WarningStyle.Render("! " + status.Label)
		default:
			statusStr = theme.StatusStopped.Render("○ " + status.Label)
		}

		name := stack.Name
		desc := stack.Desc
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		var line string
		if selected {
			pointer := theme.TitleSelectedStyle.Render("▶ ")
			nameStr := theme.TitleSelectedStyle.Render(fmt.Sprintf("%-16s", name))
			descStr := theme.TextStyle.Render(fmt.Sprintf("%-30s", desc))
			line = pointer + nameStr + " │ " + descStr + " │ " + statusStr
		} else {
			nameStr := theme.TextStyle.Render(fmt.Sprintf("  %-16s", name))
			descStr := theme.DimStyle.Render(fmt.Sprintf("%-30s", desc))
			line = nameStr + " │ " + descStr + " │ " + statusStr
		}
		b.WriteString(line + "\n")
	}

	addIdx := len(m.stacks)
	if m.cursor == addIdx {
		b.WriteString(theme.AddTitleSelectedStyle.Render("▶ + Adicionar stack") + "\n")
	} else {
		b.WriteString(theme.AddTitleStyle.Render("  + Adicionar stack") + "\n")
	}

	b.WriteString("\n")
	if !m.available {
		b.WriteString(theme.WarningStyle.Render("  Docker indisponível: verifique o daemon e as permissões do socket.") + "\n\n")
	}
	b.WriteString(theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))

	return b.String()
}

func (m *DockerModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *DockerModel) GoBack() bool        { return m.goBack }
func (m *DockerModel) WantsAddStack() bool { return m.addStack }
func (m *DockerModel) SelectedStack() *config.DockerStack {
	return m.selected
}
func (m *DockerModel) ClearSelection() { m.selected = nil }
func (m *DockerModel) ClearAddStack()  { m.addStack = false }

func (m DockerModel) refreshStatuses() tea.Cmd {
	stacks := m.stacks
	cfg := m.cfg
	dockerClient := m.docker
	return func() tea.Msg {
		statuses := make(map[int]docker.StackStatus)
		if dockerClient == nil {
			for i := range stacks {
				statuses[i] = docker.UnavailableStatus()
			}
			return stackStatusMsg{statuses: statuses}
		}
		for i, stack := range stacks {
			composePath := cfg.ResolveDockerComposePath(stack)
			projectName := filepath.Base(stack.Path)
			statuses[i] = dockerClient.GetStackStatus(composePath, projectName)
		}
		return stackStatusMsg{statuses: statuses}
	}
}
