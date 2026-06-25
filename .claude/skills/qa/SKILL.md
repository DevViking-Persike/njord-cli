---
name: qa
description: >-
  Roda a etapa de QA da esteira (disciplina 30), com validação real além de unit,
  provando que o incremento funciona e não regrediu, cobrindo critérios de
  aceitação, caminhos de erro e autorização. Use quando o usuário pedir "rodar
  QA", "validar o incremento", "testar de verdade", "RPA", ou "/qa". Gate com
  VERDICT=PASS e relatório.
---

# Skill: qa (disciplina 30)

Prova que o incremento **funciona de verdade** e que nada regrediu. Método em
`.spec/sprints/30-qa/README.md`; regras em `.claude/rules/testes.md` quando rodar
no Claude Code, ou nas regras equivalentes do projeto quando rodar no Codex. Para a
**automação** (RPA de navegador validando cada tela front+back), use a skill
**`/qa-rpa`** — este `/qa` é o gate; o `/qa-rpa` é a execução.

## Entrada
Diff aprovado no review gate (Arquitetura 10b) + build verde.

## Montar o QA
1. Traduzir **cada critério de aceitação** (Discovery) em uma checagem real.
2. Cobrir os **invariantes** tocados (regressão).
3. Cobrir **caminho de erro** (input inválido → 4xx, etc.).
4. **AuthZ/RBAC**: cada papel vê só o que deve.
5. **Smoke** no ambiente alvo (saúde + fluxo real).
> Modo **refatorar**: regressão pesada — comportamento observável **idêntico**.
> Modo **documentar**: os comandos/links da doc executam/resolvem (doc bate com código).

## Gate (DoD)
- [ ] **VERDICT=PASS** + relatório arquivado.
- [ ] Todos os critérios de aceitação cobertos.
- [ ] Invariantes tocados sem regressão. [ ] ≥1 caminho de erro por endpoint.
- [ ] AuthZ validado. [ ] Smoke verde (se deploy).
- [ ] `.spec/STATE.md` atualizado. FAIL → volta ao Dev.

## Anti-patterns
- ❌ Teste que passa com bug (falso negativo). ❌ Testar implementação interna.
- ❌ Aceitar PASS sem relatório.
