package gitlab

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
)

type cmdRunner interface {
	Output() ([]byte, error)
	Run() error
	SetStdout(io.Writer)
	SetStderr(io.Writer)
}

type execCmd struct {
	*exec.Cmd
}

func (c execCmd) SetStdout(w io.Writer) { c.Stdout = w }
func (c execCmd) SetStderr(w io.Writer) { c.Stderr = w }

var commandRunner = func(name string, args ...string) cmdRunner {
	return execCmd{Cmd: exec.Command(name, args...)}
}

type PushService struct {
	LoadConfig func(path string) (*config.Config, error)
	NewGitLab  func(token, url string) (*gitlabclient.Client, error)
	Getwd      func() (string, error)
}

func NewPushService() PushService {
	return PushService{
		LoadConfig: config.Load,
		NewGitLab:  gitlabclient.NewClient,
		Getwd:      os.Getwd,
	}
}

func (s PushService) Run(configPath string, args []string, stdout, stderr io.Writer) error {
	branchCmd := commandRunner("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("detectando branch atual: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))

	fmt.Fprintf(stderr, "📌 Branch: %s\n", branch)

	pushArgs := append([]string{"push"}, args...)
	pushCmd := commandRunner("git", pushArgs...)
	pushCmd.SetStdout(stdout)
	pushCmd.SetStderr(stderr)
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("git push falhou: %w", err)
	}

	if !strings.Contains(branch, "-subtask-") {
		return nil
	}

	fmt.Fprintf(stderr, "\n🔧 Branch subtask detectada, disparando pipeline...\n")

	cwd, err := s.Getwd()
	if err != nil {
		return fmt.Errorf("obtendo diretório atual: %w", err)
	}
	projectPath, err := gitlabclient.ParseGitLabPath(cwd)
	if err != nil {
		return fmt.Errorf("detectando projeto GitLab: %w", err)
	}

	cfgPath := configPath
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}

	cfg, err := s.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("carregando config: %w", err)
	}
	if cfg.GitLab.Token == "" {
		return fmt.Errorf("token GitLab não configurado em %s", cfgPath)
	}

	client, err := s.NewGitLab(cfg.GitLab.Token, cfg.GitLab.URL)
	if err != nil {
		return fmt.Errorf("criando client GitLab: %w", err)
	}

	pipeline, err := client.TriggerMRPipeline(projectPath, branch)
	if err != nil {
		return fmt.Errorf("disparando pipeline: %w", err)
	}

	fmt.Fprintf(stderr, "✅ Pipeline #%d disparada (MR pipeline) em %s\n", pipeline.ID, branch)
	fmt.Fprintf(stderr, "🔗 %s\n", pipeline.URL)

	return nil
}
