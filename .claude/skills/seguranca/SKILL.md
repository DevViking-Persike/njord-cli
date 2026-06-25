---
name: seguranca
description: >-
  Roda a etapa de Segurança da esteira (disciplina 40), com exploração dinâmica
  autorizada do ambiente vivo mapeada aos invariantes de segurança: token vazando,
  authz/IDOR, sessão, audit, CSP, redirect/SSRF e bypass. Use quando o usuário
  pedir "redteam", "testar segurança", "tentar invadir", "auditoria de
  segurança", ou "/seguranca". Último portão antes do release.
---

# Skill: seguranca (disciplina 40)

Tenta **quebrar/invadir** o que subiu, como um atacante. Método em
`.spec/sprints/40-seguranca/README.md`; invariantes em `.claude/rules/seguranca.md`
quando rodar no Claude Code, ou na regra equivalente do projeto quando rodar no Codex.
Para a **execução ofensiva** (pentest autorizado do próprio local/dev — SQLi, token
exposto, IDOR, bypass…), use a skill **`/redteam`** — este `/seguranca` é o gate.

## Escopo e autorização
- Alvo: ambiente vivo **autorizado** (NÃO produção sem aceite explícito). Sem DoS.
- Combina **estático** (`/security-review` quando existir) + **dinâmico** (tentar invadir).

## Cenários (mapeados aos invariantes)
| # | Cenário | Invariante violado se passar |
|---|---|---|
| A1 | Token/credencial vazando pro cliente | "nunca no browser/PageData" |
| A2 | Credencial forjada aceita | authn na borda |
| A3 | Escalonamento de privilégio / IDOR | authZ deny-by-default |
| A4 | Sessão (httpOnly, replay) | cookie cifrado |
| A5 | Audit forjado / UPDATE-DELETE | append-only + actor autenticado |
| A6 | XSS / CSP fraca | CSP nonce |
| A7 | Open redirect / SSRF / mass-assignment | borda controlada |
| A8 | Validação afrouxada / bypass em prod | nunca em produção |

## Gate (DoD)
- [ ] Cenários executados (ou N/A justificado).
- [ ] Achados classificados (Crítico/Alto/Médio/Baixo) + **PoC** + remediação.
- [ ] **0 Crítico/Alto aberto** (ou aceite de risco registrado).
- [ ] Relatório arquivado **sem segredos colados**. Achados bloqueantes → tasks (volta ao Dev).

## Anti-patterns
- ❌ Rodar contra produção/terceiros sem autorização. ❌ DoS como "teste".
- ❌ Colar segredo no relatório. ❌ "Seguro" sem tentar token/authz.
