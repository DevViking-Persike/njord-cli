# 10 Arquitetura

## Proposito

Validar abordagem antes do desenvolvimento e revisar o diff depois.

## Quando roda

- `design`: depois do discovery e antes do desenvolvimento.
- `review`: depois do diff e antes do QA.

## Definition of Ready

- Discovery aprovado.
- Escopo e criterios de aceitacao claros.
- Regras de camada conhecidas.

## Checklist

- Conferir camadas e direcao de dependencia.
- Conferir contratos CLI e saida de stdout/stderr.
- Conferir tratamento de erro e comandos externos.
- Decidir se precisa ADR.

## Definition of Done

- Veredito `aprovado` ou `reprovado`.
- Debitos e riscos registrados.
- `.spec/STATE.md` atualizado.

## Anti-patterns

- Colocar regra de negocio em `cmd/`.
- Acoplar CLI Go ao banco interno do Tauri sem contrato explicito.
- Quebrar stdout usado pelo shell wrapper.

## Template

Use `_TEMPLATE-arquitetura.md`.
