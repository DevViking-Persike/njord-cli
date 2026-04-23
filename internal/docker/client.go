package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// containerConflictRegex captura o nome do container conflitante na mensagem
// "Conflict. The container name "/foo" is already in use by container "abc..."".
// Docker sempre prefixa o nome com "/"; o grupo 1 já vem sem a barra.
var containerConflictRegex = regexp.MustCompile(`container name "/?([^"]+)" is already in use`)

type ContainerInfo struct {
	Name  string
	State string
	Ports string
}

type StackStatus struct {
	Total   int
	Running int
	Symbol  string // ●, ○, ◐
	Label   string
}

func UnavailableStatus() StackStatus {
	return StackStatus{Symbol: "!", Label: "docker indisponivel"}
}

type Client struct {
	cli *dockerclient.Client
}

func NewClient() (*Client, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	client := &Client{cli: cli}
	if !client.IsAvailable() {
		client.Close()
		return nil, fmt.Errorf("docker daemon not reachable")
	}
	return client, nil
}

func (c *Client) Close() {
	if c.cli != nil {
		c.cli.Close()
	}
}

// IsAvailable checks if docker daemon is running.
func (c *Client) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := c.cli.Ping(ctx)
	return err == nil
}

// GetStackStatus returns the status of a docker compose project.
func (c *Client) GetStackStatus(composePath, projectName string) StackStatus {
	containers := c.listByProject(projectName)

	if len(containers) == 0 {
		containers = c.listByCompose(composePath)
	}

	if len(containers) == 0 {
		return StackStatus{Symbol: "○", Label: "parado"}
	}

	total := len(containers)
	running := 0
	for _, ct := range containers {
		if ct.State == "running" {
			running++
		}
	}

	if running == 0 {
		return StackStatus{Total: total, Running: 0, Symbol: "○", Label: "parado"}
	}
	if running == total {
		return StackStatus{Total: total, Running: running, Symbol: "●", Label: fmt.Sprintf("%d/%d rodando", running, total)}
	}
	return StackStatus{Total: total, Running: running, Symbol: "◐", Label: fmt.Sprintf("%d/%d rodando", running, total)}
}

// ListContainers returns detailed container info for a project.
func (c *Client) ListContainers(projectName string) []ContainerInfo {
	return c.listByProject(projectName)
}

// StartProject starts a docker compose project.
//
// Se o compose falhar com conflito de nome de container (ex.: um container
// parado de outra stack/projeto com `container_name:` pinado está ocupando
// o nome), remove o container conflitante e repete o up. Só tenta uma vez
// pra evitar loops se o problema for outro.
func (c *Client) StartProject(composePath, projectName string) error {
	if c.isComposeValid(composePath) {
		err := c.composeExec(composePath, "up", "-d")
		if err == nil {
			return nil
		}
		if name := parseConflictContainerName(err.Error()); name != "" {
			if rmErr := c.forceRemoveContainer(name); rmErr == nil {
				return c.composeExec(composePath, "up", "-d")
			}
		}
		return err
	}
	return c.startByLabel(projectName)
}

// parseConflictContainerName extrai o nome do container conflitante do erro
// do compose. Retorna "" se a mensagem não bater com o padrão.
func parseConflictContainerName(msg string) string {
	m := containerConflictRegex.FindStringSubmatch(msg)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// forceRemoveContainer apaga um container por nome (ignora parado/rodando)
// via API do daemon. Usado pra resolver conflitos de nome antes de retry.
func (c *Client) forceRemoveContainer(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return c.cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})
}

// StopProject stops a docker compose project.
func (c *Client) StopProject(composePath, projectName string) error {
	if c.isComposeValid(composePath) {
		return c.composeExec(composePath, "down")
	}
	return c.stopByLabel(projectName)
}

// RestartProject restarts a docker compose project.
func (c *Client) RestartProject(composePath, projectName string) error {
	if c.isComposeValid(composePath) {
		return c.composeExec(composePath, "restart")
	}
	return c.restartByLabel(projectName)
}

// GetLogs returns the last N lines of logs for a project.
func (c *Client) GetLogs(composePath, projectName string, tail int) (string, error) {
	if c.isComposeValid(composePath) {
		return c.composeLogs(composePath, tail)
	}
	return c.logsByLabel(projectName, tail)
}

// --- Internal helpers ---

func (c *Client) listContainersByLabel(projectName string, all bool) ([]types.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	f := filters.NewArgs()
	f.Add("label", "com.docker.compose.project="+projectName)

	return c.cli.ContainerList(ctx, container.ListOptions{
		All:     all,
		Filters: f,
	})
}

