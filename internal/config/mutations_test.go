package config

import "testing"

func TestUpdateCategory(t *testing.T) {
	cfg := &Config{
		Categories: []Category{{ID: "core", Name: "Core", Sub: "Old"}},
	}

	if err := cfg.UpdateCategory(0, "Platform", "New"); err != nil {
		t.Fatalf("UpdateCategory() error = %v", err)
	}

	if cfg.Categories[0].Name != "Platform" || cfg.Categories[0].Sub != "New" {
		t.Fatalf("UpdateCategory() did not update category: %#v", cfg.Categories[0])
	}
}

func TestAddProjectToExistingCategory(t *testing.T) {
	cfg := &Config{
		Categories: []Category{{ID: "core", Name: "Core"}},
	}

	project := Project{Alias: "repo-a"}
	if err := cfg.AddProjectToCategory("core", project, nil); err != nil {
		t.Fatalf("AddProjectToCategory() error = %v", err)
	}

	if len(cfg.Categories[0].Projects) != 1 {
		t.Fatalf("len(Projects) = %d, want 1", len(cfg.Categories[0].Projects))
	}
}

func TestAddProjectToNewCategory(t *testing.T) {
	cfg := &Config{}
	project := Project{Alias: "repo-a"}
	newCategory := &Category{Name: "Core", Sub: "APIs"}

	if err := cfg.AddProjectToCategory("core", project, newCategory); err != nil {
		t.Fatalf("AddProjectToCategory() error = %v", err)
	}

	if len(cfg.Categories) != 1 {
		t.Fatalf("len(Categories) = %d, want 1", len(cfg.Categories))
	}
	if cfg.Categories[0].ID != "core" {
		t.Fatalf("Category ID = %q, want core", cfg.Categories[0].ID)
	}
	if len(cfg.Categories[0].Projects) != 1 || cfg.Categories[0].Projects[0].Alias != "repo-a" {
		t.Fatalf("Projects = %#v, want repo-a", cfg.Categories[0].Projects)
	}
}

func TestRemoveProject(t *testing.T) {
	cfg := &Config{
		Categories: []Category{
			{
				ID: "core",
				Projects: []Project{
					{Alias: "repo-a"},
					{Alias: "repo-b"},
				},
			},
		},
	}

	if err := cfg.RemoveProject(0, 0); err != nil {
		t.Fatalf("RemoveProject() error = %v", err)
	}

	if len(cfg.Categories[0].Projects) != 1 || cfg.Categories[0].Projects[0].Alias != "repo-b" {
		t.Fatalf("Projects = %#v, want only repo-b", cfg.Categories[0].Projects)
	}
}

func TestUpdateProjectGroupAndGitLabPath(t *testing.T) {
	cfg := &Config{
		Categories: []Category{
			{
				ID:       "core",
				Projects: []Project{{Alias: "repo-a"}},
			},
		},
	}

	if err := cfg.UpdateProjectGroup(0, 0, "backend"); err != nil {
		t.Fatalf("UpdateProjectGroup() error = %v", err)
	}
	if err := cfg.SetProjectGitLabPath(0, 0, "group/repo-a"); err != nil {
		t.Fatalf("SetProjectGitLabPath() error = %v", err)
	}

	project := cfg.Categories[0].Projects[0]
	if project.Group != "backend" {
		t.Fatalf("Group = %q, want backend", project.Group)
	}
	if project.GitLabPath != "group/repo-a" {
		t.Fatalf("GitLabPath = %q, want group/repo-a", project.GitLabPath)
	}
}

// seedConfig returns a fresh cfg with one category containing two projects.
// Used by boundary tests below. Index boundaries: catIdx valid = 0,
// projIdx valid = 0 or 1, out-of-range = 2.
func seedConfig() *Config {
	return &Config{
		Categories: []Category{
			{
				ID:       "core",
				Projects: []Project{{Alias: "a"}, {Alias: "b"}},
			},
		},
	}
}

func TestUpdateCategory_IndexErrors(t *testing.T) {
	tests := []struct {
		name   string
		catIdx int
	}{
		{"negative", -1},
		{"equals len (boundary)", 1},
		{"beyond len", 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedConfig()
			if err := cfg.UpdateCategory(tt.catIdx, "x", "y"); err == nil {
				t.Errorf("UpdateCategory(%d) err = nil, want error", tt.catIdx)
			}
		})
	}
}

func TestAddProjectToCategory_NotFoundNoNewCategory(t *testing.T) {
	cfg := seedConfig()
	err := cfg.AddProjectToCategory("missing", Project{Alias: "z"}, nil)
	if err == nil {
		t.Fatal("expected error when category not found and newCategory=nil")
	}
}

