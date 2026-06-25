# Desenvolvimento 01 - CLI companion do njord-tauri

## Tasks

- [x] Criar pacote `internal/app/tauri`.
- [x] Adicionar testes unitarios da resolucao, validacao e status.
- [x] Registrar grupo Cobra `tauri`.
- [x] Implementar `tauri status`.
- [x] Implementar `tauri dev --dry-run`.
- [x] Rodar validacoes.

## Evidencias

- `internal/app/tauri/project.go`: resolucao de root, validacao de checkout,
  leitura de `package.json`, status e montagem do comando de dev.
- `internal/app/tauri/project_test.go`: cobre precedencia de `--path`,
  `NJORD_TAURI_PATH`, fallback por diretorio irmao, checkout invalido,
  `dev:clean` e status.
- `cmd/njord/tauri.go`: wiring Cobra do grupo `tauri`.
- `cmd/njord/main.go`: comando raiz preservado e abaixo do limite de 300 linhas.
- `gofmt`: PASS.
- `go test ./...`: PASS.
- `go build -o njord ./cmd/njord/`: PASS.
- `git diff --check`: PASS.
- `bash .claude/tools/spec-check.sh`: PASS.
- `./njord tauri status --path /home/victorpersike/Persike/njord-tauri`: PASS.
- `./njord tauri dev --dry-run --path /home/victorpersike/Persike/njord-tauri`: PASS.
- `NJORD_TAURI_PATH=/home/victorpersike/Persike/njord-tauri ./njord tauri status`: PASS.
- `./njord tauri status --path /tmp`: FAIL esperado com erro claro sobre
  `/tmp/package.json` ausente.

## Debitos

- Integracao com comandos internos do `njord-tauri` fica para incremento futuro.
- `tauri dev` sem `--dry-run` nao foi executado porque inicia processo longo.

## Resultado

PASS. Diff pronto para `10b Arquitetura review`.
