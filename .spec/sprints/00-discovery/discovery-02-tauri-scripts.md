# Discovery 02 - Scripts operacionais do njord-tauri

## Contexto

O incremento 01 criou o grupo `tauri` com `status` e `dev --dry-run`. O proximo
passo para transformar o `njord-cli` em CLI companion do `njord-tauri` e permitir
rodar comandos operacionais ja definidos no `package.json` do app Tauri, sem
duplicar logica de build/teste no Go.

## Escopo

- Adicionar comandos `tauri check`, `tauri test` e `tauri build`.
- Reutilizar a validacao de checkout do incremento 01.
- Cada comando deve delegar para `npm run <script>` no root do `njord-tauri`.
- Cada comando deve aceitar `--dry-run` para validacao sem processo longo.
- Cobrir montagem de comandos e erro de script ausente com testes unitarios.

## Fora de escopo

- Rodar `npm run build` real durante QA automatizado, porque pode ser demorado.
- Instalar dependencias do `njord-tauri`.
- Interpretar saida de `npm`.
- Acessar internals Rust/Tauri, SurrealDB ou processos da GUI.

## Comportamento atual a preservar

- `tauri status` continua funcionando.
- `tauri dev --dry-run` continua preferindo `dev:clean`.
- `njord-cli` sem subcomando continua abrindo a TUI.
- `migrate` e `push` continuam inalterados.

## Criterios de aceitacao

- Given checkout valido, when `njord tauri check --dry-run --path <root>` roda,
  then imprime `cd -- '<root>' && npm run check`.
- Given checkout valido, when `njord tauri test --dry-run --path <root>` roda,
  then imprime `cd -- '<root>' && npm run test`.
- Given checkout valido, when `njord tauri build --dry-run --path <root>` roda,
  then imprime `cd -- '<root>' && npm run build`.
- Given package sem script requerido, when o comando correspondente roda, then
  retorna erro claro de script ausente.
- Given `NJORD_TAURI_PATH` definido, when comandos rodam sem `--path`, then usam
  a env var como no incremento 01.

## Riscos e premissas

- Premissa: os scripts oficiais do `njord-tauri` continuam sendo fonte de
  verdade operacional.
- Risco: comandos reais podem ser longos ou abrir toolchain externa; QA local usa
  `--dry-run` para smoke e testes unitarios para regra.
- Risco: futuras mudancas de scripts no `package.json`; mitigado por validacao
  dinamica de existencia do script.

## Veredito

Discovery aprovado para arquitetura design.