func TestAddProjectToNewCategory_PreservesExistingProjects(t *testing.T) {
	cfg := &Config{}
	seed := &Category{
		Name:     "Core",
		Projects: []Project{{Alias: "seed1"}, {Alias: "seed2"}},
	}
	if err := cfg.AddProjectToCategory("core", Project{Alias: "added"}, seed); err != nil {
		t.Fatalf("AddProjectToCategory() error = %v", err)
	}
	projects := cfg.Categories[0].Projects
	if len(projects) != 3 {
		t.Fatalf("len(Projects) = %d, want 3", len(projects))
	}
	// Confirm seed was copied, not mutated
	if len(seed.Projects) != 2 {
		t.Fatalf("seed.Projects mutated: len = %d, want 2", len(seed.Projects))
	}
}

func TestRemoveProject_IndexErrors(t *testing.T) {
	tests := []struct {
		name             string
		catIdx, projIdx  int
	}{
		{"negative catIdx", -1, 0},
		{"catIdx equals len (boundary)", 1, 0},
		{"negative projIdx", 0, -1},
		{"projIdx equals len (boundary)", 0, 2},
		{"projIdx beyond len", 0, 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedConfig()
			if err := cfg.RemoveProject(tt.catIdx, tt.projIdx); err == nil {
				t.Errorf("RemoveProject(%d,%d) err = nil, want error", tt.catIdx, tt.projIdx)
			}
		})
	}
}

func TestRemoveProject_LastIndex(t *testing.T) {
	cfg := seedConfig()
	// len-1 is valid, len is not — verifies `>=` not `>`
	if err := cfg.RemoveProject(0, 1); err != nil {
		t.Fatalf("RemoveProject(0,1) error = %v", err)
	}
	if len(cfg.Categories[0].Projects) != 1 {
		t.Fatalf("Projects after remove = %d, want 1", len(cfg.Categories[0].Projects))
	}
	if cfg.Categories[0].Projects[0].Alias != "a" {
		t.Fatalf("remaining project = %q, want a", cfg.Categories[0].Projects[0].Alias)
	}
}

func TestUpdateProjectGroup_IndexErrors(t *testing.T) {
	tests := []struct {
		name            string
		catIdx, projIdx int
	}{
		{"negative catIdx", -1, 0},
		{"catIdx equals len (boundary)", 1, 0},
		{"negative projIdx", 0, -1},
		{"projIdx equals len (boundary)", 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedConfig()
			if err := cfg.UpdateProjectGroup(tt.catIdx, tt.projIdx, "g"); err == nil {
				t.Errorf("UpdateProjectGroup(%d,%d) err = nil, want error", tt.catIdx, tt.projIdx)
			}
		})
	}
}

func TestSetProjectGitLabPath_IndexErrors(t *testing.T) {
	tests := []struct {
		name            string
		catIdx, projIdx int
	}{
		{"negative catIdx", -1, 0},
		{"catIdx equals len (boundary)", 1, 0},
		{"negative projIdx", 0, -1},
		{"projIdx equals len (boundary)", 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedConfig()
			if err := cfg.SetProjectGitLabPath(tt.catIdx, tt.projIdx, "g/p"); err == nil {
				t.Errorf("SetProjectGitLabPath(%d,%d) err = nil, want error", tt.catIdx, tt.projIdx)
			}
		})
	}
}

func TestSetProjectGitLabPath_LastIndex(t *testing.T) {
	cfg := seedConfig()
	if err := cfg.SetProjectGitLabPath(0, 1, "group/b"); err != nil {
		t.Fatalf("SetProjectGitLabPath(0,1) error = %v", err)
	}
	if cfg.Categories[0].Projects[1].GitLabPath != "group/b" {
		t.Fatalf("GitLabPath = %q, want group/b", cfg.Categories[0].Projects[1].GitLabPath)
	}
}

func TestSetProjectGitHubPath(t *testing.T) {
	cfg := seedConfig()
	if err := cfg.SetProjectGitHubPath(0, 0, "user/a"); err != nil {
		t.Fatalf("SetProjectGitHubPath() error = %v", err)
	}
	if got := cfg.Categories[0].Projects[0].GitHubPath; got != "user/a" {
		t.Fatalf("GitHubPath = %q, want user/a", got)
	}
}

func TestSetProjectGitHubPath_IndexErrors(t *testing.T) {
	tests := []struct {
		name            string
		catIdx, projIdx int
	}{
		{"negative catIdx", -1, 0},
		{"catIdx equals len", 1, 0},
		{"negative projIdx", 0, -1},
		{"projIdx equals len", 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := seedConfig()
			if err := cfg.SetProjectGitHubPath(tt.catIdx, tt.projIdx, "u/p"); err == nil {
				t.Errorf("SetProjectGitHubPath(%d,%d) err = nil, want error", tt.catIdx, tt.projIdx)
			}
		})
	}
}
