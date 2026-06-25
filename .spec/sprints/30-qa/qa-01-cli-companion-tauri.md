# QA 01 - CLI companion do njord-tauri

## Matriz

| Cenario | Status | Evidencia |
|---|---|---|
| `go test ./...` | PASS | todos os pacotes verdes |
| `go build -o njord ./cmd/njord/` | PASS | binario `njord` gerado |
| `njord tauri status --path /home/victorpersike/Persike/njord-tauri` | PASS | reconhece `njord-tauri@0.1.0` |
| `njord tauri dev --dry-run --path /home/victorpersike/Persike/njord-tauri` | PASS | imprime `npm run dev:clean` no root correto |
| `NJORD_TAURI_PATH=/home/victorpersike/Persike/njord-tauri ./njord tauri status` | PASS | env var tem precedencia quando `--path` nao existe |
| `./njord tauri status --path /tmp` | PASS | erro esperado para checkout invalido |
| `bash .claude/tools/spec-check.sh` | PASS | estrutura integra e 0 link quebrado |

## Comandos executados

```bash
/home/victorpersike/go/bin/gofmt -w cmd/njord/main.go cmd/njord/tauri.go internal/app/tauri/project.go internal/app/tauri/project_test.go
/home/victorpersike/go/bin/go test ./...
/home/victorpersike/go/bin/go build -o njord ./cmd/njord/
git diff --check
./njord tauri status --path /home/victorpersike/Persike/njord-tauri
./njord tauri dev --dry-run --path /home/victorpersike/Persike/njord-tauri
NJORD_TAURI_PATH=/home/victorpersike/Persike/njord-tauri ./njord tauri status
./njord tauri status --path /tmp
bash .claude/tools/spec-check.sh
```

## Evidencias

O comando `tauri status` imprime root, pacote, marcadores obrigatorios e scripts
do `package.json`. O comando `tauri dev --dry-run` imprime:

```text
cd -- '/home/victorpersike/Persike/njord-tauri' && npm run dev:clean
```

O path invalido retorna erro claro:

```text
validating /tmp/package.json: stat /tmp/package.json: no such file or directory
```

## Veredito

VERDICT=PASS.
