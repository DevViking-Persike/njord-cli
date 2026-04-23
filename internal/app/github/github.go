// Package github contém as regras puras para trabalhar com projetos
// hospedados no GitHub: montagem de URL, comando de clone e detecção do host.
package github

import (
	"fmt"
	"os"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
)

// Host identifica onde o repositório do projeto está hospedado.
type Host string

const (
	HostGitLab Host = "gitlab"
	HostGitHub Host = "github"
	HostNone   Host = ""
)

// DetectHost infere o host de um projeto olhando somente os campos da config.
// Prioriza gitlab_path; na ausência, usa github_path.
func DetectHost(p config.Project) Host {
	if strings.TrimSpace(p.GitLabPath) != "" {
		return HostGitLab
	}
	if strings.TrimSpace(p.GitHubPath) != "" {
		return HostGitHub
	}
	return HostNone
}

// BrowserURL devolve a URL HTTPS pública do projeto no GitHub.
// Retorna erro se o projeto não tem github_path.
func BrowserURL(p config.Project) (string, error) {
	path := strings.Trim(strings.TrimSpace(p.GitHubPath), "/")
	if path == "" {
		return "", fmt.Errorf("projeto %q não tem github_path configurado", p.Alias)
	}
	return "https://github.com/" + path, nil
}

// BuildOpenBrowserCommand monta o comando shell que abre a URL no browser padrão.
// Usa xdg-open (Linux) — suficiente pro ambiente alvo.
func BuildOpenBrowserCommand(url string) string {
	return "xdg-open " + shellQuote(url)
}

// BuildCloneCommand monta o comando shell pra clonar um projeto GitHub pra destPath,
// entrar no diretório e abrir no editor. Se destPath já existe, só entra nele.
func BuildCloneCommand(p config.Project, destPath, editor string) (string, error) {
	if strings.TrimSpace(destPath) == "" {
		return "", fmt.Errorf("destPath vazio")
	}
	editor = strings.TrimSpace(editor)
	if editor == "" {
		editor = "code"
	}
	url, err := CloneURL(p)
	if err != nil {
		return "", err
	}
	dest := shellQuote(destPath)
	clone := "git clone " + shellQuote(url) + " " + dest
	return "if [ ! -d " + dest + " ]; then " + clone + "; fi && cd -- " + dest + " && " + editor + " .", nil
}

// CloneURL devolve a URL usada pra clonar (SSH por padrão, pra reuso das chaves).
func CloneURL(p config.Project) (string, error) {
	path := strings.Trim(strings.TrimSpace(p.GitHubPath), "/")
	if path == "" {
		return "", fmt.Errorf("projeto %q não tem github_path configurado", p.Alias)
	}
	return "git@github.com:" + path + ".git", nil
}

// LocalExists reporta se a pasta resolvida do projeto existe no disco.
func LocalExists(cfg *config.Config, p config.Project) bool {
	if cfg == nil {
		return false
	}
	path := cfg.ResolveProjectPath(p)
	if strings.HasPrefix(path, "@") {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
