package jira

import (
	jiraapp "github.com/DevViking-Persike/njord-cli/internal/app/jira"
	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	tea "github.com/charmbracelet/bubbletea"
)

// createStep é a fase atual do formulário de criação.
type createStep int

const (
	createPickType createStep = iota
	createPickParent // só se type=Subtask
	createSummary
	createDesc
	createPickStatus
	createSubmitting
	createDone
)

// CreateGateway é a superfície mínima que o CreateModel precisa.
// JiraService do app/jira satisfaz.
type CreateGateway interface {
	ListProjectBacklog(projectKey string) ([]jiraclient.Issue, error)
	CreateIssueAsMe(jiraapp.CreateIssueRequest) (jiraclient.Issue, error)
}

type createLoadedMsg struct {
	backlog []jiraclient.Issue
	err     error
}

type createSubmitMsg struct {
	issue jiraclient.Issue
	err   error
}

// issueTypeOptions lista os tipos oferecidos no picker. "Subtask" cai no fluxo
// de pick parent; os outros pulam direto pro summary.
var issueTypeOptions = []string{"Task", "Bug", "Story", "Subtask"}

// statusCategoryOptions mapeia a escolha da UI pra categoria do Jira. Vazio =
// "mantém no default" (usualmente Backlog).
var statusCategoryOptions = []struct {
	label    string
	category string
}{
	{"Manter (Backlog)", ""},
	{"Em desenvolvimento", "indeterminate"},
	{"Concluído", "done"},
}

// CreateModel é o formulário multi-step de criação de issue.
type CreateModel struct {
	loader  CreateGateway
	project jiraclient.Project
	step    createStep

	issueType string
	parentKey string
	summary   string
	desc      string
	statusCat string

	cursor   int
	inputBuf string

	backlog     []jiraclient.Issue
	loading     bool
	loadErr     string
	submitErr   string
	createdKey  string
	width       int
	height      int
	goBack      bool
}

func NewCreateModel(loader CreateGateway, project jiraclient.Project) CreateModel {
	return CreateModel{loader: loader, project: project, step: createPickType}
}

func (m CreateModel) Init() tea.Cmd { return nil }

func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case createLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
		} else {
			m.backlog = msg.backlog
		}
		return m, nil
	case createSubmitMsg:
		if msg.err != nil {
			m.submitErr = msg.err.Error()
		} else {
			m.createdKey = msg.issue.Key
		}
		m.step = createDone
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m CreateModel) handleKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	switch m.step {
	case createPickType:
		return m.handlePickList(msg, len(issueTypeOptions), m.advanceFromType)
	case createPickParent:
		return m.handleParentPick(msg)
	case createSummary:
		return m.handleTextInput(msg, func(val string) (CreateModel, tea.Cmd) {
			if val == "" {
				return m, nil // summary é obrigatório
			}
			m.summary = val
			m.inputBuf = ""
			m.step = createDesc
			return m, nil
		})
	case createDesc:
		return m.handleTextInput(msg, func(val string) (CreateModel, tea.Cmd) {
			m.desc = val // pode ser vazio
			m.inputBuf = ""
			m.cursor = 0
			m.step = createPickStatus
			return m, nil
		})
	case createPickStatus:
		return m.handlePickList(msg, len(statusCategoryOptions), func() (CreateModel, tea.Cmd) {
			m.statusCat = statusCategoryOptions[m.cursor].category
			m.step = createSubmitting
			return m, m.submitCmd()
		})
	case createDone:
		if msg.String() == "enter" || msg.String() == "esc" {
			m.goBack = true
		}
	}
	return m, nil
}

// advanceFromType é chamado quando o usuário confirma o tipo. Se for Subtask,
// dispara o fetch do backlog; senão pula pro summary.
func (m CreateModel) advanceFromType() (CreateModel, tea.Cmd) {
	m.issueType = issueTypeOptions[m.cursor]
	m.cursor = 0
	if m.issueType == "Subtask" {
		m.loading = true
		m.loadErr = ""
		m.step = createPickParent
		return m, m.fetchBacklogCmd()
	}
	m.step = createSummary
	m.inputBuf = ""
	return m, nil
}

func (m CreateModel) handleParentPick(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	if m.loading {
		if msg.String() == "esc" {
			m.step = createPickType
			m.cursor = 0
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.backlog)-1 {
			m.cursor++
		}
	case "esc":
		m.step = createPickType
		m.cursor = 0
	case "enter":
		if m.cursor < len(m.backlog) {
			m.parentKey = m.backlog[m.cursor].Key
			m.inputBuf = ""
			m.step = createSummary
		}
	}
	return m, nil
}

// handlePickList roteia teclas comuns de picker (up/down/enter/esc) pra step
// que usa lista com N opções. onEnter decide como avançar.
func (m CreateModel) handlePickList(msg tea.KeyMsg, size int, onEnter func() (CreateModel, tea.Cmd)) (CreateModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < size-1 {
			m.cursor++
		}
	case "esc":
		m.goBack = true
	case "enter":
		return onEnter()
	}
	return m, nil
}

// handleTextInput cuida dos steps de campo de texto (summary/desc). onSubmit
// recebe o valor trimado e decide o próximo step.
func (m CreateModel) handleTextInput(msg tea.KeyMsg, onSubmit func(string) (CreateModel, tea.Cmd)) (CreateModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.goBack = true
	case "enter":
		return onSubmit(trim(m.inputBuf))
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

// fetchBacklogCmd é async: devolve o backlog do space pra picker de parent.
func (m CreateModel) fetchBacklogCmd() tea.Cmd {
	loader := m.loader
	key := m.project.Key
	return func() tea.Msg {
		items, err := loader.ListProjectBacklog(key)
		return createLoadedMsg{backlog: items, err: err}
	}
}

// submitCmd é async: cria a issue + aplica a transição alvo se houver.
func (m CreateModel) submitCmd() tea.Cmd {
	loader := m.loader
	req := jiraapp.CreateIssueRequest{
		ProjectKey:     m.project.Key,
		Summary:        m.summary,
		Description:    m.desc,
		Type:           m.issueType,
		ParentKey:      m.parentKey,
		TargetCategory: m.statusCat,
	}
	return func() tea.Msg {
		issue, err := loader.CreateIssueAsMe(req)
		return createSubmitMsg{issue: issue, err: err}
	}
}

func (m *CreateModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *CreateModel) GoBack() bool { return m.goBack }

// trim é igual ao strings.TrimSpace; exposto pra evitar import circular no view.
func trim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n') {
		j--
	}
	return s[i:j]
}
