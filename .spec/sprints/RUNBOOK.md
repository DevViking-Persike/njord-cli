# Runbook de Sprints - njord-cli

## Ordem obrigatoria

1. `discovery`
2. `arquitetura design`
3. `desenvolvimento`
4. `arquitetura review`
5. `review-codigo-subagents`
6. `qa` + `qa-rpa` quando houver UI real
7. `seguranca` + `redteam` quando houver superficie dinamica

## Review de codigo

`review-codigo-subagents` e a sprint 25 da esteira. Ele roda depois de
`/arquitetura review` e antes de QA, como pipeline read-only por lanes/subagents.

Registre relatorios em `.spec/sprints/25-review-codigo/` usando nomes como:

```text
.spec/sprints/25-review-codigo/review-codigo-01-cli-companion-tauri.md
```

## Retomada

1. Ler `.spec/STATE.md`.
2. Abrir o documento da etapa atual.
3. Executar a proxima task pendente.
4. Atualizar checklists e estado.

## Gates bloqueantes

- Discovery sem criterio de aceitacao verificavel bloqueia arquitetura.
- Arquitetura design reprovada bloqueia desenvolvimento.
- Build ou teste vermelho bloqueia review.
- Arquitetura review reprovada volta para desenvolvimento.
- Review de codigo FAIL volta para desenvolvimento ou arquitetura, conforme a causa.
- QA FAIL volta para desenvolvimento.
- Critico/Alto em seguranca bloqueia release.

## Comandos padrao

```bash
go test ./...
go build -o njord ./cmd/njord/
bash .claude/tools/spec-check.sh
```

Use `make test` antes de release ou quando a mudanca tocar comportamento critico.

## Paradas obrigatorias

- Acesso a producao ou sistema de terceiro.
- Acao destrutiva em dados reais.
- Decisao estrutural sem registro.
- Necessidade de segredo real em teste.
- Gate reprovado duas vezes.
