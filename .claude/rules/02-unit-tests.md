# Regra 2 — Testes unitários

Toda função pública nova precisa de teste. Funções privadas relevantes também.

## Critérios
- Cobertura por pacote: **≥ 70%**
- Eficácia de mutation testing (`gremlins`): **≥ 70%**
- Quebra de teste bloqueia commit — nunca desabilite / `t.Skip` para passar CI.
- `t.TempDir()` para filesystem; nunca escreva fora do tempdir.
- Estilo table-driven quando há múltiplos casos.

## Como verificar
```bash
make test          # roda tudo
make coverage      # % por pacote
make mutation-docker
make mutation-app
```

Pacotes sem `_test.go` são violação automática — checar com:
```bash
go test ./... 2>&1 | grep '\[no test files\]'
```

## Exceções aceitas
- `cmd/njord/main.go`: entry point fino (composition root). Lógica deve estar em `internal/app`.
- `internal/theme/`: constantes lipgloss, sem lógica.
- `internal/ui/components/render.go`: apresentação pura (testar via snapshot se crescer).
