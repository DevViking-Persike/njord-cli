# 30 QA

## Proposito

Provar que a fatia funciona e nao regrediu comportamento existente.

## Quando roda

Depois de arquitetura review aprovada.

## Definition of Ready

- Diff pronto.
- Build e testes unitarios verdes.
- Criterios de aceitacao rastreaveis.

## Checklist

- Rodar comandos de validacao do MANIFEST.
- Validar caminho feliz e erro via CLI.
- Confirmar que comando raiz ainda abre TUI (quando aplicavel manualmente).
- Registrar evidencias.

## Definition of Done

- Relatorio QA com PASS/FAIL.
- Regressao conhecida registrada.
- `.spec/STATE.md` atualizado.

## Anti-patterns

- Concluir QA apenas por leitura de codigo.
- Rodar processo longo sem `--dry-run` em automacao local.

## Template

Use `_TEMPLATE-qa.md`.
