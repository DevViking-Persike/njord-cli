# Seguranca 01 - CLI companion do njord-tauri

## Escopo

Comandos `tauri status` e `tauri dev --dry-run`.

## Checks

| Check | Status | Nota |
|---|---|---|
| Sem impressao de secrets | PASS | `status` mostra metadados de projeto e scripts, sem tokens |
| Path validado antes de execucao | PASS | exige `package.json` e `src-tauri/Cargo.toml` |
| Processo longo nao roda em teste | PASS | validacao usa `--dry-run` |
| Sem leitura do SurrealDB interno | PASS | nao acessa dados internos do app |

## Achados

Nenhum Critico/Alto aberto.

LOW: `tauri dev` executa `npm` no checkout validado. Esse e o comportamento
esperado, mas deve permanecer restrito a um root reconhecido como `njord-tauri`.

## Veredito

PASS.
