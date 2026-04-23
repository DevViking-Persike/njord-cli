package clone

import (
	"fmt"
	"strings"

	cloneapp "github.com/DevViking-Persike/njord-cli/internal/app/clone"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := theme.TitleSelectedStyle.Render("  ⤓ Clonar novo — " + m.scopeLabel())
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	b.WriteString(m.renderSearch() + "\n\n")

	if m.loading {
		b.WriteString(theme.DimStyle.Render("  Carregando..."))
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString(theme.ErrorStyle.Render("  ✗ " + m.loadErr))
		return b.String()
	}
	if len(m.repos) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum repo nesse escopo."))
		return b.String()
	}
	if len(m.filtered) == 0 {
		b.WriteString(theme.DimStyle.Render("  Nenhum repo corresponde à busca."))
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
	b.WriteString("\n" + theme.DimStyle.Render(fmt.Sprintf("  %d de %d", len(m.filtered), len(m.repos))))
	return b.String()
}

// scopeLabel dá o texto descritivo do escopo atual pra ser mostrado no header.
func (m Model) scopeLabel() string {
	switch m.scope.Host {
	case cloneapp.HostGitLab:
		if m.scope.Group != nil {
			return "GitLab · " + m.scope.Group.FullPath
		}
		return "GitLab (todos)"
	case cloneapp.HostGitHub:
		return "GitHub (todos)"
	}
	return ""
}

func (m Model) renderSearch() string {
	label := theme.DimStyle.Render("  Buscar:")
	value := m.search
	if value == "" {
		value = theme.DimStyle.Render("(digite pra filtrar)")
	} else {
		value = lipgloss.NewStyle().Foreground(theme.TitleSel).Render(value + "▌")
	}
	return label + " " + value
}

func (m Model) renderRow(i int) string {
	r := m.filtered[i]
	name := r.FullName
	desc := r.Description
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}
	if i == m.cursor {
		return "  " + theme.TitleSelectedStyle.Render("▶ "+name) + " " + theme.DimStyle.Render(desc)
	}
	return "  " + theme.TextStyle.Render("  "+name) + " " + theme.DimStyle.Render(desc)
}
