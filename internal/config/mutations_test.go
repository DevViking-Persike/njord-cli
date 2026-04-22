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
