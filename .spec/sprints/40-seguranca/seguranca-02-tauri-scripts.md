# Seguranca 02 - Scripts operacionais do njord-tauri

## Escopo

Comandos `tauri check`, `tauri test` e `tauri build`, especialmente modo
`--dry-run`.

## Checks

| Check | Status | Nota |
|---|---|---|
| Sem impressao de secrets | PASS | dry-run imprime apenas path e script |
| Script validado antes de execucao | PASS | script ausente retorna erro antes de `npm` |
| Processo longo nao roda em teste | PASS | smoke usa `--dry-run` |
| Execucao restrita ao root validado | PASS | root passa por `LoadProject` |

## Achados

Nenhum Critico/Alto aberto.

LOW: comandos sem `--dry-run` delegam para `npm` no checkout validado. Esse e o
comportamento esperado da CLI companion.

## Veredito

PASS.
