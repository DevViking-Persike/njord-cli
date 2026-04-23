package jira

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	"github.com/charmbracelet/lipgloss"
)

func (m CreateModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	header := lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue).Render(
		"   Criar card em " + m.project.Name + " (" + m.project.Key + ")")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString(header + "\n" + divider + "\n\n")

	switch m.step {
	case createPickType:
		b.WriteString(m.renderTypePicker())
	case createPickParent:
		b.WriteString(m.renderParentPicker())
	case createSummary:
		b.WriteString(m.renderTextField("Título (summary):", "(obrigatório)"))
	case createDesc:
		b.WriteString(m.renderTextField("Descrição:", "(opcional — enter vazio pula)"))
	case createPickStatus:
		b.WriteString(m.renderStatusPicker())
	case createSubmitting:
		b.WriteString("  " + theme.DimStyle.Render("Criando..."))
	case createDone:
		b.WriteString(m.renderDone())
	}
	return b.String()
}

func (m CreateModel) renderTypePicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Tipo:") + "\n\n")
	for i, opt := range issueTypeOptions {
		if i == m.cursor {
			b.WriteString("  " + theme.JiraTitleSelectedStyle.Render("▶ "+opt) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
		}
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))
	return b.String()
}

func (m CreateModel) renderParentPicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Escolha o card pai (Subtask):") + "\n\n")
	if m.loading {
		b.WriteString("  " + theme.DimStyle.Render("Carregando backlog...") + "\n")
		return b.String()
	}
	if m.loadErr != "" {
		b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.loadErr) + "\n")
		return b.String()
	}
	if len(m.backlog) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhum card no backlog.") + "\n")
		return b.String()
	}
	// Limita a 15 visíveis pra não explodir a tela; picker já ordenado pela API.
	max := 15
	end := max
	if end > len(m.backlog) {
		end = len(m.backlog)
	}
	for i := 0; i < end; i++ {
		iss := m.backlog[i]
		label := fmt.Sprintf("%s  %s", iss.Key, truncate(iss.Summary, 55))
		if i == m.cursor {
			b.WriteString("  " + theme.JiraTitleSelectedStyle.Render("▶ "+label) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+label) + "\n")
		}
	}
	if len(m.backlog) > max {
		b.WriteString("  " + theme.DimStyle.Render(fmt.Sprintf("(mostrando %d de %d)", max, len(m.backlog))) + "\n")
	}
	return b.String()
}

func (m CreateModel) renderTextField(label, hint string) string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render(label) + " " + theme.DimStyle.Render(hint) + "\n\n")
	b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))
	return b.String()
}

func (m CreateModel) renderStatusPicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Status inicial:") + "\n\n")
	for i, opt := range statusCategoryOptions {
		if i == m.cursor {
			b.WriteString("  " + theme.JiraTitleSelectedStyle.Render("▶ "+opt.label) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+opt.label) + "\n")
		}
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter criar  esc back"))
	return b.String()
}

func (m CreateModel) renderDone() string {
	var b strings.Builder
	if m.submitErr != "" {
		b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.submitErr) + "\n")
	} else {
		b.WriteString("  " + theme.SuccessStyle.Render("✓ Criado "+m.createdKey) + "\n")
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  enter/esc voltar"))
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
