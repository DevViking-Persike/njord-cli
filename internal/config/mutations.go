package config

import "fmt"

// UpdateCategory updates the display fields of an existing category.
func (cfg *Config) UpdateCategory(catIdx int, name, sub string) error {
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return fmt.Errorf("category index out of range: %d", catIdx)
	}

	cfg.Categories[catIdx].Name = name
	cfg.Categories[catIdx].Sub = sub
	return nil
}

// AddProjectToCategory appends a project to an existing category or creates a new one.
func (cfg *Config) AddProjectToCategory(categoryID string, project Project, newCategory *Category) error {
	for i := range cfg.Categories {
		if cfg.Categories[i].ID == categoryID {
			cfg.Categories[i].Projects = append(cfg.Categories[i].Projects, project)
			return nil
		}
	}

	if newCategory == nil {
		return fmt.Errorf("category not found: %s", categoryID)
	}

	category := *newCategory
	category.ID = categoryID
	category.Projects = append([]Project{}, category.Projects...)
	category.Projects = append(category.Projects, project)
	cfg.Categories = append(cfg.Categories, category)
	return nil
}

// RemoveProject removes a project from a category by index.
func (cfg *Config) RemoveProject(catIdx, projIdx int) error {
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return fmt.Errorf("category index out of range: %d", catIdx)
	}

	projects := cfg.Categories[catIdx].Projects
	if projIdx < 0 || projIdx >= len(projects) {
		return fmt.Errorf("project index out of range: %d", projIdx)
	}

	cfg.Categories[catIdx].Projects = append(projects[:projIdx], projects[projIdx+1:]...)
	return nil
}

// UpdateProjectGroup updates the group of a project by index.
func (cfg *Config) UpdateProjectGroup(catIdx, projIdx int, group string) error {
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return fmt.Errorf("category index out of range: %d", catIdx)
	}

	if projIdx < 0 || projIdx >= len(cfg.Categories[catIdx].Projects) {
		return fmt.Errorf("project index out of range: %d", projIdx)
	}

	cfg.Categories[catIdx].Projects[projIdx].Group = group
	return nil
}

// SetProjectGitLabPath updates the gitlab_path of a project by index.
func (cfg *Config) SetProjectGitLabPath(catIdx, projIdx int, gitlabPath string) error {
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return fmt.Errorf("category index out of range: %d", catIdx)
	}

	if projIdx < 0 || projIdx >= len(cfg.Categories[catIdx].Projects) {
		return fmt.Errorf("project index out of range: %d", projIdx)
	}

	cfg.Categories[catIdx].Projects[projIdx].GitLabPath = gitlabPath
	return nil
}

// SetProjectGitHubPath updates the github_path of a project by index.
func (cfg *Config) SetProjectGitHubPath(catIdx, projIdx int, githubPath string) error {
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return fmt.Errorf("category index out of range: %d", catIdx)
	}

	if projIdx < 0 || projIdx >= len(cfg.Categories[catIdx].Projects) {
		return fmt.Errorf("project index out of range: %d", projIdx)
	}

	cfg.Categories[catIdx].Projects[projIdx].GitHubPath = githubPath
	return nil
}
