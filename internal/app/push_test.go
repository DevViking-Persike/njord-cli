package app

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestPushServiceRunRequiresGitLabTokenForSubtaskBranch(t *testing.T) {
	originalCommandRunner := commandRunner
	commandRunner = func(name string, args ...string) cmdRunner {
		return stubCmdRunner{
			output: []byte("feature/BILL-123-B1-subtask-ajuste\n"),
		}
	}
	defer func() { commandRunner = originalCommandRunner }()

	repoDir := t.TempDir()
	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	configContent := []byte("[remote \"origin\"]\n\turl = git@gitlab.com:group/repo.git\n")
	if err := os.WriteFile(filepath.Join(gitDir, "config"), configContent, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	service := NewPushService()
	service.Getwd = func() (string, error) { return repoDir, nil }
	service.LoadConfig = func(path string) (*config.Config, error) {
		return &config.Config{}, nil
	}

	var stdout, stderr bytes.Buffer
	err := service.Run("", nil, &stdout, &stderr)
	if err == nil {
		t.Fatal("Run() error = nil, want token error")
	}
	if got := err.Error(); got != "token GitLab não configurado em "+config.DefaultConfigPath() {
		t.Fatalf("Run() error = %q, want token error", got)
	}
}

func TestPushServiceRunSkipsGitLabForNonSubtaskBranch(t *testing.T) {
	originalCommandRunner := commandRunner
	callCount := 0
	commandRunner = func(name string, args ...string) cmdRunner {
		callCount++
		if callCount == 1 {
			return stubCmdRunner{
				output: []byte("feature/BILL-123-B1-delivery-ajuste\n"),
			}
		}
		return stubCmdRunner{}
	}
	defer func() { commandRunner = originalCommandRunner }()

	service := NewPushService()
	service.LoadConfig = func(path string) (*config.Config, error) {
		t.Fatal("LoadConfig should not be called for non-subtask branch")
		return nil, nil
	}

	var stdout, stderr bytes.Buffer
	if err := service.Run("", nil, &stdout, &stderr); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

type stubCmdRunner struct {
	output []byte
	runErr error
}

func (s stubCmdRunner) Output() ([]byte, error) { return s.output, s.runErr }
func (s stubCmdRunner) Run() error              { return s.runErr }
func (s stubCmdRunner) SetStdout(io.Writer)     {}
func (s stubCmdRunner) SetStderr(io.Writer)     {}
