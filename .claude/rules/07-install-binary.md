# Regra 7 — Deploy do binário para teste local

O comando `njord` no shell do usuário é uma **função** definida em `.zshrc`/`.bashrc` que invoca `~/.local/bin/njord-cli`, não o binário local `./njord` do repo.

```sh
njord () {
    local result
    result=$(~/.local/bin/njord-cli "$@" 2>/dev/tty)
    local code=$?
    if [[ $code -eq 0 && -n "$result" ]]; then
        eval "$result"
    fi
}
```

A função existe porque o `njord` pode emitir um comando shell no stdout (ex.: `cd "/path" && code .`) que o wrapper faz `eval`. Sem isso a integração com TUI quebra.

## Implicações

- `go build ./cmd/njord/` **não é suficiente** para testar mudanças no terminal do usuário. Precisa **copiar o binário para `~/.local/bin/njord-cli`**.
- Não existe Go path mágico ou symlink — é `cp` mesmo.

## Como aplicar

Sempre que fizer uma mudança que o usuário vai testar via `njord`:

```bash
make install
```

O target faz `go build` + `cp` para `~/.local/bin/njord-cli`.

## Quando NÃO precisa

- Ao rodar testes: `make test`, `go test ./...` — usam o código fonte diretamente.
- Ao rodar ferramenta throwaway via `go run ./cmd/<nome>/` (ex.: debug, ping) — só precisa de `./cmd/<nome>/main.go`.

## Como detectar o problema
Sintoma típico: "a feature nova não aparece na UI" mesmo após `go build`.

Verificação rápida:
```bash
ls -la ~/.local/bin/njord-cli   # mtime deve ser recente
grep -c "<termo-da-feature>" ~/.local/bin/njord-cli   # deve retornar > 0
```

Se o binário instalado for antigo, rodar `make install`.
