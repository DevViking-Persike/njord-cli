package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

// Spinner frames for running/pending pipelines
var spinnerFrames = []string{"◐", "◓", "◑", "◒"}

// Pipeline status icons
const (
	iconSuccess  = "✓"
	iconFailed   = "✗"
	iconBlocked  = "⊘"
	iconUnknown  = "○"
)

type gitlabScreen int

const (
	gitlabProjectList gitlabScreen = iota
	gitlabConfigPath
	gitlabConfigPathInput
)

type gitlabProjectRef struct {
	catIdx  int
	projIdx int
	project config.Project
	catName string
}

// Async message for pipeline status + approval
type gitlabProjectStatusMsg struct {
	gitlabPath string
	status     string // success, failed, running, pending, canceled, blocked, ""
	lastTime   time.Time
	approval   *gitlab.MRApprovalInfo
}

// Tick message for spinner animation
type gitlabSpinnerTickMsg struct{}

type GitLabModel struct {
	cfg        *config.Config
	configPath string
	client     *gitlab.Client
	screen     gitlabScreen
	goBack     bool
	selected   *gitlabProjectRef

	projects       []gitlabProjectRef
	pipelineStatus map[string]string              // gitlab_path -> status
	approvalInfo   map[string]*gitlab.MRApprovalInfo // gitlab_path -> approval
	lastActivity   map[string]time.Time           // gitlab_path -> last pipeline time
	loadedCount    int
	sorted         bool
	cursor         int
	offset         int
	spinnerFrame   int

	inputBuf string
	message  string

	width, height int
}

func NewGitLabModel(cfg *config.Config, configPath string, client *gitlab.Client) GitLabModel {
	m := GitLabModel{
		cfg:            cfg,
		configPath:     configPath,
		client:         client,
		screen:         gitlabProjectList,
		pipelineStatus: make(map[string]string),
		approvalInfo:   make(map[string]*gitlab.MRApprovalInfo),
		lastActivity:   make(map[string]time.Time),
	}
	m.buildProjectList()
	return m
}

func (m *GitLabModel) buildProjectList() {
	m.projects = nil
	for ci, cat := range m.cfg.Categories {
		for pi, proj := range cat.Projects {
			if proj.GitLabPath != "" {
				m.projects = append(m.projects, gitlabProjectRef{
					catIdx:  ci,
					projIdx: pi,
					project: proj,
					catName: cat.Name,
				})
			}
		}
	}
}

func (m GitLabModel) Init() tea.Cmd {
	if m.client == nil || len(m.projects) == 0 {
		return nil
	}
	// Fetch pipeline status for all projects + start spinner
	cmds := []tea.Cmd{spinnerTick()}
	for _, ref := range m.projects {
		path := ref.project.GitLabPath
		cmds = append(cmds, m.fetchProjectStatus(path))
	}
	return tea.Batch(cmds...)
}

func (m GitLabModel) Update(msg tea.Msg) (GitLabModel, tea.Cmd) {
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

func (m GitLabModel) View() string {
	var b strings.Builder

	header := theme.GitLabTitleSelectedStyle.Render("  ◆ GitLab")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	switch m.screen {
	case gitlabProjectList:
		if len(m.projects) == 0 {
			b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto com gitlab_path configurado.") + "\n\n")
			b.WriteString("  " + theme.TextStyle.Render("Configure o gitlab_path dos projetos via Settings") + "\n")
			b.WriteString("  " + theme.TextStyle.Render("ou edite o njord.yaml manualmente.") + "\n")
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
				icon := m.renderStatusIcon(ref.project.GitLabPath)
				label := ref.project.Alias + " — " + ref.project.Desc
				approvalTag := m.renderApprovalTag(ref.project.GitLabPath)
				pathTag := theme.DimStyle.Render(" [" + ref.project.GitLabPath + "]")
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
		ref := m.projects[m.cursor]
		b.WriteString("  " + theme.TextStyle.Render("Projeto: "+ref.project.Alias) + "\n\n")
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

func (m *GitLabModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *GitLabModel) GoBack() bool              { return m.goBack }
func (m *GitLabModel) Selected() *gitlabProjectRef { return m.selected }
func (m *GitLabModel) ClearSelection()            { m.selected = nil }

func (m GitLabModel) visibleRows() int {
	available := m.height - 10
	if available < 3 {
		return 3
	}
	return available
}

func (m *GitLabModel) ensureVisible() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// --- Status icon rendering ---

func (m GitLabModel) renderStatusIcon(gitlabPath string) string {
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

func (m GitLabModel) hasActiveSpinners() bool {
	// Active if any project still loading or has running/pending pipeline
	for _, ref := range m.projects {
		status, ok := m.pipelineStatus[ref.project.GitLabPath]
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

func (m GitLabModel) fetchProjectStatus(gitlabPath string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		msg := gitlabProjectStatusMsg{gitlabPath: gitlabPath}

		pipelines, err := client.ListPipelines(gitlabPath, 1)
		if err == nil && len(pipelines) > 0 {
			msg.status = pipelines[0].Status
			msg.lastTime = pipelines[0].CreatedAt
		}

		approval, _ := client.GetProjectLatestMRApproval(gitlabPath)
		msg.approval = approval

		return msg
	}
}

func (m *GitLabModel) sortByRecent() {
	sort.SliceStable(m.projects, func(i, j int) bool {
		ti := m.lastActivity[m.projects[i].project.GitLabPath]
		tj := m.lastActivity[m.projects[j].project.GitLabPath]
		return ti.After(tj)
	})
}

func (m GitLabModel) renderApprovalTag(gitlabPath string) string {
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

func (m GitLabModel) handleProjectList(msg tea.KeyMsg) (GitLabModel, tea.Cmd) {
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
			m.selected = &ref
		}
	}
	return m, nil
}

func (m GitLabModel) handleConfigPath(msg tea.KeyMsg) (GitLabModel, tea.Cmd) {
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
	case "enter":
		if m.cursor == 0 {
			// Auto-detect
			ref := m.projects[m.cursor]
			path := m.cfg.ResolveProjectPath(ref.project)
			glPath, err := gitlab.ParseGitLabPath(path)
			if err != nil {
				m.message = fmt.Sprintf("Erro: %s", err)
				return m, nil
			}
			m.cfg.Categories[ref.catIdx].Projects[ref.projIdx].GitLabPath = glPath
			_ = config.Save(m.cfg, m.configPath)
			m.buildProjectList()
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

func (m GitLabModel) handleConfigPathInput(msg tea.KeyMsg) (GitLabModel, tea.Cmd) {
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
