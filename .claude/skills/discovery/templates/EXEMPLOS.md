# Exemplos de discovery "bom" (referência)

Exemplos preenchidos para calibrar qualidade. Tema fictício: **lembrete de
vencimento de fatura** num SaaS de cobrança.

---

## Exemplo A — modo PRODUTO

**Outcome:** reduzir inadimplência (churn involuntário). Métrica: % de faturas
pagas até o vencimento — hoje **71%**, alvo 85%.

**Usuário/contexto:** dono de pequena empresa, paga várias faturas no fim do mês,
**esquece** vencimentos quando estão fora do e-mail principal.

**Oportunidade (Mom Test):** "Me conta a última vez que pagou uma fatura em
atraso." → *"Vi a notificação de bloqueio do serviço, aí paguei correndo."* →
descobre alternativas hoje, frustração e custo real, não opinião.

**JTBD:** quando uma fatura está pra vencer, quero **ser lembrado no canal que eu
checo** (WhatsApp), pra **não pagar multa nem ter serviço cortado**.

**4 riscos:** Valor — 3 de 5 entrevistados já pagaram multa evitável (sinal real).
Usabilidade — opt-in simples. Viab. técnica — gateway de WhatsApp já integrado
(→ dev). Viab. negócio — custo por mensagem < multa evitada; LGPD: precisa consentimento.

**Sucesso:** "menos faturas vencem sem aviso." Leading: % de faturas com lembrete
entregue. Lagging: pagas-até-vencimento de 71% → 85%.
**MVP:** 1 lembrete, 3 dias antes, WhatsApp, opt-in. **Fora:** régua multi-toque, e-mail, SMS.

---

## Exemplo B — modo DESENVOLVIMENTO

**Escopo:** envia 1 lembrete por fatura, 3 dias antes do vencimento, via WhatsApp,
só para clientes opt-in. **NÃO:** régua de cobrança, retry inteligente, outros canais.
**Slice:** job diário que seleciona faturas D-3 e dispara via gateway.

**Requisitos:** entrada = faturas com `vencimento = hoje+3` e `cliente.optin=true`;
saída = mensagem enviada + registro de envio (idempotente: 1 por fatura).

**NFR (com número):** Performance — processar 50k faturas/dia em < 10 min.
Confiabilidade — idempotente (reprocesso não duplica); falha do gateway → retry com
backoff, no máx 3. Segurança — consentimento LGPD registrado; telefone é dado pessoal
(não logar em claro). Observabilidade — métrica de enviados/falhos.

**Restrições:** gateway WhatsApp existente (rate limit 80 msg/s; custo por msg);
janela de envio comercial (8h–20h). Legal: opt-in obrigatório.

**Premissas/riscos:** premissa = telefone cadastrado está correto (risco: bounce —
medir taxa). Risco compliance: enviar sem opt-in = multa LGPD (desenhar o gate antes).
Spike: validar rate limit real do gateway com 1k mensagens.

**Aceitação (verificável):**
1. **Dado** fatura D-3 de cliente opt-in **quando** o job roda **então** 1 lembrete é
   enviado e registrado (e reprocessar não envia de novo).
2. **Dado** cliente **sem** opt-in **quando** o job roda **então** nenhum envio ocorre.
3. **Dado** gateway fora **quando** o envio falha **então** há retry (≤3) e a falha é métrica.

**Definition of Ready:** ✅ escopo, ✅ NFR com número, ✅ restrições/compliance, ✅
aceitação verificável → pronto pra Arquitetura.
