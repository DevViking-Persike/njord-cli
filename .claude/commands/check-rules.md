---
description: Auditar o repositório contra as regras de engenharia do CLAUDE.md
---

Rode uma auditoria do repositório contra as **Regras de Engenharia** descritas em `CLAUDE.md`.

## O que checar

1. **Tamanho de arquivo** (regra 1 — máx 300 linhas)
   - Liste todo `.go` (incluindo `_test.go`) acima de 300 linhas, ordenado desc:
     `find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | sort -rn | awk '$1 > 300'`
   - Para cada violação, sugira um plano de split por responsabilidade (não edite ainda).

2. **Testes** (regra 2 — cobertura ≥ 70%, mutation ≥ 70%)
   - `go test -cover ./...` — liste pacotes < 70%.
   - Pacotes SEM `_test.go` (verificar via `go test ./... 2>&1 | grep '\[no test files\]'`) são violação.
   - Para pacotes `internal/app` e `internal/docker` (núcleo refatorável), rode:
     `gremlins unleash ./internal/app/` e `./internal/docker/` — reporte eficácia.

3. **SOLID + Clean Architecture** (regras 3 e 4)
   - Grep por violações de dependência:
     - `internal/app` importando `internal/ui` → VIOLAÇÃO (`rg -l 'njord-cli/internal/ui' internal/app/`)
     - `internal/ui` importando `cmd/` → VIOLAÇÃO
     - `internal/{docker,gitlab,git}` importando `internal/ui` ou `internal/app` → VIOLAÇÃO
   - Procure `os/exec` ou chamadas SDK fora de `internal/{docker,gitlab,git}/`:
     `rg -l '"os/exec"' internal/ui/ internal/app/` (se houver hit → lógica de I/O vazando da camada certa).

4. **Simplicidade** (regra 5)
   - Funções muito longas (> 60 linhas) em `internal/ui/`:
     `gopls` ou análise manual — destaque as 5 maiores.
   - Arquivos com comentários em excesso que descrevem o *quê* (não o *porquê*) — amostragem de 3.

5. **Build sanity**
   - `go vet ./...` deve passar sem warnings.
   - `go build ./cmd/njord/` deve compilar.

## Formato do relatório

Emita em markdown, seção por regra, com:
- ✅ conforme / ⚠️ violação pequena / ❌ violação a corrigir antes do próximo PR
- Contagem agregada no topo ("7 violações: 5 tamanho, 1 cobertura, 1 arquitetura")
- Plano priorizado de correção (top 3 próximos passos), mas **não edite arquivos** — só diagnóstico.

Se o usuário pedir para corrigir a seguir, use `/refactor <arquivo>` para cada alvo.
