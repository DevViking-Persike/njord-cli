package project

import (
	"fmt"
	"io"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// projectItem implements list.Item for the bubbles list.
type projectItem struct {
	project config.Project
}

func (i projectItem) Title() string       { return i.project.Alias }
func (i projectItem) Description() string { return i.project.Desc }
func (i projectItem) FilterValue() string {
	return i.project.Alias + " " + i.project.Desc + " " + i.project.Path
}

// groupHeaderItem is a non-selectable separator between groups.
type groupHeaderItem struct {
	name string
}

func (i groupHeaderItem) Title() string       { return i.name }
func (i groupHeaderItem) Description() string { return "" }
func (i groupHeaderItem) FilterValue() string { return "" }

// projectDelegate renders each item in the list.
type projectDelegate struct{}

func (d projectDelegate) Height() int                             { return 2 }
func (d projectDelegate) Spacing() int                            { return 0 }
func (d projectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d projectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	// Render group header
	if gh, ok := item.(groupHeaderItem); ok {
		label := theme.GroupHeaderStyle.Render("── " + strings.ToUpper(gh.name) + " ──")
		fmt.Fprintf(w, "\n%s", label)
		return
	}

	pi, ok := item.(projectItem)
	if !ok {
		return
	}

	selected := index == m.Index()

	alias := pi.project.Alias
	desc := pi.project.Desc
	path := pi.project.Path

	var line1, line2 string
	if selected {
		pointer := theme.TitleSelectedStyle.Render("▶ ")
		aliasStr := theme.TitleSelectedStyle.Render(alias)
		descStr := theme.TextStyle.Render(desc)
		line1 = pointer + aliasStr + "  " + descStr
		line2 = "    " + theme.DimStyle.Render(path)
	} else {
		aliasStr := theme.TitleStyle.Render("  " + alias)
		descStr := theme.DimStyle.Render(desc)
		line1 = aliasStr + "  " + descStr
		line2 = "    " + theme.DimStyle.Render(path)
	}

	fmt.Fprintf(w, "%s\n%s", line1, line2)
}

type Model struct {
	cfg      *config.Config
	catID    string
	list     list.Model
	projects []config.Project
	selected *config.Project
	goBack   bool
	width    int
	height   int
}

func NewModel(cfg *config.Config, catID string, projects []config.Project) Model {
	// Build items with group headers
	groups, byGroup := config.GroupedProjects(projects)

	var items []list.Item
	for _, g := range groups {
		if g != "" {
			items = append(items, groupHeaderItem{name: g})
		}
		for _, p := range byGroup[g] {
			items = append(items, projectItem{project: p})
		}
	}

	// Find category name
	catName := "Todos"
	for _, cat := range cfg.Categories {
		if cat.ID == catID {
			catName = cat.Name
			break
		}
	}

	delegate := projectDelegate{}
	l := list.New(items, delegate, 80, 20)
	l.Title = catName
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	l.Styles.Title = theme.HeaderStyle
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(theme.Title)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(theme.TitleSel)

	m := Model{
		cfg:      cfg,
		catID:    catID,
		list:     l,
		projects: projects,
	}

	// Ensure initial selection is not on a header
	m.skipToProject(1)

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	prevIdx := m.list.Index()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't intercept keys when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "esc":
			m.goBack = true
			return m, nil
		case "enter":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
				m.selected = &item.project
			}
			return m, nil
		case "q":
			if m.list.FilterState() != list.Filtering {
				m.goBack = true
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// Skip headers after navigation
	if m.list.Index() != prevIdx {
		if _, ok := m.list.SelectedItem().(groupHeaderItem); ok {
			if m.list.Index() > prevIdx {
				m.skipToProject(1)
			} else {
				m.skipToProject(-1)
			}
		}
	}

	return m, cmd
}

// skipToProject moves the cursor in the given direction (+1/-1) until a projectItem is found.
func (m *Model) skipToProject(dir int) {
	items := m.list.Items()
	idx := m.list.Index()
	for idx >= 0 && idx < len(items) {
		if _, ok := items[idx].(groupHeaderItem); !ok {
			break
		}
		idx += dir
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(items) {
		idx = len(items) - 1
	}
	m.list.Select(idx)
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.list.View())
	return b.String()
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	if m.cfg != nil {
		m.list.SetSize(w-2, h-4)
	}
}

func (m *Model) Selected() *config.Project {
	return m.selected
}

func (m *Model) GoBack() bool {
	return m.goBack
}
