---
name: discovery
description: >-
  Conduz um discovery estruturado para gerar contexto antes de
  construir/refatorar/documentar. Dois modos: produto (porquê, usuário, valor) e
  desenvolvimento (escopo, NFR, restrições, aceitação). Pode criar subagents
  independentes para pesquisa, leitura de código, riscos e lacunas quando houver
  contexto amplo. Usa The Mom Test. Use quando o usuário pedir "fazer discovery",
  "discovery de produto/desenvolvimento", "levantar requisitos/contexto", "antes
  de começar a desenvolver", ou "/discovery [produto|desenvolvimento]". É a
  disciplina 00 da esteira .spec.
---

# Skill: discovery

Conduz o **discovery** (disciplina 00 da esteira) fazendo **boas perguntas** para
gerar um contexto que sustente as decisões seguintes. Dois modos:

| Modo | Gera | Alimenta |
|---|---|---|
| **produto** | o **porquê**: outcome, usuário, dor, valor, riscos de produto | documentação de **produto** + justificativa de feature (scaffold `criar`/`documentar`) |
| **desenvolvimento** | o **como/escopo**: requisitos, NFR, restrições, riscos técnicos, aceitação | **construir/refatorar** (scaffold `criar`/`refatorar`) |

> Em `criar` use **os dois** (produto → desenvolvimento). Em `refatorar`, foco em
> **desenvolvimento** (estado atual + não-regressão). Em `documentar`, **produto**
> (se for doc de produto) ou **desenvolvimento/engenharia reversa** (se for doc técnica).

## Como perguntar (método — The Mom Test)

A qualidade do discovery vem de **como** se pergunta. Regras:

1. **Fale do passado, não do futuro.** Pergunte o que a pessoa **já fez**, não o
   que ela faria. ("Me conta a **última vez** que…" › "Você usaria…?")
2. **Peça histórias e exemplos concretos**, não opiniões. Opinião e elogio são
   ruído; comportamento é dado.
3. **Pergunte "o quê" e "como", evite "por quê"** direto (gera racionalização).
4. **Não venda nem valide a sua ideia.** Não conte a solução antes de entender o
   problema — senão a pessoa só concorda.
5. **Cave o problema**: frequência, impacto, o que já tentaram, quanto custa hoje.
6. **Pergunte em ondas** (3–6 por vez), siga os follow-ups, não despeje tudo.

> Fontes: The Mom Test (Rob Fitzpatrick); Continuous Discovery / Opportunity
> Solution Tree (Teresa Torres); The Four Big Risks (Marty Cagan/SVPG); Jobs to be
> Done; NFRs (Volere/arc42).

---

## Subagents de discovery

Use subagents quando o discovery depender de investigação paralela que pode ser
feita sem bloquear a entrevista: leitura de código legado, inventário de
integrações, comparação de documentação, pesquisa de riscos, análise de NFR ou
levantamento de lacunas por domínio.

### Quando criar
- Contexto espalhado em muitos arquivos, módulos, docs ou sistemas.
- `refatorar` ou `documentar`, onde é preciso medir estado atual antes de decidir.
- Produto + desenvolvimento no mesmo incremento, separando oportunidade,
  viabilidade técnica e riscos.
- Risco alto de viés: pedir leituras independentes evita uma conclusão prematura.

### Passes recomendados
- **Produto/oportunidade:** sintetizar outcome, usuário, dor, evidências e lacunas.
- **Código/legado:** mapear módulos, fluxos, integrações, pontos frágeis e testes
  existentes. Read-only.
- **NFR/riscos:** levantar performance, segurança, confiabilidade, compliance,
  dependências externas e spikes necessários.
- **Docs/operabilidade:** comparar README, runbooks, scripts e realidade do código,
  marcando stale, ausente ou não verificável.

### Contrato de cada subagent
Ao criar um subagent, dê escopo estreito, indique arquivos/diretórios permitidos e
peça saída em Markdown com:
- Evidências com caminho e linha quando houver.
- Lacunas e perguntas que precisam de usuário.
- Riscos e premissas separadas de fatos observados.
- Nenhuma edição de arquivo, nenhum comando destrutivo e nenhum acesso a produção.

Não use subagents para substituir a entrevista com o usuário. Consolide os
achados como **evidência auxiliar** e deixe claro o que foi confirmado pelo
usuário, o que veio do código/docs e o que ainda é hipótese.

---

## Modo PRODUTO — banco de perguntas

### 1. Outcome (resultado de negócio — a raiz)
- Que **resultado** queremos mover? (não a feature — o efeito: retenção, ativação, receita, custo, NPS…)
- Como esse resultado se liga à estratégia / North Star?
- Como saberemos que mexemos nele? (métrica + baseline atual)

### 2. Usuário & contexto
- Para **quem** é? (persona, papel, contexto de uso)
- Em que **situação** o problema aparece? (gatilho, frequência)

