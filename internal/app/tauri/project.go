package tauri

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	EnvPath     = "NJORD_TAURI_PATH"
	packageFile = "package.json"
	cargoFile   = "src-tauri/Cargo.toml"
)

type Project struct {
	Root    string
	Name    string
	Version string
	Scripts map[string]string
}

type CommandSpec struct {
	Dir  string
	Name string
	Args []string
}

type packageJSON struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Scripts map[string]string `json:"scripts"`
}

func ResolveRoot(explicit string, getwd func() (string, error), lookupEnv func(string) (string, bool)) (string, error) {
	if path := strings.TrimSpace(explicit); path != "" {
		return cleanPath(path), nil
	}

	if lookupEnv != nil {
		if path, ok := lookupEnv(EnvPath); ok && strings.TrimSpace(path) != "" {
			return cleanPath(path), nil
		}
	}

	if getwd == nil {
		getwd = os.Getwd
	}
	cwd, err := getwd()
	if err != nil {
		return "", fmt.Errorf("detecting current directory: %w", err)
	}

	return filepath.Join(filepath.Dir(cwd), "njord-tauri"), nil
}

func LoadProject(root string) (Project, error) {
	root = cleanPath(root)
	if root == "" {
		return Project{}, fmt.Errorf("njord-tauri path is required")
	}

	if err := requireFile(root, packageFile); err != nil {
		return Project{}, err
	}
	if err := requireFile(root, cargoFile); err != nil {
		return Project{}, err
	}

	data, err := os.ReadFile(filepath.Join(root, packageFile))
	if err != nil {
		return Project{}, fmt.Errorf("reading %s: %w", filepath.Join(root, packageFile), err)
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return Project{}, fmt.Errorf("parsing %s: %w", filepath.Join(root, packageFile), err)
	}
	if strings.TrimSpace(pkg.Name) != "njord-tauri" {
		return Project{}, fmt.Errorf("%s is not a njord-tauri checkout: package name is %q", root, pkg.Name)
	}
	if pkg.Scripts == nil {
		pkg.Scripts = map[string]string{}
	}

	return Project{
		Root:    root,
		Name:    pkg.Name,
		Version: pkg.Version,
		Scripts: pkg.Scripts,
	}, nil
}

func BuildDevCommand(project Project) CommandSpec {
	if _, ok := project.Scripts["dev:clean"]; ok {
		return CommandSpec{Dir: project.Root, Name: "npm", Args: []string{"run", "dev:clean"}}
	}
	return CommandSpec{Dir: project.Root, Name: "npm", Args: []string{"run", "tauri", "dev"}}
}

func BuildScriptCommand(project Project, script string) (CommandSpec, error) {
	script = strings.TrimSpace(script)
	if script == "" {
		return CommandSpec{}, fmt.Errorf("script name is required")
	}
	if _, ok := project.Scripts[script]; !ok {
		return CommandSpec{}, fmt.Errorf("script %q not found in %s", script, filepath.Join(project.Root, packageFile))
	}
	return CommandSpec{Dir: project.Root, Name: "npm", Args: []string{"run", script}}, nil
}

func FormatStatus(project Project) string {
	scriptNames := make([]string, 0, len(project.Scripts))
	for name := range project.Scripts {
		scriptNames = append(scriptNames, name)
	}
	sort.Strings(scriptNames)

	var b strings.Builder
	fmt.Fprintf(&b, "Njord Tauri\n")
	fmt.Fprintf(&b, "  root: %s\n", project.Root)
	fmt.Fprintf(&b, "  package: %s@%s\n", project.Name, project.Version)
	fmt.Fprintf(&b, "  %s: ok\n", packageFile)
	fmt.Fprintf(&b, "  %s: ok\n", cargoFile)
	fmt.Fprintf(&b, "  scripts:")
	if len(scriptNames) == 0 {
		fmt.Fprintf(&b, " none\n")
	} else {
		fmt.Fprintf(&b, " %s\n", strings.Join(scriptNames, ", "))
	}
	dev := BuildDevCommand(project)
	fmt.Fprintf(&b, "  dev command: %s\n", dev.String())
	return b.String()
}

func (c CommandSpec) String() string {
	if len(c.Args) == 0 {
		return c.Name
	}
	return c.Name + " " + strings.Join(c.Args, " ")
}

func (c CommandSpec) ShellString() string {
	return "cd -- " + shellQuote(c.Dir) + " && " + c.String()
}

func requireFile(root, rel string) error {
	path := filepath.Join(root, rel)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validating %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("validating %s: expected file, got directory", path)
	}
	return nil
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			path = filepath.Join(home, path[2:])
		}
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
