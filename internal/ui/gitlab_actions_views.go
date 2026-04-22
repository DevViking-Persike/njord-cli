package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/shared"
	"github.com/DevViking-Persike/njord-cli/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Merge Requests ---

func (m GitLabActionsModel) viewMRList() string {
	var b strings.Builder
	if len(m.mrs) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhum MR aberto encontrado.") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render(fmt.Sprintf("Merge Requests abertos (%d):", len(m.mrs))) + "\n\n")
		start, end := m.scroll.Bounds(len(m.mrs), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			mr := m.mrs[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "▶ "
			}

			stateStyle := mrStateStyle(mr.State)
			stateTag := stateStyle.Render("[" + mr.State + "]")

			pipelineTag := ""
			if mr.Pipeline != "" {
				ps := pipelineStateStyle(mr.Pipeline)
				pipelineTag = " " + ps.Render("⬤ "+mr.Pipeline)
			}

			title := fmt.Sprintf("!%d %s", mr.IID, mr.Title)
			if i == m.cursor {
				title = theme.GitLabTitleSelectedStyle.Render(prefix+title) + " " + stateTag + pipelineTag
			} else {
				title = theme.TextStyle.Render(prefix+title) + " " + stateTag + pipelineTag
			}
			b.WriteString("  " + title + "\n")

			branchInfo := theme.DimStyle.Render(fmt.Sprintf("     %s → %s  por %s  %s",
				mr.Branch, mr.Target, mr.Author, shared.TimeAgo(mr.CreatedAt)))
			b.WriteString("  " + branchInfo + "\n")
		}
		components.RenderScrollDown(&b, end, len(m.mrs))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleMRList(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.mrs)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	}
	return m, nil
}

// --- Pipelines ---

func (m GitLabActionsModel) viewPipelineList() string {
	var b strings.Builder
	if len(m.pipelines) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Nenhuma pipeline encontrada.") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render(fmt.Sprintf("Pipelines recentes (%d):", len(m.pipelines))) + "\n\n")
		start, end := m.scroll.Bounds(len(m.pipelines), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			p := m.pipelines[i]
			prefix := "  "
			if i == m.cursor {
				prefix = "▶ "
			}

			ps := pipelineStateStyle(p.Status)
			statusTag := ps.Render("⬤ " + p.Status)

			line := fmt.Sprintf("#%d  %s  ref: %s  %s", p.ID, statusTag, p.Ref, shared.TimeAgo(p.CreatedAt))
			if i == m.cursor {
				b.WriteString("  " + theme.GitLabTitleSelectedStyle.Render(prefix) + line + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render(prefix) + line + "\n")
			}
		}
		components.RenderScrollDown(&b, end, len(m.pipelines))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handlePipelineList(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.pipelines)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	}
	return m, nil
}

// --- Trigger Pipeline ---

func (m GitLabActionsModel) viewTriggerBranch() string {
	var b strings.Builder
	if len(m.branches) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Carregando branches...") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render("Selecione a branch para disparar pipeline:") + "\n\n")
		start, end := m.scroll.Bounds(len(m.branches), glActionsChromeLines)
		components.RenderScrollUp(&b, start)
		for i := start; i < end; i++ {
			b.WriteString(m.renderBranchLine(i))
		}
		components.RenderScrollDown(&b, end, len(m.branches))
	}
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleTriggerBranch(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	if len(m.branches) == 0 {
		if msg.String() == "esc" {
			m.screen = glActionsMenu
			m.cursor = 0
		}
		return m, nil
	}
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.branches)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	case "enter":
		m.branchName = m.branches[m.cursor].Name
		m.cursor = 0
		m.screen = glActionsTriggerConfirm
	}
	return m, nil
}

func (m GitLabActionsModel) viewTriggerConfirm() string {
	var b strings.Builder
	b.WriteString("  " + theme.WarningStyle.Render("Disparar pipeline na branch:") + "\n\n")
	b.WriteString("  " + theme.GitLabTitleStyle.Render(m.branchName) + "\n\n")
	opts := []string{"Sim, disparar", "Não, cancelar"}
	selectedFn := func(s string) string { return theme.TitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString(components.RenderMenuOptions(opts, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter confirmar"))
	return b.String()
}

func (m GitLabActionsModel) handleTriggerConfirm(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 2); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = glActionsTriggerBranch
		m.cursor = 0
	case "enter":
		if m.cursor == 0 {
			m.loading = true
			return m, m.triggerPipeline(m.branchName)
		}
		m.screen = glActionsTriggerBranch
		m.cursor = 0
	}
	return m, nil
}
