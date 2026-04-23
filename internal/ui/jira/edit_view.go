package jira

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	"github.com/charmbracelet/lipgloss"
)

func (m EditModel) View() string {
	var b strings.Builder
	b.WriteString(shared.NjordTitle() + "\n\n")

	title := "   Editar card em " + m.project.Name + " (" + m.project.Key + ")"
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(theme.JiraBlue).Render(title) + "\n")
	b.WriteString(theme.DimStyle.Render("  "+strings.Repeat("─", 50)) + "\n\n")

	switch m.step {
	case editPickIssue:
		b.WriteString(m.renderIssuePicker())
	case editSummary:
		b.WriteString(m.renderText("Summary:", "(edite ou deixe vazio pra não mudar)"))
	case editDesc:
		b.WriteString(m.renderText("Descrição:", "(vazio mantém a atual)"))
	case editPickStatus:
		b.WriteString(m.renderStatusPicker())
	case editSubmitting:
		b.WriteString("  " + theme.DimStyle.Render("Salvando..."))
	case editDone:
		b.WriteString(m.renderDone())
	}
	return b.String()
}

func (m EditModel) renderIssuePicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Escolha o card a editar:") + "\n\n")
	if m.loadingBacklog {
		b.WriteString("  " + theme.DimStyle.Render("Carregando backlog..."))
		return b.String()
	}
	if m.backlogErr != "" {
		b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.backlogErr))
		return b.String()
	}
	if len(m.backlog) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhum card no backlog."))
		return b.String()
	}
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
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))
	return b.String()
}

func (m EditModel) renderText(label, hint string) string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render(label) + " " + theme.DimStyle.Render(hint) + "\n\n")
	b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))
	return b.String()
}

func (m EditModel) renderStatusPicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Status:") + "\n\n")
	if m.loadingTrans {
		b.WriteString("  " + theme.DimStyle.Render("Carregando transições..."))
		return b.String()
	}
	if m.transErr != "" {
		b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.transErr))
		return b.String()
	}
	// Primeiro item é sempre "Manter" (não aplica transição).
	opts := append([]string{"Manter status atual"}, transitionLabels(m.transitions)...)
	for i, opt := range opts {
		if i == m.cursor {
			b.WriteString("  " + theme.JiraTitleSelectedStyle.Render("▶ "+opt) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
		}
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter salvar  esc back"))
	return b.String()
}

func transitionLabels(ts []jiraclient.Transition) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, t.Name+" → "+t.ToStatus)
	}
	return out
}

func (m EditModel) renderDone() string {
	var b strings.Builder
	if m.submitErr != "" {
		b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.submitErr) + "\n")
	} else {
		b.WriteString("  " + theme.SuccessStyle.Render("✓ "+m.okKey+" atualizado") + "\n")
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  enter/esc voltar"))
	return b.String()
}
