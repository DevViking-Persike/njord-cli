---
name: arquitetura
description: >-
  Roda o gate de Arquitetura da esteira (disciplina 10) — design gate antes do
  desenvolvimento e review gate depois do desenvolvimento, validando plano,
  ADR, camadas e contratos. Use quando o usuário pedir "revisar arquitetura",
  "design gate", "review do que o dev fez", "validar a abordagem", ou
  "/arquitetura [design|review]". Gate bloqueante.
---

# Skill: arquitetura (gate transversal — disciplina 10)

Roda o gate de Arquitetura **2×** por incremento. Método em
`.spec/sprints/10-arquitetura/README.md`; regras em `.claude/rules/` quando rodar
no Claude Code, ou nas regras equivalentes do projeto quando rodar no Codex
(arquitetura, seguranca). Cada gate é **bloqueante**: reprovou → volta uma casa.

## Entrada
- `design` (antes do dev) ou `review` (depois do dev), em ARGUMENTS.
- Discovery aprovado (design) ou branch/diff pronto (review).

## design gate (antes do dev)
1. A abordagem respeita **camadas/contratos/stack** do projeto?
2. Impacto em **segurança** e invariantes mapeado (`seguranca.md`)?
3. Dependência cross-módulo só pelo contrato definido?
4. Precisa de **ADR**? (decisão estrutural → `.spec/reference/ADR-NNN`).
→ Veredito: aprovado (segue p/ Dev) ou reprovado (volta à Discovery/Dev).

## review gate (depois do dev) — revisar o DIFF
1. **0 violação de camada / direção de dependência** (lint de camadas verde).
2. Sem segredo vazando; nenhuma regra de `seguranca.md` quebrada.
3. Lógica na camada certa (não no controller/handler/componente).
4. Bate com os **critérios de aceitação** da Discovery.
5. Débito técnico **registrado** (não escondido).
→ Veredito: aprovado p/ QA, ou lista de correções (volta ao Dev).

## Saída
- Preencher a instância `arquitetura-NN-<tema>.md` (template da disciplina).
- Atualizar `.spec/STATE.md` (status + veredito). Reprovou 2× → parada (pedir humano).

> Para o review do diff, apoie-se em `/code-review` quando existir; o gate de
> arquitetura é o humano-no-loop sobre o resultado.
