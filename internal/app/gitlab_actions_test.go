package app

import (
	"errors"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
)

type stubGitLabActionsClient struct {
	mrs        []gitlab.MergeRequestInfo
	pipelines  []gitlab.PipelineInfo
	branches   []gitlab.BranchInfo
	triggered  *gitlab.PipelineInfo
	triggerErr error
	createErr  error
}

func (s stubGitLabActionsClient) ListMergeRequests(projectPath string, state string) ([]gitlab.MergeRequestInfo, error) {
	return s.mrs, nil
}

func (s stubGitLabActionsClient) ListPipelines(projectPath string, limit int) ([]gitlab.PipelineInfo, error) {
	return s.pipelines, nil
}

func (s stubGitLabActionsClient) ListBranchesDetailed(projectPath string) ([]gitlab.BranchInfo, error) {
	return s.branches, nil
}

func (s stubGitLabActionsClient) TriggerPipeline(projectPath, ref string) (*gitlab.PipelineInfo, error) {
	return s.triggered, s.triggerErr
}

func (s stubGitLabActionsClient) CreateBranch(projectPath, branchName, ref string) error {
	return s.createErr
}

func TestLoadGitLabActionsData(t *testing.T) {
	client := stubGitLabActionsClient{
		mrs:       []gitlab.MergeRequestInfo{{IID: 1}},
		pipelines: []gitlab.PipelineInfo{{ID: 2}},
		branches:  []gitlab.BranchInfo{{Name: "main"}},
	}

	mrs, err := LoadMergeRequests(client, "group/repo")
	if err != nil || len(mrs) != 1 {
		t.Fatalf("LoadMergeRequests() = %#v, %v", mrs, err)
	}

	pipelines, err := LoadPipelines(client, "group/repo", 20)
	if err != nil || len(pipelines) != 1 {
		t.Fatalf("LoadPipelines() = %#v, %v", pipelines, err)
	}

	branches, err := LoadBranches(client, "group/repo")
	if err != nil || len(branches) != 1 {
		t.Fatalf("LoadBranches() = %#v, %v", branches, err)
	}
}

func TestTriggerProjectPipeline(t *testing.T) {
	message, err := TriggerProjectPipeline(stubGitLabActionsClient{
		triggered: &gitlab.PipelineInfo{ID: 42},
	}, "group/repo", "main")
	if err != nil {
		t.Fatalf("TriggerProjectPipeline() error = %v", err)
	}
	if message != "Pipeline #42 disparada na branch main" {
		t.Fatalf("TriggerProjectPipeline() = %q", message)
	}
}

func TestCreateProjectBranch(t *testing.T) {
	message, err := CreateProjectBranch(stubGitLabActionsClient{}, "group/repo", "feature/a", "main")
	if err != nil {
		t.Fatalf("CreateProjectBranch() error = %v", err)
	}
	if message != "Branch 'feature/a' criada a partir de 'main'" {
		t.Fatalf("CreateProjectBranch() = %q", message)
	}
}

func TestGitLabActionsErrors(t *testing.T) {
	_, err := TriggerProjectPipeline(stubGitLabActionsClient{triggerErr: errors.New("boom")}, "group/repo", "main")
	if err == nil {
		t.Fatal("TriggerProjectPipeline() error = nil, want error")
	}

	_, err = CreateProjectBranch(stubGitLabActionsClient{createErr: errors.New("boom")}, "group/repo", "feature/a", "main")
	if err == nil {
		t.Fatal("CreateProjectBranch() error = nil, want error")
	}
}
