package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	mrs []gitlab.MergeRequestInfo
	err error
}

type gitlabPipelinesMsg struct {
	pipelines []gitlab.PipelineInfo
	err       error
}

type gitlabBranchesMsg struct {
	branches []gitlab.BranchInfo
	err      error
}

type gitlabActionDoneMsg struct {
	success bool
	message string
	err     error
}

type GitLabActionsModel struct {
	client      *gitlab.Client
	projectPath string
	projectName string
	gitlabURL   string
	goBack      bool

	screen  gitlabActionsScreen
	cursor  int
	scroll  components.ScrollState
	loading bool

	// Data
	mrs       []gitlab.MergeRequestInfo
	pipelines []gitlab.PipelineInfo
	branches  []gitlab.BranchInfo

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

func NewGitLabActionsModel(client *gitlab.Client, projectPath, projectName, gitlabURL string) GitLabActionsModel {
	return GitLabActionsModel{
		client:      client,
		projectPath: projectPath,
		projectName: projectName,
		gitlabURL:   gitlabURL,
		screen:      glActionsMenu,
	}
}

func (m GitLabActionsModel) Init() tea.Cmd {
	return nil
}

func (m GitLabActionsModel) Update(msg tea.Msg) (GitLabActionsModel, tea.Cmd) {
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

func (m GitLabActionsModel) View() string {
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

func (m *GitLabActionsModel) SetSize(w, h int) {
	m.width = w
	m.scroll.Height = h
}

func (m *GitLabActionsModel) GoBack() bool { return m.goBack }

// --- Menu ---

func (m GitLabActionsModel) viewMenu() string {
	var b strings.Builder
	options := []string{"Merge Requests", "Pipelines", "Disparar Pipeline", "Criar Branch", "Abrir no Navegador"}
	selectedFn := func(s string) string { return theme.GitLabTitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString("  " + theme.TextStyle.Render("O que deseja fazer?") + "\n\n")
	b.WriteString(components.RenderMenuOptions(options, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleMenu(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
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

// --- Merge Requests ---

func (m GitLabActionsModel) viewMRList() string {
	var b strings.Builder
	if len(m.mrs) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhum MR aberto encontrado.") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render(fmt.Sprintf("Merge Requests abertos (%d):", len(m.mrs))) + "\n\n")
		start, end := m.scroll.Bounds(len(m.mrs), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			mr := m.mrs[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "▶ "
			}

			stateStyle := mrStateStyle(mr.State)
			stateTag := stateStyle.Render("[" + mr.State + "]")

			pipelineTag := ""
			if mr.Pipeline != "" {
				ps := pipelineStateStyle(mr.Pipeline)
				pipelineTag = " " + ps.Render("⬤ "+mr.Pipeline)
			}

			title := fmt.Sprintf("!%d %s", mr.IID, mr.Title)
			if i == m.cursor {
				title = theme.GitLabTitleSelectedStyle.Render(prefix+title) + " " + stateTag + pipelineTag
			} else {
				title = theme.TextStyle.Render(prefix+title) + " " + stateTag + pipelineTag
			}
			b.WriteString("  " + title + "\n")

			branchInfo := theme.DimStyle.Render(fmt.Sprintf("     %s → %s  por %s  %s",
				mr.Branch, mr.Target, mr.Author, timeAgo(mr.CreatedAt)))
			b.WriteString("  " + branchInfo + "\n")
		}
		components.RenderScrollDown(&b, end, len(m.mrs))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleMRList(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.mrs)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	}
	return m, nil
}

// --- Pipelines ---

func (m GitLabActionsModel) viewPipelineList() string {
	var b strings.Builder
	if len(m.pipelines) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhuma pipeline encontrada.") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render(fmt.Sprintf("Pipelines recentes (%d):", len(m.pipelines))) + "\n\n")
		start, end := m.scroll.Bounds(len(m.pipelines), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			p := m.pipelines[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "▶ "
			}

			ps := pipelineStateStyle(p.Status)
			statusTag := ps.Render("⬤ " + p.Status)

			line := fmt.Sprintf("#%d  %s  ref: %s  %s", p.ID, statusTag, p.Ref, timeAgo(p.CreatedAt))
			if i == m.cursor {
				b.WriteString("  " + theme.GitLabTitleSelectedStyle.Render(prefix) + line + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render(prefix) + line + "\n")
			}
		}
		components.RenderScrollDown(&b, end, len(m.pipelines))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handlePipelineList(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.pipelines)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	}
	return m, nil
}

// --- Trigger Pipeline ---

func (m GitLabActionsModel) viewTriggerBranch() string {
	var b strings.Builder
	if len(m.branches) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Carregando branches...") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render("Selecione a branch para disparar pipeline:") + "\n\n")
		start, end := m.scroll.Bounds(len(m.branches), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			b.WriteString(m.renderBranchLine(i))
		}
		components.RenderScrollDown(&b, end, len(m.branches))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleTriggerBranch(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	if len(m.branches) == 0 {
		if msg.String() == "esc" {
			m.screen = glActionsMenu
			m.cursor = 0
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.branches)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	case "enter":
		m.branchName = m.branches[m.cursor].Name
		m.cursor = 0
		m.screen = glActionsTriggerConfirm
	}
	return m, nil
}

func (m GitLabActionsModel) viewTriggerConfirm() string {
	var b strings.Builder
	b.WriteString("  " + theme.WarningStyle.Render("Disparar pipeline na branch:") + "\n\n")
	b.WriteString("  " + theme.GitLabTitleStyle.Render(m.branchName) + "\n\n")
	opts := []string{"Sim, disparar", "Não, cancelar"}
	selectedFn := func(s string) string { return theme.TitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString(components.RenderMenuOptions(opts, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter confirmar"))
	return b.String()
}

func (m GitLabActionsModel) handleTriggerConfirm(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 2); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = glActionsTriggerBranch
		m.cursor = 0
	case "enter":
		if m.cursor == 0 {
			m.loading = true
			return m, m.triggerPipeline(m.branchName)
		}
		m.screen = glActionsTriggerBranch
		m.cursor = 0
	}
	return m, nil
}

// --- Create Branch (Jira flow) ---

// Step 1: Select Jira sigla

func (m GitLabActionsModel) viewCreateBranchSigla() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Selecione a sigla Jira:") + "\n\n")
	start, end := m.scroll.Bounds(len(jiraTeams), glActionsChromeLines)
	components.RenderScrollUp(&b, start)
	for i := start; i < end; i++ {
		t := jiraTeams[i]
		label := fmt.Sprintf("%s — %s (%s)", t.Sigla, t.Name, t.Code)
		if i == m.cursor {
			b.WriteString("  " + theme.GitLabTitleSelectedStyle.Render("▶ "+label) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+label) + "\n")
		}
	}
	components.RenderScrollDown(&b, end, len(jiraTeams))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchSigla(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(jiraTeams)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	case "enter":
		m.branchSiglaIdx = m.cursor
		m.inputBuf = ""
		m.message = ""
		m.screen = glActionsCreateBranchNumber
	}
	return m, nil
}

// Step 2: Enter ticket number

func (m GitLabActionsModel) viewCreateBranchNumber() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	preview := fmt.Sprintf("feature/%s-___-%s-...", t.Sigla, t.Code)
	b.WriteString("  " + theme.DimStyle.Render(preview) + "\n\n")
	b.WriteString("  " + theme.TextStyle.Render("Número do ticket Jira:") + "\n\n")
	b.WriteString("  " + theme.TitleStyle.Render(t.Sigla+"-") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
	b.WriteString(components.RenderError(m.message))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchNumber(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = glActionsCreateBranchSigla
		m.cursor = m.branchSiglaIdx
		m.message = ""
	case "enter":
		val := strings.TrimSpace(m.inputBuf)
		if val == "" {
			m.message = "Número não pode ser vazio"
			return m, nil
		}
		m.branchNumber = val
		m.inputBuf = ""
		m.message = ""
		m.branchTypeIdx = 0
		m.cursor = 0
		m.screen = glActionsCreateBranchType
	default:
		if newBuf, handled := components.TextInput(msg, m.inputBuf, components.DigitsOnly); handled {
			m.inputBuf = newBuf
		}
	}
	return m, nil
}

// Step 3: Select type (delivery/subtask)

func (m GitLabActionsModel) viewCreateBranchType() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	preview := fmt.Sprintf("feature/%s-%s-%s-___-...", t.Sigla, m.branchNumber, t.Code)
	b.WriteString("  " + theme.DimStyle.Render(preview) + "\n\n")
	b.WriteString("  " + theme.TextStyle.Render("Tipo da branch:") + "\n\n")
	types := []string{"delivery", "subtask"}
	selectedFn := func(s string) string { return theme.GitLabTitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString(components.RenderMenuOptions(types, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchType(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 2); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = glActionsCreateBranchNumber
		m.inputBuf = m.branchNumber
	case "enter":
		m.branchTypeIdx = m.cursor
		m.inputBuf = ""
		m.message = ""
		m.screen = glActionsCreateBranchDesc
	}
	return m, nil
}

// Step 4: Enter description

func (m GitLabActionsModel) viewCreateBranchDesc() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	branchTypes := []string{"delivery", "subtask"}
	prefix := fmt.Sprintf("feature/%s-%s-%s-%s", t.Sigla, m.branchNumber, t.Code, branchTypes[m.branchTypeIdx])

	normalized := normalizeBranchDesc(m.inputBuf)
	if normalized != "" {
		b.WriteString("  " + theme.DimStyle.Render(prefix+"-"+normalized) + "\n\n")
	} else {
		b.WriteString("  " + theme.DimStyle.Render(prefix+"-...") + "\n\n")
	}

	b.WriteString("  " + theme.TextStyle.Render("Descrição da branch:") + "\n\n")
	b.WriteString(components.RenderTextInput(m.inputBuf, ""))
	b.WriteString(components.RenderError(m.message))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchDesc(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = glActionsCreateBranchType
		m.cursor = m.branchTypeIdx
	case "enter":
		desc := normalizeBranchDesc(m.inputBuf)
		if desc == "" {
			m.message = "Descrição não pode ser vazia"
			return m, nil
		}
		t := jiraTeams[m.branchSiglaIdx]
		branchTypes := []string{"delivery", "subtask"}
		m.branchName = fmt.Sprintf("feature/%s-%s-%s-%s-%s",
			t.Sigla, m.branchNumber, t.Code, branchTypes[m.branchTypeIdx], desc)
		m.inputBuf = ""
		m.message = ""
		m.cursor = 0
		m.scroll.Offset = 0
		m.loading = true
		m.screen = glActionsCreateBranchBase
		return m, m.fetchBranches()
	default:
		if newBuf, handled := components.TextInput(msg, m.inputBuf, nil); handled {
			m.inputBuf = newBuf
		}
	}
	return m, nil
}

// normalizeBranchDesc normalizes branch description: lowercase, no accents, spaces to dashes
func normalizeBranchDesc(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = removeAccents(s)
	s = strings.ReplaceAll(s, " ", "-")
	var buf strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			buf.WriteRune(r)
		}
	}
	// Collapse multiple dashes
	result := buf.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}

func removeAccents(s string) string {
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
	}
	var buf strings.Builder
	for _, r := range s {
		if rep, ok := replacements[r]; ok {
			buf.WriteRune(rep)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// Step 5: Select base branch

func (m GitLabActionsModel) viewCreateBranchBase() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Nova branch: ") + theme.GitLabTitleStyle.Render(m.branchName) + "\n\n")
	if len(m.branches) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Carregando branches...") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render("Selecione a branch base:") + "\n\n")
		start, end := m.scroll.Bounds(len(m.branches), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			b.WriteString(m.renderBranchLine(i))
		}
		components.RenderScrollDown(&b, end, len(m.branches))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchBase(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	if len(m.branches) == 0 {
		if msg.String() == "esc" {
			m.screen = glActionsCreateBranchDesc
			m.inputBuf = ""
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.branches)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsCreateBranchDesc
		m.inputBuf = ""
	case "enter":
		baseBranch := m.branches[m.cursor].Name
		m.loading = true
		return m, m.createBranch(m.branchName, baseBranch)
	}
	return m, nil
}

// --- Result ---

func (m GitLabActionsModel) viewResult() string {
	var b strings.Builder
	b.WriteString(components.RenderMessage(m.message, m.msgType))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleResult(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	if msg.String() == "enter" || msg.String() == "esc" {
		m.screen = glActionsMenu
		m.cursor = 0
		m.message = ""
	}
	return m, nil
}

// --- Async commands ---

func (m GitLabActionsModel) fetchMRs() tea.Cmd {
	return func() tea.Msg {
		mrs, err := m.client.ListMergeRequests(m.projectPath, "opened")
		return gitlabMRsMsg{mrs: mrs, err: err}
	}
}

func (m GitLabActionsModel) fetchPipelines() tea.Cmd {
	return func() tea.Msg {
		pipelines, err := m.client.ListPipelines(m.projectPath, 20)
		return gitlabPipelinesMsg{pipelines: pipelines, err: err}
	}
}

func (m GitLabActionsModel) fetchBranches() tea.Cmd {
	return func() tea.Msg {
		branches, err := m.client.ListBranchesDetailed(m.projectPath)
		return gitlabBranchesMsg{branches: branches, err: err}
	}
}

func (m GitLabActionsModel) triggerPipeline(ref string) tea.Cmd {
	return func() tea.Msg {
		p, err := m.client.TriggerPipeline(m.projectPath, ref)
		if err != nil {
			return gitlabActionDoneMsg{err: err}
		}
		return gitlabActionDoneMsg{
			success: true,
			message: fmt.Sprintf("Pipeline #%d disparada na branch %s", p.ID, ref),
		}
	}
}

func (m GitLabActionsModel) createBranch(name, ref string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.CreateBranch(m.projectPath, name, ref)
		if err != nil {
			return gitlabActionDoneMsg{err: err}
		}
		return gitlabActionDoneMsg{
			success: true,
			message: fmt.Sprintf("Branch '%s' criada a partir de '%s'", name, ref),
		}
	}
}

// --- Helpers ---

func mrStateStyle(state string) lipgloss.Style {
	switch state {
	case "opened":
		return theme.MROpenStyle
	case "merged":
		return theme.MRMergedStyle
	case "closed":
		return theme.MRClosedStyle
	default:
		return theme.DimStyle
	}
}

func pipelineStateStyle(status string) lipgloss.Style {
	switch status {
	case "success":
		return theme.PipelineSuccessStyle
	case "failed":
		return theme.PipelineFailedStyle
	case "running":
		return theme.PipelineRunningStyle
	case "pending":
		return theme.PipelinePendingStyle
	case "canceled":
		return theme.DimStyle
	default:
		return theme.DimStyle
	}
}

func (m GitLabActionsModel) renderBranchLine(i int) string {
	branch := m.branches[i]
	prefix := "  "
	if i == m.cursor {
		prefix = "▶ "
	}

	var namePart string
	if i == m.cursor {
		namePart = theme.GitLabTitleSelectedStyle.Render(prefix + branch.Name)
	} else {
		namePart = theme.TextStyle.Render(prefix + branch.Name)
	}

	tags := ""
	if branch.Default {
		tags += " " + theme.DimStyle.Render("[default]")
	}
	if branch.Protected {
		tags += " " + theme.DimStyle.Render("[protected]")
	}

	approvalTag := ""
	if branch.MRApproval != nil {
		approvalTag = " " + renderApprovalTag(branch.MRApproval)
	}

	ago := ""
	if !branch.CommitDate.IsZero() {
		ago = " " + theme.DimStyle.Render(timeAgo(branch.CommitDate))
	}

	return "  " + namePart + tags + approvalTag + ago + "\n"
}

func renderApprovalTag(a *gitlab.MRApprovalInfo) string {
	if a.Approved {
		return theme.PipelineSuccessStyle.Render("✓ aprovado")
	}
	left := a.ApprovalsRequired - a.ApprovalsGiven
	if left <= 0 {
		left = 1
	}
	label := fmt.Sprintf("⏳ %d/%d", a.ApprovalsGiven, a.ApprovalsRequired)
	if a.RuleName != "" {
		label += " " + a.RuleName
	}
	return theme.PipelinePendingStyle.Render(label)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "agora"
	case d < time.Hour:
		return fmt.Sprintf("%dm atrás", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh atrás", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "ontem"
		}
		return fmt.Sprintf("%dd atrás", days)
	}
}
