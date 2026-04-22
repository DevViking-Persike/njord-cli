package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/git"
)

type AddProjectInput struct {
	GitURL          string
	ClonePath       string
	Alias           string
	Description     string
	CategoryID      string
	NewCategoryName string
	NewCategorySub  string
	Group           string
}

func CloneProject(url, clonePath string) error {
	return git.Clone(url, clonePath)
}

func RelativeProjectPath(cfg *config.Config, clonePath string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is required")
	}

	avitaBase := config.ExpandPath(cfg.Settings.ProjectsBase)
	persBase := config.ExpandPath(cfg.Settings.PersonalBase)

	if strings.HasPrefix(clonePath, avitaBase) {
		relPath, err := filepath.Rel(avitaBase, clonePath)
		if err != nil {
			return "", fmt.Errorf("resolving path relative to projects base: %w", err)
		}
		return relPath, nil
	}

	if strings.HasPrefix(clonePath, persBase) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("getting user home dir: %w", err)
		}
		relPath, err := filepath.Rel(home, clonePath)
		if err != nil {
			return "", fmt.Errorf("resolving path relative to home dir: %w", err)
		}
		return relPath, nil
	}

	return clonePath, nil
}

func SaveProject(cfg *config.Config, configPath string, input AddProjectInput) error {
	relPath, err := RelativeProjectPath(cfg, input.ClonePath)
	if err != nil {
		return err
	}

	project := config.Project{
		Alias: input.Alias,
		Desc:  input.Description,
		Path:  relPath,
		Group: input.Group,
	}

	var newCategory *config.Category
	if input.NewCategoryName != "" {
		newCategory = &config.Category{
			Name: input.NewCategoryName,
			Sub:  input.NewCategorySub,
		}
	}

	if err := cfg.AddProjectToCategory(input.CategoryID, project, newCategory); err != nil {
		return err
	}

	return config.Save(cfg, configPath)
}
