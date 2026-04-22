// Package shared holds UI helpers used by multiple screens in internal/ui.
// Put here what would otherwise force cross-subpackage stutter (NjordTitle,
// TimeAgo, card layout constants).
package shared

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Card layout constants. Subscreens that render a grid of cards reuse these
// to keep the visual rhythm consistent across screens.
const (
	MinCardWidth   = 30 // minimum card content width (without borders)
	BorderOverhead = 2  // left + right borders
	CardHeight     = 6  // default card height used for scroll/visibility math
)

// NjordTitle renders the branded "ᚾ N J O R D" rune+text used at the top of
// every top-level screen.
func NjordTitle() string {
	runeStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff9800"))
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#daa520"))
	return "  " + runeStyle.Render("ᚾ") + " " + nameStyle.Render("N J O R D")
}

// TimeAgo formats a past time.Time as a human-readable pt-BR string
// ("agora", "5m atrás", "3h atrás", "ontem", "4d atrás").
func TimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "agora"
	case d < time.Hour:
		return fmt.Sprintf("%dm atrás", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh atrás", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "ontem"
		}
		return fmt.Sprintf("%dd atrás", days)
	}
}
