# Regras de engenharia

Cada arquivo aqui define uma regra. Skills e commands (`.claude/commands/`)
referenciam regras específicas.

| # | Regra | Verificação automatizada |
|---|-------|--------------------------|
| 1 | [Tamanho de arquivo (alvo ~300, teto ~500)](01-file-size.md) | sim |
| 2 | [Testes unitários (≥ 84% cov + mutation)](02-unit-tests.md) | sim |
| 3 | [SOLID](03-solid.md) | parcial (grep de violações) |
| 4 | [Clean Architecture](04-clean-architecture.md) | sim (grep de imports) |
| 5 | [Simplicidade](05-simplicity.md) | não (code review) |
| 6 | [Refatoração contínua](06-continuous-refactoring.md) | não (disciplina) |
| 7 | [Build e execução do desktop app](07-install-binary.md) | sim (`npm run tauri build`) |
| 8 | [Delegar execução ao usuário](08-delegate-execution.md) | não (disciplina) |
| 9 | [UI responsiva (mobile-first)](09-responsive-ui.md) | parcial (grep de larguras fixas + DevTools) |
| 10 | [Arquitetura de frontend (MVVM + Atomic)](10-frontend-architecture.md) | parcial (grep de camadas) |
| — | [Segurança](seguranca.md) | parcial |
| — | [Fluxo de desenvolvimento](fluxo-desenvolvimento.md) | não (disciplina) |

## Comandos instalados
- `/check-rules` — audita o repo contra todas as regras
- `/refactor <arquivo>` — refatora um arquivo aplicando as regras relevantes
- `/responsive-pass <rota>` — audita e refatora UI aplicando Regra 9
- `/dead-code-cleansing` — identifica e remove código morto após confirmação

Violação exige justificativa explícita no commit/PR.
