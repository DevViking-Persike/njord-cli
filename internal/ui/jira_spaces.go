package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// jiraSpacesLoadedMsg carries the async result of listing Jira projects.
type jiraSpacesLoadedMsg struct {
	projects []jiraclient.Project
	err      error
}

// JiraSpacesLoader is the minimum surface for loading Jira spaces.
// Kept as an interface so the UI never depends directly on internal/app.
type JiraSpacesLoader interface {
	ListSpaces() ([]jiraclient.Project, error)
}

// JiraSpacesModel renders a grid of Jira projects (espaços).
type JiraSpacesModel struct {
	loader    JiraSpacesLoader
	projects  []jiraclient.Project
	loading   bool
	loadErr   string
	cursor    int
	cols      int
	cardWidth int
	width     int
	height    int
	offset    int // first visible row (scroll position)
	selected  *jiraclient.Project
	goBack    bool
}

func NewJiraSpacesModel(loader JiraSpacesLoader) JiraSpacesModel {
	return JiraSpacesModel{
		loader:    loader,
		loading:   true,
		cols:      2,
		cardWidth: 36,
	}
}

func (m JiraSpacesModel) Init() tea.Cmd {
	loader := m.loader
	return func() tea.Msg {
		projects, err := loader.ListSpaces()
		return jiraSpacesLoadedMsg{projects: projects, err: err}
	}
}

func (m JiraSpacesModel) Update(msg tea.Msg) (JiraSpacesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case jiraSpacesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.projects = msg.projects
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			if msg.String() == "esc" || msg.String() == "q" {
				m.goBack = true
			}
			return m, nil
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m JiraSpacesModel) handleKey(msg tea.KeyMsg) (JiraSpacesModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor >= m.cols {
			m.cursor -= m.cols
		}
		m.ensureVisible()
	case "down", "j":
		if m.cursor+m.cols < len(m.projects) {
			m.cursor += m.cols
		}
		m.ensureVisible()
	case "left", "h":
		if m.cursor%m.cols > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor%m.cols < m.cols-1 && m.cursor+1 < len(m.projects) {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(m.projects) {
			p := m.projects[m.cursor]
			m.selected = &p
		}
	case "esc", "q":
		m.goBack = true
	}
	return m, nil
}

func (m JiraSpacesModel) View() string {
	var b strings.Builder

	b.WriteString(njordTitle() + "\n\n")
	section := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue).Render("  Jira — Espaços")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(section + "\n" + divider + "\n\n")

	if m.loading {
		b.WriteString(theme.DimStyle.Render("  Carregando projetos do Jira..."))
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString(theme.ErrorStyle.Render("  ✗ " + m.loadErr))
		b.WriteString("\n\n" + theme.HelpStyle.Render("  esc back"))
		return b.String()
	}
	if len(m.projects) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum projeto encontrado."))
		b.WriteString("\n\n" + theme.HelpStyle.Render("  esc back"))
		return b.String()
	}

	b.WriteString(m.renderGrid())

	rows := (len(m.projects) + m.cols - 1) / m.cols
	visible := m.visibleRows()
	if rows > visible {
		b.WriteString(theme.DimStyle.Render(fmt.Sprintf("  [%d/%d]", m.offset+1, rows-visible+1)))
	}
	return b.String()
}

func (m JiraSpacesModel) renderGrid() string {
	var rowBuf strings.Builder
	rows := (len(m.projects) + m.cols - 1) / m.cols
	visible := m.visibleRows()
	end := m.offset + visible
	if end > rows {
		end = rows
	}
	for row := m.offset; row < end; row++ {
		var cards []string
		for col := 0; col < m.cols; col++ {
			idx := row*m.cols + col
			if idx >= len(m.projects) {
				cards = append(cards, strings.Repeat(" ", m.cardWidth+borderOverhead))
				continue
			}
			cards = append(cards, m.renderCard(m.projects[idx], idx == m.cursor))
		}
		rowBuf.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cards...))
		rowBuf.WriteString("\n")
	}
	return rowBuf.String()
}

// visibleRows returns how many card rows fit in the viewport after the
// njord title, section header, divider, scroll indicator and app-level help.
func (m JiraSpacesModel) visibleRows() int {
	const cardHeight = 4
	// Vertical chrome: njord title (2) + section+divider (3) + help (2) + slack (1)
	const chromeHeight = 8
	available := m.height - chromeHeight
	if available < cardHeight {
		return 1
	}
	return available / cardHeight
}

func (m *JiraSpacesModel) ensureVisible() {
	row := m.cursor / m.cols
	visible := m.visibleRows()
	if row < m.offset {
		m.offset = row
	}
	if row >= m.offset+visible {
		m.offset = row - visible + 1
	}
}

func (m JiraSpacesModel) renderCard(p jiraclient.Project, selected bool) string {
	cardStyle, titleStyle, subStyle := theme.CardStyle, theme.TitleStyle, theme.SubStyle
	if selected {
		cardStyle, titleStyle, subStyle = theme.CardSelectedStyle, theme.TitleSelectedStyle, theme.SubSelectedStyle
	}
	name := titleStyle.Render(p.Name)
	key := subStyle.Render(fmt.Sprintf("key: %s", p.Key))
	content := lipgloss.JoinVertical(lipgloss.Left, name, key)
	return cardStyle.Width(m.cardWidth).Render(content)
}

func (m *JiraSpacesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.recalcLayout()
	if len(m.projects) > 0 && m.cursor >= len(m.projects) {
		m.cursor = len(m.projects) - 1
	}
	m.ensureVisible()
}

func (m *JiraSpacesModel) recalcLayout() {
	if m.width <= 0 {
		return
	}
	maxCols := m.width / (minCardWidth + borderOverhead)
	if maxCols < 1 {
		maxCols = 1
	}
	if maxCols > 5 {
		maxCols = 5
	}
	m.cols = maxCols
	m.cardWidth = (m.width / m.cols) - borderOverhead
}

// GoBack reports whether the user pressed esc/q.
func (m *JiraSpacesModel) GoBack() bool { return m.goBack }

// Selected returns the project the user picked, or nil.
func (m *JiraSpacesModel) Selected() *jiraclient.Project { return m.selected }

// ClearSelection clears the picked project.
func (m *JiraSpacesModel) ClearSelection() { m.selected = nil }
