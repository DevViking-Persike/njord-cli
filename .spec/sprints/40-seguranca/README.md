# 40 Seguranca

## Proposito

Checar riscos de seguranca antes de release.

## Quando roda

Depois de QA aprovado.

## Definition of Ready

- QA PASS.
- Superficie de comando e IO conhecida.

## Checklist

- Confirmar que nenhum segredo foi logado.
- Confirmar que `tauri status` nao imprime tokens nem arquivos sensiveis.
- Confirmar que `tauri dev` executa apenas no path validado.
- Confirmar que erros nao vazam conteudo de arquivos sensiveis.

## Definition of Done

- Relatorio de seguranca com PASS/FAIL.
- Critico/Alto zero aberto.
- `.spec/STATE.md` atualizado.

## Anti-patterns

- Ler secrets do app Tauri sem necessidade.
- Executar comandos destrutivos.
- Usar paths sem validacao.

## Template

Use `_TEMPLATE-seguranca.md`.
