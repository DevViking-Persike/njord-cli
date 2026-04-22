package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/gitlab"
)

type stubGitLabStatusLoader struct {
	pipelines []gitlab.PipelineInfo
	approval  *gitlab.MRApprovalInfo
}

func (s stubGitLabStatusLoader) ListPipelines(projectPath string, limit int) ([]gitlab.PipelineInfo, error) {
	return s.pipelines, nil
}

func (s stubGitLabStatusLoader) GetProjectLatestMRApproval(projectPath string) (*gitlab.MRApprovalInfo, error) {
	return s.approval, nil
}

func TestLoadGitLabProjectStatus(t *testing.T) {
	now := time.Now()
	status := LoadGitLabProjectStatus(stubGitLabStatusLoader{
		pipelines: []gitlab.PipelineInfo{{Status: "success", CreatedAt: now}},
		approval:  &gitlab.MRApprovalInfo{Approved: true},
	}, "group/repo")

	if status.Status != "success" {
		t.Fatalf("Status = %q, want success", status.Status)
	}
	if !status.LastTime.Equal(now) {
		t.Fatalf("LastTime = %v, want %v", status.LastTime, now)
	}
	if status.Approval == nil || !status.Approval.Approved {
		t.Fatalf("Approval = %#v, want approved", status.Approval)
	}
}

func TestDetectGitLabPath(t *testing.T) {
	repoDir := t.TempDir()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	content := []byte("[remote \"origin\"]\n\turl = git@gitlab.com:group/repo.git\n")
	if err := os.WriteFile(filepath.Join(gitDir, "config"), content, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg := &config.Config{
		Categories: []config.Category{
			{Projects: []config.Project{{Path: repoDir}}},
		},
	}

	got, err := DetectGitLabPath(cfg, 0, 0)
	if err != nil {
		t.Fatalf("DetectGitLabPath() error = %v", err)
	}
	if got != "group/repo" {
		t.Fatalf("DetectGitLabPath() = %q, want group/repo", got)
	}
}

func TestSaveGitLabPath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "njord.yaml")
	cfg := &config.Config{
		Categories: []config.Category{
			{Projects: []config.Project{{Alias: "repo-a"}}},
		},
	}

	if err := SaveGitLabPath(cfg, configPath, 0, 0, "group/repo-a"); err != nil {
		t.Fatalf("SaveGitLabPath() error = %v", err)
	}
	if got := cfg.Categories[0].Projects[0].GitLabPath; got != "group/repo-a" {
		t.Fatalf("GitLabPath = %q, want group/repo-a", got)
	}
}
