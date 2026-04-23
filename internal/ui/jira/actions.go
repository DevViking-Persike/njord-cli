package jira

import (
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpaceAction enumera o que o usuário pode fazer ao entrar num espaço Jira.
type SpaceAction int

const (
	ActionList SpaceAction = iota
	ActionCreate
	ActionEdit
)

// SpaceActionsModel é o menu (3 botões) que aparece depois de escolher um
// espaço Jira na tela de spaces. Mantemos enxuto porque ele só roteia.
type SpaceActionsModel struct {
	project  jiraclient.Project
	cursor   int
	width    int
	height   int
	selected *SpaceAction
	goBack   bool
}

func NewSpaceActionsModel(project jiraclient.Project) SpaceActionsModel {
	return SpaceActionsModel{project: project}
}

func (m SpaceActionsModel) Init() tea.Cmd { return nil }

func (m SpaceActionsModel) Update(msg tea.Msg) (SpaceActionsModel, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 2 {
			m.cursor++
		}
	case "enter":
		a := SpaceAction(m.cursor)
		m.selected = &a
	case "esc", "q":
		m.goBack = true
	}
	return m, nil
}

func (m SpaceActionsModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue).Render(
		"   " + m.project.Name + " (" + m.project.Key + ")")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	opts := []struct {
		label string
		sub   string
	}{
		{"Listar issues", "backlog e suas"},
		{"Criar card", "nova issue no seu nome"},
		{"Editar card", "alterar summary, desc ou status"},
	}
	for i, opt := range opts {
		sub := theme.DimStyle.Render(" — " + opt.sub)
		if i == m.cursor {
			b.WriteString("  " + theme.JiraTitleSelectedStyle.Render("▶ "+opt.label) + sub + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+opt.label) + sub + "\n")
		}
	}
	return b.String()
}

func (m *SpaceActionsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *SpaceActionsModel) GoBack() bool          { return m.goBack }
func (m *SpaceActionsModel) Selected() *SpaceAction { return m.selected }
func (m *SpaceActionsModel) ClearSelection()        { m.selected = nil }
func (m *SpaceActionsModel) Project() jiraclient.Project { return m.project }
