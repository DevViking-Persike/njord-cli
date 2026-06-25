# Regras de Fluxo de Desenvolvimento

> Como o trabalho atravessa a esteira de `.spec/sprints/` (Discovery → Arquitetura
> → Dev → Review de Código → QA → Segurança). Há **3 modos**; o modo escolhido
> muda a ênfase de cada disciplina e os critérios de aceitação. Defina o modo
> **na Discovery**.

## A esteira (comum aos 3 modos)

```
00 DISCOVERY → 10 ARQUITETURA(design) → 20 DEV → 10 ARQUITETURA(review) → 25 REVIEW CÓDIGO → 30 QA → 40 SEGURANÇA → release
```

- **Arquitetura é gate transversal** (roda 2×: valida o plano antes do dev e
  revisa o que o dev entregou). Cada gate é **bloqueante**: reprovou, volta uma casa.
- **Mesmo `NN`** em todas as disciplinas de um incremento (rastreia ponta a ponta).
- Estado vivo em `.spec/STATE.md`; como rodar em `.spec/sprints/RUNBOOK.md`.

---

## Modo CRIAR (sistema novo / greenfield)

Construir algo que não existe.

| Etapa | Ênfase |
|---|---|
| Discovery | escopo contra o contrato/objetivo; critérios de aceitação verificáveis; o que é **baseline** vs **aditivo** |
| Arquitetura (design) | desenho do zero: camadas, contratos, stack, ADR das decisões estruturais |
| Dev | implementar por camada + testes junto; build/lint verdes |
| Arquitetura (review) | o entregue bate com o design/ADR? 0 violação de camada |
| Review de Código | subagents auditam diff, regras locais, testes, segurança básica, operabilidade e lacunas antes do QA |
| QA | cobre cada critério de aceitação + caminho de erro |
| Segurança | invadir pelo navegador o que subiu (token, authz, audit, CSP) |

**DoD do incremento:** funciona, testado, sem violação de camada, review de
código sem `FAIL`, sem achado crítico de segurança.

---

## Modo REFATORAR (sistema existente)

Mudar a estrutura interna **sem mudar o comportamento observável**.

| Etapa | Ênfase |
|---|---|
| Discovery | **inventário do estado atual** (medido, não suposto); metas do refactor; **critérios de não-regressão** (o que NÃO pode mudar) |
| Arquitetura (design) | **atual × alvo**: o que muda, o que se preserva; plano incremental (Strangler Fig se grande); ADR se decisão estrutural |
| Dev | mudanças **pequenas e reversíveis**; testes de caracterização cobrindo o comportamento antes de mexer |
| Arquitetura (review) | a refatoração atingiu a meta sem violar camada nem vazar comportamento? |
| Review de Código | subagents focam regressão, acoplamento novo, dívida criada e candidatos a código morto/deps |
| QA | **regressão pesada**: a suíte/RPA prova que o comportamento observável é idêntico |
| Segurança | re-auditoria das superfícies tocadas |

**DoD do incremento:** meta de refactor atingida, **0 regressão** comprovada,
reversível.

> ❌ Anti-pattern: refatorar e adicionar feature no mesmo incremento — separar.

---

## Modo DOCUMENTAR (sistema existente sem/com pouca doc)

Tornar o sistema entendível e operável, sem mudar código.

| Etapa | Ênfase |
|---|---|
| Discovery | **engenharia reversa**: mapear módulos, fluxos, integrações, infra reais (ler o código, não os docs antigos) |
| Arquitetura (design) | montar o **mapa de arquitetura** vigente (camadas, comunicação, deploy) → `.spec/reference/` |
| Dev → **escrever docs** | gerar `reference/` (arquitetura, roadmap, deploy, observabilidade), READMEs por módulo, runbooks |
| Arquitetura (review) | a doc **bate com o código real**? sem afirmação stale (stack morta, infra antiga) |
| Review de Código | subagents verificam docs contra código, comandos, exemplos, links e lacunas de operabilidade |
| QA | verificar comandos/links dos docs (executam? resolvem? smoke real) |
| Segurança | documentar os invariantes de segurança + 1 passada `/security-review` |

**DoD do incremento:** doc fiel ao código atual, sem referência quebrada/stale,
verificável; o roteador do agente (`CLAUDE.md`, `AGENTS.md` ou equivalente)
aponta para o `.spec/`.

> ❌ Anti-pattern: tratar doc histórica como verdade atual — validar contra o
> código; remover/arquivar o que está superado.

---

## Paradas obrigatórias (em qualquer modo)

Pare e peça decisão humana quando:
1. Item fora do escopo/contrato sem aprovação escrita (registrar em `STATE.md`).
2. Ação destrutiva/irreversível ou deploy em **produção**.
3. Gate reprovado 2× seguidas na mesma etapa (não converge).
4. Decisão estrutural nova sem ADR.
5. Qualquer passo que exigiria abrir/expor segredo (`.claude/rules/seguranca.md`
   ou regra equivalente do projeto).
