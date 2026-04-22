package jira

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/app/jira"
	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type issuesLoadedMsg struct {
	issues []jiraclient.Issue
	mode   viewMode
	err    error
}

type viewMode int

const (
	modeBacklog viewMode = iota // todos do projeto não-concluídos
	modeMine                    // todas as minhas (qualquer status)
)

// IssuesLoader fornece backlog + histórico do usuário num projeto.
type IssuesLoader interface {
	ListProjectBacklog(projectKey string) ([]jiraclient.Issue, error)
	ListMyProjectIssues(projectKey string) ([]jiraclient.Issue, error)
}

// status filter order: "" (todos) → "indeterminate" → "done" → "new" → ""
var statusFilterCycle = []string{"", "indeterminate", "done", "new"}

var statusFilterLabel = map[string]string{
	"":              "Todos",
	"indeterminate": "Em andamento",
	"done":          "Concluído",
	"new":           "A fazer",
}

// IssuesModel mostra issues do projeto agrupadas por status, com toggle
// backlog/minhas e filtro por categoria de status.
type IssuesModel struct {
	loader       IssuesLoader
	projectKey   string
	projectName  string
	mode         viewMode
	issues       []jiraclient.Issue
	search       string
	statusFilter string // "" | "indeterminate" | "done" | "new"
	statuses     []string
	byStatus     map[string][]jiraclient.Issue
	totalShown   int
	loading      bool
	loadErr      string
	width        int
	height       int
	offset       int
	goBack       bool
}

func NewIssuesModel(loader IssuesLoader, project jiraclient.Project) IssuesModel {
	return IssuesModel{
		loader:      loader,
		projectKey:  project.Key,
		projectName: project.Name,
		mode:        modeBacklog,
		loading:     true,
	}
}

func (m IssuesModel) Init() tea.Cmd {
	return loadCmd(m.loader, m.projectKey, m.mode)
}

func loadCmd(loader IssuesLoader, key string, mode viewMode) tea.Cmd {
	return func() tea.Msg {
		var issues []jiraclient.Issue
		var err error
		if mode == modeMine {
			issues, err = loader.ListMyProjectIssues(key)
		} else {
			issues, err = loader.ListProjectBacklog(key)
		}
		return issuesLoadedMsg{issues: issues, mode: mode, err: err}
	}
}

func (m IssuesModel) Update(msg tea.Msg) (IssuesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case issuesLoadedMsg:
		// Ignora respostas de modo obsoleto (usuário trocou antes do fetch voltar).
		if msg.mode != m.mode {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.loadErr = ""
		m.issues = msg.issues
		m.regroup()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m IssuesModel) handleKey(msg tea.KeyMsg) (IssuesModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Primeiro esc limpa busca/filtro, segundo volta.
		if m.search != "" || m.statusFilter != "" {
			m.search = ""
			m.statusFilter = ""
			m.regroup()
			m.offset = 0
			return m, nil
		}
		m.goBack = true
		return m, nil
	case tea.KeyLeft:
		return m.toggleMode(modeBacklog)
	case tea.KeyRight:
		return m.toggleMode(modeMine)
	case tea.KeyTab:
		m.statusFilter = nextInCycle(statusFilterCycle, m.statusFilter)
		m.regroup()
		m.offset = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.regroup()
			m.offset = 0
		}
		return m, nil
	case tea.KeyUp:
		if m.offset > 0 {
			m.offset--
		}
		return m, nil
	case tea.KeyDown:
		if m.offset < m.maxOffset() {
			m.offset++
		}
		return m, nil
	case tea.KeyPgUp:
		m.offset -= m.visibleLines()
		if m.offset < 0 {
			m.offset = 0
		}
		return m, nil
	case tea.KeyPgDown:
		m.offset += m.visibleLines()
		if m.offset > m.maxOffset() {
			m.offset = m.maxOffset()
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		for _, r := range msg.Runes {
			m.search += string(r)
		}
		m.regroup()
		m.offset = 0
		return m, nil
	}
	return m, nil
}

func (m IssuesModel) toggleMode(target viewMode) (IssuesModel, tea.Cmd) {
	if m.mode == target {
		return m, nil
	}
	m.mode = target
	m.loading = true
	m.loadErr = ""
	m.issues = nil
	m.statuses = nil
	m.byStatus = nil
	m.totalShown = 0
	m.offset = 0
	return m, loadCmd(m.loader, m.projectKey, target)
}

func nextInCycle(cycle []string, current string) string {
	for i, v := range cycle {
		if v == current {
			return cycle[(i+1)%len(cycle)]
		}
	}
	return cycle[0]
}

