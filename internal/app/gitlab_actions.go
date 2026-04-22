package app

import (
	"fmt"

	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
)

type GitLabActionsClient interface {
	ListMergeRequests(projectPath string, state string) ([]gitlab.MergeRequestInfo, error)
	ListPipelines(projectPath string, limit int) ([]gitlab.PipelineInfo, error)
	ListBranchesDetailed(projectPath string) ([]gitlab.BranchInfo, error)
	TriggerPipeline(projectPath, ref string) (*gitlab.PipelineInfo, error)
	CreateBranch(projectPath, branchName, ref string) error
}

func LoadMergeRequests(client GitLabActionsClient, projectPath string) ([]gitlab.MergeRequestInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("gitlab client is required")
	}
	return client.ListMergeRequests(projectPath, "opened")
}

func LoadPipelines(client GitLabActionsClient, projectPath string, limit int) ([]gitlab.PipelineInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("gitlab client is required")
	}
	return client.ListPipelines(projectPath, limit)
}

func LoadBranches(client GitLabActionsClient, projectPath string) ([]gitlab.BranchInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("gitlab client is required")
	}
	return client.ListBranchesDetailed(projectPath)
}

func TriggerProjectPipeline(client GitLabActionsClient, projectPath, ref string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("gitlab client is required")
	}
	pipeline, err := client.TriggerPipeline(projectPath, ref)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Pipeline #%d disparada na branch %s", pipeline.ID, ref), nil
}

func CreateProjectBranch(client GitLabActionsClient, projectPath, branchName, ref string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("gitlab client is required")
	}
	if err := client.CreateBranch(projectPath, branchName, ref); err != nil {
		return "", err
	}
	return fmt.Sprintf("Branch '%s' criada a partir de '%s'", branchName, ref), nil
}
