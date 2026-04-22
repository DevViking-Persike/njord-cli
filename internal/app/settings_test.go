package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestUpdateCategoryPersists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	cfg := &config.Config{
		Categories: []config.Category{{ID: "core", Name: "Core", Sub: "Old"}},
	}

	if err := UpdateCategory(cfg, configPath, 0, "Platform", "New"); err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}

	if cfg.Categories[0].Name != "Platform" || cfg.Categories[0].Sub != "New" {
		t.Fatalf("category = %#v, want updated values", cfg.Categories[0])
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config not persisted: %v", err)
	}
}

func TestUpdateProjectGroupPersists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	cfg := &config.Config{
		Categories: []config.Category{
			{ID: "core", Projects: []config.Project{{Alias: "repo-a"}}},
		},
	}

	if err := UpdateProjectGroup(cfg, configPath, 0, 0, "backend"); err != nil {
		t.Fatalf("UpdateProjectGroup() error = %v", err)
	}

	if got := cfg.Categories[0].Projects[0].Group; got != "backend" {
		t.Fatalf("Group = %q, want backend", got)
	}
}

func TestRemoveProjectPersists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	cfg := &config.Config{
		Categories: []config.Category{
			{ID: "core", Projects: []config.Project{{Alias: "repo-a"}, {Alias: "repo-b"}}},
		},
	}

	if err := RemoveProject(cfg, configPath, 0, 0); err != nil {
		t.Fatalf("RemoveProject() error = %v", err)
	}

	if len(cfg.Categories[0].Projects) != 1 || cfg.Categories[0].Projects[0].Alias != "repo-b" {
		t.Fatalf("Projects = %#v, want only repo-b", cfg.Categories[0].Projects)
	}
}

func TestUpdatePathsAndTokenPersist(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	cfg := &config.Config{}

	if err := UpdateProjectsBase(cfg, configPath, "/work/projects"); err != nil {
		t.Fatalf("UpdateProjectsBase() error = %v", err)
	}
	if err := UpdatePersonalBase(cfg, configPath, "/work/personal"); err != nil {
		t.Fatalf("UpdatePersonalBase() error = %v", err)
	}
	if err := UpdateGitLabToken(cfg, configPath, "glpat-123"); err != nil {
		t.Fatalf("UpdateGitLabToken() error = %v", err)
	}

	if cfg.Settings.ProjectsBase != "/work/projects" {
		t.Fatalf("ProjectsBase = %q, want /work/projects", cfg.Settings.ProjectsBase)
	}
	if cfg.Settings.PersonalBase != "/work/personal" {
		t.Fatalf("PersonalBase = %q, want /work/personal", cfg.Settings.PersonalBase)
	}
	if cfg.GitLab.Token != "glpat-123" {
		t.Fatalf("Token = %q, want glpat-123", cfg.GitLab.Token)
	}
}
