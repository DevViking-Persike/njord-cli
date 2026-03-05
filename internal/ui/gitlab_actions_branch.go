package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Create Branch (Jira flow) ---

// Step 1: Select Jira sigla

func (m GitLabActionsModel) viewCreateBranchSigla() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Selecione a sigla Jira:") + "\n\n")
	start, end := m.scroll.Bounds(len(jiraTeams), glActionsChromeLines)
	components.RenderScrollUp(&b, start)
	for i := start; i < end; i++ {
		t := jiraTeams[i]
		label := fmt.Sprintf("%s — %s (%s)", t.Sigla, t.Name, t.Code)
		if i == m.cursor {
			b.WriteString("  " + theme.GitLabTitleSelectedStyle.Render("▶ "+label) + "\n")
		} else {
			b.WriteString("  " + theme.TextStyle.Render("  "+label) + "\n")
		}
	}
	components.RenderScrollDown(&b, end, len(jiraTeams))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchSigla(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(jiraTeams)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, glActionsChromeLines)
		}
	case "esc":
		m.screen = glActionsMenu
		m.cursor = 0
	case "enter":
		m.branchSiglaIdx = m.cursor
		m.inputBuf = ""
		m.message = ""
		m.screen = glActionsCreateBranchNumber
	}
	return m, nil
}

// Step 2: Enter ticket number

func (m GitLabActionsModel) viewCreateBranchNumber() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	preview := fmt.Sprintf("feature/%s-___-%s-...", t.Sigla, t.Code)
	b.WriteString("  " + theme.DimStyle.Render(preview) + "\n\n")
	b.WriteString("  " + theme.TextStyle.Render("Número do ticket Jira:") + "\n\n")
	b.WriteString("  " + theme.TitleStyle.Render(t.Sigla+"-") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
	b.WriteString(components.RenderError(m.message))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchNumber(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = glActionsCreateBranchSigla
		m.cursor = m.branchSiglaIdx
		m.message = ""
	case "enter":
		val := strings.TrimSpace(m.inputBuf)
		if val == "" {
			m.message = "Número não pode ser vazio"
			return m, nil
		}
		m.branchNumber = val
		m.inputBuf = ""
		m.message = ""
		m.branchTypeIdx = 0
		m.cursor = 0
		m.screen = glActionsCreateBranchType
	default:
		if newBuf, handled := components.TextInput(msg, m.inputBuf, components.DigitsOnly); handled {
			m.inputBuf = newBuf
		}
	}
	return m, nil
}

// Step 3: Select type (delivery/subtask)

func (m GitLabActionsModel) viewCreateBranchType() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	preview := fmt.Sprintf("feature/%s-%s-%s-___-...", t.Sigla, m.branchNumber, t.Code)
	b.WriteString("  " + theme.DimStyle.Render(preview) + "\n\n")
	b.WriteString("  " + theme.TextStyle.Render("Tipo da branch:") + "\n\n")
	types := []string{"delivery", "subtask"}
	selectedFn := func(s string) string { return theme.GitLabTitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }
	b.WriteString(components.RenderMenuOptions(types, m.cursor, selectedFn, normalFn))
	b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchType(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 2); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = glActionsCreateBranchNumber
		m.inputBuf = m.branchNumber
	case "enter":
		m.branchTypeIdx = m.cursor
		m.inputBuf = ""
		m.message = ""
		m.screen = glActionsCreateBranchDesc
	}
	return m, nil
}

// Step 4: Enter description

func (m GitLabActionsModel) viewCreateBranchDesc() string {
	var b strings.Builder
	t := jiraTeams[m.branchSiglaIdx]
	branchTypes := []string{"delivery", "subtask"}
	prefix := fmt.Sprintf("feature/%s-%s-%s-%s", t.Sigla, m.branchNumber, t.Code, branchTypes[m.branchTypeIdx])

	normalized := normalizeBranchDesc(m.inputBuf)
	if normalized != "" {
		b.WriteString("  " + theme.DimStyle.Render(prefix+"-"+normalized) + "\n\n")
	} else {
		b.WriteString("  " + theme.DimStyle.Render(prefix+"-...") + "\n\n")
	}

	b.WriteString("  " + theme.TextStyle.Render("Descrição da branch:") + "\n\n")
	b.WriteString(components.RenderTextInput(m.inputBuf, ""))
	b.WriteString(components.RenderError(m.message))
	b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc voltar"))
	return b.String()
}

func (m GitLabActionsModel) handleCreateBranchDesc(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = glActionsCreateBranchType
		m.cursor = m.branchTypeIdx
	case "enter":
		desc := normalizeBranchDesc(m.inputBuf)
		if desc == "" {
			m.message = "Descrição não pode ser vazia"
			return m, nil
		}
		t := jiraTeams[m.branchSiglaIdx]
		branchTypes := []string{"delivery", "subtask"}
		m.branchName = fmt.Sprintf("feature/%s-%s-%s-%s-%s",
			t.Sigla, m.branchNumber, t.Code, branchTypes[m.branchTypeIdx], desc)
		m.inputBuf = ""
		m.message = ""
		m.cursor = 0
		m.scroll.Offset = 0
		m.loading = true
		m.screen = glActionsCreateBranchBase
		return m, m.fetchBranches()
	default:
		if newBuf, handled := components.TextInput(msg, m.inputBuf, nil); handled {
			m.inputBuf = newBuf
		}
	}
	return m, nil
}

// normalizeBranchDesc normalizes branch description: lowercase, no accents, spaces to dashes
func normalizeBranchDesc(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = removeAccents(s)
	s = strings.ReplaceAll(s, " ", "-")
	var buf strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			buf.WriteRune(r)
		}
	}
	// Collapse multiple dashes
	result := buf.String()
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	return strings.Trim(result, "-")
}

func removeAccents(s string) string {
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
	}
	var buf strings.Builder
	for _, r := range s {
		if rep, ok := replacements[r]; ok {
			buf.WriteRune(rep)
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// Step 5: Select base branch

func (m GitLabActionsModel) viewCreateBranchBase() string {
	var b strings.Builder
	b.WriteString("  " + theme.TextStyle.Render("Nova branch: ") + theme.GitLabTitleStyle.Render(m.branchName) + "\n\n")
	if len(m.branches) == 0 {
		b.WriteString("  " + theme.DimStyle.Render("Carregando branches...") + "\n")
	} else {
		b.WriteString("  " + theme.TextStyle.Render("Selecione a branch base:") + "\n\n")
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

func (m GitLabActionsModel) handleCreateBranchBase(msg tea.KeyMsg) (GitLabActionsModel, tea.Cmd) {
	if len(m.branches) == 0 {
		if msg.String() == "esc" {
			m.screen = glActionsCreateBranchDesc
			m.inputBuf = ""
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
		m.screen = glActionsCreateBranchDesc
		m.inputBuf = ""
	case "enter":
		baseBranch := m.branches[m.cursor].Name
		m.loading = true
		return m, m.createBranch(m.branchName, baseBranch)
	}
	return m, nil
}
