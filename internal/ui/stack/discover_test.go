package stack

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

func writeCompose(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("services:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverComposeFiles_TwoLayers(t *testing.T) {
	base := t.TempDir()

	// camada 1: baseDir/<proj>/compose
	writeCompose(t, filepath.Join(base, "gap-stack"))
	writeCompose(t, filepath.Join(base, "outro-top"))

	// camada 2: baseDir/Dockers/<proj>/compose
	writeCompose(t, filepath.Join(base, dockersSubdir, "mysql"))
	writeCompose(t, filepath.Join(base, dockersSubdir, "mongo"))

	// template deve ser skipado
	writeCompose(t, filepath.Join(base, dockersSubdir, templateStack))

	// uma já cadastrada deve sair da lista
	existing := []config.DockerStack{{Path: "Dockers/mongo"}}

	got := discoverComposeFiles(base, existing)
	sort.Strings(got)
	want := []string{"Dockers/mysql", "gap-stack", "outro-top"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, g := range got {
		if g != want[i] {
			t.Fatalf("[%d] = %q, want %q", i, g, want[i])
		}
	}
}

func TestDiscoverComposeFiles_EmptyBase(t *testing.T) {
	if got := discoverComposeFiles(t.TempDir(), nil); got != nil {
		t.Fatalf("expected nil on empty dir, got %v", got)
	}
}

func TestDiscoverComposeFiles_IgnoresFoldersWithoutCompose(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "sem-compose"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := discoverComposeFiles(base, nil); got != nil {
		t.Fatalf("expected nil when no compose files, got %v", got)
	}
}
