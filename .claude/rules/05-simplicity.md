# Regra 5 — Simplicidade

## Princípios
- **Não antecipe abstração.** 3 linhas duplicadas são melhores que uma abstração prematura.
- Sem flags booleanas que mudam comportamento interno da função — prefira duas funções (`StartProject` vs `StartProjectDetached`).
- Sem camadas wrapper "por segurança" (interface → struct → interface → struct). Uma indireção resolve.
- Sem comentários que descrevem o *quê* — só o *porquê* não-óbvio (bug conhecido, invariante sutil, workaround).
- Sem error handling para casos que não podem acontecer. Confie em garantias internas/framework.
- Sem backwards-compat shims para código ainda não lançado.

## Idioma
- pt-BR: mensagens de UI e comentários explicativos.
- Inglês: identificadores de código (nomes de função, tipo, pacote) e mensagens de erro internas.

## Sinais de que está complicado demais
- Função > 60 linhas
- 3+ níveis de ifs aninhados
- Nome com "Manager", "Helper", "Util" (geralmente indica SRP fraco)
- Teste precisa de 20 linhas de setup pra um caso — função está fazendo muita coisa.

## Como verificar
Manual, no code review. Sem automação infalível — use julgamento.
