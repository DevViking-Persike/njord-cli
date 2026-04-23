package config

import (
	"strings"
	"testing"
)

func TestMarshalYAML_IncludesAllSections(t *testing.T) {
	cfg := &Config{
		Settings: Settings{Editor: "vim", ProjectsBase: "~/work", PersonalBase: "~/me"},
		GitLab:   GitLabSettings{Token: "glpat-xyz", URL: "https://gitlab.example.com"},
		Jira:     JiraSettings{Token: "jt", URL: "https://x.atlassian.net", Email: "a@b.com"},
		Categories: []Category{{
			ID: "core", Name: "Core", Sub: "APIs",
			Projects: []Project{{Alias: "r", Desc: "d", Path: "p", Group: "g", GitLabPath: "g/r"}},
		}},
		DockerStacks: []DockerStack{{Name: "db", Desc: "mysql", Path: "Dockers/db"}},
	}

	out, err := marshalYAML(cfg)
	if err != nil {
		t.Fatalf("marshalYAML() error = %v", err)
	}
	s := string(out)

	must := []string{
		`editor: "vim"`,
		`projects_base: "~/work"`,
		`personal_base: "~/me"`,
		`token: "glpat-xyz"`,
		`url: "https://gitlab.example.com"`,
		"jira:",
		`token: "jt"`,
		`email: "a@b.com"`,
		"- id: core",
		`name: "Core"`,
		`sub: "APIs"`,
		"- alias: r",
		`group: "g"`,
		`gitlab_path: "g/r"`,
		`- name: "db"`,
	}
	for _, frag := range must {
		if !strings.Contains(s, frag) {
			t.Errorf("output missing %q\n--- got ---\n%s", frag, s)
		}
	}
}

func TestWriteGitLab_OmitsURLWhenEmpty(t *testing.T) {
	var b strings.Builder
	writeGitLab(&b, GitLabSettings{Token: "t"})
	s := b.String()
	if !strings.Contains(s, `token: "t"`) {
		t.Errorf("expected token in output, got %q", s)
	}
	if strings.Contains(s, "url:") {
		t.Errorf("expected url omitted, got %q", s)
	}
}

func TestWriteJira_OmitsSectionWhenAllEmpty(t *testing.T) {
	var b strings.Builder
	writeJira(&b, JiraSettings{})
	if s := b.String(); s != "" {
		t.Errorf("expected empty output for empty JiraSettings, got %q", s)
	}
}

func TestWriteJira_OmitsFieldsIndividually(t *testing.T) {
	tests := []struct {
		name   string
		jira   JiraSettings
		want   []string
		absent []string
	}{
		{
			name:   "only token",
			jira:   JiraSettings{Token: "t"},
			want:   []string{"jira:", `token: "t"`},
			absent: []string{"url:", "email:"},
		},
		{
			name:   "only url",
			jira:   JiraSettings{URL: "https://x"},
			want:   []string{"jira:", `url: "https://x"`},
			absent: []string{"token:", "email:"},
		},
		{
			name:   "only email",
			jira:   JiraSettings{Email: "a@b"},
			want:   []string{"jira:", `email: "a@b"`},
			absent: []string{"token:", "url:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeJira(&b, tt.jira)
			s := b.String()
			for _, w := range tt.want {
				if !strings.Contains(s, w) {
					t.Errorf("expected %q in %q", w, s)
				}
			}
			for _, a := range tt.absent {
				if strings.Contains(s, a) {
					t.Errorf("expected %q NOT in %q", a, s)
				}
			}
		})
	}
}

func TestWriteProject_OmitsOptionalFields(t *testing.T) {
	var b strings.Builder
	writeProject(&b, Project{Alias: "a", Desc: "d", Path: "p"})
	s := b.String()
	if !strings.Contains(s, "- alias: a") {
		t.Errorf("missing alias: %q", s)
	}
	if strings.Contains(s, "group:") {
		t.Errorf("expected group omitted: %q", s)
	}
	if strings.Contains(s, "gitlab_path:") {
		t.Errorf("expected gitlab_path omitted: %q", s)
	}
	if strings.Contains(s, "github_path:") {
		t.Errorf("expected github_path omitted: %q", s)
	}
}

func TestWriteProject_IncludesGitHubPath(t *testing.T) {
	var b strings.Builder
	writeProject(&b, Project{Alias: "a", Desc: "d", Path: "p", GitHubPath: "user/repo"})
	s := b.String()
	if !strings.Contains(s, `github_path: "user/repo"`) {
		t.Errorf("expected github_path present: %q", s)
	}
}

func TestWriteGitHub_OmitsSectionWhenEmpty(t *testing.T) {
	var b strings.Builder
	writeGitHub(&b, GitHubSettings{})
	if got := b.String(); got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestWriteGitHub_WritesToken(t *testing.T) {
	var b strings.Builder
	writeGitHub(&b, GitHubSettings{Token: "ghp_abc"})
	s := b.String()
	if !strings.Contains(s, "github:") {
		t.Errorf("expected github: section, got %q", s)
	}
	if !strings.Contains(s, `token: "ghp_abc"`) {
		t.Errorf("expected token line, got %q", s)
	}
}

func TestGitLabURL_DefaultWhenEmpty(t *testing.T) {
	if got := (GitLabSettings{}).GitLabURL(); got != "https://gitlab.com" {
		t.Errorf("GitLabURL() = %q, want default", got)
	}
	if got := (GitLabSettings{URL: "https://custom"}).GitLabURL(); got != "https://custom" {
		t.Errorf("GitLabURL() = %q, want custom", got)
	}
}
