package app

import (
	"fmt"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
)

type GitLabStatusLoader interface {
	ListPipelines(projectPath string, limit int) ([]gitlab.PipelineInfo, error)
	GetProjectLatestMRApproval(projectPath string) (*gitlab.MRApprovalInfo, error)
}

type GitLabProjectStatus struct {
	Status   string
	LastTime time.Time
	Approval *gitlab.MRApprovalInfo
}

func LoadGitLabProjectStatus(client GitLabStatusLoader, gitlabPath string) GitLabProjectStatus {
	status := GitLabProjectStatus{}
	if client == nil || gitlabPath == "" {
		return status
	}

	pipelines, err := client.ListPipelines(gitlabPath, 1)
	if err == nil && len(pipelines) > 0 {
		status.Status = pipelines[0].Status
		status.LastTime = pipelines[0].CreatedAt
	}

	approval, _ := client.GetProjectLatestMRApproval(gitlabPath)
	status.Approval = approval
	return status
}

func DetectGitLabPath(cfg *config.Config, catIdx, projIdx int) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is required")
	}
	if catIdx < 0 || catIdx >= len(cfg.Categories) {
		return "", fmt.Errorf("category index out of range: %d", catIdx)
	}
	if projIdx < 0 || projIdx >= len(cfg.Categories[catIdx].Projects) {
		return "", fmt.Errorf("project index out of range: %d", projIdx)
	}

	project := cfg.Categories[catIdx].Projects[projIdx]
	path := cfg.ResolveProjectPath(project)
	return gitlab.ParseGitLabPath(path)
}

func SaveGitLabPath(cfg *config.Config, configPath string, catIdx, projIdx int, gitlabPath string) error {
	if err := cfg.SetProjectGitLabPath(catIdx, projIdx, gitlabPath); err != nil {
		return err
	}
	return config.Save(cfg, configPath)
}
