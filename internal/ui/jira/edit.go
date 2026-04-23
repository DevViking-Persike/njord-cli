package jira

import (
	jiraapp "github.com/DevViking-Persike/njord-cli/internal/app/jira"
	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	tea "github.com/charmbracelet/bubbletea"
)

// editStep enumera as fases do formulário de edição.
type editStep int

const (
	editPickIssue editStep = iota
	editSummary
	editDesc
	editPickStatus
	editSubmitting
	editDone
)

// EditGateway é a superfície mínima que o EditModel precisa. JiraService
// do app/jira satisfaz.
type EditGateway interface {
	ListProjectBacklog(projectKey string) ([]jiraclient.Issue, error)
	ListTransitions(key string) ([]jiraclient.Transition, error)
	UpdateIssue(jiraapp.UpdateIssueRequest) error
}

type editBacklogMsg struct {
	issues []jiraclient.Issue
	err    error
}

type editTransitionsMsg struct {
	transitions []jiraclient.Transition
	err         error
}

type editSubmittedMsg struct {
	err error
}

// EditModel: pick issue → edita summary → desc → status → submit.
type EditModel struct {
	loader  EditGateway
	project jiraclient.Project
	step    editStep

	backlog        []jiraclient.Issue
	loadingBacklog bool
	backlogErr     string

	picked *jiraclient.Issue

	summary string
	desc    string

	transitions  []jiraclient.Transition
	loadingTrans bool
	transErr     string
	transitionID string

	cursor    int
	inputBuf  string
	submitErr string
	okKey     string

	width, height int
	goBack        bool
}

func NewEditModel(loader EditGateway, project jiraclient.Project) EditModel {
	return EditModel{
		loader:         loader,
		project:        project,
		step:           editPickIssue,
		loadingBacklog: true,
	}
}

func (m EditModel) Init() tea.Cmd {
	loader := m.loader
	key := m.project.Key
	return func() tea.Msg {
		items, err := loader.ListProjectBacklog(key)
		return editBacklogMsg{issues: items, err: err}
	}
}

func (m EditModel) Update(msg tea.Msg) (EditModel, tea.Cmd) {
	switch msg := msg.(type) {
	case editBacklogMsg:
		m.loadingBacklog = false
		if msg.err != nil {
			m.backlogErr = msg.err.Error()
		} else {
			m.backlog = msg.issues
		}
		return m, nil
	case editTransitionsMsg:
		m.loadingTrans = false
		if msg.err != nil {
			m.transErr = msg.err.Error()
		} else {
			m.transitions = msg.transitions
		}
		return m, nil
	case editSubmittedMsg:
		if msg.err != nil {
			m.submitErr = msg.err.Error()
		} else if m.picked != nil {
			m.okKey = m.picked.Key
		}
		m.step = editDone
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m EditModel) handleKey(msg tea.KeyMsg) (EditModel, tea.Cmd) {
	switch m.step {
	case editPickIssue:
		return m.handlePickIssue(msg)
	case editSummary:
		return m.handleTextField(msg, func(val string) (EditModel, tea.Cmd) {
			m.summary = val
			m.inputBuf = ""
			m.step = editDesc
			return m, nil
		})
	case editDesc:
		return m.handleTextField(msg, func(val string) (EditModel, tea.Cmd) {
			m.desc = val
			m.inputBuf = ""
			m.cursor = 0
			m.step = editPickStatus
			m.loadingTrans = true
			m.transErr = ""
			return m, m.fetchTransitionsCmd()
		})
	case editPickStatus:
		return m.handlePickStatus(msg)
	case editDone:
		if s := msg.String(); s == "enter" || s == "esc" {
			m.goBack = true
		}
	}
	return m, nil
}

func (m EditModel) handlePickIssue(msg tea.KeyMsg) (EditModel, tea.Cmd) {
	if m.loadingBacklog {
		if msg.String() == "esc" {
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
		if m.cursor < len(m.backlog)-1 {
			m.cursor++
		}
	case "esc":
		m.goBack = true
	case "enter":
		if m.cursor < len(m.backlog) {
			iss := m.backlog[m.cursor]
			m.picked = &iss
			m.summary = iss.Summary
			m.inputBuf = iss.Summary // pré-preenche o campo pra editar
			m.step = editSummary
		}
	}
	return m, nil
}

func (m EditModel) handleTextField(msg tea.KeyMsg, onSubmit func(string) (EditModel, tea.Cmd)) (EditModel, tea.Cmd) {
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

func (m EditModel) handlePickStatus(msg tea.KeyMsg) (EditModel, tea.Cmd) {
	if m.loadingTrans {
		if msg.String() == "esc" {
			m.goBack = true
		}
		return m, nil
	}
	size := len(m.transitions) + 1 // +1 do "Manter"
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
		// cursor 0 = manter; depois disso, m.cursor-1 é índice em m.transitions
		if m.cursor > 0 {
			m.transitionID = m.transitions[m.cursor-1].ID
		}
		m.step = editSubmitting
		return m, m.submitCmd()
	}
	return m, nil
}

func (m EditModel) fetchTransitionsCmd() tea.Cmd {
	loader := m.loader
	key := m.picked.Key
	return func() tea.Msg {
		ts, err := loader.ListTransitions(key)
		return editTransitionsMsg{transitions: ts, err: err}
	}
}

func (m EditModel) submitCmd() tea.Cmd {
	loader := m.loader
	req := jiraapp.UpdateIssueRequest{
		Key:          m.picked.Key,
		Summary:      m.summary,
		Description:  m.desc,
		TransitionID: m.transitionID,
	}
	return func() tea.Msg {
		err := loader.UpdateIssue(req)
		return editSubmittedMsg{err: err}
	}
}

func (m *EditModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *EditModel) GoBack() bool { return m.goBack }
