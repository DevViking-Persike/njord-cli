package gitlab

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/app/gitlab"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

type gitlabActionsScreen int

const (
	glActionsMenu gitlabActionsScreen = iota
	glActionsMRList
	glActionsPipelineList
	glActionsTriggerBranch
	glActionsTriggerConfirm
	glActionsCreateBranchSigla
	glActionsCreateBranchNumber
	glActionsCreateBranchType
	glActionsCreateBranchDesc
	glActionsCreateBranchBase
	glActionsResult
)

// Chrome lines used for scroll calculations in gitlab actions.
const glActionsChromeLines = 10

// Team definitions for Jira branch naming
type jiraTeam struct {
	Code  string // A1, B1, C1, etc.
	Sigla string // PLA, BILL, SIE, etc.
	Name  string // Full team name
}

var jiraTeams = []jiraTeam{
	{Code: "A1", Sigla: "PLA", Name: "Plataforma"},
	{Code: "B1", Sigla: "BILL", Name: "Billing - Financeiro"},
	{Code: "C1", Sigla: "SIE", Name: "Gestão de Apólice"},
	{Code: "D1", Sigla: "GAP", Name: "Consistência dos Dados"},
	{Code: "E1", Sigla: "SBO", Name: "Backoffice"},
	{Code: "F1", Sigla: "FOPS", Name: "Ops - Novos Clientes"},
	{Code: "H1", Sigla: "HOT", Name: "Hotfix"},
	{Code: "L1", Sigla: "LOW", Name: "Low Priority"},
	{Code: "S1", Sigla: "SPAVT", Name: "Suporte"},
}

// Async messages
type gitlabMRsMsg struct {
	mrs []gitlabclient.MergeRequestInfo
	err error
}

type gitlabPipelinesMsg struct {
	pipelines []gitlabclient.PipelineInfo
	err       error
}

type gitlabBranchesMsg struct {
	branches []gitlabclient.BranchInfo
	err      error
}

type gitlabActionDoneMsg struct {
	success bool
	message string
	err     error
}

type ActionsModel struct {
	client      *gitlabclient.Client
	projectPath string
	projectName string
	gitlabURL   string
	goBack      bool

	screen  gitlabActionsScreen
	cursor  int
	scroll  components.ScrollState
	loading bool

	// Data
	mrs       []gitlabclient.MergeRequestInfo
	pipelines []gitlabclient.PipelineInfo
	branches  []gitlabclient.BranchInfo

	// Input state
	inputBuf       string
	branchName     string
	branchSiglaIdx int
	branchNumber   string
	branchTypeIdx  int // 0=delivery, 1=subtask
	message        string
	msgType        string

	width int
}

func NewActionsModel(client *gitlabclient.Client, projectPath, projectName, gitlabURL string) ActionsModel {
	return ActionsModel{
		client:      client,
		projectPath: projectPath,
		projectName: projectName,
		gitlabURL:   gitlabURL,
		screen:      glActionsMenu,
	}
}

func (m ActionsModel) Init() tea.Cmd {
	return nil
}

func (m ActionsModel) Update(msg tea.Msg) (ActionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case gitlabMRsMsg:
		m.loading = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro: %s", msg.err)
			m.msgType = "error"
			m.screen = glActionsResult
			return m, nil
		}
		m.mrs = msg.mrs
		m.cursor = 0
		m.scroll.Offset = 0
		m.screen = glActionsMRList
		return m, nil

	case gitlabPipelinesMsg:
		m.loading = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro: %s", msg.err)
			m.msgType = "error"
			m.screen = glActionsResult
			return m, nil
		}
		m.pipelines = msg.pipelines
		m.cursor = 0
		m.scroll.Offset = 0
		m.screen = glActionsPipelineList
		return m, nil

	case gitlabBranchesMsg:
		m.loading = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro: %s", msg.err)
			m.msgType = "error"
			m.screen = glActionsResult
			return m, nil
		}
		m.branches = msg.branches
		m.cursor = 0
		m.scroll.Offset = 0
		return m, nil

	case gitlabActionDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro: %s", msg.err)
			m.msgType = "error"
		} else {
			m.message = msg.message
			m.msgType = "ok"
		}
		m.screen = glActionsResult
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch m.screen {
		case glActionsMenu:
			return m.handleMenu(msg)
		case glActionsMRList:
			return m.handleMRList(msg)
		case glActionsPipelineList:
			return m.handlePipelineList(msg)
		case glActionsTriggerBranch:
			return m.handleTriggerBranch(msg)
		case glActionsTriggerConfirm:
			return m.handleTriggerConfirm(msg)
		case glActionsCreateBranchSigla:
			return m.handleCreateBranchSigla(msg)
		case glActionsCreateBranchNumber:
			return m.handleCreateBranchNumber(msg)
		case glActionsCreateBranchType:
			return m.handleCreateBranchType(msg)
		case glActionsCreateBranchDesc:
			return m.handleCreateBranchDesc(msg)
		case glActionsCreateBranchBase:
			return m.handleCreateBranchBase(msg)
		case glActionsResult:
			return m.handleResult(msg)
		}
	}
	return m, nil
}

