# Arquitetura 02 - Scripts operacionais do njord-tauri

## Entrada

Discovery aprovado em `.spec/sprints/00-discovery/discovery-02-tauri-scripts.md`.

## Plano tecnico

- Adicionar em `internal/app/tauri` uma funcao para montar `npm run <script>`
  validando se o script existe no `package.json`.
- Manter `BuildDevCommand` como caso especial que prefere `dev:clean`.
- Adicionar testes unitarios para scripts `check`, `test`, `build` e script
  ausente.
- Em `cmd/njord/tauri.go`, registrar subcomandos `check`, `test` e `build`.
- Reutilizar helper comum para executar/dry-run `CommandSpec`.

## Camadas e contratos

| Camada | Responsabilidade |
|---|---|
| `cmd/njord/tauri.go` | wiring Cobra, flags, IO e execucao de processo |
| `internal/app/tauri` | validacao de checkout e montagem de comandos npm |
| `package.json` do `njord-tauri` | fonte de verdade dos scripts disponiveis |

Contratos:

- `--path` > `NJORD_TAURI_PATH` > `../njord-tauri`.
- Script ausente retorna erro antes de executar processo.
- `--dry-run` nao executa `npm`.
- Comandos reais sempre rodam com `process.Dir` igual ao root validado.

## Riscos

- Processo longo em `build`; mitigado com `--dry-run` nos smokes.
- Dependencia de `npm`; a CLI delega e propaga erro.
- Saida de `npm` nao e parseada neste incremento.

## ADR

Nao necessario. Mantem a decisao do incremento 01: CLI companion por scripts e
contratos explicitos, sem acoplar ao storage interno.

## Veredito design

Aprovado.

## Veredito review

PASS.

## Review do diff

- Camadas: PASS. A regra de montagem/validacao de scripts esta em
  `internal/app/tauri`; `cmd/njord/tauri.go` faz wiring e execucao.
- Contrato de scripts: PASS. O `package.json` do `njord-tauri` e a fonte de
  verdade; script ausente falha antes de `npm`.
- Dry-run: PASS. `check`, `test` e `build` imprimem o comando sem executar.
- Tamanho de arquivo: PASS. Arquivos tocados permanecem abaixo de 300 linhas.
- Seguranca: PASS. Sem leitura de secrets, sem SurrealDB, execucao restrita ao
  root validado.

Debito registrado: saida estruturada de `npm` e comandos reais completos ficam
para ciclos futuros.
