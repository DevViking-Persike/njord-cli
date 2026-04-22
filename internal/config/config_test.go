package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	got := ExpandPath("~/workspace")
	want := filepath.Join(home, "workspace")
	if got != want {
		t.Fatalf("ExpandPath() = %q, want %q", got, want)
	}
}

func TestResolveProjectPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	cfg := &Config{
		Settings: Settings{
			ProjectsBase: "~/Avita",
			PersonalBase: "~/Persike",
		},
	}

	tests := []struct {
		name    string
		project Project
		want    string
	}{
		{
			name:    "special alias path",
			project: Project{Path: "@rdp"},
			want:    "@rdp",
		},
		{
			name:    "absolute path",
			project: Project{Path: "/tmp/repo"},
			want:    "/tmp/repo",
		},
		{
			name:    "personal repo",
			project: Project{Path: "Persike/tools/njord-cli"},
			want:    filepath.Join(home, "Persike/tools/njord-cli"),
		},
		{
			name:    "env repo",
			project: Project{Path: "env/service-a"},
			want:    filepath.Join(home, "Avita", "env/service-a"),
		},
		{
			name:    "default repo",
			project: Project{Path: "service-a"},
			want:    filepath.Join(home, "Avita", "service-a"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.ResolveProjectPath(tt.project)
			if got != tt.want {
				t.Fatalf("ResolveProjectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveDockerComposePath(t *testing.T) {
	tempDir := t.TempDir()
	stackDir := filepath.Join(tempDir, "stack-a")
	if err := os.MkdirAll(stackDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	composePath := filepath.Join(stackDir, "compose.yaml")
	if err := os.WriteFile(composePath, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &Config{
		Settings: Settings{
			ProjectsBase: tempDir,
		},
	}

	got := cfg.ResolveDockerComposePath(DockerStack{Path: "stack-a"})
	want := composePath
	if got != want {
		t.Fatalf("ResolveDockerComposePath() = %q, want %q", got, want)
	}
}
