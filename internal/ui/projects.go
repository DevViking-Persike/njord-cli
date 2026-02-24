package ui

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

// projectDelegate renders each item in the list.
type projectDelegate struct{}

func (d projectDelegate) Height() int                             { return 2 }
func (d projectDelegate) Spacing() int                            { return 0 }
func (d projectDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d projectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
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

type ProjectsModel struct {
	cfg      *config.Config
	catID    string
	list     list.Model
	projects []config.Project
	selected *config.Project
	goBack   bool
	width    int
	height   int
}

func NewProjectsModel(cfg *config.Config, catID string, projects []config.Project) ProjectsModel {
	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = projectItem{project: p}
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

	return ProjectsModel{
		cfg:      cfg,
		catID:    catID,
		list:     l,
		projects: projects,
	}
}

func (m ProjectsModel) Init() tea.Cmd {
	return nil
}

func (m ProjectsModel) Update(msg tea.Msg) (ProjectsModel, tea.Cmd) {
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
	return m, cmd
}

func (m ProjectsModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.list.View())
	return b.String()
}

func (m *ProjectsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.list.SetSize(w-2, h-4)
}

func (m *ProjectsModel) Selected() *config.Project {
	return m.selected
}

func (m *ProjectsModel) GoBack() bool {
	return m.goBack
}
