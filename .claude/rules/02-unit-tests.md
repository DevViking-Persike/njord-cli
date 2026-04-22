# Regra 2 — Testes unitários + mutation

Toda função pública nova precisa de teste. Funções privadas relevantes também.

## Critérios (bloqueantes)
- Cobertura por pacote testável: **≥ 84%**
- Eficácia de mutation testing (`gremlins`): **≥ 84%**
- **Mutation roda sempre junto com os testes** (`make test` executa ambos). Se a eficácia cair abaixo de 84%, o teste precisa ser fortalecido antes do commit.
- Quebra de teste bloqueia commit — nunca desabilite / `t.Skip` para passar CI.
- `t.TempDir()` para filesystem; nunca escreva fora do tempdir.
- Estilo table-driven quando há múltiplos casos.

## Como verificar
```bash
make test          # roda unit tests + mutation (ambos obrigatórios)
make test-unit     # só unit tests — uso durante dev loop
make coverage      # % por pacote
```

Pacotes sem `_test.go` são violação automática — checar com:
```bash
go test ./... 2>&1 | grep '\[no test files\]'
```

## Tratando mutantes sobreviventes (LIVED)
Quando `gremlins` reporta `LIVED CONDITIONALS_NEGATION at <arquivo>:<linha>`:
1. Abrir o arquivo na linha indicada.
2. Identificar qual condição não é coberta por teste.
3. Adicionar caso de teste que falharia se a condição invertesse.
4. Rodar `make test` de novo — esperar eficácia ≥ 84%.

## Exceções aceitas (não contam para o threshold)
- `cmd/njord/main.go`: entry point fino (composition root). Lógica deve estar em `internal/app`.
- `internal/theme/`: constantes lipgloss, sem lógica.
- `internal/ui/components/render.go`: apresentação pura (testar via snapshot se crescer).
- Pacotes com ≥ 80% de chamadas a SDK externo (ex.: `internal/docker`, `internal/gitlab`) — aplicar threshold apenas nas funções puras do pacote.
