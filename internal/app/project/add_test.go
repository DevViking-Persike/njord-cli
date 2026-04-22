package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestRelativeProjectPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	cfg := &config.Config{
		Settings: config.Settings{
			ProjectsBase: "~/Avita",
			PersonalBase: "~/Persike",
		},
	}

	tests := []struct {
		name      string
		clonePath string
		want      string
	}{
		{
			name:      "projects base",
			clonePath: filepath.Join(home, "Avita", "repo-a"),
			want:      "repo-a",
		},
		{
			name:      "personal base",
			clonePath: filepath.Join(home, "Persike", "tools", "njord-cli"),
			want:      filepath.Join("Persike", "tools", "njord-cli"),
		},
		{
			name:      "external absolute path",
			clonePath: "/tmp/repo-a",
			want:      "/tmp/repo-a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RelativeProjectPath(cfg, tt.clonePath)
			if err != nil {
				t.Fatalf("RelativeProjectPath() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("RelativeProjectPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSaveProjectToExistingCategory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")

	cfg := &config.Config{
		Settings: config.Settings{
			ProjectsBase: filepath.Join(tempDir, "Avita"),
			PersonalBase: filepath.Join(tempDir, "Persike"),
		},
		Categories: []config.Category{
			{ID: "core", Name: "Core"},
		},
	}

	err := SaveProject(cfg, configPath, AddProjectInput{
		ClonePath:   filepath.Join(tempDir, "Avita", "repo-a"),
		Alias:       "repo-a",
		Description: "Repo A",
		CategoryID:  "core",
		Group:       "backend",
	})
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	if len(cfg.Categories[0].Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(cfg.Categories[0].Projects))
	}
	if cfg.Categories[0].Projects[0].Path != "repo-a" {
		t.Fatalf("Project.Path = %q, want repo-a", cfg.Categories[0].Projects[0].Path)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not written: %v", err)
	}
}

func TestSaveProjectToNewCategory(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	cfg := &config.Config{
		Settings: config.Settings{
			ProjectsBase: filepath.Join(tempDir, "Avita"),
			PersonalBase: "~/Persike",
		},
	}

	err = SaveProject(cfg, configPath, AddProjectInput{
		ClonePath:       filepath.Join(home, "Persike", "repo-a"),
		Alias:           "repo-a",
		Description:     "Repo A",
		CategoryID:      "tools",
		NewCategoryName: "Tools",
		NewCategorySub:  "Utilities",
	})
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	if len(cfg.Categories) != 1 {
		t.Fatalf("len(Categories) = %d, want 1", len(cfg.Categories))
	}
	if cfg.Categories[0].Projects[0].Path != filepath.Join("Persike", "repo-a") {
		t.Fatalf("Project.Path = %q, want Persike/repo-a", cfg.Categories[0].Projects[0].Path)
	}
}
