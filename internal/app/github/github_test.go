package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestDetectHost(t *testing.T) {
	cases := []struct {
		name string
		proj config.Project
		want Host
	}{
		{"gitlab wins over github", config.Project{GitLabPath: "g/p", GitHubPath: "u/r"}, HostGitLab},
		{"github only", config.Project{GitHubPath: "u/r"}, HostGitHub},
		{"none", config.Project{}, HostNone},
		{"whitespace only", config.Project{GitHubPath: "  "}, HostNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectHost(tc.proj); got != tc.want {
				t.Fatalf("DetectHost = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBrowserURL(t *testing.T) {
	url, err := BrowserURL(config.Project{Alias: "x", GitHubPath: "owner/repo"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if url != "https://github.com/owner/repo" {
		t.Fatalf("got %q", url)
	}

	if _, err := BrowserURL(config.Project{Alias: "x"}); err == nil {
		t.Fatal("expected error for missing github_path")
	}
}

func TestBrowserURLTrimsSlashes(t *testing.T) {
	url, err := BrowserURL(config.Project{GitHubPath: "/owner/repo/"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if url != "https://github.com/owner/repo" {
		t.Fatalf("got %q", url)
	}
}

func TestCloneURL(t *testing.T) {
	url, err := CloneURL(config.Project{GitHubPath: "owner/repo"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if url != "git@github.com:owner/repo.git" {
		t.Fatalf("got %q", url)
	}
	if _, err := CloneURL(config.Project{}); err == nil {
		t.Fatal("expected error for missing github_path")
	}
}

func TestBuildOpenBrowserCommand(t *testing.T) {
	got := BuildOpenBrowserCommand("https://github.com/a/b")
	want := "xdg-open 'https://github.com/a/b'"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestBuildCloneCommand(t *testing.T) {
	p := config.Project{Alias: "x", GitHubPath: "owner/repo"}
	cmd, err := BuildCloneCommand(p, "/tmp/dest", "nvim")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(cmd, "git clone 'git@github.com:owner/repo.git' '/tmp/dest'") {
		t.Fatalf("clone segment missing: %q", cmd)
	}
	if !strings.Contains(cmd, "cd -- '/tmp/dest' && nvim .") {
		t.Fatalf("cd+editor segment missing: %q", cmd)
	}
	if !strings.HasPrefix(cmd, "if [ ! -d '/tmp/dest' ]") {
		t.Fatalf("guard segment missing: %q", cmd)
	}
}

func TestBuildCloneCommandDefaultsEditor(t *testing.T) {
	p := config.Project{GitHubPath: "o/r"}
	cmd, err := BuildCloneCommand(p, "/tmp/d", "")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(cmd, "code .") {
		t.Fatalf("expected 'code .' fallback, got %q", cmd)
	}
}

func TestBuildCloneCommandErrors(t *testing.T) {
	if _, err := BuildCloneCommand(config.Project{GitHubPath: "o/r"}, "", "code"); err == nil {
		t.Fatal("expected error for empty destPath")
	}
	if _, err := BuildCloneCommand(config.Project{}, "/tmp/x", "code"); err == nil {
		t.Fatal("expected error for missing github_path")
	}
}

func TestLocalExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "repo"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Settings: config.Settings{ProjectsBase: dir},
	}
	present := config.Project{Path: "repo"}
	absent := config.Project{Path: "missing"}
	rdp := config.Project{Path: "@rdp"}

	if !LocalExists(cfg, present) {
		t.Fatal("expected present project to exist")
	}
	if LocalExists(cfg, absent) {
		t.Fatal("expected missing project to not exist")
	}
	if LocalExists(cfg, rdp) {
		t.Fatal("expected @rdp special path to never exist")
	}
	if LocalExists(nil, present) {
		t.Fatal("expected nil cfg to return false")
	}
}
