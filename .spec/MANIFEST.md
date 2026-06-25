# Spec Manifest - njord-cli

Este arquivo e o ponto de entrada operacional para adaptar o `njord-cli` como
CLI companion do `njord-tauri`.

## Bootstrap de sessao

1. Ler `CLAUDE.md`.
2. Ler `.spec/MANIFEST.md`.
3. Ler `.spec/STATE.md`.
4. Ler `.spec/sprints/RUNBOOK.md`.
5. Ler a disciplina ou incremento ativo.

## Regra-mae

O `njord-cli` deve evoluir como CLI companion do `njord-tauri`, preservando o
comportamento existente de TUI/wrapper enquanto adiciona comandos pequenos,
testaveis e sem acesso direto a persistencia interna SurrealDB do app Tauri.

## Mapa do `.spec`

| Caminho | Papel |
|---|---|
| `.spec/MANIFEST.md` | ponto de entrada operacional |
| `.spec/STATE.md` | estado vivo do incremento atual |
| `.spec/reference/` | inventario, decisoes e referencias do projeto |
| `.spec/sprints/RUNBOOK.md` | ordem dos gates e retomada |
| `.spec/sprints/00-discovery/` | discovery e criterios de aceitacao |
| `.spec/sprints/10-arquitetura/` | design gate e review gate |
| `.spec/sprints/20-desenvolvimento/` | tasks e evidencias de dev |
| `.spec/sprints/25-review-codigo/` | review de codigo por lanes/subagents |
| `.spec/sprints/30-qa/` | plano e relatorio de QA |
| `.spec/sprints/40-seguranca/` | verificacoes de seguranca |

## Disciplinas

| Etapa | Onde olhar |
|---|---|
| 00 Discovery | `.spec/sprints/00-discovery/README.md` |
| 10 Arquitetura | `.spec/sprints/10-arquitetura/README.md` |
| 20 Desenvolvimento | `.spec/sprints/20-desenvolvimento/README.md` |
| 25 Review de Codigo | `.spec/sprints/25-review-codigo/README.md` |
| 30 QA | `.spec/sprints/30-qa/README.md` |
| 40 Seguranca | `.spec/sprints/40-seguranca/README.md` |

## Regras de execucao

| Tema | Fonte |
|---|---|
| Tamanho de arquivo | `.claude/rules/01-file-size.md` |
| Testes | `.claude/rules/02-unit-tests.md` |
| SOLID | `.claude/rules/03-solid.md` |
| Clean Architecture | `.claude/rules/04-clean-architecture.md` |
| Simplicidade | `.claude/rules/05-simplicity.md` |
| Refatoracao continua | `.claude/rules/06-continuous-refactoring.md` |
| Build/install | `.claude/rules/07-install-binary.md` |
| Delegar execucao | `.claude/rules/08-delegate-execution.md` |
| Fluxo de desenvolvimento | `.claude/rules/fluxo-desenvolvimento.md` |
| Seguranca | `.claude/rules/seguranca.md` |

## Skills da esteira

| Skill | Quando usar | Registro |
|---|---|---|
| `review-codigo-subagents` | sprint 25, depois de `/arquitetura review` e antes de QA | `.spec/sprints/25-review-codigo/` |

## Maquinario de validacao

```bash
go test ./...
go build -o njord ./cmd/njord/
bash .claude/tools/spec-check.sh
```

`make test` tambem existe, mas inclui mutation testing e pode ser usado quando
o incremento mexer em comportamento critico ou antes de release.

## Projetos relacionados

| Projeto | Papel |
|---|---|
| `/home/victorpersike/Persike/njord-cli` | repo atual; CLI Go/Cobra/Bubbletea |
| `/home/victorpersike/Persike/njord-tauri` | app alvo; Tauri 2 + SvelteKit + Rust |
