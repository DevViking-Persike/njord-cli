package tauri

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRootPrecedence(t *testing.T) {
	root, err := ResolveRoot("/explicit", func() (string, error) {
		return "/cwd/njord-cli", nil
	}, func(string) (string, bool) {
		return "/env", true
	})
	if err != nil {
		t.Fatalf("ResolveRoot() error = %v", err)
	}
	if root != "/explicit" {
		t.Fatalf("ResolveRoot() = %q, want explicit path", root)
	}

	root, err = ResolveRoot("", func() (string, error) {
		return "/cwd/njord-cli", nil
	}, func(string) (string, bool) {
		return "/env", true
	})
	if err != nil {
		t.Fatalf("ResolveRoot() env error = %v", err)
	}
	if root != "/env" {
		t.Fatalf("ResolveRoot() env = %q, want env path", root)
	}
}

func TestResolveRootFallsBackToSibling(t *testing.T) {
	root, err := ResolveRoot("", func() (string, error) {
		return "/work/njord-cli", nil
	}, func(string) (string, bool) {
		return "", false
	})
	if err != nil {
		t.Fatalf("ResolveRoot() error = %v", err)
	}
	if root != "/work/njord-tauri" {
		t.Fatalf("ResolveRoot() = %q, want sibling path", root)
	}
}

func TestLoadProjectReadsValidTauriCheckout(t *testing.T) {
	root := makeProject(t, "njord-tauri", map[string]string{
		"dev":       "vite dev",
		"dev:clean": "bash scripts/dev.sh",
		"tauri":     "tauri",
	})

	project, err := LoadProject(root)
	if err != nil {
		t.Fatalf("LoadProject() error = %v", err)
	}

	if project.Name != "njord-tauri" || project.Version != "0.1.0" {
		t.Fatalf("LoadProject() = %#v", project)
	}
	if _, ok := project.Scripts["dev:clean"]; !ok {
		t.Fatalf("LoadProject() scripts missing dev:clean")
	}
}

func TestLoadProjectRejectsInvalidCheckout(t *testing.T) {
	root := makeProject(t, "other-app", nil)

	_, err := LoadProject(root)
	if err == nil {
		t.Fatal("LoadProject() error = nil, want invalid checkout error")
	}
	if !strings.Contains(err.Error(), "not a njord-tauri checkout") {
		t.Fatalf("LoadProject() error = %q", err)
	}
}

func TestBuildDevCommandPrefersCleanScript(t *testing.T) {
	cmd := BuildDevCommand(Project{
		Root: "/work/njord-tauri",
		Scripts: map[string]string{
			"dev:clean": "bash scripts/dev.sh",
			"tauri":     "tauri",
		},
	})

	if cmd.String() != "npm run dev:clean" {
		t.Fatalf("BuildDevCommand() = %q", cmd.String())
	}
	if cmd.ShellString() != "cd -- '/work/njord-tauri' && npm run dev:clean" {
		t.Fatalf("ShellString() = %q", cmd.ShellString())
	}
}

func TestBuildDevCommandFallsBackToTauriDev(t *testing.T) {
	cmd := BuildDevCommand(Project{
		Root:    "/work/njord-tauri",
		Scripts: map[string]string{"tauri": "tauri"},
	})

	if cmd.String() != "npm run tauri dev" {
		t.Fatalf("BuildDevCommand() = %q", cmd.String())
	}
}

func TestBuildScriptCommandUsesExistingPackageScript(t *testing.T) {
	cmd, err := BuildScriptCommand(Project{
		Root: "/work/njord-tauri",
		Scripts: map[string]string{
			"check": "svelte-kit sync && svelte-check",
			"test":  "vitest run",
			"build": "vite build",
		},
	}, "check")
	if err != nil {
		t.Fatalf("BuildScriptCommand() error = %v", err)
	}

	if cmd.String() != "npm run check" {
		t.Fatalf("BuildScriptCommand() = %q", cmd.String())
	}
	if cmd.ShellString() != "cd -- '/work/njord-tauri' && npm run check" {
		t.Fatalf("ShellString() = %q", cmd.ShellString())
	}
}

func TestBuildScriptCommandRejectsMissingScript(t *testing.T) {
	_, err := BuildScriptCommand(Project{
		Root:    "/work/njord-tauri",
		Scripts: map[string]string{"check": "svelte-check"},
	}, "build")
	if err == nil {
		t.Fatal("BuildScriptCommand() error = nil, want missing script error")
	}
	if !strings.Contains(err.Error(), `script "build" not found`) {
		t.Fatalf("BuildScriptCommand() error = %q", err)
	}
}

func TestFormatStatusIncludesProjectMetadata(t *testing.T) {
	status := FormatStatus(Project{
		Root:    "/work/njord-tauri",
		Name:    "njord-tauri",
		Version: "0.1.0",
		Scripts: map[string]string{
			"tauri":     "tauri",
			"dev:clean": "bash scripts/dev.sh",
		},
	})

	for _, want := range []string{
		"Njord Tauri",
		"root: /work/njord-tauri",
		"package: njord-tauri@0.1.0",
		"src-tauri/Cargo.toml: ok",
		"dev command: npm run dev:clean",
	} {
		if !strings.Contains(status, want) {
			t.Fatalf("FormatStatus() missing %q in:\n%s", want, status)
		}
	}
}

func makeProject(t *testing.T, name string, scripts map[string]string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src-tauri"), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src-tauri", "Cargo.toml"), []byte("[package]\nname = \"njord-tauri\"\n"), 0644); err != nil {
		t.Fatalf("WriteFile(Cargo.toml) error = %v", err)
	}

	scriptJSON := ""
	if scripts != nil {
		parts := make([]string, 0, len(scripts))
		for key, value := range scripts {
			parts = append(parts, `"`+key+`":"`+value+`"`)
		}
		scriptJSON = `,"scripts":{` + strings.Join(parts, ",") + `}`
	}

	data := []byte(`{"name":"` + name + `","version":"0.1.0"` + scriptJSON + `}`)
	if err := os.WriteFile(filepath.Join(root, packageFile), data, 0644); err != nil {
		t.Fatalf("WriteFile(package.json) error = %v", err)
	}

	return root
}
