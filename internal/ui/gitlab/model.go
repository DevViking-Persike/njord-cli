package gitlab

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/app/gitlab"
	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

// Spinner frames for running/pending pipelines
var spinnerFrames = []string{"◐", "◓", "◑", "◒"}

// Pipeline status icons
const (
	iconSuccess = "✓"
	iconFailed  = "✗"
	iconBlocked = "⊘"
	iconUnknown = "○"
)

type gitlabScreen int

const (
	gitlabProjectList gitlabScreen = iota
	gitlabConfigPath
	gitlabConfigPathInput
)

// Filter resolve a lista de projetos exibidos na tela a partir da config.
// O model pede essa função tanto na inicialização quanto após mutação da YAML.
type Filter func(*config.Config) []githubapp.ProjectRef

// Async message for pipeline status + approval
type gitlabProjectStatusMsg struct {
	gitlabPath string
	status     string // success, failed, running, pending, canceled, blocked, ""
	lastTime   time.Time
	approval   *gitlabclient.MRApprovalInfo
}

// Tick message for spinner animation
type gitlabSpinnerTickMsg struct{}

type Model struct {
	cfg        *config.Config
	configPath string
	client     *gitlabclient.Client
	filter     Filter
	screen     gitlabScreen
	goBack     bool
	selected   *githubapp.ProjectRef
	configRef  *githubapp.ProjectRef

	projects       []githubapp.ProjectRef
	pipelineStatus map[string]string                       // gitlab_path -> status
	approvalInfo   map[string]*gitlabclient.MRApprovalInfo // gitlab_path -> approval
	lastActivity   map[string]time.Time                    // gitlab_path -> last pipeline time
	loadedCount    int
	sorted         bool
	cursor         int
	offset         int
	spinnerFrame   int

	inputBuf string
	message  string

	width, height int
}

// NewModel constrói a tela com a lista já filtrada. Passar githubapp.FilterGitLab
// pra comportamento pré-hub; outro filtro pra variar o conjunto exibido.
func NewModel(cfg *config.Config, configPath string, client *gitlabclient.Client, filter Filter) Model {
	if filter == nil {
		filter = githubapp.FilterGitLab
	}
	m := Model{
		cfg:            cfg,
		configPath:     configPath,
		client:         client,
		filter:         filter,
		screen:         gitlabProjectList,
		pipelineStatus: make(map[string]string),
		approvalInfo:   make(map[string]*gitlabclient.MRApprovalInfo),
		lastActivity:   make(map[string]time.Time),
	}
	m.buildProjectList()
	return m
}

func (m *Model) buildProjectList() {
	m.projects = m.filter(m.cfg)
}

func (m Model) Init() tea.Cmd {
	if m.client == nil || len(m.projects) == 0 {
		return nil
	}
	// Fetch pipeline status for all projects + start spinner
	cmds := []tea.Cmd{spinnerTick()}
	for _, ref := range m.projects {
		path := ref.Project.GitLabPath
		if path == "" {
			continue
		}
		cmds = append(cmds, m.fetchProjectStatus(path))
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case gitlabProjectStatusMsg:
		m.pipelineStatus[msg.gitlabPath] = msg.status
		if !msg.lastTime.IsZero() {
			m.lastActivity[msg.gitlabPath] = msg.lastTime
		}
		if msg.approval != nil {
			m.approvalInfo[msg.gitlabPath] = msg.approval
		}
		m.loadedCount++
		if !m.sorted && m.loadedCount >= len(m.projects) {
			m.sortByRecent()
			m.sorted = true
		}
		return m, nil

	case gitlabSpinnerTickMsg:
		m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
		// Only keep ticking if we have running/pending pipelines
		if m.hasActiveSpinners() {
			return m, spinnerTick()
		}
		return m, nil

	case tea.KeyMsg:
		switch m.screen {
		case gitlabProjectList:
			return m.handleProjectList(msg)
		case gitlabConfigPath:
			return m.handleConfigPath(msg)
		case gitlabConfigPathInput:
			return m.handleConfigPathInput(msg)
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	header := theme.GitLabTitleSelectedStyle.Render("  ◆ GitLab")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	switch m.screen {
	case gitlabProjectList:
		if len(m.projects) == 0 {
			b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto cadastrado.") + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("Selecione o projeto:") + "\n\n")
			visible := m.visibleRows()
			start := m.offset
			end := start + visible
			if end > len(m.projects) {
				end = len(m.projects)
			}
			if start > 0 {
				b.WriteString("  " + theme.DimStyle.Render("  ↑ mais projetos...") + "\n")
			}
			for i := start; i < end; i++ {
				ref := m.projects[i]
				icon := m.renderStatusIcon(ref.Project.GitLabPath)
				label := ref.Project.Alias + " — " + ref.Project.Desc
				approvalTag := m.renderApprovalTag(ref.Project.GitLabPath)
				pathLabel := ref.Project.GitLabPath
				if pathLabel == "" {
					pathLabel = "sem gitlab_path"
				}
				pathTag := theme.DimStyle.Render(" [" + pathLabel + "]")
				if i == m.cursor {
					b.WriteString("  " + icon + " " + theme.GitLabTitleSelectedStyle.Render("▶ "+label) + pathTag + approvalTag + "\n")
				} else {
					b.WriteString("  " + icon + " " + theme.TextStyle.Render("  "+label) + pathTag + approvalTag + "\n")
				}
			}
			if end < len(m.projects) {
				b.WriteString("  " + theme.DimStyle.Render("  ↓ mais projetos...") + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))

	case gitlabConfigPath:
		if m.configRef == nil {
			b.WriteString("  " + theme.ErrorStyle.Render("Projeto não identificado.") + "\n")
			return b.String()
		}
		ref := m.configRef
		b.WriteString("  " + theme.TextStyle.Render("Projeto: "+ref.Project.Alias) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("GitLab path não configurado.") + "\n\n")
		opts := []string{"Auto-detectar do git remote", "Digitar manualmente"}
		for i, opt := range opts {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))

	case gitlabConfigPathInput:
		b.WriteString("  " + theme.TextStyle.Render("GitLab path (ex: grupo/repo):") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))
	}

	return b.String()
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) GoBack() bool                { return m.goBack }
// Selected returns the config.Project the user picked in the screen, or nil.
func (m *Model) Selected() *config.Project {
	if m.selected == nil {
		return nil
	}
	p := m.selected.Project
	return &p
}
func (m *Model) ClearSelection()             { m.selected = nil }

func (m Model) visibleRows() int {
	available := m.height - 10
	if available < 3 {
		return 3
	}
	return available
}

func (m *Model) ensureVisible() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// --- Status icon rendering ---

func (m Model) renderStatusIcon(gitlabPath string) string {
	if gitlabPath == "" {
		return theme.DimStyle.Render("·")
	}

	status, ok := m.pipelineStatus[gitlabPath]
	if !ok {
		// Still loading
		return theme.DimStyle.Render(spinnerFrames[m.spinnerFrame])
	}

	switch status {
	case "success":
		return theme.PipelineSuccessStyle.Render(iconSuccess)
	case "failed":
		return theme.PipelineFailedStyle.Render(iconFailed)
	case "running":
		return theme.PipelineRunningStyle.Render(spinnerFrames[m.spinnerFrame])
	case "pending", "waiting_for_resource", "created":
		return theme.PipelinePendingStyle.Render(spinnerFrames[m.spinnerFrame])
	case "canceled", "skipped", "blocked":
		return theme.DimStyle.Render(iconBlocked)
	default:
		return theme.DimStyle.Render(iconUnknown)
	}
}

func (m Model) hasActiveSpinners() bool {
	// Active if any project still loading or has running/pending pipeline
	for _, ref := range m.projects {
		if ref.Project.GitLabPath == "" {
			continue
		}
		status, ok := m.pipelineStatus[ref.Project.GitLabPath]
		if !ok {
			return true // still loading
		}
		if status == "running" || status == "pending" || status == "waiting_for_resource" || status == "created" {
			return true
		}
	}
	return false
}

// --- Async commands ---

func (m Model) fetchProjectStatus(gitlabPath string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		msg := gitlabProjectStatusMsg{gitlabPath: gitlabPath}
		status := gitlab.LoadGitLabProjectStatus(client, gitlabPath)
		msg.status = status.Status
		msg.lastTime = status.LastTime
		msg.approval = status.Approval

		return msg
	}
}

