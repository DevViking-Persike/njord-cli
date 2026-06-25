# 20 Desenvolvimento

## Proposito

Implementar a fatia aprovada pela arquitetura com testes junto.

## Quando roda

Depois do gate `10 Arquitetura design`.

## Definition of Ready

- Discovery aprovado.
- Plano tecnico aprovado.
- Criterios de aceitacao verificaveis.

## Checklist

- Quebrar em tasks.
- Implementar por camada.
- Adicionar testes de caminho feliz e erro.
- Rodar build/testes locais.
- Atualizar `.spec/STATE.md`.

## Definition of Done

- Tasks completas.
- Testes novos verdes.
- `go test ./...` verde.
- Build local verde.
- Diff pronto para arquitetura review.

## Anti-patterns

- Implementar feature grande junto com refatoracao ampla.
- Esconder regra de negocio no wiring Cobra.
- Deixar processo longo em teste automatizado.

## Template

Use `_TEMPLATE-desenvolvimento.md`.
