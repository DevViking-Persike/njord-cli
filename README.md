# ᚾ Njord CLI

TUI (Terminal User Interface) em Go para gerenciar projetos, Docker stacks e integrar com GitLab. Navegue entre repositorios, gerencie containers, acompanhe pipelines e crie branches padronizadas — tudo sem sair do terminal.

## Stack

- **Go 1.23+**
- **Bubbletea** + **Lipgloss** — TUI framework
- **Cobra** — CLI commands
- **Koanf** — YAML config
- **Docker SDK v27** — gerenciamento de containers
- **GitLab API client-go** — integracao com GitLab

## Instalacao

### Build local

```bash
go build -o ~/.local/bin/njord-cli ./cmd/njord/
```

### Shell wrapper

O Njord precisa de um wrapper no shell para executar comandos como `cd` no contexto do terminal. Adicione ao seu `~/.zshrc` ou `~/.bashrc`:

```bash
njord() {
    local result
    result=$(njord-cli "$@" 2>/dev/tty)
    local code=$?
    if [[ $code -eq 0 && -n "$result" ]]; then
        eval "$result"
    fi
}
```

Depois: `source ~/.zshrc` e use `njord` para abrir a TUI.

## Configuracao

Config em `~/.config/njord/njord.yaml`:

```yaml
settings:
  editor: "code"          # code, cursor, nvim, vim
  projects_base: "~/Avita"
  personal_base: "~/Persike"

gitlab:
  token: "glpat-xxxx"     # Personal Access Token (scope: api)
  url: ""                 # opcional, default: https://gitlab.com

categories:
  - id: financeiro
    name: "Financeiro"
    sub: "Modulos financeiros"
    projects:
      - alias: avita-fin
        desc: "SGA Modulo Financeiro Frontend"
        path: "sga-modulo-financeiro-angular-typescript"
        group: "frontend"                                    # opcional
        gitlab_path: "avitaseg/bill/bibliotecas/angular/..." # opcional

docker_stacks:
  - name: "GAP Stack"
    desc: "MySQL Database (porta 3306)"
    path: "gap-stack-desenvolvimento"
```

### Campos do projeto

| Campo | Descricao |
|-------|-----------|
| `alias` | Nome curto exibido na TUI |
| `desc` | Descricao do projeto |
| `path` | Caminho relativo a `projects_base` (ou absoluto com `~/`) |
| `group` | Agrupamento visual na lista (opcional) |
| `gitlab_path` | Path do projeto no GitLab, ex: `grupo/subgrupo/repo` (opcional) |

### Paths especiais

- `@rdp` — Abre conexao RDP via Cloudflare Tunnel
- `env/...` — Projetos em subdiretorio `env/` do `projects_base`
- `Persike/...` — Projetos pessoais resolvidos a partir de `~/`

## Funcionalidades

### Grid principal

A tela inicial exibe um grid de cards:

- **Categorias** — Cada categoria do config aparece como card com contagem de projetos
- **Todos** — Card especial que lista todos os projetos de todas as categorias
- **Docker** — Gerenciamento de Docker stacks
- **GitLab** — Integracao com GitLab (MRs, pipelines, branches)
- **+ Adicionar** — Wizard para adicionar novo projeto
- **Configuracoes** — Editar settings, categorias, token GitLab

### Header

- Titulo **ᚾ N J O R D** no canto direito
- Box **"Aprovacoes recentes"** no canto esquerdo (se GitLab configurado):
  - Mostra projetos com push nas ultimas 6 horas
  - Icone de aprovacao: `✓` aprovado ou `⏳ 0/1 Code Review B1` pendente

### Projetos

Ao selecionar uma categoria, a lista de projetos aparece agrupada por `group`. Selecionar um projeto executa:

```
cd "<projects_base>/<path>" && <editor> .
```

### Docker

- Lista todas as stacks configuradas com status dos containers (running/stopped)
- Acoes por stack: **Up**, **Down**, **Restart**, **Logs**
- Opcao de adicionar nova stack

### GitLab

Requer `gitlab.token` configurado (Personal Access Token com scope `api`).

#### Lista de projetos

- Mostra apenas projetos com `gitlab_path` configurado
- Icone de status da pipeline mais recente:
  - `✓` success (verde)
  - `✗` failed (vermelho)
  - `◐` running/pending (spinner animado)
  - `⊘` blocked/canceled
  - `○` desconhecido