### 3. Oportunidade / problema (Opportunity Solution Tree + Mom Test)
- Qual **dor/necessidade/desejo** específico? (não a solução)
- **Como o usuário resolve isso hoje?** Me conta a **última vez** que precisou. (alternativas atuais)
- O que é **mais frustrante** nesse processo hoje? Quanto custa (tempo/dinheiro/risco)?
- O que já tentaram pra resolver? Por que não resolveu?

### 4. Jobs to be Done
- Quando [situação], o usuário quer [motivação], **pra** [resultado esperado]?
- O que ele está **ultimamente tentando realizar**?

### 5. Os 4 riscos (Cagan — matar antes de construir)
- **Valor:** ele vai **usar/pagar**? Que evidência temos? (não opinião — sinal de comportamento)
- **Usabilidade:** vai **conseguir usar**? onde costuma travar?
- **Viabilidade técnica:** dá pra **construir** com o time/stack/prazo? (passa pro modo desenvolvimento)
- **Viabilidade de negócio:** funciona pro **negócio**? (legal, financeiro, operacional, suporte, marca)

### 6. Sucesso & escopo
- Como é o **sucesso** em 1 frase? Qual a métrica (leading + lagging)?
- O que está **fora** de escopo agora? O que é a **menor fatia** que entrega valor (MVP/slice)?

---

## Modo DESENVOLVIMENTO — banco de perguntas

### 1. Escopo
- O que o sistema **faz** (casos de uso principais)? O que **NÃO** faz?
- Qual a **menor fatia** entregável (vertical slice)? O que é baseline × aditivo?

### 2. Requisitos funcionais
- Entradas, saídas, regras de negócio, estados, caminhos de erro.
- Atores/papéis e o que cada um pode fazer (autorização).

### 3. NFR — atributos de qualidade (escolher os **top 3–5** e dar número)
- **Performance/escala:** quantos usuários/req? latência alvo (p95)? volume de dados?
- **Disponibilidade/confiabilidade:** SLO? o que acontece em falha? recuperação?
- **Segurança:** dados sensíveis? authn/authz? auditoria? compliance (LGPD/…)?
- **Manutenibilidade:** quem mantém? testabilidade? observabilidade?
- **Compatibilidade/portabilidade:** plataformas, navegadores, integrações.
- **Usabilidade/acessibilidade/i18n** quando aplicável.

### 4. Restrições (constraints)
- **Stack/arquitetura** obrigatória ou existente? **Legado** a respeitar?
- **Integrações de terceiros** (limites, custos, SLAs, rate limits)?
- **Prazo, equipe, orçamento, legal/compliance.**

### 5. Premissas & riscos
- O que estamos **assumindo** (e qual premissa, se falsa, derruba o plano)?
- **Riscos** técnicos, de compliance (retrofitar compliance custa ~3× — desenhar antes), de dependência.
- O que precisa de **spike** (provar viabilidade antes de comprometer)?

### 6. Dependências & critérios de aceitação
- De quais sistemas/serviços/equipes depende?
- **Critérios de aceitação verificáveis** (Given/When/Then) — viram teste no QA.
- **Definition of Ready:** tudo acima respondido = pronto pra Arquitetura/Dev.

> Para **refatorar**: acrescente — qual o **comportamento atual** a preservar
> (critérios de **não-regressão**)? que **caracterização** (testes) cobre isso hoje?
> Para **documentar técnico** (engenharia reversa): use **Diátaxis** — que docs
> faltam? tutorial (aprender), how-to (tarefa), reference (consulta), explicação
> (entender)?

---

## Como rodar

1. Confirme o **modo** (produto/desenvolvimento) e o **scaffold-mode**
   (criar/refatorar/documentar) — se não souber, pergunte.
2. Se o contexto exigir investigação paralela, crie subagents read-only antes ou
   durante as ondas de perguntas. Use os passes acima e continue a entrevista em
   paralelo quando não depender do resultado.
3. Faça as perguntas **em ondas** (use o método Mom Test). **Não** invente as
   respostas: se for entrevista real, conduza; se o usuário já tem o contexto,
   colete e **confronte lacunas** (aponte o que falta responder).
4. Quando houver contexto suficiente, **sintetize** na instância de discovery
   (`.spec/sprints/00-discovery/discovery-NN-<tema>.md`) — use os templates desta
   skill (`templates/discovery-produto.md` / `templates/discovery-desenvolvimento.md`).
5. **Gate da Discovery (DoD):** escopo confirmado, aditivos aprovados, critérios
   de aceitação verificáveis, riscos/NFR mapeados. Só então passa pra Arquitetura.

Exemplos preenchidos (o que é "bom"): `templates/EXEMPLOS.md`.

## Anti-patterns

- ❌ Perguntar opinião/hipótese ("você gostaria…?", "isso é útil?") — viola o Mom Test.
- ❌ Pular pra solução antes de entender a oportunidade/problema.
- ❌ NFR sem número ("rápido", "escalável") — exija alvo mensurável.
- ❌ Critério de aceitação não verificável.
- ❌ Fechar o discovery com premissa arriscada não validada nem marcada como risco.
