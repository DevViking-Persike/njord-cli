# Desenvolvimento 02 - Scripts operacionais do njord-tauri

## Tasks

- [x] Adicionar montagem validada de scripts em `internal/app/tauri`.
- [x] Cobrir scripts e erro ausente com testes unitarios.
- [x] Registrar `tauri check`, `tauri test` e `tauri build`.
- [x] Reutilizar helper de execucao/dry-run.
- [x] Rodar validacoes.

## Evidencias

- `internal/app/tauri/project.go`: `BuildScriptCommand` valida existencia do
  script em `package.json` antes de montar `npm run <script>`.
- `internal/app/tauri/project_test.go`: cobre script existente e script ausente.
- `cmd/njord/tauri.go`: adiciona subcomandos `check`, `test` e `build` com
  `--dry-run` e helper comum de execucao.
- `gofmt`: PASS.
- `go test ./...`: PASS.
- `go build -o njord ./cmd/njord/`: PASS.
- `git diff --check`: PASS.
- `wc -l`: todos os arquivos tocados abaixo de 300 linhas.
- `./njord tauri check --dry-run --path /home/victorpersike/Persike/njord-tauri`: PASS.
- `./njord tauri test --dry-run --path /home/victorpersike/Persike/njord-tauri`: PASS.
- `./njord tauri build --dry-run --path /home/victorpersike/Persike/njord-tauri`: PASS.
- Fixture temporaria sem script `build`: FAIL esperado com erro claro.

## Debitos

- Interpretacao estruturada da saida de `npm` fica para incremento futuro.
- Comandos reais sem `--dry-run` nao foram executados para evitar processos
  longos neste gate.

## Resultado

PASS. Diff pronto para `10b Arquitetura review`.
