package stack

import (
	"os"
	"path/filepath"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

// discoverComposeFiles varre compose files em duas camadas sob baseDir:
//  1. baseDir/<folder>/compose*              (projetos top-level)
//  2. baseDir/Dockers/<folder>/compose*      (pasta convencional de stacks)
//
// Descarta o skeleton _template e o que já está cadastrado em cfg.DockerStacks.
// Paths retornados são relativos a baseDir, no mesmo formato usado pela YAML
// (ex.: "mysql-local", "Dockers/mysql") — casam direto com DockerStack.Path.
func discoverComposeFiles(baseDir string, existing []config.DockerStack) []string {
	registered := make(map[string]bool)
	for _, stack := range existing {
		registered[stack.Path] = true
	}

	var out []string
	out = append(out, scanDir(baseDir, "", registered)...)
	out = append(out, scanDir(filepath.Join(baseDir, dockersSubdir), dockersSubdir, registered)...)
	return out
}

// scanDir lista pastas imediatas dentro de dir que contêm um compose file e
// devolve os paths relativos (usando relPrefix como prefixo). Skipa a pasta
// _template e qualquer item já em registered.
func scanDir(dir, relPrefix string, registered map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var found []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == templateStack {
			continue
		}
		if !hasCompose(filepath.Join(dir, entry.Name())) {
			continue
		}
		rel := entry.Name()
		if relPrefix != "" {
			rel = filepath.Join(relPrefix, entry.Name())
		}
		if registered[rel] {
			continue
		}
		found = append(found, rel)
	}
	return found
}

// hasCompose retorna true se dir contém qualquer variante de compose file.
func hasCompose(dir string) bool {
	names := []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}
	for _, n := range names {
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			return true
		}
	}
	return false
}
