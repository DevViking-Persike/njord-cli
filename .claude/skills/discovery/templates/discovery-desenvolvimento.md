# Discovery NN — <tema> `[DISCOVERY · DESENVOLVIMENTO]`

> Modo desenvolvimento: o **como/escopo**. Gera contexto técnico (base pra
> construir/refatorar). Critérios verificáveis viram teste no QA.

## 1. Escopo
- **Faz:** <casos de uso principais> · **NÃO faz:** <fora>
- **Menor fatia entregável (vertical slice):** <...> · baseline × aditivo: <...>

## 2. Requisitos funcionais
- Entradas/saídas/regras/estados/erros: <...>
- Atores e autorização (quem pode o quê): <...>

## 3. NFR — top 3–5 atributos de qualidade (com número)
| Atributo | Alvo mensurável |
|---|---|
| Performance/escala | <X req/s, p95 < Y ms, N usuários, volume de dados> |
| Disponibilidade | <SLO %, comportamento em falha, recuperação> |
| Segurança | <dados sensíveis, authn/authz, auditoria, LGPD/compliance> |
| Manutenibilidade | <testabilidade, observabilidade, quem mantém> |
| <outro> | <...> |

## 4. Restrições
- Stack/arquitetura obrigatória/existente: <...> · Legado a respeitar: <...>
- Integrações de terceiros (limites/custos/SLA/rate limit): <...>
- Prazo / equipe / orçamento / legal: <...>

## 5. Premissas & riscos
- **Premissas** (a que, se falsa, derruba o plano): <...>
- **Riscos** (técnico / compliance — retrofit ~3× / dependência): <...>
- **Spikes** necessários (provar viabilidade antes): <...>

## 6. Achados de subagents (se usados)
- **Subagents criados:** <código/legado, NFR/riscos, docs/operabilidade, outro>
- **Evidências úteis:** <fatos observados, com caminho/linha quando houver>
- **Lacunas para usuário:** <perguntas ainda abertas>
- **Hipóteses não confirmadas:** <não tratar como verdade>

## 7. Dependências & aceitação
- **Depende de:** <sistemas/serviços/equipes>
- **Critérios de aceitação (verificáveis):**
  1. **Dado** <contexto> **quando** <ação> **então** <resultado observável>.

> **Refatorar:** comportamento atual a preservar (não-regressão): <...> ·
> caracterização (testes) que cobre: <...>

## Definition of Ready (DoD da Discovery dev)
- [ ] Escopo (faz × não faz × slice) · [ ] requisitos funcionais
- [ ] NFR top 3–5 **com número** · [ ] restrições/premissas/riscos mapeados
- [ ] dependências · [ ] critérios de aceitação verificáveis · [ ] (refatorar) não-regressão definida
