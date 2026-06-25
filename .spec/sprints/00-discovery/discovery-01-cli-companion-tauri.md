# Discovery 01 - CLI companion do njord-tauri

## Contexto

O `njord-cli` e um CLI/TUI em Go com Cobra e Bubbletea. O `njord-tauri` e o
sucessor desktop GUI em Tauri 2 + SvelteKit + Rust, com backend modular e uma
`.spec` ativa. O objetivo deste incremento e iniciar a adaptacao do CLI para
atuar como companion operacional do app Tauri.

## Escopo

- Criar a base `.spec` do `njord-cli` em modo refatorar.
- Adicionar um grupo Cobra `tauri`.
- Implementar `tauri status` para localizar e validar um checkout do
  `njord-tauri`.
- Implementar `tauri dev` para executar o script de desenvolvimento do
  `njord-tauri` no diretorio correto.
- Cobrir a logica nova com testes unitarios.

## Fora de escopo

- Ler ou escrever o SurrealDB interno do `njord-tauri`.
- Portar a TUI Bubbletea para Rust/Tauri.
- Alterar o repo `njord-tauri`.
- Criar integracao IPC com app desktop rodando.

## Comportamento atual a preservar

- Rodar `njord-cli` sem subcomando continua abrindo a TUI existente.
- `migrate` continua migrando `data.sh` para `njord.yaml`.
- `push` continua delegando ao servico GitLab atual.
- Stdout continua reservado para comandos de shell quando a TUI seleciona um
  projeto.

## Criterios de aceitacao

- Given um checkout valido do `njord-tauri`, when `njord-cli tauri status` roda,
  then o comando imprime nome, versao, path e scripts relevantes.
- Given um path sem `package.json` ou `src-tauri/Cargo.toml`, when `tauri status`
  roda, then o comando retorna erro claro.
- Given `NJORD_TAURI_PATH` definido, when comandos `tauri` rodam, then esse path
  tem precedencia sobre a descoberta por diretorio irmao.
- Given `tauri dev --dry-run`, when o comando roda, then imprime o comando que
  executaria sem iniciar processo longo.
- Given o comando raiz sem subcomando, when executado, then o fluxo da TUI atual
  permanece intacto.

## Riscos e premissas

- Premissa: a primeira fatia deve ser operacional e reversivel, sem acoplar o Go
  ao armazenamento interno do app Tauri.
- Risco: rodar `tauri dev` abre processo longo; por isso a validacao automatica
  usa `--dry-run`.
- Risco: docs antigas do `njord-tauri` citam modulos removidos; a referencia
  usada deve ser `CLAUDE.md`, `.spec` e codigo atual.

## Veredito

Discovery aprovado para arquitetura design.
