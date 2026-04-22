package app

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

const rdpCommand = "gnome-terminal --title='TRON - VPS RDP' -- bash -c 'echo \"Conectando à VPS via RDP...\"; echo \"Use o cliente RDP em localhost:3390\"; echo \"\"; cloudflared access rdp --hostname tron.victorpersike.dev.br --url localhost:3390; exec bash' &"

// BuildProjectCommand returns the shell command produced when a project is selected in the TUI.
func BuildProjectCommand(cfg *config.Config, project config.Project) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is required")
	}

	path := cfg.ResolveProjectPath(project)
	if path == "@rdp" {
		return rdpCommand, nil
	}

	editor := strings.TrimSpace(cfg.Settings.Editor)
	if editor == "" {
		editor = "code"
	}

	return "cd -- " + shellQuote(path) + " && " + editor + " .", nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
