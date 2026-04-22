# Regra 1 — Tamanho de arquivo

**Máximo 300 linhas por arquivo `.go`** (incluindo `_test.go`, excluindo linhas em branco).

## Motivação
Arquivo grande mistura responsabilidades, dificulta revisão e esconde acoplamento.

## Como aplicar
- Ao abrir um arquivo, se já tem > 280 linhas, refatorar antes de adicionar feature.
- Split por responsabilidade: `<feature>.go`, `<feature>_<subresponsibility>.go` (padrão já presente em `internal/ui/gitlab_actions_branch.go`).
- Teste também — se `_test.go` crescer, dividir por cenário (`foo_happy_test.go`, `foo_error_test.go`).

## Como verificar
```bash
find . -name '*.go' -not -path './vendor/*' -exec wc -l {} + | sort -rn | awk '$1 > 300'
```
