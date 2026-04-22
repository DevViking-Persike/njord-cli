# njord-cli

TUI em Go para gerenciar projetos, GitLab e Docker.

## Build & Run

```bash
export PATH="$HOME/go/bin:/usr/local/go/bin:$PATH"
make install   # compila e instala em ~/.local/bin/njord-cli
njord          # no shell: chama a função que invoca o binário instalado
```

Detalhe importante: o comando `njord` é uma função do shell que roda `~/.local/bin/njord-cli`, não `./njord` do repo. Ver `.claude/rules/07-install-binary.md`.

## Estrutura

- `cmd/njord/main.go` - Entry point (cobra CLI)
- `internal/ui/` - TUI (bubbletea). `app.go` roteia telas, `grid.go` tela principal
- `internal/gitlab/` - Client GitLab API (`client.go` + `types.go`)
- `internal/config/` - Config YAML (`~/.config/njord/njord.yaml`)
- `internal/theme/` - Estilos lipgloss
- `internal/docker/` - Client Docker

## Convenções

- Idioma: pt-BR para UI e comentários
- Padrão bubbletea: Model/Init/Update/View
- Fetches async: `tea.Cmd` closure → retorna `tea.Msg` privada
- GitLab client: lazy init, criado no primeiro fetch ou navegação
- `gitlab_path` em Project mapeia para path no GitLab
- `pathToAlias`: mapa `gitlab_path → alias` construído iterando categories
- `timeAgo()` está em `internal/ui/gitlab_actions.go`

## Header da Grid

O header tem até 3 elementos lado a lado:
1. Box "Aprovações recentes" - pushes recentes com approval status
2. Box "MRs pendentes" - MRs abertos do usuário (scope=created_by_me)
3. Título "ᚾ N J O R D"

Boxes só aparecem quando têm dados. Layout se adapta automaticamente.

## go-gitlab

- Lib: `gitlab.com/gitlab-org/api/client-go` v1.46.0
- Instance-level MRs: `ListMergeRequests(opts)` - scopes: `created_by_me`, `assigned_to_me`, `all`
- Project-level MRs: `ListProjectMergeRequests(path, opts)`

## Regras de Engenharia

Regras completas em `.claude/rules/` — uma por arquivo, índice em `.claude/rules/README.md`.

Resumo:
1. **Tamanho** — arquivo `.go` ≤ 300 linhas
2. **Testes** — cobertura ≥ 84%, mutation ≥ 84% (rodam juntos via `make test`)
3. **SOLID** — SRP, OCP, LSP, ISP, DIP
4. **Clean Architecture** — `app` (negócio) → gateways → `ui` → `cmd`
5. **Simplicidade** — sem abstração prematura, sem flags booleanas
6. **Refatoração contínua** — teste antes de mexer, commit separado do bugfix
7. **Deploy** — `make install` depois de qualquer mudança testável via `njord`
8. **Delegar execução** — pedir pro usuário rodar comandos cujo output não influencia a próxima decisão (TUI, testes de UX, make install pós-rodada)

Skills: `/check-rules` (audita), `/refactor <arquivo>` (refatora guiado).

Violação exige justificativa explícita no commit/PR.
