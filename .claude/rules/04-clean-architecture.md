# Regra 4 — Clean Architecture

## Camadas (de dentro pra fora)
1. `internal/app/` — regras de negócio puras (testáveis sem infra)
2. `internal/{docker,gitlab,git,config}/` — gateways/infra (exec, HTTP, FS)
3. `internal/ui/` — entrega (bubbletea TUI)
4. `cmd/njord/` — composition root (wiring)

## Regras de dependência
- **Fluxo aponta sempre para dentro**: ui → app → (nada externo importa app).
- `internal/app` nunca importa `internal/ui`.
- `internal/ui` nunca importa `cmd/`.
- `internal/{docker,gitlab,git}` nunca importa `internal/ui` nem `internal/app`.

## Onde colocar o quê
- **Regra de negócio** (ex.: "se branch é subtask, disparar pipeline depois do push"): `internal/app/`.
- **Chamada ao Docker/GitLab/Git/FS**: gateway correspondente.
- **Renderização e estado de tela**: `internal/ui/`.
- **Wiring/bootstrap** (cobra, carregar config, injetar dependências): `cmd/njord/main.go`.

## Teste seco
Se `internal/app/*.go` importa `os/exec`, `net/http`, `github.com/docker/...`, é violação — mover a chamada pra gateway.

## Como verificar
```bash
rg -l 'njord-cli/internal/ui' internal/app/
rg -l 'njord-cli/internal/(ui|app)' internal/docker/ internal/gitlab/ internal/git/
```
Saída esperada: vazia.
