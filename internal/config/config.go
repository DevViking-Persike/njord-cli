package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Settings struct {
	Editor       string `koanf:"editor" yaml:"editor"`
	ProjectsBase string `koanf:"projects_base" yaml:"projects_base"`
	PersonalBase string `koanf:"personal_base" yaml:"personal_base"`
}

type Project struct {
	Alias      string `koanf:"alias" yaml:"alias"`
	Desc       string `koanf:"desc" yaml:"desc"`
	Path       string `koanf:"path" yaml:"path"`
	Group      string `koanf:"group" yaml:"group,omitempty"`
	GitLabPath string `koanf:"gitlab_path" yaml:"gitlab_path,omitempty"`
}

type Category struct {
	ID       string    `koanf:"id" yaml:"id"`
	Name     string    `koanf:"name" yaml:"name"`
	Sub      string    `koanf:"sub" yaml:"sub"`
	Projects []Project `koanf:"projects" yaml:"projects"`
}

type DockerStack struct {
	Name string `koanf:"name" yaml:"name"`
	Desc string `koanf:"desc" yaml:"desc"`
	Path string `koanf:"path" yaml:"path"`
}

type GitLabSettings struct {
	Token string `koanf:"token" yaml:"token"`
	URL   string `koanf:"url" yaml:"url,omitempty"`
}

func (g GitLabSettings) GitLabURL() string {
	if g.URL != "" {
		return g.URL
	}
	return "https://gitlab.com"
}

type Config struct {
	Settings     Settings       `koanf:"settings" yaml:"settings"`
	GitLab       GitLabSettings `koanf:"gitlab" yaml:"gitlab"`
	Categories   []Category     `koanf:"categories" yaml:"categories"`
	DockerStacks []DockerStack  `koanf:"docker_stacks" yaml:"docker_stacks"`
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "njord", "njord.yaml")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Set defaults
	if cfg.Settings.Editor == "" {
		cfg.Settings.Editor = "code"
	}
	if cfg.Settings.ProjectsBase == "" {
		cfg.Settings.ProjectsBase = "~/Avita"
	}
	if cfg.Settings.PersonalBase == "" {
		cfg.Settings.PersonalBase = "~/Persike"
	}

	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := marshalYAML(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// ResolveProjectPath returns the full filesystem path for a project.
func (cfg *Config) ResolveProjectPath(p Project) string {
	path := p.Path

	// Special handlers
	if strings.HasPrefix(path, "@") {
		return path
	}

	// Personal projects (contain "Persike/")
	if strings.Contains(path, "Persike/") || strings.HasPrefix(path, "Persike/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path)
	}

	// Environment projects
	if strings.HasPrefix(path, "env/") {
		base := expandPath(cfg.Settings.ProjectsBase)
		return filepath.Join(base, path)
	}

	// Default: projects base
	base := expandPath(cfg.Settings.ProjectsBase)
	return filepath.Join(base, path)
}

// ResolveDockerComposePath returns the full path to docker-compose.yml for a stack.
func (cfg *Config) ResolveDockerComposePath(stack DockerStack) string {
	base := expandPath(cfg.Settings.ProjectsBase)
	return filepath.Join(base, stack.Path, "docker-compose.yml")
}

// TotalProjects returns the total number of projects across all categories.
func (cfg *Config) TotalProjects() int {
	total := 0
	for _, cat := range cfg.Categories {
		total += len(cat.Projects)
	}
	return total
}

// AllProjects returns all projects from all categories.
func (cfg *Config) AllProjects() []Project {
	var all []Project
	for _, cat := range cfg.Categories {
		all = append(all, cat.Projects...)
	}
	return all
}

// GitLabProjectCount returns the number of projects with a gitlab_path configured.
func (cfg *Config) GitLabProjectCount() int {
	count := 0
	for _, cat := range cfg.Categories {
		for _, p := range cat.Projects {
			if p.GitLabPath != "" {
				count++
			}
		}
	}
	return count
}

// PathToAliasMap returns a map from gitlab_path to project alias.
func (cfg *Config) PathToAliasMap() map[string]string {
	m := make(map[string]string)
	for _, cat := range cfg.Categories {
		for _, p := range cat.Projects {
			if p.GitLabPath != "" {
				m[p.GitLabPath] = p.Alias
			}
		}
	}
	return m
}

// GroupedProjects returns projects sorted by group, preserving order within groups.
// Projects without a group come last under an empty-string key.
func GroupedProjects(projects []Project) (groups []string, byGroup map[string][]Project) {
	byGroup = make(map[string][]Project)
	seen := make(map[string]bool)
	for _, p := range projects {
		g := p.Group
		if !seen[g] {
			seen[g] = true
			groups = append(groups, g)
		}
		byGroup[g] = append(byGroup[g], p)
	}
	// Sort: named groups first (alphabetically), then ungrouped
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i] == "" {
			return false
		}
		if groups[j] == "" {
			return true
		}
		return groups[i] < groups[j]
	})
	return groups, byGroup
}

// marshalYAML produces a clean YAML output for the config.
func marshalYAML(cfg *Config) ([]byte, error) {
	var b strings.Builder

	b.WriteString("settings:\n")
	b.WriteString(fmt.Sprintf("  editor: %q\n", cfg.Settings.Editor))
	b.WriteString(fmt.Sprintf("  projects_base: %q\n", cfg.Settings.ProjectsBase))
	b.WriteString(fmt.Sprintf("  personal_base: %q\n", cfg.Settings.PersonalBase))

	b.WriteString("\ngitlab:\n")
	b.WriteString(fmt.Sprintf("  token: %q\n", cfg.GitLab.Token))
	if cfg.GitLab.URL != "" {
		b.WriteString(fmt.Sprintf("  url: %q\n", cfg.GitLab.URL))
	}

	b.WriteString("\ncategories:\n")
	for _, cat := range cfg.Categories {
		b.WriteString(fmt.Sprintf("  - id: %s\n", cat.ID))
		b.WriteString(fmt.Sprintf("    name: %q\n", cat.Name))
		b.WriteString(fmt.Sprintf("    sub: %q\n", cat.Sub))
		b.WriteString("    projects:\n")
		for _, p := range cat.Projects {
			b.WriteString(fmt.Sprintf("      - alias: %s\n", p.Alias))
			b.WriteString(fmt.Sprintf("        desc: %q\n", p.Desc))
			b.WriteString(fmt.Sprintf("        path: %q\n", p.Path))
			if p.Group != "" {
				b.WriteString(fmt.Sprintf("        group: %q\n", p.Group))
			}
			if p.GitLabPath != "" {
				b.WriteString(fmt.Sprintf("        gitlab_path: %q\n", p.GitLabPath))
			}
		}
	}

	b.WriteString("\ndocker_stacks:\n")
	for _, s := range cfg.DockerStacks {
		b.WriteString(fmt.Sprintf("  - name: %q\n", s.Name))
		b.WriteString(fmt.Sprintf("    desc: %q\n", s.Desc))
		b.WriteString(fmt.Sprintf("    path: %q\n", s.Path))
	}

	return []byte(b.String()), nil
}
