package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// dockersSubdir é a pasta sob ProjectsBase onde as stacks novas são criadas.
// Mantida como constante porque casa com o padrão `Dockers/<nome>` já salvo
// nos registros existentes em docker_stacks.
const dockersSubdir = "Dockers"

// templateStack é o nome da pasta template dentro de Dockers/. Se existir,
// o compose.yml dela é usado como skeleton pra stack nova.
const templateStack = "_template"

// fallbackCompose é usado quando _template/docker-compose.yml não existe.
const fallbackCompose = `# Stack gerado pelo njord-cli.
# Ajuste a imagem e as variáveis antes de rodar "Iniciar".

services:
  app:
    image: registry.gitlab.com/avitaseg/avita-registry/NOME-DA-IMAGEM
    network_mode: host
    restart: unless-stopped
`

func (m AddModel) handleCreateDirInput(msg tea.KeyMsg) (AddModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.step = addStackSelectPath
		m.inputBuf = ""
		m.message = ""
		return m, nil
	case "enter":
		name := sanitizeStackFolder(m.inputBuf)
		if name == "" {
			m.message = "Nome inválido"
			return m, nil
		}
		relPath, err := createStackFolder(m.cfg, name)
		if err != nil {
			m.message = err.Error()
			return m, nil
		}
		m.stackPath = relPath
		m.inputBuf = name
		m.message = ""
		m.step = addStackName
		return m, nil
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
	case "ctrl+u":
		m.inputBuf = ""
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.inputBuf += string(msg.Runes)
		}
	}
	return m, nil
}

// sanitizeStackFolder devolve um nome de pasta seguro: trim, sem separadores
// de caminho, sem ponto inicial. String vazia quando inválido.
func sanitizeStackFolder(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" || strings.ContainsAny(s, "/\\") || strings.HasPrefix(s, ".") {
		return ""
	}
	return s
}

// createStackFolder cria a pasta ~/Avita/Dockers/<name>/ e um compose.yml
// derivado do _template (ou fallback). Devolve o path relativo pronto pra
// gravar em DockerStack.Path.
//
// Falha se a pasta já existir — evita sobrescrever trabalho do usuário.
func createStackFolder(cfg *config.Config, name string) (string, error) {
	baseDir := config.ExpandPath(cfg.Settings.ProjectsBase)
	dockersDir := filepath.Join(baseDir, dockersSubdir)
	newDir := filepath.Join(dockersDir, name)

	if _, err := os.Stat(newDir); err == nil {
		return "", fmt.Errorf("pasta %s já existe", newDir)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return "", fmt.Errorf("criando pasta: %w", err)
	}

	content := templateContent(dockersDir)
	composePath := filepath.Join(newDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("gravando compose: %w", err)
	}

	return filepath.Join(dockersSubdir, name), nil
}

// templateContent lê o compose do _template se existir; senão cai pro fallback.
func templateContent(dockersDir string) string {
	tpl := filepath.Join(dockersDir, templateStack, "docker-compose.yml")
	if data, err := os.ReadFile(tpl); err == nil {
		return string(data)
	}
	return fallbackCompose
}
