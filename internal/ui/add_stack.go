package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addStackStep int

const (
	addStackSelectPath addStackStep = iota
	addStackName
	addStackDesc
	addStackConfirm
	addStackDone
)

type AddStackModel struct {
	cfg        *config.Config
	configPath string
	step       addStackStep
	goBack     bool

	discovered []string
	stackPath  string
	stackName  string
	stackDesc  string

	inputBuf    string
	cursor      int
	message     string
	messageType string
	width       int
	height      int
}

func NewAddStackModel(cfg *config.Config, configPath string) AddStackModel {
	// Discover compose files not yet registered
	baseDir := config.ExpandPath(cfg.Settings.ProjectsBase)
	discovered := discoverComposeFiles(baseDir, cfg.DockerStacks)

	return AddStackModel{
		cfg:        cfg,
		configPath: configPath,
		step:       addStackSelectPath,
		discovered: discovered,
	}
}

func (m AddStackModel) Init() tea.Cmd {
	return nil
}

func (m AddStackModel) Update(msg tea.Msg) (AddStackModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case addStackSelectPath:
			return m.handlePathSelect(msg)
		case addStackName:
			return m.handleNameInput(msg)
		case addStackDesc:
			return m.handleDescInput(msg)
		case addStackConfirm:
			return m.handleStackConfirm(msg)
		case addStackDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.goBack = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m AddStackModel) View() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(theme.DockerBlue).Render("  + Adicionar Stack")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	switch m.step {
	case addStackSelectPath:
		if len(m.discovered) == 0 {
			b.WriteString("  " + theme.ErrorStyle.Render("Nenhum docker-compose.yml novo encontrado em ~/Avita/") + "\n")
			b.WriteString("  " + theme.DimStyle.Render("Todos já estão registrados.") + "\n")
			b.WriteString("\n" + theme.HelpStyle.Render("  esc voltar"))
		} else {
			b.WriteString("  " + theme.TextStyle.Render("Selecionar projeto com docker-compose:") + "\n\n")
			for i, path := range m.discovered {
				if i == m.cursor {
					b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+path) + "\n")
				} else {
					b.WriteString("  " + theme.TextStyle.Render("  "+path) + "\n")
				}
			}
			b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))
		}

	case addStackName:
		b.WriteString("  " + theme.TextStyle.Render("Nome da stack (exibição no menu):") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case addStackDesc:
		b.WriteString("  " + theme.TextStyle.Render("Descrição curta:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case addStackConfirm:
		b.WriteString("  " + theme.TextStyle.Render("Confirmar:") + "\n\n")
		b.WriteString("  " + theme.DimStyle.Render("Path:  ") + theme.TextStyle.Render(m.stackPath) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Nome:  ") + theme.TextStyle.Render(m.stackName) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Desc:  ") + theme.TextStyle.Render(m.stackDesc) + "\n\n")
		opts := []string{"Confirmar", "Cancelar"}
		for i, opt := range opts {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}

	case addStackDone:
		if m.messageType == "ok" {
			b.WriteString("  " + theme.SuccessStyle.Render("✓ Stack adicionada!") + "\n")
		} else {
			b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	}

	return b.String()
}

func (m *AddStackModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *AddStackModel) GoBack() bool { return m.goBack }

func (m AddStackModel) handlePathSelect(msg tea.KeyMsg) (AddStackModel, tea.Cmd) {
	if len(m.discovered) == 0 {
		if msg.String() == "esc" || msg.String() == "q" {
			m.goBack = true
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.discovered)-1 {
			m.cursor++
		}
	case "esc", "q":
		m.goBack = true
	case "enter":
		m.stackPath = m.discovered[m.cursor]
		m.inputBuf = ""
		m.step = addStackName
	}
	return m, nil
}

func (m AddStackModel) handleNameInput(msg tea.KeyMsg) (AddStackModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = addStackSelectPath
		m.cursor = 0
	case "enter":
		name := strings.TrimSpace(m.inputBuf)
		if name == "" {
			return m, nil
		}
		m.stackName = name
		m.inputBuf = ""
		m.step = addStackDesc
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
	case "ctrl+u":
		m.inputBuf = ""
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.inputBuf += string(msg.Runes)
		}
	}
	return m, nil
}

func (m AddStackModel) handleDescInput(msg tea.KeyMsg) (AddStackModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = addStackName
		m.inputBuf = m.stackName
	case "enter":
		desc := strings.TrimSpace(m.inputBuf)
		if desc == "" {
			return m, nil
		}
		m.stackDesc = desc
		m.inputBuf = ""
		m.cursor = 0
		m.step = addStackConfirm
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
	case "ctrl+u":
		m.inputBuf = ""
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.inputBuf += string(msg.Runes)
		}
	}
	return m, nil
}

func (m AddStackModel) handleStackConfirm(msg tea.KeyMsg) (AddStackModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "esc":
		m.step = addStackDesc
		m.inputBuf = m.stackDesc
	case "enter":
		if m.cursor == 0 {
			// Save
			m.cfg.DockerStacks = append(m.cfg.DockerStacks, config.DockerStack{
				Name: m.stackName,
				Desc: m.stackDesc,
				Path: m.stackPath,
			})
			if err := config.Save(m.cfg, m.configPath); err != nil {
				m.message = fmt.Sprintf("Erro: %s", err)
				m.messageType = "error"
			} else {
				m.messageType = "ok"
			}
			m.step = addStackDone
		} else {
			m.goBack = true
		}
	}
	return m, nil
}

func discoverComposeFiles(baseDir string, existing []config.DockerStack) []string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	registeredPaths := make(map[string]bool)
	for _, stack := range existing {
		registeredPaths[stack.Path] = true
	}

	var discovered []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(baseDir, entry.Name())
		candidates := []string{
			filepath.Join(dir, "docker-compose.yml"),
			filepath.Join(dir, "docker-compose.yaml"),
			filepath.Join(dir, "compose.yml"),
			filepath.Join(dir, "compose.yaml"),
		}
		foundCompose := false
		for _, composePath := range candidates {
			if _, err := os.Stat(composePath); err == nil {
				foundCompose = true
				break
			}
		}
		if foundCompose && !registeredPaths[entry.Name()] {
			discovered = append(discovered, entry.Name())
		}
	}
	return discovered
}
