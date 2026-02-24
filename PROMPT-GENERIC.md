# Prompt: Tornar Njord CLI genérico para qualquer PC

## Contexto
O Njord CLI (`~/Persike/njord-cli/`) é uma TUI em Go (bubbletea + lipgloss + cobra) para gerenciar projetos e Docker stacks. Atualmente funciona apenas no meu PC com paths hardcoded. Preciso torná-lo genérico para rodar em qualquer máquina.

## Stack atual
- Go 1.23+ (binário em `cmd/njord/main.go`)
- Config YAML em `~/.config/njord/njord.yaml` (koanf)
- Docker SDK v27 (`github.com/docker/docker`)
- Charm stack: bubbletea, lipgloss, bubbles, cobra

## O que implementar

### 1. Comando `njord-cli init` (Setup interativo)

Adicionar subcomando cobra `init` que roda um wizard de primeiro uso:

```
njord-cli init
```

**Wizard steps:**
1. **Boas-vindas** - "Bem-vindo ao Njord! Vamos configurar seu ambiente."
2. **Editor** - Perguntar qual editor usar (code, cursor, nvim, vim, custom)
3. **Diretório de projetos** - Perguntar path base dos projetos (default: `~/Projects`)
4. **Diretório pessoal** - Perguntar path base pessoal (default: `~/Personal`, ou skip)
5. **Criar diretórios** - Criar os diretórios se não existirem
6. **Gerar config** - Salvar `~/.config/njord/njord.yaml` com settings + categorias vazias
7. **Shell wrapper** - Detectar shell (zsh/bash/fish) e oferecer instalar o wrapper automaticamente:
   - zsh: adicionar função no `~/.zshrc`
   - bash: adicionar função no `~/.bashrc`
   - fish: criar `~/.config/fish/functions/njord.fish`
8. **Confirmação** - Mostrar resumo e próximos passos

**Implementação:**
- Arquivo: `cmd/njord/init.go` (novo)
- Usar `charmbracelet/huh` para forms interativos OU inputs simples com bubbletea
- Detectar shell: `os.Getenv("SHELL")`
- Gerar config mínima (sem projetos pré-cadastrados, só settings)

**Shell wrappers por shell:**

```zsh
# ~/.zshrc
njord() {
    local result
    result=$(njord-cli "$@" 2>/dev/tty)
    local code=$?
    if [[ $code -eq 0 && -n "$result" ]]; then
        eval "$result"
    fi
}
```

```bash
# ~/.bashrc
njord() {
    local result
    result=$(njord-cli "$@" 2>/dev/tty)
    local code=$?
    if [[ $code -eq 0 && -n "$result" ]]; then
        eval "$result"
    fi
}
```

```fish
# ~/.config/fish/functions/njord.fish
function njord
    set result (njord-cli $argv 2>/dev/tty)
    set code $status
    if test $code -eq 0 -a -n "$result"
        eval $result
    end
end
```

### 2. Auto-detecção no primeiro uso

Em `cmd/njord/main.go`, no `runTUI()`:
- Se config não existe → rodar `init` automaticamente ao invés de criar config hardcoded
- Remover o `defaultConfig()` com projetos Avita hardcoded
- Config padrão deve ser vazia (só settings)

### 3. GoReleaser para distribuição

Criar `.goreleaser.yaml` na raiz do projeto:

```yaml
version: 2

project_name: njord-cli

before:
  hooks:
    - go mod tidy

builds:
  - id: njord-cli
    main: ./cmd/njord/
    binary: njord-cli
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc

release:
  github:
    owner: DevViking-Persike
    name: njord-cli
```

### 4. GitHub Actions para release automático

Criar `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 5. Atualizar `cmd/njord/main.go`

- Adicionar variáveis de build: `version`, `commit` (injetadas pelo goreleaser via ldflags)
- Registrar comando `init` no cobra
- Alterar lógica de config missing: chamar init ao invés de criar default hardcoded
- Remover `defaultConfig()` com projetos Avita

### 6. Makefile para build local

Criar `Makefile`:

```makefile
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: build install clean

build:
	go build -ldflags "$(LDFLAGS)" -o njord-cli ./cmd/njord/

install: build
	cp njord-cli ~/.local/bin/njord-cli

clean:
	rm -f njord-cli
```

### 7. Atualizar README.md

Criar `README.md` com:
- O que é o Njord
- Screenshot/GIF da TUI
- Instalação (3 formas):
  1. Download do release: `curl -sL github.com/.../releases/latest/...`
  2. Go install: `go install github.com/DevViking-Persike/njord-cli@latest`
  3. Build manual: `git clone && make install`
- Primeiro uso: `njord-cli init`
- Uso: `njord`
- Configuração: explicar njord.yaml
- Shell wrapper

## Arquivos a criar/modificar

| Arquivo | Ação |
|---------|------|
| `cmd/njord/init.go` | **CRIAR** - Comando init com wizard |
| `cmd/njord/main.go` | **MODIFICAR** - Registrar init, remover defaultConfig hardcoded, adicionar ldflags vars |
| `.goreleaser.yaml` | **CRIAR** - Config GoReleaser |
| `.github/workflows/release.yml` | **CRIAR** - CI/CD release |
| `Makefile` | **CRIAR** - Build helpers |
| `README.md` | **CRIAR** - Documentação |

## Verificação

1. `njord-cli init` → wizard roda, cria config, instala shell wrapper
2. `njord` → funciona com config criada pelo init (sem projetos, grid vazio com Docker + Add)
3. Adicionar projeto via TUI → funciona
4. `make build` → compila binário
5. `make install` → instala em ~/.local/bin
6. `git tag v0.2.0 && git push --tags` → GitHub Action cria release com binários
7. Em outro PC: download binário → `njord-cli init` → `njord` funciona
