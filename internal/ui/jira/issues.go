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
	err    error
}

// IssuesLoader carrega o backlog completo de um projeto.
type IssuesLoader interface {
	ListProjectBacklog(projectKey string) ([]jiraclient.Issue, error)
}

// IssuesModel mostra o backlog do projeto agrupado por status, com busca
// por nome ou key (GAP-123) em tempo real.
type IssuesModel struct {
	loader      IssuesLoader
	projectKey  string
	projectName string
	issues      []jiraclient.Issue
	search      string
	statuses    []string
	byStatus    map[string][]jiraclient.Issue
	totalShown  int
	loading     bool
	loadErr     string
	width       int
	height      int
	offset      int
	goBack      bool
}

func NewIssuesModel(loader IssuesLoader, project jiraclient.Project) IssuesModel {
	return IssuesModel{
		loader:      loader,
		projectKey:  project.Key,
		projectName: project.Name,
		loading:     true,
	}
}

func (m IssuesModel) Init() tea.Cmd {
	loader := m.loader
	key := m.projectKey
	return func() tea.Msg {
		issues, err := loader.ListProjectBacklog(key)
		return issuesLoadedMsg{issues: issues, err: err}
	}
}

func (m IssuesModel) Update(msg tea.Msg) (IssuesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case issuesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.issues = msg.issues
		m.regroup()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg), nil
	}
	return m, nil
}

func (m IssuesModel) handleKey(msg tea.KeyMsg) IssuesModel {
	switch msg.Type {
	case tea.KeyEsc:
		// Primeiro esc limpa busca; segundo esc volta.
		if m.search != "" {
			m.search = ""
			m.regroup()
			m.offset = 0
			return m
		}
		m.goBack = true
		return m
	case tea.KeyBackspace:
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.regroup()
			m.offset = 0
		}
		return m
	case tea.KeyUp:
		if m.offset > 0 {
			m.offset--
		}
		return m
	case tea.KeyDown:
		if m.offset < m.maxOffset() {
			m.offset++
		}
		return m
	case tea.KeyPgUp:
		m.offset -= m.visibleLines()
		if m.offset < 0 {
			m.offset = 0
		}
		return m
	case tea.KeyPgDown:
		m.offset += m.visibleLines()
		if m.offset > m.maxOffset() {
			m.offset = m.maxOffset()
		}
		return m
	case tea.KeyRunes, tea.KeySpace:
		for _, r := range msg.Runes {
			m.search += string(r)
		}
		m.regroup()
		m.offset = 0
		return m
	}
	return m
}

// regroup recalcula os grupos por status a partir das issues filtradas.
func (m *IssuesModel) regroup() {
	filtered := jira.FilterIssues(m.issues, m.search)
	m.statuses, m.byStatus = jira.GroupedByStatus(filtered)
	m.totalShown = len(filtered)
}

func (m IssuesModel) View() string {
	var b strings.Builder

	b.WriteString(shared.NjordTitle() + "\n\n")
	header := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue).
		Render(fmt.Sprintf("  %s — Backlog", m.projectName))
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n")
	b.WriteString(m.renderSearch() + "\n")
	b.WriteString(divider + "\n\n")

	if m.loading {
		b.WriteString(theme.DimStyle.Render("  Carregando backlog..."))
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString(theme.ErrorStyle.Render("  ✗ " + m.loadErr))
		return b.String()
	}
	if len(m.issues) == 0 {
		b.WriteString(theme.DimStyle.Render("  Backlog vazio neste projeto."))
		return b.String()
	}
	if m.totalShown == 0 {
		b.WriteString(theme.DimStyle.Render(fmt.Sprintf("  Nenhuma issue corresponde a %q.", m.search)))
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

// renderSearch exibe o input de busca atual. Campo sempre visível — digitar
// letras/números filtra live, backspace apaga.
func (m IssuesModel) renderSearch() string {
	label := theme.DimStyle.Render("  Buscar:")
	value := m.search
	if value == "" {
		value = theme.DimStyle.Render("(digite nome ou key, ex. GAP-42)")
	} else {
		value = lipgloss.NewStyle().Foreground(theme.JiraBlueSel).Render(value + "▌")
	}
	total := ""
	if m.search != "" && !m.loading && m.loadErr == "" {
		total = theme.DimStyle.Render(fmt.Sprintf("   (%d de %d)", m.totalShown, len(m.issues)))
	}
	return label + " " + value + total
}

// buildLines produces one line per issue and per status header, in display
// order. Kept separate so scroll logic only tracks line indexes.
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
	const chromeHeight = 10 // title + header + search + divider + help + padding
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
