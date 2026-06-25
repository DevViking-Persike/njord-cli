# Sprints - njord-cli

## Disciplinas

| Etapa | Papel |
|---|---|
| 00 Discovery | definir escopo, nao-regressao e criterios verificaveis |
| 10 Arquitetura | aprovar abordagem antes do dev e revisar o diff depois |
| 20 Desenvolvimento | implementar em fatias pequenas com testes junto |
| 25 Review de Codigo | revisar diff por lanes/subagents read-only |
| 30 QA | provar comportamento real e ausencia de regressao |
| 40 Seguranca | checar segredos, comandos externos e superficie de risco |

## Fluxo

`00 Discovery -> 10 Arquitetura design -> 20 Desenvolvimento -> 10 Arquitetura review -> 25 Review de Codigo -> 30 QA -> 40 Seguranca -> release`

Arquitetura roda duas vezes por incremento. Gate reprovado bloqueia a proxima
etapa.

## Handoffs

| Origem | Entrega |
|---|---|
| Discovery | escopo, nao-regressao, criterios de aceitacao |
| Arquitetura design | plano tecnico, camadas, contratos, ADR se necessario |
| Desenvolvimento | diff, testes, validacao local, debitos anotados |
| Arquitetura review | veredito do diff contra plano e regras de camada |
| Review de Codigo | achados por severidade, evidencias e veredito geral |
| QA | evidencias de funcionamento e regressao |
| Seguranca | verificacao de segredos, comandos e permissoes |

## Convencao

Use o mesmo numero em todas as disciplinas: `*-NN-<tema>.md`.
