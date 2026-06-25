# 00 Discovery

## Proposito

Levantar escopo, comportamento atual, criterios de nao-regressao e criterios de
aceitacao antes de mudar codigo.

## Quando roda

Primeira etapa de cada incremento. Em modo refatorar, foca no estado atual e em
preservar comportamento.

## Definition of Ready

- Projeto alvo identificado.
- Problema e menor fatia definidos.
- Comportamento a preservar listado.

## Checklist

- Mapear estado atual do `njord-cli`.
- Mapear contrato relevante do `njord-tauri`.
- Definir escopo pequeno e testavel.
- Definir criterios Given/When/Then.
- Registrar riscos e premissas.

## Definition of Done

- Escopo confirmado ou premissas explicitas.
- Criterios verificaveis.
- Nao-regressao mapeada.
- Documento de discovery criado.

## Anti-patterns

- Pular direto para implementacao sem criterio verificavel.
- Misturar port completo com primeira fatia.
- Acessar segredos ou banco real sem necessidade.

## Template

Use `_TEMPLATE-discovery.md`.
