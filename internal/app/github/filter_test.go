package github

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func fixture(t *testing.T) *config.Config {
	t.Helper()
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "alfa"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &config.Config{
		Settings: config.Settings{ProjectsBase: base, PersonalBase: base},
		Categories: []config.Category{
			{
				ID: "work", Name: "Work",
				Projects: []config.Project{
					{Alias: "alfa", Path: "alfa", GitLabPath: "grp/alfa"},
					{Alias: "beta", Path: "beta-missing", GitLabPath: "grp/beta"},
				},
			},
			{
				ID: PersonalCategoryID, Name: "Pessoal",
				Projects: []config.Project{
					{Alias: "dots", Path: "dots-missing", GitHubPath: "user/dots"},
					{Alias: "legacy", Path: "legacy-missing"},
				},
			},
			{
				ID: "nohost", Name: "Sem host",
				Projects: []config.Project{
					{Alias: "orphan", Path: "alfa"},
				},
			},
		},
	}
}

func TestFilterGitLab(t *testing.T) {
	cfg := fixture(t)
	got := FilterGitLab(cfg)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Project.Alias != "alfa" || got[1].Project.Alias != "beta" {
		t.Fatalf("unexpected order: %+v", got)
	}
	if got[0].CatIdx != 0 || got[0].ProjIdx != 0 {
		t.Fatalf("indices not preserved")
	}
}

func TestFilterGitHub(t *testing.T) {
	cfg := fixture(t)
	got := FilterGitHub(cfg)
	aliases := []string{}
	for _, r := range got {
		aliases = append(aliases, r.Project.Alias)
	}
	// "dots" tem github_path; "legacy" é da categoria pessoal
	if len(got) != 2 || aliases[0] != "dots" || aliases[1] != "legacy" {
		t.Fatalf("unexpected github set: %v", aliases)
	}
}

func TestFilterLocal(t *testing.T) {
	cfg := fixture(t)
	got := FilterLocal(cfg)
	// Só "alfa" (e "orphan", que aponta pro mesmo path "alfa") existem no disco
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestAllProjectRefsNil(t *testing.T) {
	if got := AllProjectRefs(nil); got != nil {
		t.Fatalf("expected nil for nil cfg, got %v", got)
	}
}
