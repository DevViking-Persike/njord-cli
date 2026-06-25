---
name: scaffold-spec
description: >-
  Monta a base operacional .spec de um projeto e orquestra as skills da esteira
  (discovery, arquitetura, desenvolvimento, qa/qa-rpa, seguranca/redteam, deploy)
  para construir um projeto inteiro e bem estruturado. Cria cadência de sprints,
  MANIFEST, STATE, RUNBOOK, reference, rules, commands, tools e hooks. Use quando
  o usuário pedir "criar estrutura .spec", "scaffold spec", "montar um projeto do
  zero", "bootstrap operacional", "preparar projeto pra criar/refatorar/documentar
  um sistema", ou "/scaffold-spec [criar|refatorar|documentar]".
---

# Skill: scaffold-spec

Monta a **base operacional** de um projeto: o diretório `.spec/` (cadência de
sprints por disciplina + índice + estado + runbook), as **rules** de engenharia,
os **commands** do Claude Code e a **skill de deploy**. Generaliza o padrão
validado em produção. Funciona em projeto novo (greenfield) ou existente.

> **Princípio do projeto:** `rule` = fonte de verdade (conhecimento); `skill` =
> runbook que aplica. O `.spec/` é o manual operacional; roteadores como
> `CLAUDE.md` só apontam pra ele.

## Compatibilidade Claude Code + Codex

- Mantenha as skills do Claude Code como fonte de verdade.
- Para Codex, no projeto consumidor, crie `.codex/skills` como symlink para
  `.claude/skills`.
- Não crie cópias divergentes de `SKILL.md`; se precisar ajustar uma skill,
  ajuste a fonte e deixe o symlink refletir.
- Arquivos TOML não são necessários para skills Codex neste formato. Use
  `SKILL.md` com frontmatter YAML e, quando existir, `agents/openai.yaml`.

## Ecossistema — `scaffold-spec` é o hub

Esta skill **monta a base e relaciona todas as outras** para construir um projeto
inteiro e bem estruturado. Cada disciplina da esteira tem sua skill; o `scaffold`
as **instala** e o `RUNBOOK` as **invoca na ordem**:

```
/scaffold-spec [criar|refatorar|documentar]   ← monta .spec/ + rules + commands + skills + tools + hooks
        │
        ▼   a esteira, dirigida pelas skills (gates bloqueantes):
00  /discovery [produto|desenvolvimento]   → contexto (Mom Test / JTBD / 4 riscos / NFR)
10  /arquitetura design                    → gate: a abordagem é sã?
20  /desenvolvimento                       → implementa (testes junto)
10  /arquitetura review                    → gate: o diff bate com plano/ADR? 0 violação de camada
25  /review-codigo-subagents               → review técnico por lanes/subagents
30  /qa  →  /qa-rpa                         → validação real front+back de cada tela (RPA)
40  /seguranca  →  /redteam                → pentest autorizado (próprio local/dev)
    /deploy                                → build → registry → apply → smoke
    .claude/tools/spec-check.sh            → valida a entrega (estrutura + links)
```

| Skill | Papel | Etapa |
|---|---|---|
| **scaffold-spec** | monta a base + orquestra (o **hub**) | — |
| `discovery` | levanta o contexto (modos **produto** / **desenvolvimento**) | 00 |
| `arquitetura` | gate de **design** (antes) e **review** (depois do dev) | 10 |
| `desenvolvimento` | implementa conforme spec + plano | 20 |
| `review-codigo-subagents` | sprint de review de código por subagents independentes | 25 |
| `qa` / `qa-rpa` | gate de QA / **RPA front+back** automatizado de cada tela | 30 |
| `seguranca` / `redteam` | gate de segurança / **pentest** do próprio local-dev | 40 |
| `deploy` | sobe o projeto (build → apply → smoke) | transversal |

> **Relação bidirecional:** o `scaffold` instala e referencia todas; cada skill de
> etapa aponta de volta pra sua disciplina em `.spec/sprints/` e pras regras do projeto.
> Resultado: o projeto fica **íntegro da base à entrega validada**. O `RUNBOOK.md`
> gerado deve **invocar a skill de cada etapa** na ordem (ver blueprint no Passo 1).

## Entrada — MODO

A skill aceita um modo em ARGUMENTS (default: perguntar):

| Modo | Quando | Ênfase da esteira |
|---|---|---|
| **criar** | sistema novo (greenfield) | Discovery (escopo) → Arquitetura (design do zero) → Dev → QA → Segurança |
| **refatorar** | sistema existente | Discovery = inventário do estado atual + metas + critérios de **não-regressão**; Arquitetura = atual×alvo; Dev incremental; QA pesado em regressão; Segurança = re-auditoria |
| **documentar** | sistema existente sem docs | Discovery = engenharia reversa/inventário; "Dev" vira **escrever docs**; QA = doc bate com o código; produz `reference/` + mapa de arquitetura |