// regroup recalcula os grupos aplicando filtro de categoria + busca textual.
func (m *IssuesModel) regroup() {
	filtered := jira.FilterByStatusCategory(m.issues, m.statusFilter)
	filtered = jira.FilterIssues(filtered, m.search)
	m.statuses, m.byStatus = jira.GroupedByStatus(filtered)
	m.totalShown = len(filtered)
}

func (m IssuesModel) View() string {
	var b strings.Builder

	b.WriteString(shared.NjordTitle() + "\n\n")
	b.WriteString(m.renderHeader() + "\n")
	b.WriteString(m.renderFilters() + "\n")
	b.WriteString(m.renderSearch() + "\n")
	b.WriteString(theme.DimStyle.Render("  "+strings.Repeat("─", 50)) + "\n\n")

	if m.loading {
		b.WriteString(theme.DimStyle.Render("  Carregando..."))
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString(theme.ErrorStyle.Render("  ✗ " + m.loadErr))
		return b.String()
	}
	if len(m.issues) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhuma issue neste modo."))
		return b.String()
	}
	if m.totalShown == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhuma issue corresponde aos filtros."))
		return b.String()
	}

	lines := m.buildLines()
	visible := m.visibleLines()
	end := m.offset + visible
	if end > len(lines) {
		end = len(lines)
	}
	for i := m.offset; i < end; i++ {
		b.WriteString(lines[i] + "\n")
	}
	if len(lines) > visible {
		b.WriteString(theme.DimStyle.Render(fmt.Sprintf("  [%d/%d]", m.offset+1, m.maxOffset()+1)))
	}
	return b.String()
}

func (m IssuesModel) renderHeader() string {
	base := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue)
	label := fmt.Sprintf("  %s —", m.projectName)
	tabActive := base.Background(theme.JiraBgSel).Render
	tabInactive := theme.DimStyle.Render
	var backlog, mine string
	if m.mode == modeBacklog {
		backlog = tabActive(" Backlog ")
		mine = tabInactive(" Minhas ")
	} else {
		backlog = tabInactive(" Backlog ")
		mine = tabActive(" Minhas ")
	}
	hint := theme.DimStyle.Render("  ← →")
	return label + backlog + "/" + mine + hint
}

func (m IssuesModel) renderFilters() string {
	hint := theme.DimStyle.Render("tab")
	var parts []string
	active := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlueSel).Background(theme.JiraBgSel).Render
	inactive := theme.DimStyle.Render
	for _, key := range statusFilterCycle {
		label := " " + statusFilterLabel[key] + " "
		if key == m.statusFilter {
			parts = append(parts, active(label))
		} else {
			parts = append(parts, inactive(label))
		}
	}
	return "  " + hint + " " + strings.Join(parts, " ")
}

func (m IssuesModel) renderSearch() string {
	label := theme.DimStyle.Render("  Buscar:")
	value := m.search
	if value == "" {
		value = theme.DimStyle.Render("(digite nome ou key, ex. GAP-42)")
	} else {
		value = lipgloss.NewStyle().Foreground(theme.JiraBlueSel).Render(value + "▌")
	}
	counter := ""
	if !m.loading && m.loadErr == "" && (m.search != "" || m.statusFilter != "") {
		counter = theme.DimStyle.Render(fmt.Sprintf("   (%d de %d)", m.totalShown, len(m.issues)))
	}
	return label + " " + value + counter
}

func (m IssuesModel) buildLines() []string {
	var lines []string
	countStyle := theme.DimStyle
	statusStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlueSel)

	for i, status := range m.statuses {
		issues := m.byStatus[status]
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, statusStyle.Render("  "+status)+countStyle.Render(fmt.Sprintf("  (%d)", len(issues))))
		for _, iss := range issues {
			lines = append(lines, formatIssueLine(iss))
		}
	}
	return lines
}

func formatIssueLine(iss jiraclient.Issue) string {
	keyStyle := lipgloss.NewStyle().Foreground(theme.JiraBlue)
	typeStyle := theme.DimStyle
	summary := iss.Summary
	if len(summary) > 80 {
		summary = summary[:77] + "..."
	}
	return fmt.Sprintf("    %s  %s  %s",
		keyStyle.Render(iss.Key),
		typeStyle.Render(fmt.Sprintf("[%s]", iss.Type)),
		summary,
	)
}

func (m IssuesModel) visibleLines() int {
	const chromeHeight = 12 // title + header + filters + search + divider + help + padding
	v := m.height - chromeHeight
	if v < 3 {
		return 3
	}
	return v
}

func (m IssuesModel) maxOffset() int {
	max := len(m.buildLines()) - m.visibleLines()
	if max < 0 {
		return 0
	}
	return max
}

func (m *IssuesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if max := m.maxOffset(); m.offset > max {
		m.offset = max
	}
}

func (m *IssuesModel) GoBack() bool { return m.goBack }
