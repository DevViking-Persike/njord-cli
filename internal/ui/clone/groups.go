package clone

import (
	"fmt"
	"strings"

	cloneapp "github.com/DevViking-Persike/njord-cli/internal/app/clone"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type groupsLoadedMsg struct {
	groups []cloneapp.Group
	err    error
}

// GroupsModel é o scope picker: escolhe a fonte (GitLab/GitHub) e, no caso do
// GitLab, qual grupo navegar. GitHub não tem grupos aninhados — a listagem
// lá é sempre plana, então vai direto pra Model com Scope{Host:GitHub}.
type GroupsModel struct {
	glClient *gitlabclient.Client

	source cloneapp.Host

	groups   []cloneapp.Group
	loading  bool
	loadErr  string
	filtered []cloneapp.Group

	search string
	cursor int
	offset int
	width  int
	height int

	selected *Scope
	goBack   bool
}

// NewGroupsModel aceita só o client GitLab — GitHub não precisa de picker.
func NewGroupsModel(gl *gitlabclient.Client) GroupsModel {
	return GroupsModel{glClient: gl, source: cloneapp.HostGitLab, loading: true}
}

func (m GroupsModel) Init() tea.Cmd { return m.fetchGroups() }

func (m GroupsModel) fetchGroups() tea.Cmd {
	gl := m.glClient
	return func() tea.Msg {
		if gl == nil {
			return groupsLoadedMsg{err: fmt.Errorf("GitLab token não configurado")}
		}
		items, err := gl.ListMyGroups()
		if err != nil {
			return groupsLoadedMsg{err: err}
		}
		groups := make([]cloneapp.Group, 0, len(items))
		for _, it := range items {
			groups = append(groups, cloneapp.GroupFromGitLab(it))
		}
		return groupsLoadedMsg{groups: groups}
	}
}

func (m GroupsModel) Update(msg tea.Msg) (GroupsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case groupsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.loadErr = ""
		// Colapsa pro "top bucket" (ex.: avitaseg/bill cobre bibliotecas/angular/dotnet).
		m.groups = cloneapp.CollapseToTopBuckets(msg.groups)
		m.recomputeFiltered()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m GroupsModel) handleKey(msg tea.KeyMsg) (GroupsModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.search != "" {
			m.search = ""
			m.recomputeFiltered()
			m.offset = 0
			return m, nil
		}
		m.goBack = true
		return m, nil
	case tea.KeyLeft:
		m.source = cloneapp.HostGitLab
		return m, nil
	case tea.KeyRight:
		m.source = cloneapp.HostGitHub
		return m, nil
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, nil
	case tea.KeyDown:
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.ensureVisible()
		}
		return m, nil
	case tea.KeyEnter:
		return m.triggerEnter()
	case tea.KeyBackspace:
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.recomputeFiltered()
			m.offset = 0
			m.cursor = 0
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		if m.source == cloneapp.HostGitLab {
			m.search += string(msg.Runes)
			m.recomputeFiltered()
			m.offset = 0
			m.cursor = 0
		}
		return m, nil
	}
	return m, nil
}

func (m GroupsModel) triggerEnter() (GroupsModel, tea.Cmd) {
	if m.source == cloneapp.HostGitHub {
		// GitHub pula direto pra Model, sem grupo.
		scope := Scope{Host: cloneapp.HostGitHub}
		m.selected = &scope
		return m, nil
	}
	if m.cursor < len(m.filtered) {
		g := m.filtered[m.cursor]
		scope := Scope{Host: cloneapp.HostGitLab, Group: &g}
		m.selected = &scope
	}
	return m, nil
}

func (m *GroupsModel) recomputeFiltered() {
	m.filtered = cloneapp.FilterGroups(m.groups, m.search)
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		m.offset = 0
	}
}

func (m *GroupsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ensureVisible()
}

func (m *GroupsModel) GoBack() bool      { return m.goBack }
func (m *GroupsModel) Selected() *Scope  { return m.selected }
func (m *GroupsModel) ClearSelection()   { m.selected = nil }

func (m GroupsModel) visibleRows() int {
	v := m.height - 12
	if v < 3 {
		return 3
	}
	return v
}

func (m *GroupsModel) ensureVisible() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// --- View ---

func (m GroupsModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := theme.TitleSelectedStyle.Render("  ⤓ Clonar novo — escolha a fonte")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	b.WriteString(m.renderToggle() + "\n")
	if m.source == cloneapp.HostGitLab {
		b.WriteString(m.renderSearch() + "\n\n")
		b.WriteString(m.renderGitLabBody())
	} else {
		b.WriteString("\n  " + theme.TextStyle.Render("GitHub não tem grupos aninhados.") + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Enter pra listar todos os seus repos.") + "\n")
	}
	return b.String()
}

func (m GroupsModel) renderToggle() string {
	active := lipgloss.NewStyle().Bold(true).Foreground(theme.TitleSel).Background(theme.BgSel).Render
	inactive := theme.DimStyle.Render
	gl := " GitLab "
	gh := " GitHub "
	if m.source == cloneapp.HostGitLab {
		gl = active(gl)
		gh = inactive(gh)
	} else {
		gl = inactive(gl)
		gh = active(gh)
	}
	hint := theme.DimStyle.Render("  ← →")
	return "  " + gl + "/" + gh + hint
}

func (m GroupsModel) renderSearch() string {
	label := theme.DimStyle.Render("  Buscar:")
	value := m.search
	if value == "" {
		value = theme.DimStyle.Render("(digite pra filtrar grupos)")
	} else {
		value = lipgloss.NewStyle().Foreground(theme.TitleSel).Render(value + "▌")
	}
	return label + " " + value
}

func (m GroupsModel) renderGitLabBody() string {
	var b strings.Builder
	if m.loading {
		b.WriteString(theme.DimStyle.Render("  Carregando grupos..."))
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString(theme.ErrorStyle.Render("  ✗ " + m.loadErr))
		return b.String()
	}
	if len(m.groups) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum grupo retornado."))
		return b.String()
	}
	if len(m.filtered) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum grupo corresponde à busca."))
		return b.String()
	}
	visible := m.visibleRows()
	end := m.offset + visible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	if m.offset > 0 {
		b.WriteString(theme.DimStyle.Render("  ↑ mais...") + "\n")
	}
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderRow(i) + "\n")
	}
	if end < len(m.filtered) {
		b.WriteString(theme.DimStyle.Render("  ↓ mais...") + "\n")
	}
	b.WriteString("\n" + theme.DimStyle.Render(fmt.Sprintf("  %d de %d", len(m.filtered), len(m.groups))))
	return b.String()
}

func (m GroupsModel) renderRow(i int) string {
	g := m.filtered[i]
	label := g.FullPath
	if i == m.cursor {
		return "  " + theme.TitleSelectedStyle.Render("▶ "+label)
	}
	return "  " + theme.TextStyle.Render("  "+label)
}
