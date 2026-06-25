# QA 02 - Scripts operacionais do njord-tauri

## Matriz

| Cenario | Status | Evidencia |
|---|---|---|
| `go test ./...` | PASS | todos os pacotes verdes |
| `go build -o njord ./cmd/njord/` | PASS | binario gerado |
| `njord tauri check --dry-run --path /home/victorpersike/Persike/njord-tauri` | PASS | imprime `npm run check` |
| `njord tauri test --dry-run --path /home/victorpersike/Persike/njord-tauri` | PASS | imprime `npm run test` |
| `njord tauri build --dry-run --path /home/victorpersike/Persike/njord-tauri` | PASS | imprime `npm run build` |
| script ausente em fixture temporaria | PASS | erro esperado antes de executar `npm` |
| `bash .claude/tools/spec-check.sh` | PASS | estrutura integra e 0 link quebrado |

## Comandos executados

```bash
/home/victorpersike/go/bin/gofmt -w cmd/njord/tauri.go internal/app/tauri/project.go internal/app/tauri/project_test.go
/home/victorpersike/go/bin/go test ./...
/home/victorpersike/go/bin/go build -o njord ./cmd/njord/
git diff --check
./njord tauri check --dry-run --path /home/victorpersike/Persike/njord-tauri
./njord tauri test --dry-run --path /home/victorpersike/Persike/njord-tauri
./njord tauri build --dry-run --path /home/victorpersike/Persike/njord-tauri
NJORD_TAURI_PATH=/home/victorpersike/Persike/njord-tauri ./njord tauri check --dry-run
bash .claude/tools/spec-check.sh
```

## Evidencias

Dry-runs:

```text
cd -- '/home/victorpersike/Persike/njord-tauri' && npm run check
cd -- '/home/victorpersike/Persike/njord-tauri' && npm run test
cd -- '/home/victorpersike/Persike/njord-tauri' && npm run build
```

Erro de script ausente:

```text
script "build" not found in /tmp/<fixture>/package.json
```

## Veredito

VERDICT=PASS.
