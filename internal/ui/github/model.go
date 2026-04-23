// Package github é a tela TUI que lista projetos considerados GitHub
// (por github_path ou por pertencerem à categoria pessoal) e permite
// abrir no browser ou clonar.
package github

import (
	"strings"

	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
)

// Model lista os projetos GitHub e deixa o usuário navegar pra tela de ações.
type Model struct {
	cfg      *config.Config
	projects []githubapp.ProjectRef
	cursor   int
	offset   int
	width    int
	height   int
	selected *githubapp.ProjectRef
	goBack   bool
}

func NewModel(cfg *config.Config) Model {
	return Model{
		cfg:      cfg,
		projects: githubapp.FilterGitHub(cfg),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(key)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
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
	case "enter":
		if m.cursor < len(m.projects) {
			ref := m.projects[m.cursor]
			m.selected = &ref
		}
	case "esc", "q":
		m.goBack = true
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := theme.TitleSelectedStyle.Render("   GitHub — projetos")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	if len(m.projects) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto GitHub encontrado.") + "\n")
		return b.String()
	}

	visible := m.visibleRows()
	end := m.offset + visible
	if end > len(m.projects) {
		end = len(m.projects)
	}
	if m.offset > 0 {
		b.WriteString("  " + theme.DimStyle.Render("↑ mais projetos...") + "\n")
	}
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderRow(i) + "\n")
	}
	if end < len(m.projects) {
		b.WriteString("  " + theme.DimStyle.Render("↓ mais projetos...") + "\n")
	}
	return b.String()
}

func (m Model) renderRow(i int) string {
	ref := m.projects[i]
	label := ref.Project.Alias + " — " + ref.Project.Desc
	tag := ""
	if ref.Project.GitHubPath != "" {
		tag = theme.DimStyle.Render(" [" + ref.Project.GitHubPath + "]")
	} else {
		tag = theme.WarningStyle.Render(" [sem github_path]")
	}
	if i == m.cursor {
		return "  " + theme.TitleSelectedStyle.Render("▶ "+label) + tag
	}
	return "  " + theme.TextStyle.Render("  "+label) + tag
}

func (m Model) visibleRows() int {
	v := m.height - 8
	if v < 3 {
		return 3
	}
	return v
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

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) GoBack() bool                   { return m.goBack }
func (m *Model) Selected() *githubapp.ProjectRef { return m.selected }
func (m *Model) ClearSelection()                { m.selected = nil }
