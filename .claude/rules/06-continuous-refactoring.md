# Regra 6 — Refatoração contínua

## Regra do escoteiro
Deixe o código melhor do que encontrou. Mas **no escopo apropriado** — nunca misture refatoração grande com bugfix/feature.

## Antes de adicionar feature
- Se o arquivo está > 280 linhas, refatorar primeiro (commit separado), feature depois.
- Se a função alvo não tem teste, escrever teste de caracterização (cobre comportamento atual), só então modificar.

## Antes de refatorar
- `go test ./<pkg>/` precisa passar.
- Testes existentes são contrato — não deletar. Se teste ficou obsoleto, substituir por equivalente no novo código.

## Commits
- Um motivo por commit. Mensagens em pt-BR, conventional commits:
  - `refactor: ...` para mudança estrutural sem mudar comportamento
  - `test: ...` para testes isolados
  - `fix: ...` para bugfix
  - `feat: ...` para nova feature
  - `docs: ...` para documentação/CLAUDE.md/rules

## Bug descoberto no meio de refatoração
Parar, reportar ao usuário, perguntar se cria commit separado. **Não corrigir no mesmo commit** — ruído no histórico e dificulta revert.
