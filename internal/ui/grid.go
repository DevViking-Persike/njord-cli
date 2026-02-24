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
	GridItemAdd
)

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
	items    []GridItem
	cursor   int
	cols     int
	width    int
	height   int
	selected *GridSelection
	offset   int // scroll offset in rows
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

	// Add card
	items = append(items, GridItem{
		Type: GridItemAdd,
		Name: "+ Adicionar",
		Sub:  "Novo projeto",
	})

	return GridModel{
		items: items,
		cols:  2,
	}
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

	// Header
	header := theme.HeaderStyle.Render("  Njord")
	b.WriteString(header)
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
				rowCards = append(rowCards, strings.Repeat(" ", 38))
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

func (m *GridModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *GridModel) Selected() *GridSelection {
	return m.selected
}

func (m *GridModel) ClearSelection() {
	m.selected = nil
}

func (m GridModel) renderCard(item GridItem, selected bool) string {
	var cardStyle, titleStyle lipgloss.Style

	switch item.Type {
	case GridItemDocker:
		if selected {
			cardStyle = theme.DockerCardSelectedStyle
			titleStyle = theme.DockerTitleSelectedStyle
		} else {
			cardStyle = theme.DockerCardStyle
			titleStyle = theme.DockerTitleStyle
		}
	case GridItemAdd:
		if selected {
			cardStyle = theme.AddCardSelectedStyle
			titleStyle = theme.AddTitleSelectedStyle
		} else {
			cardStyle = theme.AddCardStyle
			titleStyle = theme.AddTitleStyle
		}
	default:
		if selected {
			cardStyle = theme.CardSelectedStyle
			titleStyle = theme.TitleSelectedStyle
		} else {
			cardStyle = theme.CardStyle
			titleStyle = theme.TitleStyle
		}
	}

	name := titleStyle.Render(item.Name)
	sub := theme.SubStyle.Render(item.Sub)

	var count string
	if item.Type == GridItemAdd {
		count = ""
	} else {
		count = theme.CountStyle.Render(fmt.Sprintf("%d projetos", item.Count))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, name, sub, count)
	return cardStyle.Render(content)
}

func (m GridModel) visibleRows() int {
	// Each card row is ~5 lines (border + padding + content)
	// Header takes ~3 lines, help takes ~2 lines
	cardHeight := 6
	available := m.height - 5
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
