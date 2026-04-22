package project

import (
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestBuildProjectCommandUsesResolvedPathAndEditor(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{
			Editor:       "cursor",
			ProjectsBase: "/workspace",
		},
	}

	command, err := BuildProjectCommand(cfg, config.Project{Path: "repo-a"})
	if err != nil {
		t.Fatalf("BuildProjectCommand() error = %v", err)
	}

	want := "cd -- '/workspace/repo-a' && cursor ."
	if command != want {
		t.Fatalf("BuildProjectCommand() = %q, want %q", command, want)
	}
}

func TestBuildProjectCommandHandlesQuotesInPath(t *testing.T) {
	cfg := &config.Config{
		Settings: config.Settings{
			Editor:       "code",
			ProjectsBase: "/workspace",
		},
	}

	command, err := BuildProjectCommand(cfg, config.Project{Path: "repo's"})
	if err != nil {
		t.Fatalf("BuildProjectCommand() error = %v", err)
	}

	want := "cd -- '/workspace/repo'\\''s' && code ."
	if command != want {
		t.Fatalf("BuildProjectCommand() = %q, want %q", command, want)
	}
}

func TestBuildProjectCommandHandlesRDPShortcut(t *testing.T) {
	cfg := &config.Config{}

	command, err := BuildProjectCommand(cfg, config.Project{Path: "@rdp"})
	if err != nil {
		t.Fatalf("BuildProjectCommand() error = %v", err)
	}

	if command != rdpCommand {
		t.Fatalf("BuildProjectCommand() = %q, want rdp command", command)
	}
}