func (m ActionsModel) View() string {
	var b strings.Builder

	header := theme.GitLabTitleSelectedStyle.Render("  ◆ GitLab — " + m.projectName)
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	if m.loading {
		b.WriteString("  " + theme.DimStyle.Render("Carregando...") + "\n")
		return b.String()
	}

	switch m.screen {
	case glActionsMenu:
		b.WriteString(m.viewMenu())
	case glActionsMRList:
		b.WriteString(m.viewMRList())
	case glActionsPipelineList:
		b.WriteString(m.viewPipelineList())
	case glActionsTriggerBranch:
		b.WriteString(m.viewTriggerBranch())
	case glActionsTriggerConfirm:
		b.WriteString(m.viewTriggerConfirm())
	case glActionsCreateBranchSigla:
		b.WriteString(m.viewCreateBranchSigla())
	case glActionsCreateBranchNumber:
		b.WriteString(m.viewCreateBranchNumber())
	case glActionsCreateBranchType:
		b.WriteString(m.viewCreateBranchType())
	case glActionsCreateBranchDesc:
		b.WriteString(m.viewCreateBranchDesc())
	case glActionsCreateBranchBase:
		b.WriteString(m.viewCreateBranchBase())
	case glActionsResult:
		b.WriteString(m.viewResult())
	}

	return b.String()
}

func (m *ActionsModel) SetSize(w, h int) {
	m.width = w
	m.scroll.Height = h
}

func (m *ActionsModel) GoBack() bool { return m.goBack }

// --- Menu ---

func (m ActionsModel) viewMenu() string {
	var b strings.Builder
	options := []string{"Merge Requests", "Pipelines", "Disparar Pipeline", "Criar Branch", "Abrir no Navegador"}
	selectedFn := func(s string) string { return theme.GitLabTitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString("  " + theme.TextStyle.Render("O que deseja fazer?") + "\n\n")
	b.WriteString(components.RenderMenuOptions(options, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m ActionsModel) handleMenu(msg tea.KeyMsg) (ActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 5); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.goBack = true
	case "enter":
		switch m.cursor {
		case 0: // Merge Requests
			m.loading = true
			return m, m.fetchMRs()
		case 1: // Pipelines
			m.loading = true
			return m, m.fetchPipelines()
		case 2: // Disparar Pipeline
			m.loading = true
			m.screen = glActionsTriggerBranch
			return m, m.fetchBranches()
		case 3: // Criar Branch
			m.branchSiglaIdx = 0
			m.cursor = 0
			m.screen = glActionsCreateBranchSigla
		case 4: // Abrir no Navegador
			url := strings.TrimRight(m.gitlabURL, "/") + "/" + m.projectPath
			openBrowser(url)
			m.message = "Navegador aberto: " + url
			m.msgType = "ok"
			m.screen = glActionsResult
		}
	}
	return m, nil
}

// --- Result ---

func (m ActionsModel) viewResult() string {
	var b strings.Builder
	b.WriteString(components.RenderMessage(m.message, m.msgType))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	return b.String()
}

func (m ActionsModel) handleResult(msg tea.KeyMsg) (ActionsModel, tea.Cmd) {
	if msg.String() == "enter" || msg.String() == "esc" {
		m.screen = glActionsMenu
		m.cursor = 0
		m.message = ""
	}
	return m, nil
}

// --- Async commands ---

func (m ActionsModel) fetchMRs() tea.Cmd {
	return func() tea.Msg {
		mrs, err := gitlab.LoadMergeRequests(m.client, m.projectPath)
		return gitlabMRsMsg{mrs: mrs, err: err}
	}
}

func (m ActionsModel) fetchPipelines() tea.Cmd {
	return func() tea.Msg {
		pipelines, err := gitlab.LoadPipelines(m.client, m.projectPath, 20)
		return gitlabPipelinesMsg{pipelines: pipelines, err: err}
	}
}

func (m ActionsModel) fetchBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := gitlab.LoadBranches(m.client, m.projectPath)
		return gitlabBranchesMsg{branches: branches, err: err}
	}
}

func (m ActionsModel) triggerPipeline(ref string) tea.Cmd {
	return func() tea.Msg {
		message, err := gitlab.TriggerProjectPipeline(m.client, m.projectPath, ref)
		if err != nil {
			return gitlabActionDoneMsg{err: err}
		}
		return gitlabActionDoneMsg{
			success: true,
			message: message,
		}
	}
}

func (m ActionsModel) createBranch(name, ref string) tea.Cmd {
	return func() tea.Msg {
		message, err := gitlab.CreateProjectBranch(m.client, m.projectPath, name, ref)
		if err != nil {
			return gitlabActionDoneMsg{err: err}
		}
		return gitlabActionDoneMsg{
			success: true,
			message: message,
		}
	}
}