- Icone de aprovacao do MR aberto:
  - `✓ aprovado` (verde)
  - `⏳ 0/1 Code Review B1` (amarelo, com nome da regra)
- Lista ordenada por atividade mais recente

#### Acoes por projeto

1. **Merge Requests** — Lista MRs abertos com status, branch, autor e tempo
2. **Pipelines** — Lista pipelines recentes filtradas pelo seu usuario
3. **Disparar Pipeline** — Seleciona branch e dispara pipeline
4. **Criar Branch** — Fluxo padronizado Jira (ver abaixo)
5. **Abrir no Navegador** — Abre o projeto no GitLab via browser

### Criar Branch (fluxo Jira)

O fluxo de criacao de branch segue a convencao:

```
feature/<SIGLA>-<NUMERO>-<EQUIPE>-<TIPO>-<descricao>
```

**Exemplo:** `feature/BILL-1633-B1-subtask-ajuste-modulo-financeiro`

#### Passos

1. **Selecionar sigla Jira** — Lista de equipes:

| Equipe | Sigla | Codigo |
|--------|-------|--------|
| Plataforma | PLA | A1 |
| Billing - Financeiro | BILL | B1 |
| Gestao de Apolice | SIE | C1 |
| Consistencia dos Dados | GAP | D1 |
| Backoffice | SBO | E1 |
| Ops - Novos Clientes | FOPS | F1 |
| Hotfix | HOT | H1 |
| Low Priority | LOW | L1 |
| Suporte | SPAVT | S1 |

2. **Digitar numero do ticket** — Apenas digitos, preview em tempo real
3. **Selecionar tipo** — `delivery` ou `subtask`
4. **Digitar descricao** — Auto-normalizada:
   - Converte para lowercase
   - Remove acentos (ç→c, ã→a, é→e, etc.)
   - Substitui espacos por hifens
   - Remove caracteres especiais
5. **Selecionar branch base** — Lista ordenada por mais recente, com icones de aprovacao

#### Lista de branches

A lista de branches exibe:

- Nome da branch
- Tags: `[default]`, `[protected]`
- Aprovacao do MR (se existir): `✓ aprovado` ou `⏳ 0/1 Code Review B1`
- Tempo desde o ultimo commit

### Configuracoes

Via menu Settings:

1. **Editor** — code, cursor, nvim, vim, custom
2. **Projects base** — Diretorio base dos projetos
3. **Personal base** — Diretorio base pessoal
4. **Adicionar categoria** — Nova categoria de projetos
5. **GitLab Token** — Configurar/atualizar o PAT

## Navegacao

| Tecla | Acao |
|-------|------|
| `↑` `↓` `←` `→` / `h` `j` `k` `l` | Navegar |
| `Enter` | Selecionar |
| `Esc` | Voltar |
| `q` | Sair |
| `Ctrl+C` | Sair forcado |

## Estrutura do projeto

```
njord-cli/
├── cmd/njord/
│   └── main.go              # Entry point, cobra commands
├── internal/
│   ├── config/
│   │   ├── config.go         # Config structs, Load/Save, marshal
│   │   └── migrate.go        # Migracao de data.sh legado
│   ├── docker/
│   │   └── client.go         # Docker SDK wrapper
│   ├── gitlab/
│   │   ├── client.go         # GitLab API client
│   │   ├── remote.go         # Git remote URL parser
│   │   └── types.go          # GitLab data types
│   ├── theme/
│   │   └── theme.go          # Lipgloss styles e cores
│   └── ui/
│       ├── app.go            # AppModel principal, wiring de telas
│       ├── grid.go           # Grid de cards na tela inicial
│       ├── projects.go       # Lista de projetos
│       ├── docker.go         # Tela Docker stacks
│       ├── docker_actions.go # Acoes Docker (up/down/restart/logs)
│       ├── gitlab.go         # Lista projetos GitLab
│       ├── gitlab_actions.go # Acoes GitLab (MRs, pipelines, branches)
│       ├── add_project.go    # Wizard adicionar projeto
│       ├── add_stack.go      # Wizard adicionar stack
│       └── settings.go       # Tela de configuracoes
└── go.mod
```

## GitLab Token

Para obter um Personal Access Token:

1. Acesse **GitLab** → **Settings** → **Access Tokens**
2. Crie um token com scope **`api`**
3. O token começa com `glpat-`
4. Configure via Settings no Njord ou edite `~/.config/njord/njord.yaml`

## Licenca

Uso interno.
