package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
)

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

type Client struct {
	cli *dockerclient.Client
}

func NewClient() (*Client, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &Client{cli: cli}, nil
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
func (c *Client) StartProject(composePath, projectName string) error {
	if c.isComposeValid(composePath) {
		return c.composeExec(composePath, "up", "-d")
	}
	return c.startByLabel(projectName)
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
	return cmd.Run() == nil
}

func (c *Client) composeExec(composePath string, args ...string) error {
	cmdArgs := append([]string{"compose", "-f", composePath}, args...)
	cmd := exec.Command("docker", cmdArgs...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func (c *Client) composeLogs(composePath string, tail int) (string, error) {
	cmd := exec.Command("docker", "compose", "-f", composePath, "logs",
		"--tail", fmt.Sprintf("%d", tail), "--no-log-prefix")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
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