func (m *Model) sortByRecent() {
	sort.SliceStable(m.projects, func(i, j int) bool {
		ti := m.lastActivity[m.projects[i].Project.GitLabPath]
		tj := m.lastActivity[m.projects[j].Project.GitLabPath]
		return ti.After(tj)
	})
}

func (m Model) renderApprovalTag(gitlabPath string) string {
	a, ok := m.approvalInfo[gitlabPath]
	if !ok || a == nil {
		return ""
	}
	return " " + renderApprovalTag(a)
}

func spinnerTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return gitlabSpinnerTickMsg{}
	})
}

// --- Key handlers ---

func (m Model) handleProjectList(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
	case "down", "j":
		if m.cursor < len(m.projects)-1 {
			m.cursor++
			m.ensureVisible()
		}
	case "esc":
		m.goBack = true
	case "enter":
		if len(m.projects) > 0 && m.cursor < len(m.projects) {
			ref := m.projects[m.cursor]
			if ref.Project.GitLabPath == "" {
				m.configRef = &ref
				m.screen = gitlabConfigPath
				m.cursor = 0
				m.message = ""
			} else {
				m.selected = &ref
			}
		}
	}
	return m, nil
}

func (m Model) handleConfigPath(msg tea.KeyMsg) (Model, tea.Cmd) {
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
		m.screen = gitlabProjectList
		m.cursor = 0
		m.message = ""
		m.configRef = nil
	case "enter":
		if m.configRef == nil {
			m.message = "Projeto não identificado"
			return m, nil
		}
		if m.cursor == 0 {
			glPath, err := gitlab.DetectGitLabPath(m.cfg, m.configRef.CatIdx, m.configRef.ProjIdx)
			if err != nil {
				m.message = fmt.Sprintf("Erro: %s", err)
				return m, nil
			}
			if err := gitlab.SaveGitLabPath(m.cfg, m.configPath, m.configRef.CatIdx, m.configRef.ProjIdx, glPath); err != nil {
				m.message = fmt.Sprintf("Erro ao salvar config: %s", err)
				return m, nil
			}
			m.buildProjectList()
			m.configRef = nil
			m.screen = gitlabProjectList
			m.cursor = 0
		} else {
			m.inputBuf = ""
			m.message = ""
			m.screen = gitlabConfigPathInput
		}
	}
	return m, nil
}

func (m Model) handleConfigPathInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = gitlabConfigPath
		m.cursor = 0
		m.message = ""
	case "enter":
		val := strings.TrimSpace(m.inputBuf)
		if val == "" {
			m.message = "Path não pode ser vazio"
			return m, nil
		}
		if m.configRef == nil {
			m.message = "Projeto não identificado"
			return m, nil
		}
		if err := gitlab.SaveGitLabPath(m.cfg, m.configPath, m.configRef.CatIdx, m.configRef.ProjIdx, val); err != nil {
			m.message = fmt.Sprintf("Erro ao salvar config: %s", err)
			return m, nil
		}
		m.buildProjectList()
		m.configRef = nil
		m.screen = gitlabProjectList
		m.cursor = 0
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
