package stack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func TestSanitizeStackFolder(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"my-stack", "my-stack"},
		{"  spaces  ", "spaces"},
		{"", ""},
		{"  ", ""},
		{"bad/slash", ""},
		{"bad\\backslash", ""},
		{".hidden", ""},
		{"ok_under_score", "ok_under_score"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := sanitizeStackFolder(tc.in); got != tc.want {
				t.Errorf("sanitizeStackFolder(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCreateStackFolder(t *testing.T) {
	base := t.TempDir()
	cfg := &config.Config{Settings: config.Settings{ProjectsBase: base}}

	// Template skeleton in Dockers/_template/docker-compose.yml pra simular o
	// layout real do usuário e garantir que a gente copie o conteúdo certo.
	tplDir := filepath.Join(base, dockersSubdir, templateStack)
	if err := os.MkdirAll(tplDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tplContent := "services:\n  app:\n    image: from-template\n"
	if err := os.WriteFile(filepath.Join(tplDir, "docker-compose.yml"), []byte(tplContent), 0o644); err != nil {
		t.Fatal(err)
	}

	relPath, err := createStackFolder(cfg, "novo")
	if err != nil {
		t.Fatalf("createStackFolder() err = %v", err)
	}
	want := filepath.Join(dockersSubdir, "novo")
	if relPath != want {
		t.Fatalf("relPath = %q, want %q", relPath, want)
	}

	newCompose := filepath.Join(base, relPath, "docker-compose.yml")
	data, err := os.ReadFile(newCompose)
	if err != nil {
		t.Fatalf("reading new compose: %v", err)
	}
	if string(data) != tplContent {
		t.Fatalf("content = %q, want %q", data, tplContent)
	}
}

func TestCreateStackFolder_FailsIfExists(t *testing.T) {
	base := t.TempDir()
	cfg := &config.Config{Settings: config.Settings{ProjectsBase: base}}

	dir := filepath.Join(base, dockersSubdir, "existe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := createStackFolder(cfg, "existe")
	if err == nil {
		t.Fatal("expected error for existing folder")
	}
	if !strings.Contains(err.Error(), "já existe") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateStackFolder_FallbackWhenTemplateMissing(t *testing.T) {
	base := t.TempDir()
	cfg := &config.Config{Settings: config.Settings{ProjectsBase: base}}

	relPath, err := createStackFolder(cfg, "semtpl")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(base, relPath, "docker-compose.yml"))
	if !strings.Contains(string(data), "NOME-DA-IMAGEM") {
		t.Fatalf("expected fallback content, got %q", data)
	}
}
