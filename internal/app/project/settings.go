package project

import (
	"fmt"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func UpdateCategory(cfg *config.Config, configPath string, catIdx int, name, sub string) error {
	if err := cfg.UpdateCategory(catIdx, name, sub); err != nil {
		return err
	}
	return config.Save(cfg, configPath)
}

func UpdateProjectsBase(cfg *config.Config, configPath, path string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	cfg.Settings.ProjectsBase = path
	return config.Save(cfg, configPath)
}

func UpdatePersonalBase(cfg *config.Config, configPath, path string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	cfg.Settings.PersonalBase = path
	return config.Save(cfg, configPath)
}

func RemoveProject(cfg *config.Config, configPath string, catIdx, projIdx int) error {
	if err := cfg.RemoveProject(catIdx, projIdx); err != nil {
		return err
	}
	return config.Save(cfg, configPath)
}

func UpdateProjectGroup(cfg *config.Config, configPath string, catIdx, projIdx int, group string) error {
	if err := cfg.UpdateProjectGroup(catIdx, projIdx, group); err != nil {
		return err
	}
	return config.Save(cfg, configPath)
}

func UpdateGitLabToken(cfg *config.Config, configPath, token string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	cfg.GitLab.Token = token
	return config.Save(cfg, configPath)
}