Se o usuário não passou o modo, **pergunte qual** antes de gerar (muda os
critérios de aceitação e a ênfase).

## Passo 1 — Gerar o esqueleto `.spec/`

Crie esta árvore na raiz do projeto-alvo (não sobrescreva o que já existir sem
confirmar):

```
.spec/
├── MANIFEST.md              # mapa read-first (índice de tudo)
├── STATE.md                 # estado vivo do incremento atual
├── reference/               # docs de referência do projeto (arquitetura, roadmap, etc.)
│   └── README.md
└── sprints/
    ├── README.md            # framework das 5 disciplinas + fluxo da esteira
    ├── RUNBOOK.md           # como rodar a esteira (ordem + gates bloqueantes)
    ├── 00-discovery/        { README.md, _TEMPLATE-discovery.md }
    ├── 10-arquitetura/      { README.md, _TEMPLATE-arquitetura.md }
    ├── 20-desenvolvimento/  { README.md, _TEMPLATE-desenvolvimento.md }
    ├── 25-review-codigo/    { README.md, _TEMPLATE-review-codigo.md }
    ├── 30-qa/               { README.md, _TEMPLATE-qa.md }
    └── 40-seguranca/        { README.md, _TEMPLATE-seguranca.md }
```

### Conteúdo de cada arquivo (blueprint)

**`MANIFEST.md`** — ponto de entrada único. Seções: *Bootstrap de sessão* (ordem
de leitura: MANIFEST → STATE → RUNBOOK → disciplina atual); *Mapa do `.spec/`*
(tabela caminho→o quê); *Disciplinas → onde olhar* (tabela etapa→README→docs de
referência); *Regras de execução* (tabela apontando `.claude/rules/*` e, no
Codex, para a regra equivalente do projeto);
*Maquinário de validação* (comandos de teste/build/lint do projeto); *Regra-mãe*
(o que governa o escopo — preencher com o contrato/escopo do projeto).

**`STATE.md`** — estado vivo. Campos: incremento ativo (NN, tema, etapa, branch,
atualizado em); tabela de progresso da esteira
(00→10→20→10-review→25-review-codigo→30→40 com
status ⬜🟡✅🔴); último resultado de validação; pendências; itens aguardando
aprovação; histórico de incrementos; protocolo de atualização (atualizar ao
entrar/sair de cada etapa; nunca avançar com gate reprovado).

**`sprints/README.md`** — as 6 disciplinas (00 Discovery, 10 Arquitetura [gate
transversal], 20 Desenvolvimento, 25 Review de Código, 30 QA, 40 Segurança), o
fluxo da esteira
(`00 → 10-design → 20 → 10-review → 25-review-codigo → 30 → 40 → release`,
Arquitetura roda 2× como gate bloqueante), os handoffs (contrato entre disciplinas) e a convenção de
instância (`<disciplina>-NN-<tema>.md`, mesmo NN em toda a esteira).

**`sprints/RUNBOOK.md`** — como rodar a esteira autonomamente: ler STATE → retomar
etapa; loop pelas etapas **invocando a skill de cada uma** (`/discovery` → `/arquitetura`
→ `/desenvolvimento` → `/arquitetura review` → `/review-codigo-subagents`
→ `/qa`+`/qa-rpa` → `/seguranca`+`/redteam`
→ `/deploy`), com os **gates bloqueantes**; comandos reais por etapa; **paradas
obrigatórias** (pedir humano): item fora do escopo sem aprovação, ação destrutiva/
produção, gate reprovado 2×, decisão estrutural sem registro, segredo.

**Disciplinas (`NN-*/README.md`)** — cada uma com: Propósito; Quando roda
(gate/ordem); Definition of Ready (entrada); Atividades/Checklist; Definition of
Done (saída); Anti-patterns; link pro `_TEMPLATE`. Adapte a ênfase ao MODO
escolhido (ver tabela de modos). O `_TEMPLATE-*.md` é o molde fill-in-the-blank
de uma instância.

> Use o `.spec/` de referência (um projeto já estruturado) como referência de qualidade do
> conteúdo, **generalizando** o que for específico de domínio (regras fiscais,
> Zitadel, etc.) para placeholders `<...>`.

## Passo 2 — Instalar rules, commands, skills de etapa, deploy, tools e hooks

A esteira é **dirigida por skills** (uma por etapa) que o `scaffold` orquestra.
Copie tudo para o projeto-alvo em `.claude/skills` e, quando usar Codex, crie um
symlink para essa árvore canônica:

```bash
S=.claude/skills/scaffold-spec/templates
# rules e commands (fonte de verdade operacional)
mkdir -p .claude/rules .claude/commands
cp -R $S/rules/. .claude/rules/
cp -R $S/commands/. .claude/commands/
# skills da esteira e transversais — copiar do repo-fonte, ou já globais em ~/.claude/skills/
cp -R .claude/skills/{discovery,arquitetura,desenvolvimento,qa,qa-rpa,seguranca,redteam,review-codigo-subagents} <dest>/.claude/skills/ 2>/dev/null || true
# skill de deploy (copiar a pasta para preservar agents/openai.yaml)
mkdir -p .claude/skills && cp -R $S/skills/deploy .claude/skills/
# Codex: manter .claude/skills como fonte canônica e apontar para ela
mkdir -p .codex
[ -e .codex/skills ] || ln -s ../.claude/skills .codex/skills
# tool de validação ("entrega funcionando")
mkdir -p .claude/tools && cp $S/tools/spec-check.sh .claude/tools/ && chmod +x .claude/tools/spec-check.sh
# hooks (opt-in): ver $S/hooks/README.md
```

**Skills de etapa — a esteira chama em ordem:**
| Etapa | Skill |
|---|---|
| 00 Discovery | `/discovery [produto\|desenvolvimento]` — banco de perguntas (Mom Test / JTBD / 4 riscos / NFR) |
| 10 Arquitetura | `/arquitetura [design\|review]` — gate 2× |
| 20 Desenvolvimento | `/desenvolvimento` |
| 25 Review de Código | `/review-codigo-subagents` — pipeline read-only por lanes/subagents |
| 30 QA | `/qa` (gate) + `/qa-rpa` (automação RPA front+back de cada tela) |
| 40 Segurança | `/seguranca` (gate) + `/redteam` (pentest autorizado do próprio local/dev) |

> Se as skills de etapa já estiverem **globais** (`~/.claude/skills/` ou
> `~/.codex/skills/`), não precisa copiar — só garanta que existem. O `RUNBOOK.md`
> invoca cada uma na etapa certa.

- **`rules/`** — regras de engenharia, segurança, fluxo e review. Preencha ou
  ajuste placeholders `<...>` conforme o projeto.
- **`commands/`** — comandos Claude Code para auditar rules, refatorar alvo,
  responsividade e código morto.
- **`deploy/SKILL.md`** — runbook de deploy (build→registry→apply→smoke); aponte
  pra uma `staging-deploy.md` quando existir.
- **`tools/spec-check.sh`** — valida a `.spec/` (arquivos obrigatórios + links +
  STATE). Rode ao fechar: `bash .claude/tools/spec-check.sh`.
- **hooks** (`templates/hooks/`) — rodam o `spec-check` automaticamente (Stop /
  PostToolUse). **Opt-in:** mesclar no `.claude/settings.json` no Claude Code
  (não auto-aplicar). No Codex, use validação manual ou mecanismo equivalente.

## Passo 3 — Cabear o roteador do agente

Crie ou ajuste o roteador do agente (`CLAUDE.md` no Claude Code, ou equivalente
no Codex quando existir) para ser um **roteador fino** que aponta para a base nova
(não duplicar conteúdo):

- **Regra-mãe** (1 linha — o que governa o escopo).
- **Bootstrap — ordem de leitura:** `.spec/MANIFEST.md` → `.spec/STATE.md` →
  `.spec/sprints/RUNBOOK.md`.
- **Seguir a esteira do `.spec/`** (00→10→20→10→25→30→40, gates do RUNBOOK) —
  cada etapa via a skill: `/discovery`, `/arquitetura`, `/desenvolvimento`,
  `/review-codigo-subagents`, `/qa`, `/seguranca`.
- **Autoridades:** rules em `.claude/rules/` ou equivalente do projeto; skills em
  `.claude/skills/` e, para Codex, via symlink `.codex/skills`.
- **Segurança:** os invariantes irredutíveis (ver `seguranca.md`).

## Passo 4 — Fechar

- **Validar a entrega:** `bash .claude/tools/spec-check.sh` — deve dar OK (0 link
  quebrado, arquivos obrigatórios presentes). Corrija o que apontar.
- Liste o que foi criado e o que tem placeholder `<...>` a preencher.
- Atualize `STATE.md` com o 1º incremento (ou deixe "nenhum ativo").
- Registre o que **não** rodou e por quê.
- **Não** commitar automaticamente — deixar para o usuário revisar.

## Anti-patterns

- ❌ Sobrescrever um `.spec/`/roteador existente sem confirmar.
- ❌ Gerar a esteira sem definir o MODO (criar/refatorar/documentar muda tudo).
- ❌ Copiar conteúdo específico de domínio de outro projeto
  para o projeto atual — generalize.
- ❌ Deixar o roteador gordo — ele só aponta; o manual vive no `.spec/`.
