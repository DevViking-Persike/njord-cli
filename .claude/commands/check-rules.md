---
description: Auditar o repositório contra as regras em .claude/rules/
---

Rode uma auditoria do repositório contra as regras de engenharia em `.claude/rules/`.

Leia os arquivos de regras antes de começar:
- `.claude/rules/01-file-size.md`
- `.claude/rules/02-unit-tests.md`
- `.claude/rules/03-solid.md`
- `.claude/rules/04-clean-architecture.md`
- `.claude/rules/05-simplicity.md`
- `.claude/rules/06-continuous-refactoring.md`
- `.claude/rules/07-install-binary.md`

## Checagens

### Regra 1 — Tamanho
```bash
find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | sort -rn | awk '$1 > 300'
```
Liste violações ordenadas. Sugira split por responsabilidade (sem editar).

### Regra 2 — Testes
```bash
go test -cover ./...
go test ./... 2>&1 | grep '\[no test files\]'
gremlins unleash ./internal/app/
gremlins unleash ./internal/docker/
```
Reporte pacotes < 70% de cobertura e eficácia de mutation < 70%.

### Regra 3 — SOLID (sinais automáticos)
```bash
rg -l '"os/exec"' internal/ui/ internal/app/
```
Qualquer hit é candidato a mover para gateway.

### Regra 4 — Clean Architecture (imports)
```bash
rg -l 'njord-cli/internal/ui' internal/app/
rg -l 'njord-cli/internal/(ui|app)' internal/docker/ internal/gitlab/ internal/git/
rg -l 'njord-cli/cmd' internal/
```
Qualquer hit é violação.

### Regra 5 — Simplicidade
- 5 maiores funções por linhas em `internal/ui/` — destacar (não editar).
- Amostragem de comentários que descrevem o *quê*.

### Build sanity
```bash
go vet ./...
go build ./cmd/njord/
```

## Formato do relatório

Markdown, uma seção por regra:
- ✅ conforme / ⚠️ violação pequena / ❌ bloqueante
- Contagem agregada no topo
- Top 3 próximos passos priorizados

**Não edite arquivos.** Para corrigir, o usuário chama `/refactor <arquivo>`.
