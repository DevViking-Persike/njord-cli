---
name: desenvolvimento
description: >-
  Roda a etapa de Desenvolvimento da esteira (disciplina 20), implementando
  conforme a spec aceita e o plano aprovado no gate de Arquitetura, com testes
  junto e validação local verde antes do review. Use quando o usuário pedir
  "implementar", "desenvolver o incremento", "codar a sprint NN", "começar o
  dev", ou "/desenvolvimento". Não começa sem plano aprovado.
---

# Skill: desenvolvimento (disciplina 20)

Implementa o incremento. Método em `.spec/sprints/20-desenvolvimento/README.md`;
regras em `.claude/rules/` quando rodar no Claude Code, ou nas regras equivalentes
do projeto quando rodar no Codex (arquitetura, testes, seguranca,
fluxo-desenvolvimento).

## Definition of Ready (não começar sem)
- Spec aceita + critérios de aceitação (Discovery).
- Plano técnico aprovado (Arquitetura **10a design**): camadas, contratos, ADR.

## Fluxo
1. Quebrar em **tasks** (`task-NN-*.md`).
2. Implementar **por camada** (respeitar a direção de dependência).
3. **Testes junto** (não depois) — caminho feliz + erro; cobrir invariantes.
4. **Validação local verde** antes de pedir review: build + lint + teste + RPA
   (comandos no `.spec/MANIFEST.md`).
> Modo **refatorar**: mudanças pequenas/reversíveis + teste de caracterização
> antes de mexer (não-regressão). Modo **documentar**: o "dev" é escrever os docs.

## Definition of Done
- [ ] Tasks por camada · [ ] testes novos verdes; sem regressão
- [ ] build/lint/teste verdes · [ ] validação local **PASS**
- [ ] diff pronto p/ review (Arquitetura 10b) · [ ] débito anotado
- [ ] `.spec/STATE.md` atualizado

## Anti-patterns
- ❌ Pedir review com build vermelho ou validação falhando.
- ❌ Lógica fora da camada certa. ❌ Refatoração + feature no mesmo incremento.
