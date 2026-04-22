# Regras de engenharia — njord-cli

Cada arquivo aqui define uma regra. Skills (`.claude/commands/`) referenciam regras específicas.

| # | Regra | Verificação automatizada |
|---|-------|--------------------------|
| 1 | [Tamanho de arquivo (≤ 300 linhas)](01-file-size.md) | sim |
| 2 | [Testes unitários (≥ 70% cov + mutation)](02-unit-tests.md) | sim |
| 3 | [SOLID](03-solid.md) | parcial (grep de violações) |
| 4 | [Clean Architecture](04-clean-architecture.md) | sim (grep de imports) |
| 5 | [Simplicidade](05-simplicity.md) | não (code review) |
| 6 | [Refatoração contínua](06-continuous-refactoring.md) | não (disciplina) |

## Comandos
- `/check-rules` — audita o repo contra todas as regras
- `/refactor <arquivo>` — refatora um arquivo aplicando as regras relevantes

Violação exige justificativa explícita no commit/PR.