func (c *Client) listByProject(projectName string) []ContainerInfo {
	containers, err := c.listContainersByLabel(projectName, true)
	if err != nil {
		return nil
	}

	var infos []ContainerInfo
	for _, ct := range containers {
		name := ""
		if len(ct.Names) > 0 {
			name = strings.TrimPrefix(ct.Names[0], "/")
		}

		var ports []string
		for _, p := range ct.Ports {
			if p.PublicPort > 0 {
				ports = append(ports, fmt.Sprintf("%d->%d", p.PublicPort, p.PrivatePort))
			}
		}

		infos = append(infos, ContainerInfo{
			Name:  name,
			State: ct.State,
			Ports: strings.Join(ports, ", "),
		})
	}
	return infos
}

func (c *Client) listByCompose(composePath string) []ContainerInfo {
	cmd := exec.Command("docker", "compose", "-f", composePath, "ps", "--format", "{{.Name}}\t{{.State}}\t{{.Ports}}")
	cmd.Dir = filepath.Dir(composePath)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var infos []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		info := ContainerInfo{}
		if len(parts) >= 1 {
			info.Name = parts[0]
		}
		if len(parts) >= 2 {
			info.State = parts[1]
		}
		if len(parts) >= 3 {
			info.Ports = parts[2]
		}
		infos = append(infos, info)
	}
	return infos
}

func (c *Client) isComposeValid(composePath string) bool {
	cmd := exec.Command("docker", "compose", "-f", composePath, "config", "-q")
	cmd.Dir = filepath.Dir(composePath)
	return cmd.Run() == nil
}

func (c *Client) composeExec(composePath string, args ...string) error {
	cmdArgs := append([]string{"compose", "-f", composePath}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Dir = filepath.Dir(composePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if msg := extractDockerError(string(output)); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}
	return nil
}

func (c *Client) composeLogs(composePath string, tail int) (string, error) {
	cmd := exec.Command("docker", "compose", "-f", composePath, "logs",
		"--tail", fmt.Sprintf("%d", tail), "--no-log-prefix")
	cmd.Dir = filepath.Dir(composePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if msg := extractDockerError(string(out)); msg != "" {
			return "", fmt.Errorf("%s", msg)
		}
		return "", err
	}
	return string(out), nil
}

// extractDockerError picks the most meaningful error line from docker CLI output.
// Docker compose prints progress (with ANSI codes) and the real failure is either
// a line containing "error" / "fail" or the last non-progress line.
func extractDockerError(output string) string {
	clean := ansiRegexp.ReplaceAllString(output, "")
	lines := strings.Split(strings.TrimRight(clean, "\n"), "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") || strings.Contains(lower, "failed") ||
			strings.Contains(lower, "cannot") || strings.Contains(lower, "could not") {
			return cleanDockerLine(line)
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" || isBenignProgressLine(line) {
			continue
		}
		return cleanDockerLine(line)
	}
	return ""
}

func isBenignProgressLine(line string) bool {
	suffixes := []string{" Created", " Started", " Running", " Pulled", " Pulling"}
	for _, s := range suffixes {
		if strings.HasSuffix(line, s) {
			return true
		}
	}
	return strings.HasPrefix(line, "[+]")
}

func cleanDockerLine(line string) string {
	line = strings.TrimPrefix(line, "Error response from daemon: ")
	line = strings.TrimPrefix(line, "Error: ")
	line = strings.TrimPrefix(line, "ERROR: ")
	return line
}

func (c *Client) startByLabel(projectName string) error {
	ids := c.containerIDsByLabel(projectName, true)
	if len(ids) == 0 {
		return fmt.Errorf("no containers found for project %s", projectName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, id := range ids {
		if err := c.cli.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) stopByLabel(projectName string) error {
	ids := c.containerIDsByLabel(projectName, false)
	if len(ids) == 0 {
		return fmt.Errorf("no running containers found for project %s", projectName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, id := range ids {
		if err := c.cli.ContainerStop(ctx, id, container.StopOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) restartByLabel(projectName string) error {
	ids := c.containerIDsByLabel(projectName, false)
	if len(ids) == 0 {
		return fmt.Errorf("no running containers found for project %s", projectName)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, id := range ids {
		if err := c.cli.ContainerRestart(ctx, id, container.StopOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) logsByLabel(projectName string, tail int) (string, error) {
	ids := c.containerIDsByLabel(projectName, false)
	if len(ids) == 0 {
		return "", fmt.Errorf("no running containers for project %s", projectName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var buf bytes.Buffer
	for _, id := range ids {
		reader, err := c.cli.ContainerLogs(ctx, id, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       fmt.Sprintf("%d", tail),
		})
		if err != nil {
			continue
		}
		io.Copy(&buf, reader)
		reader.Close()
	}
	return buf.String(), nil
}

func (c *Client) containerIDsByLabel(projectName string, all bool) []string {
	containers, err := c.listContainersByLabel(projectName, all)
	if err != nil {
		return nil
	}

	var ids []string
	for _, ct := range containers {
		ids = append(ids, ct.ID)
	}
	return ids
}
