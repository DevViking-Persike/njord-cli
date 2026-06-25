# 25 Review de Codigo

## Proposito

Revisar o diff por lanes independentes antes de QA, priorizando bugs,
violacoes de regras locais, riscos de arquitetura, testes, seguranca,
operabilidade e documentacao.

## Quando roda

Depois de `10b Arquitetura review` aprovado e antes de `30 QA`.

## Definition of Ready

- Diff pronto.
- Arquitetura review aprovada.
- Plano/spec/ADR relevantes disponiveis em `.spec/`.
- Regras locais disponiveis em `.claude/rules/`.
- Comandos de validacao identificados ou lacuna registrada.

## Checklist

- Definir alvo do review.
- Selecionar lanes proporcionais ao diff.
- Rodar pipeline read-only com `review-codigo-subagents`.
- Consolidar achados por severidade.
- Registrar comandos executados e nao executados.
- Atualizar `.spec/STATE.md`.

## Definition of Done

- Relatorio `review-codigo-NN-<tema>.md` criado.
- Veredito geral `PASS`, `PASS_WITH_WARNINGS` ou `FAIL`.
- Achados comprovados com evidencia e proximo passo.
- `FAIL` encaminhado para Desenvolvimento ou Arquitetura.

## Anti-patterns

- Corrigir durante review sem confirmacao.
- Reportar suspeita como violacao comprovada.
- Rodar comando destrutivo ou acessar producao.
- Instalar ferramenta nova sem aprovacao.

## Template

Use `_TEMPLATE-review-codigo.md`.
