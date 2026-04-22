# Regra 8 — Delegar execução de comandos ao usuário

Output de comando roda direto no contexto do Claude. Em tarefas longas, isso queima tokens sem agregar valor quando o resultado não influencia a próxima decisão.

Preferência: **pedir pro usuário rodar** comandos cujo output não muda o que eu vou fazer em seguida.

## Delego (pede pro usuário rodar)

- `njord` ou qualquer execução interativa da TUI (não vejo a tela mesmo)
- `make install` após a última mudança da rodada
- Aberturas de browser, IDE, editor de texto
- Testes de UX (navegação, cor, layout)
- Commits e pushes quando o usuário já aprovou o escopo em linguagem natural
- Mutation testing completo quando só quero confirmar que passou depois de uma refatoração pequena (resposta "passou ou não" basta)

## Executo eu mesmo

- `go test ./...`, `go build`, `go vet` durante iteração (preciso do erro exato pra corrigir)
- `git status`, `git diff`, `git log` antes de commitar (preciso decidir o que stagear e como redigir a mensagem)
- Greps, Reads, Globs (investigação que me orienta)
- Comandos curtos cujo output redireciona o próximo passo
- `make install` quando é parte de um sanity check no meio de um refactor maior

## Formato ao delegar

Sempre explícito sobre o que esperar:

> "Roda `make install` e me fala se o card Jira apareceu. Se deu erro, cola o output inteiro."

Não fica ambíguo ("você pode testar") — fica instrução direta ("roda X, me diz Y").

## Quando em dúvida

Se o comando é rápido e o output é pequeno, rodo eu. Se é lento ou o output é grande e só importa o veredito (passou/falhou), delego.
