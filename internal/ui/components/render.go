package components

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
)

// RenderMenuOptions renders a list of options with cursor highlight.
func RenderMenuOptions(options []string, cursor int, selectedStyle, normalStyle func(string) string) string {
	var b strings.Builder
	for i, opt := range options {
		if i == cursor {
			b.WriteString("  " + selectedStyle("\u25b6 "+opt) + "\n")
		} else {
			b.WriteString("  " + normalStyle("  "+opt) + "\n")
		}
	}
	return b.String()
}

// RenderTextInput renders "> buf█" style input prompt.
func RenderTextInput(buf, prefix string) string {
	return "  " + theme.TitleStyle.Render("> ") + prefix + buf + theme.DimStyle.Render("\u2588") + "\n"
}

// RenderScrollUp renders the up scroll indicator.
func RenderScrollUp(b *strings.Builder, start int) {
	if start > 0 {
		b.WriteString("  " + theme.DimStyle.Render("  \u2191 mais...") + "\n")
	}
}

// RenderScrollDown renders the down scroll indicator.
func RenderScrollDown(b *strings.Builder, end, total int) {
	if end < total {
		b.WriteString("  " + theme.DimStyle.Render("  \u2193 mais...") + "\n")
	}
}

// RenderMessage renders success or error message.
func RenderMessage(msg, msgType string) string {
	if msgType == "ok" {
		return "  " + theme.SuccessStyle.Render("\u2713 "+msg) + "\n"
	}
	return "  " + theme.ErrorStyle.Render("\u2717 "+msg) + "\n"
}

// RenderError renders an inline error if non-empty.
func RenderError(msg string) string {
	if msg == "" {
		return ""
	}
	return "\n  " + theme.ErrorStyle.Render(msg) + "\n"
}

// SaveConfig wraps config.Save with standard error/success message pattern.
func SaveConfig(cfg *config.Config, path, successMsg string) (message, msgType string) {
	if err := config.Save(cfg, path); err != nil {
		return fmt.Sprintf("Erro ao salvar: %s", err), "error"
	}
	return successMsg, "ok"
}
