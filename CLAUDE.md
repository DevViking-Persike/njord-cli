# njord-cli

TUI em Go para gerenciar projetos, GitLab e Docker.

## Build & Run

```bash
export PATH="$HOME/go/bin:/usr/local/go/bin:$PATH"
go build ./cmd/njord/
./njord
```

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

Aplicam-se a qualquer mudança no repositório. Violação exige justificativa explícita no commit/PR.

### 1. Tamanho de arquivo
- **Máximo 300 linhas por arquivo `.go`** (excluindo comentários em branco).
- Se passar, refatorar por responsabilidade antes de commitar. Separar em arquivos coesos (`<feature>.go`, `<feature>_<subresponsibility>.go`).
- Teste (`*_test.go`) segue a mesma regra; se crescer, dividir por cenário.

### 2. Testes unitários
- **Toda função pública nova precisa de teste.** Funções privadas relevantes também.
- Quebra de teste bloqueia commit — nunca desabilite/`t.Skip` para passar CI.
- Alvo de cobertura por pacote: **≥ 70%**; eficácia de mutation (`make mutation-<pkg>`) **≥ 70%**.
- Testes devem seguir estilo *table-driven* quando há múltiplos casos.
- Use `t.TempDir()` para filesystem; nunca escreva fora do tempdir do teste.

### 3. SOLID
- **SRP**: um arquivo, um motivo para mudar. Separar regras de negócio de I/O, UI de estado.
- **OCP**: prefira injetar dependências (interfaces) a importar struct concreta quando o ponto de extensão é previsível (ex.: `gitlab.Client` injetado no TUI).
- **LSP**: interfaces pequenas; não quebre contratos em implementações.
- **ISP**: uma interface por papel. Evite interfaces "gordas" agregando responsabilidades.
- **DIP**: pacotes de alto nível (`internal/ui`, `internal/app`) dependem de abstrações, não de `os/exec` ou SDKs diretamente — coloque a chamada concreta em `internal/{docker,gitlab,git}` e injete.

### 4. Clean Architecture
- Camadas (de dentro pra fora): `internal/app` (regras de negócio) → `internal/{docker,gitlab,git,config}` (gateways/infra) → `internal/ui` (entrega) → `cmd/njord` (composition root).
- **Fluxo de dependência aponta sempre pra dentro.** `internal/app` nunca importa `internal/ui`; `internal/ui` nunca importa `cmd`.
- Lógica de negócio **não fica em TUI**. Se `internal/ui/*.go` tem regra que não é apresentação, extrair pra `internal/app/`.
- Efeitos colaterais (exec, HTTP, FS) ficam em gateways concretos — mockáveis via interface.

### 5. Simplicidade
- **Não antecipe abstração.** 3 linhas duplicadas são melhores que uma abstração prematura.
- Sem flags booleanas que mudam comportamento interno da função — prefira duas funções.
- Sem camadas wrapper "por segurança" (interface → struct → interface). Uma indireção resolve.
- Sem comentários que descrevem o *quê* — só o *porquê* não-óbvio (bug conhecido, invariante sutil).
- Pt-BR em nomes de UI e mensagens; inglês em identificadores de código e erros internos.

### 6. Refatoração contínua
- Antes de adicionar feature em arquivo > 280 linhas: refatorar primeiro, feature depois.
- Antes de tocar função sem teste: escrever teste que cobre o comportamento atual, depois modificar.
- Nunca deletar teste "que estava passando". Se obsoleto, substituir por equivalente.

### 7. Invocação
- Auditoria: `/check-rules` (skill local) roda todas as checagens automatizáveis (tamanho, cobertura, mutation, `go vet`).
- Refatoração guiada: `/refactor <arquivo>` (skill local) analisa violações e propõe plano antes de editar.
