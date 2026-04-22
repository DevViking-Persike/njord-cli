package gitlab

import (
	"fmt"

	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
)

type GitLabActionsClient interface {
	ListMergeRequests(projectPath string, state string) ([]gitlabclient.MergeRequestInfo, error)
	ListPipelines(projectPath string, limit int) ([]gitlabclient.PipelineInfo, error)
	ListBranchesDetailed(projectPath string) ([]gitlabclient.BranchInfo, error)
	TriggerPipeline(projectPath, ref string) (*gitlabclient.PipelineInfo, error)
	CreateBranch(projectPath, branchName, ref string) error
}

func LoadMergeRequests(client GitLabActionsClient, projectPath string) ([]gitlabclient.MergeRequestInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("gitlab client is required")
	}
	return client.ListMergeRequests(projectPath, "opened")
}

func LoadPipelines(client GitLabActionsClient, projectPath string, limit int) ([]gitlabclient.PipelineInfo, error) {
	if client == nil {
		return nil, fmt.Errorf("gitlab client is required")
	}
	return client.ListPipelines(projectPath, limit)
}

func LoadBranches(client GitLabActionsClient, projectPath string) ([]gitlabclient.BranchInfo, error) {
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
