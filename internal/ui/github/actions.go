package github

import (
	"fmt"
	"strings"

	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
)

// ActionsModel mostra as ações disponíveis pra um projeto GitHub selecionado:
// abrir URL no browser, clonar (se não existir) ou abrir no editor.
type ActionsModel struct {
	cfg        *config.Config
	configPath string
	ref        githubapp.ProjectRef
	actions    []actionItem
	cursor     int
	width      int
	height     int
	command    string // comando shell a ser eval pelo wrapper njord()
	goBack     bool
	message    string
	input      string
	inputting  bool // true quando o usuário está digitando github_path em falta
}

type actionItem struct {
	label    string
	subLabel string
	disabled bool
	reason   string
}

func NewActionsModel(cfg *config.Config, configPath string, ref githubapp.ProjectRef) ActionsModel {
	m := ActionsModel{cfg: cfg, configPath: configPath, ref: ref}
	m.rebuildActions()
	return m
}

func (m *ActionsModel) rebuildActions() {
	hasPath := m.ref.Project.GitHubPath != ""
	localExists := githubapp.LocalExists(m.cfg, m.ref.Project)

	m.actions = []actionItem{
		{
			label:    "Abrir no browser",
			subLabel: m.describeBrowser(),
			disabled: !hasPath,
			reason:   "preencha github_path",
		},
		{
			label:    "Clonar",
			subLabel: m.describeClone(localExists),
			disabled: !hasPath || localExists,
			reason:   m.cloneReason(hasPath, localExists),
		},
		{
			label:    "Abrir no editor",
			subLabel: "cd + editor (só se a pasta existe)",
			disabled: !localExists,
			reason:   "ainda não foi clonado",
		},
	}
	if !hasPath {
		m.actions = append(m.actions, actionItem{
			label:    "Preencher github_path",
			subLabel: "salva na config (~/.config/njord/njord.yaml)",
		})
	}
}

func (m ActionsModel) describeBrowser() string {
	if url, err := githubapp.BrowserURL(m.ref.Project); err == nil {
		return url
	}
	return "—"
}

func (m ActionsModel) describeClone(localExists bool) string {
	if localExists {
		return "já existe no disco"
	}
	return m.cfg.ResolveProjectPath(m.ref.Project)
}

func (m ActionsModel) cloneReason(hasPath, localExists bool) string {
	if !hasPath {
		return "preencha github_path"
	}
	if localExists {
		return "pasta já existe"
	}
	return ""
}

func (m ActionsModel) Init() tea.Cmd { return nil }

func (m ActionsModel) Update(msg tea.Msg) (ActionsModel, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if m.inputting {
		return m.handleInput(key)
	}
	return m.handleActions(key)
}

func (m ActionsModel) handleActions(msg tea.KeyMsg) (ActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.actions)-1 {
			m.cursor++
		}
	case "esc", "q":
		m.goBack = true
	case "enter":
		return m.triggerCurrent()
	}
	return m, nil
}

func (m ActionsModel) triggerCurrent() (ActionsModel, tea.Cmd) {
	if m.cursor >= len(m.actions) {
		return m, nil
	}
	act := m.actions[m.cursor]
	if act.disabled {
		m.message = "Indisponível: " + act.reason
		return m, nil
	}
	switch act.label {
	case "Abrir no browser":
		url, err := githubapp.BrowserURL(m.ref.Project)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.command = githubapp.BuildOpenBrowserCommand(url)
	case "Clonar":
		dest := m.cfg.ResolveProjectPath(m.ref.Project)
		cmd, err := githubapp.BuildCloneCommand(m.ref.Project, dest, m.cfg.Settings.Editor)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.command = cmd
	case "Abrir no editor":
		dest := m.cfg.ResolveProjectPath(m.ref.Project)
		editor := strings.TrimSpace(m.cfg.Settings.Editor)
		if editor == "" {
			editor = "code"
		}
		m.command = "cd -- '" + strings.ReplaceAll(dest, "'", `'\''`) + "' && " + editor + " ."
	case "Preencher github_path":
		m.inputting = true
		m.input = ""
		m.message = ""
	}
	return m, nil
}

func (m ActionsModel) handleInput(msg tea.KeyMsg) (ActionsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputting = false
		m.input = ""
		m.message = ""
	case "enter":
		val := strings.TrimSpace(m.input)
		if val == "" {
			m.message = "Path não pode ser vazio"
			return m, nil
		}
		if err := m.cfg.SetProjectGitHubPath(m.ref.CatIdx, m.ref.ProjIdx, val); err != nil {
			m.message = err.Error()
			return m, nil
		}
		if err := config.Save(m.cfg, m.configPath); err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.ref.Project.GitHubPath = val
		m.rebuildActions()
		m.inputting = false
		m.input = ""
		m.message = "github_path salvo"
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case "ctrl+u":
		m.input = ""
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m ActionsModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")
	header := fmt.Sprintf("   GitHub — %s", m.ref.Project.Alias)
	b.WriteString(theme.TitleSelectedStyle.Render(header) + "\n")
	b.WriteString(theme.DimStyle.Render("  "+strings.Repeat("─", 50)) + "\n\n")

	if m.inputting {
		b.WriteString("  " + theme.TextStyle.Render("github_path (ex: user/repo):") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.input + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		return b.String()
	}

	for i, act := range m.actions {
		label := act.label
		sub := theme.DimStyle.Render(" — " + act.subLabel)
		if act.disabled {
			label = theme.DimStyle.Render(label)
		}
		if i == m.cursor {
			b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+label) + sub + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+label) + sub + "\n")
		}
	}
	if m.message != "" {
		b.WriteString("\n  " + theme.WarningStyle.Render(m.message) + "\n")
	}
	return b.String()
}

func (m *ActionsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Command devolve o comando shell pendente que o wrapper njord() deve eval.
// Após ler, o usuário do ActionsModel deve chamar ClearCommand.
func (m ActionsModel) Command() string { return m.command }

// ClearCommand zera o comando pendente (ex.: após consumi-lo).
func (m *ActionsModel) ClearCommand() { m.command = "" }

func (m *ActionsModel) GoBack() bool { return m.goBack }
