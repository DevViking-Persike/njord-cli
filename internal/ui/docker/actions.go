package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dockerActionDoneMsg struct {
	action string
	err    error
}

type dockerLogsMsg struct {
	logs string
	err  error
}

type DockerAction int

const (
	ActionStart DockerAction = iota
	ActionStop
	ActionRestart
	ActionLogs
)

type ActionsModel struct {
	cfg         *config.Config
	docker      *docker.Client
	stack       config.DockerStack
	available   bool
	cursor      int
	goBack      bool
	containers  []docker.ContainerInfo
	message     string
	messageType string // "ok", "error", "info"
	logs        string
	showLogs    bool
	running     bool
	width       int
	height      int
}

var actionLabels = []struct {
	icon   string
	label  string
	desc   string
	action DockerAction
}{
	{"▶", "Iniciar", "docker compose up -d", ActionStart},
	{"■", "Parar", "docker compose down", ActionStop},
	{"↻", "Reiniciar", "docker compose restart", ActionRestart},
	{"☰", "Logs", "docker compose logs --tail 50", ActionLogs},
}

func NewActionsModel(cfg *config.Config, dockerClient *docker.Client, stack config.DockerStack) ActionsModel {
	var containers []docker.ContainerInfo
	if dockerClient != nil {
		projectName := filepath.Base(stack.Path)
		containers = dockerClient.ListContainers(projectName)
	}

	return ActionsModel{
		cfg:        cfg,
		docker:     dockerClient,
		stack:      stack,
		available:  dockerClient != nil,
		containers: containers,
	}
}

func (m ActionsModel) Init() tea.Cmd {
	return nil
}

func (m ActionsModel) Update(msg tea.Msg) (ActionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dockerActionDoneMsg:
		m.running = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro: %s", msg.err)
			m.messageType = "error"
		} else {
			m.message = fmt.Sprintf("%s executado!", msg.action)
			m.messageType = "ok"
		}
		// Refresh containers
		if m.docker != nil {
			projectName := filepath.Base(m.stack.Path)
			m.containers = m.docker.ListContainers(projectName)
		}
		return m, nil

	case dockerLogsMsg:
		m.running = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro ao obter logs: %s", msg.err)
			m.messageType = "error"
		} else {
			m.logs = msg.logs
			m.showLogs = true
		}
		return m, nil

	case tea.KeyMsg:
		if m.running {
			return m, nil
		}

		if m.showLogs {
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "enter" {
				m.showLogs = false
				m.logs = ""
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(actionLabels)-1 {
				m.cursor++
			}
		case "esc", "q":
			m.goBack = true
			return m, nil
		case "enter":
			if !m.available {
				m.message = "Docker indisponível"
				m.messageType = "error"
				return m, nil
			}
			return m, m.executeAction(actionLabels[m.cursor].action)
		}
	}
	return m, nil
}

func (m ActionsModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().Bold(true).Foreground(theme.DockerBlue).Render("  Docker Stacks")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	// Stack info
	nameStr := lipgloss.NewStyle().Bold(true).Foreground(theme.DockerBlue).Render(m.stack.Name)
	descStr := theme.DimStyle.Render("(" + m.stack.Desc + ")")
	b.WriteString("  " + nameStr + " " + descStr + "\n")

	composePath := m.cfg.ResolveDockerComposePath(m.stack)
	b.WriteString("  " + theme.DimStyle.Render(composePath) + "\n\n")

	if !m.available {
		b.WriteString("  " + theme.WarningStyle.Render("Docker indisponível. Verifique o daemon e as permissões do socket.") + "\n\n")
		if m.message != "" {
			b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.message) + "\n\n")
		}
		b.WriteString(theme.HelpStyle.Render("  esc back"))
		return b.String()
	}

	// Container details
	if len(m.containers) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum container encontrado (stack parada)") + "\n")
	} else {
		for _, ct := range m.containers {
			var icon string
			var stateStyle lipgloss.Style
			if ct.State == "running" {
				icon = "●"
				stateStyle = theme.StatusRunning
			} else {
				icon = "○"
				stateStyle = theme.StatusStopped
			}
			line := fmt.Sprintf("  %s  %-30s %s  %s",
				stateStyle.Render(icon),
				theme.TextStyle.Render(ct.Name),
				stateStyle.Render(translateDockerState(ct.State)),
				theme.DimStyle.Render(ct.Ports))
			b.WriteString(line + "\n")
		}
	}

	b.WriteString("\n")

	// Show logs if active
	if m.showLogs {
		b.WriteString(theme.DimStyle.Render("  ─── Logs ───") + "\n")
		lines := strings.Split(m.logs, "\n")
		max := 30
		if len(lines) > max {
			lines = lines[len(lines)-max:]
		}
		for _, line := range lines {
			if len(line) > m.width-4 && m.width > 10 {
				line = line[:m.width-7] + "..."
			}
			b.WriteString("  " + theme.DimStyle.Render(line) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  esc/enter fechar logs"))
		return b.String()
	}

	// Actions menu
	for i, action := range actionLabels {
		selected := i == m.cursor
		var line string
		if selected {
			line = theme.TitleSelectedStyle.Render(fmt.Sprintf("  ▶ %s  %s", action.icon, action.label))
			line += "  " + theme.DimStyle.Render(action.desc)
		} else {
			line = theme.TextStyle.Render(fmt.Sprintf("    %s  %s", action.icon, action.label))
			line += "  " + theme.DimStyle.Render(action.desc)
		}
		b.WriteString(line + "\n")
	}

	// Message
	if m.message != "" {
		b.WriteString("\n")
		switch m.messageType {
		case "ok":
			b.WriteString("  " + theme.SuccessStyle.Render("✓ "+m.message))
		case "error":
			b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.message))
		default:
			b.WriteString("  " + theme.TextStyle.Render(m.message))
		}
		b.WriteString("\n")
	}

	if m.running {
		b.WriteString("\n" + theme.DimStyle.Render("  Executando..."))
	}

	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter execute  esc back"))

	return b.String()
}

func translateDockerState(state string) string {
	switch state {
	case "running":
		return "rodando"
	case "created":
		return "criado"
	case "exited":
		return "parado"
	case "paused":
		return "pausado"
	case "restarting":
		return "reiniciando"
	case "dead":
		return "morto"
	default:
		return state
	}
}

func (m *ActionsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *ActionsModel) GoBack() bool { return m.goBack }

func (m *ActionsModel) executeAction(action DockerAction) tea.Cmd {
	if m.docker == nil {
		m.running = false
		m.message = "Docker indisponível"
		m.messageType = "error"
		return nil
	}

	m.running = true
	m.message = ""
	m.showLogs = false

	composePath := m.cfg.ResolveDockerComposePath(m.stack)
	projectName := filepath.Base(m.stack.Path)

	switch action {
	case ActionStart:
		return func() tea.Msg {
			err := m.docker.StartProject(composePath, projectName)
			return dockerActionDoneMsg{action: "Iniciar", err: err}
		}
	case ActionStop:
		return func() tea.Msg {
			err := m.docker.StopProject(composePath, projectName)
			return dockerActionDoneMsg{action: "Parar", err: err}
		}
	case ActionRestart:
		return func() tea.Msg {
			err := m.docker.RestartProject(composePath, projectName)
			return dockerActionDoneMsg{action: "Reiniciar", err: err}
		}
	case ActionLogs:
		return func() tea.Msg {
			logs, err := m.docker.GetLogs(composePath, projectName, 50)
			return dockerLogsMsg{logs: logs, err: err}
		}
	}
	return nil
}
