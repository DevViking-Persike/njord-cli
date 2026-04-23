package gitlab

import (
	"fmt"
	"strings"

	githubapp "github.com/DevViking-Persike/njord-cli/internal/app/github"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HubKind identifica qual card do hub foi escolhido.
type HubKind int

const (
	HubGitLab HubKind = iota
	HubGitHub
	HubLocal
	HubClone
)

// HubModel é a tela de cards (GitLab / GitHub / Local / Clonar) que aparece
// ao entrar no card Repositórios da grid principal. Layout responsivo: quantas
// colunas couberem na largura, o resto vai pra linha de baixo.
type HubModel struct {
	cfg       *config.Config
	cursor    int
	cols      int
	cardWidth int
	width     int
	height    int
	selected  *HubKind
	goBack    bool
}

type hubItem struct {
	kind   HubKind
	title  string
	sub    string
	count  int
	styles hubStyles
}

type hubStyles struct {
	card         lipgloss.Style
	cardSel      lipgloss.Style
	titleStyle   lipgloss.Style
	titleSelStyle lipgloss.Style
}

func NewHubModel(cfg *config.Config) HubModel {
	return HubModel{cfg: cfg, cols: 1, cardWidth: shared.MinCardWidth}
}

func (m HubModel) Init() tea.Cmd { return nil }

func (m HubModel) Update(msg tea.Msg) (HubModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		return m.handleKey(key)
	}
	return m, nil
}

func (m HubModel) handleKey(msg tea.KeyMsg) (HubModel, tea.Cmd) {
	items := m.items()
	cols := m.cols
	if cols < 1 {
		cols = 1
	}
	switch msg.String() {
	case "left", "h":
		if m.cursor%cols > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor%cols < cols-1 && m.cursor+1 < len(items) {
			m.cursor++
		}
	case "up", "k":
		if m.cursor >= cols {
			m.cursor -= cols
		}
	case "down", "j":
		if m.cursor+cols < len(items) {
			m.cursor += cols
		}
	case "enter":
		if m.cursor < len(items) {
			k := items[m.cursor].kind
			m.selected = &k
		}
	case "esc", "q":
		m.goBack = true
	}
	return m, nil
}

func (m HubModel) items() []hubItem {
	gl := githubapp.FilterGitLab(m.cfg)
	gh := githubapp.FilterGitHub(m.cfg)
	local := githubapp.FilterLocal(m.cfg)
	return []hubItem{
		{
			kind:  HubGitLab,
			title: "GitLab",
			sub:   "Pipelines, MRs, branches",
			count: len(gl),
			styles: hubStyles{
				card:          theme.GitLabCardStyle,
				cardSel:       theme.GitLabCardSelectedStyle,
				titleStyle:    theme.GitLabTitleStyle,
				titleSelStyle: theme.GitLabTitleSelectedStyle,
			},
		},
		{
			kind:  HubGitHub,
			title: "GitHub",
			sub:   "Abrir no browser, clonar",
			count: len(gh),
			styles: hubStyles{
				card:          theme.CardStyle,
				cardSel:       theme.CardSelectedStyle,
				titleStyle:    theme.TitleStyle,
				titleSelStyle: theme.TitleSelectedStyle,
			},
		},
		{
			kind:  HubLocal,
			title: "Local",
			sub:   "Arquivos no PC (GL/GH/—)",
			count: len(local),
			styles: hubStyles{
				card:          theme.SettingsCardStyle,
				cardSel:       theme.SettingsCardSelectedStyle,
				titleStyle:    theme.SettingsTitleStyle,
				titleSelStyle: theme.SettingsTitleSelectedStyle,
			},
		},
		{
			kind:  HubClone,
			title: "Clonar novo",
			sub:   "Buscar repos nas APIs",
			count: -1, // -1 sinaliza "sem contador" pro render
			styles: hubStyles{
				card:          theme.AddCardStyle,
				cardSel:       theme.AddCardSelectedStyle,
				titleStyle:    theme.AddTitleStyle,
				titleSelStyle: theme.AddTitleSelectedStyle,
			},
		},
	}
}

func (m HubModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := theme.GitLabTitleSelectedStyle.Render("  ◆ Repositórios — escolha a origem")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	items := m.items()
	cols := m.cols
	if cols < 1 {
		cols = 1
	}
	rows := (len(items) + cols - 1) / cols
	for row := 0; row < rows; row++ {
		rowCards := make([]string, 0, cols)
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx >= len(items) {
				rowCards = append(rowCards, strings.Repeat(" ", m.cardWidth+shared.BorderOverhead))
				continue
			}
			rowCards = append(rowCards, m.renderCard(items[idx], idx == m.cursor))
		}
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowCards...))
		b.WriteString("\n")
	}
	return b.String()
}

func (m HubModel) renderCard(it hubItem, selected bool) string {
	cardStyle := it.styles.card
	titleStyle := it.styles.titleStyle
	subStyle := theme.SubStyle
	countStyle := theme.CountStyle
	if selected {
		cardStyle = it.styles.cardSel
		titleStyle = it.styles.titleSelStyle
		subStyle = theme.SubSelectedStyle
		countStyle = theme.CountSelectedStyle
	}
	name := titleStyle.Render(it.title)
	sub := subStyle.Render(it.sub)
	var content string
	if it.count < 0 {
		content = lipgloss.JoinVertical(lipgloss.Left, name, sub)
	} else {
		count := countStyle.Render(fmt.Sprintf("%d projetos", it.count))
		content = lipgloss.JoinVertical(lipgloss.Left, name, sub, count)
	}
	return cardStyle.Width(m.cardWidth).Render(content)
}

func (m *HubModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.recalcLayout()
	if n := len(m.items()); n > 0 && m.cursor >= n {
		m.cursor = n - 1
	}
}

// recalcLayout define cols/cardWidth a partir da largura disponível, usando o
// mesmo critério da grid principal (MinCardWidth + BorderOverhead). Cap de 4
// colunas porque o hub tem 4 cards — não faz sentido esticar além.
func (m *HubModel) recalcLayout() {
	if m.width <= 0 {
		m.cols = 1
		m.cardWidth = shared.MinCardWidth
		return
	}
	n := len(m.items())
	cols := m.width / (shared.MinCardWidth + shared.BorderOverhead)
	if cols < 1 {
		cols = 1
	}
	if cols > n {
		cols = n
	}
	m.cols = cols
	m.cardWidth = (m.width / cols) - shared.BorderOverhead
	if m.cardWidth < shared.MinCardWidth {
		m.cardWidth = shared.MinCardWidth
	}
}

func (m *HubModel) GoBack() bool { return m.goBack }

// Selected devolve qual card foi escolhido, ou nil.
func (m *HubModel) Selected() *HubKind { return m.selected }

// ClearSelection zera a seleção pra reuso da mesma instância.
func (m *HubModel) ClearSelection() { m.selected = nil }
