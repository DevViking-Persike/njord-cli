package ui

import "github.com/charmbracelet/lipgloss"

// njordTitle renders the branded header "ᚾ N J O R D" used at the top of
// every top-level screen. Kept here so grid, jira, docker etc. share the
// same visual identity without duplicating styles.
func njordTitle() string {
	runeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff9800"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#daa520"))
	return "  " + runeStyle.Render("ᚾ") + " " + nameStyle.Render("N J O R D")
}
