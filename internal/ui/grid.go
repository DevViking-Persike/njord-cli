package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type GridItemType int

const (
	GridItemCategory GridItemType = iota
	GridItemDocker
	GridItemGitLab
	GridItemJira
	GridItemAdd
	GridItemSettings
)

const (
	minCardWidth   = 30 // minimum card content width (without borders)
	borderOverhead = 2  // left + right borders
)

// njordTitle is rendered at runtime with styles, not as const art

// RecentPushAlias holds alias + time + approval info for display
type RecentPushAlias struct {
	Alias    string
	Ago      string
	Approval string // "✓", "⏳ 0/1 Rule", or ""
}

// PendingMRAlias holds info for a pending MR in the header box
type PendingMRAlias struct {
	Alias    string // project alias or path
	IID      int64  // MR IID (!123)
	Title    string // MR title (truncated)
	Ago      string // "Xm atrás"
	Approval string // "✓", "⏳ 0/1", ""
}

type GridItem struct {
	Type  GridItemType
	CatID string
	Name  string
	Sub   string
	Count int
}

type GridSelection struct {
	Type  GridItemType
	CatID string
}

type GridModel struct {
	items        []GridItem
	cursor       int
	cols         int
	cardWidth    int
	width        int
	height       int
	selected     *GridSelection
	offset       int // scroll offset in rows
	recentPushes []RecentPushAlias
	pendingMRs   []PendingMRAlias
	pushError    string
	mrsError     string
}

func NewGridModel(cfg *config.Config) GridModel {
	var items []GridItem

	// "Todos" category
	items = append(items, GridItem{
		Type:  GridItemCategory,
		CatID: "*",
		Name:  "Todos",
		Sub:   "Todos os projetos",
		Count: cfg.TotalProjects(),
	})

	// Regular categories
	for _, cat := range cfg.Categories {
		items = append(items, GridItem{
			Type:  GridItemCategory,
			CatID: cat.ID,
			Name:  cat.Name,
			Sub:   cat.Sub,
			Count: len(cat.Projects),
		})
	}

	// Docker card
	items = append(items, GridItem{
		Type:  GridItemDocker,
		Name:  "Docker",
		Sub:   "Gerenciar stacks",
		Count: len(cfg.DockerStacks),
	})

	// GitLab card
	items = append(items, GridItem{
		Type:  GridItemGitLab,
		Name:  "GitLab",
		Sub:   "MRs, Pipelines, Branches",
		Count: cfg.GitLabProjectCount(),
	})

	// Jira card
	if cfg.Jira.Token != "" && cfg.Jira.URL != "" {
		items = append(items, GridItem{
			Type: GridItemJira,
			Name: "Jira",
			Sub:  "Espaços, Tasks, Epics",
		})
	}

	// Add card
	items = append(items, GridItem{
		Type: GridItemAdd,
		Name: "+ Adicionar",
		Sub:  "Novo projeto",
	})

	// Settings card
	items = append(items, GridItem{
		Type: GridItemSettings,
		Name: "Configurações",
		Sub:  "Editar categorias e paths",
	})

	return GridModel{
		items:     items,
		cols:      2,
		cardWidth: 36,
	}
}

func (m *GridModel) SetRecentPushes(pushes []RecentPushAlias) {
	m.recentPushes = pushes
}

func (m *GridModel) SetPendingMRs(mrs []PendingMRAlias) {
	m.pendingMRs = mrs
}

func (m GridModel) Init() tea.Cmd {
	return nil
}

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor >= m.cols {
				m.cursor -= m.cols
			}
			m.ensureVisible()
		case "down", "j":
			if m.cursor+m.cols < len(m.items) {
				m.cursor += m.cols
			}
			m.ensureVisible()
		case "left", "h":
			if m.cursor%m.cols > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor%m.cols < m.cols-1 && m.cursor+1 < len(m.items) {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				m.selected = &GridSelection{
					Type:  item.Type,
					CatID: item.CatID,
				}
			}
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m GridModel) View() string {
	var b strings.Builder

	// Header: [Aprovações recentes] [MRs pendentes] ᚾ N J O R D
	runeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff9800"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#daa520"))
	title := "  " + runeStyle.Render("ᚾ") + " " + nameStyle.Render("N J O R D")

	hasData := len(m.recentPushes) > 0 || len(m.pendingMRs) > 0 || m.pushError != "" || m.mrsError != ""

	if hasData {
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, m.renderApprovalBox(), "  ", title))
	} else {
		b.WriteString(title)
	}
	b.WriteString("\n\n")

	// Render cards in 2-column grid
	rows := (len(m.items) + m.cols - 1) / m.cols
	visibleRows := m.visibleRows()

	startRow := m.offset
	endRow := startRow + visibleRows
	if endRow > rows {
		endRow = rows
	}

	for row := startRow; row < endRow; row++ {
		var rowCards []string
		for col := 0; col < m.cols; col++ {
			idx := row*m.cols + col
			if idx >= len(m.items) {
				// Empty cell
				rowCards = append(rowCards, strings.Repeat(" ", m.cardWidth+borderOverhead))
				continue
			}
			item := m.items[idx]
			selected := idx == m.cursor
			rowCards = append(rowCards, m.renderCard(item, selected))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowCards...))
		b.WriteString("\n")
	}

	// Scroll indicator
	if rows > visibleRows {
		scrollInfo := theme.DimStyle.Render(fmt.Sprintf("  [%d/%d]", m.offset+1, rows-visibleRows+1))
		b.WriteString(scrollInfo)
	}

	return b.String()
}

