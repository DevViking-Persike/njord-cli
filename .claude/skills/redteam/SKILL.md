---
name: redteam
description: >-
  Pentest autorizado do próprio sistema em local ou dev para achar brechas reais
  antes de um atacante: token vazando, bypass de auth, IDOR, SQL injection, XSS,
  CSRF, SSRF, segredos expostos e sessão. Gera achados com PoC, severidade e
  remediação. Use quando o usuário pedir "hackear o projeto (próprio)", "testar
  segurança", "tentar invadir", "redteam", "pentest local/dev", ou "/redteam".
  Executor da disciplina 40 (Segurança).
---

# Skill: redteam — pentest autorizado (achar a brecha, remediar)

Exploração **dinâmica** do **próprio** sistema para encontrar aberturas reais.
Disciplina 40 → `/seguranca` é o gate; esta skill é a execução ofensiva controlada.
Objetivo é **defensivo**: achar → PoC mínimo → remediar.

## ⚠️ Autorização (inegociável)
- **Só infra do próprio projeto:** `localhost`/dev autorizados. **Nunca produção**
  nem terceiros sem aceite **escrito**.
- **Sem DoS/stress.** Sem exfiltração real de dados (PoC mínimo prova a falha — não
  baixe a base). **Nenhum segredo no relatório** (mascarar).
- Se o alvo não for claramente seu/autorizado, **pare e pergunte**.

## Método
1. **Recon:** mapear superfícies — rotas, endpoints de API/BFF, forms, params, headers,
   cookies, upload, integrações. Ver o bundle JS, source maps, páginas de erro.
2. **Atacar cada superfície** com os vetores abaixo (navegador + API).
3. Para cada brecha: **PoC reproduzível** (passos), **impacto**, **invariante violado**
   (`.claude/rules/seguranca.md` no Claude Code, ou regra equivalente no Codex),
   **remediação**, **severidade** (Crítico/Alto/Médio/Baixo).
4. Achado bloqueante → vira task de Dev (volta uma casa).

## Vetores (mapeados aos invariantes)

| # | Brecha | Como testar | Esperado (seguro) |
|---|---|---|---|
| T1 | **Token exposto** | inspecionar `window.__*`/PageData, bundle JS, source maps, Network, páginas de erro | nenhum JWT/Bearer/accessToken no client |
| T2 | **Bypass de auth** | chamar endpoint protegido sem sessão; `Authorization: Bearer dev-mock`/forjado; flags `*_BYPASS` | 401/403 |
| T3 | **IDOR / escalonamento** | logar como papel baixo → acessar rota/endpoint de papel alto; trocar `id` de outro tenant na API | 403 / só os próprios dados |
| T4 | **SQL injection** | inputs/params/headers com vetores: `' OR '1'='1`, `'--`, `1;SELECT pg_sleep(5)--` (time-based), aspas que causam erro 500 | rejeita/parametriza; **sem** 500, **sem** atraso, **sem** vazar erro de SQL |
| T5 | **XSS / CSP** | injetar `<script>`/`"><img src=x onerror=alert(1)>` em campos refletidos/persistidos; conferir header CSP | sanitizado; CSP nonce bloqueia inline |
| T6 | **CSRF** | request que muda estado **sem** token/SameSite a partir de origem externa | recusado (token/SameSite) |
| T7 | **SSRF / mass-assignment** | forçar o back a buscar URL arbitrária; enviar campos extras no payload (ex.: `role:"admin"`) | recusa URL externa; ignora campos não esperados |
| T8 | **Segredos expostos** | `grep` no bundle/`build/`, `.env`/`.git` acessíveis via HTTP, source maps, mensagens de erro com stack/credenciais | nada sensível servido |
| T9 | **Sessão** | flags do cookie (httpOnly/Secure/SameSite); ler cookie via JS; replay após logout | httpOnly+Secure+SameSite; replay inválido |
| T10 | **Rate limit / enumeração** | tentativas repetidas de login; mensagens que revelam se o user existe (sem DoS) | rate limit; mensagem genérica |

## Ferramentas
- **Navegador:** Playwright/Chromium (login real, ler DOM/PageData, adulterar request).
  Template: `templates/probe-redteam.mjs`.
- **API:** `curl`/`httpie` (Bearer forjado, IDOR, injection, headers). Exemplos:
```bash
# T2 Bearer forjado → deve 401
curl -s -o /dev/null -w '%{http_code}\n' -H 'Authorization: Bearer dev-mock-token' "$HOST/api/<recurso>"
# T4 SQLi time-based → NÃO pode atrasar ~5s
time curl -s "$HOST/api/<recurso>?busca=1%3BSELECT%20pg_sleep(5)--" >/dev/null
# T8 arquivos sensíveis servidos → deve 404
for p in /.env /.git/config /server/.env; do echo -n "$p "; curl -s -o /dev/null -w '%{http_code}\n' "$HOST$p"; done
```
- Opcional (com cautela, só no próprio dev): `sqlmap`, OWASP ZAP, Burp — para varredura mais ampla.

## Gate (DoD)
- [ ] T1–T10 executados (ou N/A justificado).
- [ ] Achados com **PoC + severidade + remediação**; segredos mascarados.
- [ ] **0 Crítico/Alto aberto** (ou aceite de risco registrado).
- [ ] Relatório arquivado sem segredos; bloqueantes viraram tasks de Dev.

## Anti-patterns
- ❌ Rodar fora do próprio local/dev sem autorização escrita. ❌ DoS. ❌ Baixar a base.
- ❌ Colar segredo/credencial real no relatório. ❌ "Seguro" sem ter tentado T1–T4.
