package gitlab

import (
	"strings"

	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
)

// LocalModel lista os projetos cuja pasta existe no disco, com a tag do host
// detectado ao lado (GL/GH/—). Enter seleciona pra abrir no editor via fluxo
// normal de projeto (BuildProjectCommand).
type LocalModel struct {
	cfg      *config.Config
	projects []githubapp.ProjectRef
	cursor   int
	offset   int
	width    int
	height   int
	selected *githubapp.ProjectRef
	goBack   bool
}

func NewLocalModel(cfg *config.Config) LocalModel {
	return LocalModel{
		cfg:      cfg,
		projects: githubapp.FilterLocal(cfg),
	}
}

func (m LocalModel) Init() tea.Cmd { return nil }

func (m LocalModel) Update(msg tea.Msg) (LocalModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(key)
	}
	return m, nil
}

func (m LocalModel) handleKey(msg tea.KeyMsg) (LocalModel, tea.Cmd) {
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

func (m LocalModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")
	header := theme.SettingsTitleSelectedStyle.Render("  💾 Local — arquivos no PC")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	if len(m.projects) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhuma pasta de projeto no disco.") + "\n")
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

func (m LocalModel) renderRow(i int) string {
	ref := m.projects[i]
	hostTag := renderHostTag(githubapp.DetectHost(ref.Project))
	label := ref.Project.Alias + " — " + ref.Project.Desc
	if i == m.cursor {
		return "  " + hostTag + " " + theme.TitleSelectedStyle.Render("▶ "+label)
	}
	return "  " + hostTag + " " + theme.TextStyle.Render("  "+label)
}

func renderHostTag(h githubapp.Host) string {
	switch h {
	case githubapp.HostGitLab:
		return theme.GitLabTitleStyle.Render("GL")
	case githubapp.HostGitHub:
		return theme.TitleStyle.Render("GH")
	default:
		return theme.DimStyle.Render("—")
	}
}

func (m LocalModel) visibleRows() int {
	v := m.height - 8
	if v < 3 {
		return 3
	}
	return v
}

func (m *LocalModel) ensureVisible() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func (m *LocalModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *LocalModel) GoBack() bool { return m.goBack }

// Selected devolve o projeto escolhido (ou nil), pra ser aberto via
// BuildProjectCommand pelo app layer.
func (m *LocalModel) Selected() *config.Project {
	if m.selected == nil {
		return nil
	}
	p := m.selected.Project
	return &p
}

func (m *LocalModel) ClearSelection() { m.selected = nil }
