package config

import (
	"fmt"
	"strings"
)

// marshalYAML produces a clean YAML output for the config.
func marshalYAML(cfg *Config) ([]byte, error) {
	var b strings.Builder

	writeSettings(&b, cfg.Settings)
	writeGitLab(&b, cfg.GitLab)
	writeJira(&b, cfg.Jira)
	writeCategories(&b, cfg.Categories)
	writeDockerStacks(&b, cfg.DockerStacks)

	return []byte(b.String()), nil
}

func writeSettings(b *strings.Builder, s Settings) {
	b.WriteString("settings:\n")
	fmt.Fprintf(b, "  editor: %q\n", s.Editor)
	fmt.Fprintf(b, "  projects_base: %q\n", s.ProjectsBase)
	fmt.Fprintf(b, "  personal_base: %q\n", s.PersonalBase)
}

func writeGitLab(b *strings.Builder, g GitLabSettings) {
	b.WriteString("\ngitlab:\n")
	fmt.Fprintf(b, "  token: %q\n", g.Token)
	if g.URL != "" {
		fmt.Fprintf(b, "  url: %q\n", g.URL)
	}
}

func writeJira(b *strings.Builder, j JiraSettings) {
	if j.Token == "" && j.URL == "" && j.Email == "" {
		return
	}
	b.WriteString("\njira:\n")
	if j.Token != "" {
		fmt.Fprintf(b, "  token: %q\n", j.Token)
	}
	if j.URL != "" {
		fmt.Fprintf(b, "  url: %q\n", j.URL)
	}
	if j.Email != "" {
		fmt.Fprintf(b, "  email: %q\n", j.Email)
	}
}

func writeCategories(b *strings.Builder, cats []Category) {
	b.WriteString("\ncategories:\n")
	for _, cat := range cats {
		fmt.Fprintf(b, "  - id: %s\n", cat.ID)
		fmt.Fprintf(b, "    name: %q\n", cat.Name)
		fmt.Fprintf(b, "    sub: %q\n", cat.Sub)
		b.WriteString("    projects:\n")
		for _, p := range cat.Projects {
			writeProject(b, p)
		}
	}
}

func writeProject(b *strings.Builder, p Project) {
	fmt.Fprintf(b, "      - alias: %s\n", p.Alias)
	fmt.Fprintf(b, "        desc: %q\n", p.Desc)
	fmt.Fprintf(b, "        path: %q\n", p.Path)
	if p.Group != "" {
		fmt.Fprintf(b, "        group: %q\n", p.Group)
	}
	if p.GitLabPath != "" {
		fmt.Fprintf(b, "        gitlab_path: %q\n", p.GitLabPath)
	}
}

func writeDockerStacks(b *strings.Builder, stacks []DockerStack) {
	b.WriteString("\ndocker_stacks:\n")
	for _, s := range stacks {
		fmt.Fprintf(b, "  - name: %q\n", s.Name)
		fmt.Fprintf(b, "    desc: %q\n", s.Desc)
		fmt.Fprintf(b, "    path: %q\n", s.Path)
	}
}
