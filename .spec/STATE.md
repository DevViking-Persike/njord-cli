# Spec State - njord-cli

## Incremento ativo

| Campo | Valor |
|---|---|
| Incremento | 02 |
| Tema | Scripts operacionais do njord-tauri |
| Modo scaffold | refatorar |
| Branch | atual |
| Etapa atual | incremento 02 concluido; pronto para abrir incremento 03 |
| Atualizado em | 2026-06-25 |

## Progresso da esteira

| Etapa | Status | Documento |
|---|---|---|
| 00 Discovery | concluido | `.spec/sprints/00-discovery/discovery-02-tauri-scripts.md` |
| 10a Arquitetura design | aprovado | `.spec/sprints/10-arquitetura/arquitetura-02-tauri-scripts.md` |
| 20 Desenvolvimento | concluido | `.spec/sprints/20-desenvolvimento/desenvolvimento-02-tauri-scripts.md` |
| 10b Arquitetura review | aprovado | `.spec/sprints/10-arquitetura/arquitetura-02-tauri-scripts.md` |
| 25 Review de Codigo | aprovado com avisos | `.spec/sprints/25-review-codigo/review-codigo-02-tauri-scripts.md` |
| 30 QA/RPA | PASS | `.spec/sprints/30-qa/qa-02-tauri-scripts.md` |
| 40 Seguranca/Redteam | PASS | `.spec/sprints/40-seguranca/seguranca-02-tauri-scripts.md` |

## Ultimo resultado de validacao

2026-06-25 — Reinstalacao da esteira a partir de
`/home/victorpersike/Persike/esteira-skills` (`d0c5be3`).

- Skills copiadas para `.claude/skills/`.
- `.codex/skills` confirmado como symlink para `../.claude/skills`.
- Rules, commands, deploy e `spec-check.sh` atualizados a partir dos templates
  de `scaffold-spec`.
- Sprint 25 `review-codigo-subagents` adicionada em
  `.spec/sprints/25-review-codigo/`.
- `bash .claude/tools/spec-check.sh`: PASS, estrutura integra e 0 link quebrado.

2026-06-25 — Incremento 01 fechado.

- `gofmt`: PASS.
- `go test ./...`: PASS.
- `go build -o njord ./cmd/njord/`: PASS.
- `git diff --check`: PASS.
- Arquivos Go tocados abaixo de 300 linhas.
- Smokes `tauri status`, `tauri dev --dry-run`, env var e path invalido: PASS.
- QA: VERDICT=PASS.
- Seguranca: PASS, 0 Critico/Alto aberto.

2026-06-25 — Incremento 02 fechado.

- `tauri check`, `tauri test` e `tauri build` adicionados.
- `gofmt`: PASS.
- `go test ./...`: PASS.
- `go build -o njord ./cmd/njord/`: PASS.
- `git diff --check`: PASS.
- Arquivos Go tocados abaixo de 300 linhas.
- Dry-runs `check/test/build`: PASS.
- Fixture com script ausente: PASS, erro esperado.
- QA: VERDICT=PASS.
- Seguranca: PASS, 0 Critico/Alto aberto.

## Pendencias

- Abrir incremento 03 para comando `tauri open` ou comando generico controlado.

## Itens aguardando aprovacao

- Nenhum.

## Historico de incrementos

| Incremento | Tema | Resultado |
|---|---|---|
| 01 | CLI companion do njord-tauri | PASS |
| 02 | Scripts operacionais do njord-tauri | PASS |

## Protocolo

- Atualizar este arquivo ao entrar ou sair de cada etapa.
- Gate reprovado nao avanca.
- Decisao estrutural deve virar nota em `.spec/reference/` ou ADR.
- Segredo nunca deve aparecer em relatorio, log ou fixture.