func (m GridModel) renderApprovalBox() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8b6508")).
		Padding(0, 1).
		Width(40)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Title)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888888"))
	var lines []string

	if m.pushError != "" {
		lines = append(lines, titleStyle.Render("Aprovados"))
		lines = append(lines, theme.DimStyle.Render("erro: "+m.pushError))
	} else if len(m.recentPushes) > 0 {
		lines = append(lines, titleStyle.Render("Aprovados"))
		for _, p := range m.recentPushes {
			icon := ""
			if p.Approval != "" {
				icon = p.Approval + " "
			}
			alias := theme.GitLabTitleStyle.Render(p.Alias)
			ago := theme.DimStyle.Render(" " + p.Ago)
			lines = append(lines, icon+alias+ago)
		}
	}

	if m.mrsError != "" {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, sectionStyle.Render("Pendentes"))
		lines = append(lines, theme.DimStyle.Render("erro: "+m.mrsError))
	} else if len(m.pendingMRs) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, sectionStyle.Render("Pendentes"))
		for _, mr := range m.pendingMRs {
			icon := ""
			if mr.Approval != "" {
				icon = mr.Approval + " "
			}
			alias := theme.GitLabTitleStyle.Render(mr.Alias)
			iid := theme.DimStyle.Render(fmt.Sprintf(" !%d", mr.IID))
			ago := theme.DimStyle.Render(" " + mr.Ago)
			lines = append(lines, icon+alias+iid+ago)
		}
	}

	content := strings.Join(lines, "\n")
	return boxStyle.Render(content)
}

func (m *GridModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.recalcLayout()
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	m.ensureVisible()
}

func (m *GridModel) recalcLayout() {
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
	if maxCols > len(m.items) {
		maxCols = len(m.items)
	}

	m.cols = maxCols
	m.cardWidth = (m.width / m.cols) - borderOverhead
}

func (m *GridModel) Selected() *GridSelection {
	return m.selected
}

func (m *GridModel) ClearSelection() {
	m.selected = nil
}

func (m GridModel) renderCard(item GridItem, selected bool) string {
	var cardStyle, titleStyle, subStyle, countStyle lipgloss.Style

	switch item.Type {
	case GridItemDocker:
		if selected {
			cardStyle = theme.DockerCardSelectedStyle
			titleStyle = theme.DockerTitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.DockerCardStyle
			titleStyle = theme.DockerTitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	case GridItemGitLab:
		if selected {
			cardStyle = theme.GitLabCardSelectedStyle
			titleStyle = theme.GitLabTitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.GitLabCardStyle
			titleStyle = theme.GitLabTitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	case GridItemJira:
		if selected {
			cardStyle = theme.JiraCardSelectedStyle
			titleStyle = theme.JiraTitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.JiraCardStyle
			titleStyle = theme.JiraTitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	case GridItemAdd:
		if selected {
			cardStyle = theme.AddCardSelectedStyle
			titleStyle = theme.AddTitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.AddCardStyle
			titleStyle = theme.AddTitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	case GridItemSettings:
		if selected {
			cardStyle = theme.SettingsCardSelectedStyle
			titleStyle = theme.SettingsTitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.SettingsCardStyle
			titleStyle = theme.SettingsTitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	default:
		if selected {
			cardStyle = theme.CardSelectedStyle
			titleStyle = theme.TitleSelectedStyle
			subStyle = theme.SubSelectedStyle
			countStyle = theme.CountSelectedStyle
		} else {
			cardStyle = theme.CardStyle
			titleStyle = theme.TitleStyle
			subStyle = theme.SubStyle
			countStyle = theme.CountStyle
		}
	}

	name := titleStyle.Render(item.Name)
	sub := subStyle.Render(item.Sub)

	var count string
	if item.Type == GridItemAdd || item.Type == GridItemSettings || item.Type == GridItemJira {
		count = ""
	} else {
		count = countStyle.Render(fmt.Sprintf("%d projetos", item.Count))
	}

	cardStyle = cardStyle.Width(m.cardWidth)
	content := lipgloss.JoinVertical(lipgloss.Left, name, sub, count)
	return cardStyle.Render(content)
}

func (m GridModel) visibleRows() int {
	// Header ~3 lines (title + blank), help ~2 lines
	cardHeight := 6
	available := m.height - 7
	if available < cardHeight {
		return 1
	}
	return available / cardHeight
}

func (m *GridModel) ensureVisible() {
	row := m.cursor / m.cols
	visible := m.visibleRows()

	if row < m.offset {
		m.offset = row
	}
	if row >= m.offset+visible {
		m.offset = row - visible + 1
	}
}
