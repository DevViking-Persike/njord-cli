# Regra 10 — Arquitetura de frontend (MVVM + Atomic Design)

O frontend segue **MVVM** (Model / ViewModel / View) combinado com **Atomic
Design** para a camada de View. Quando o projeto tiver ADRs de frontend,
registre esta decisão lá e mantenha as regras auditáveis.

## Motivação
Separar estado/regra (ViewModel + Model) de apresentação (View) torna a UI testável sem Tauri, dá fronteiras grepáveis e permite paralelismo seguro entre features. Atomic classifica as Views por acoplamento a domínio.

## Camadas MVVM

| Papel | O que é | Onde mora | Arquivo |
|-------|---------|-----------|---------|
| **Model** | tipos/domínio puros, use cases, clients `invoke`/`listen` | `domain/`, `application/`, `infrastructure/` | `.ts` |
| **ViewModel** | `$state`/`$derived` + ações; orquestra o Model via **port injetado**; sem markup, sem `invoke` direto | `presentation/view-models/<assunto>/` | `*-view-model.svelte.ts` |
| **View** | markup + binding + eventos; lê do VM, chama métodos do VM | `presentation/{pages,components}`, `shared/components` | `.svelte` |

- **VM = factory function** por padrão (closure expondo runes por getters). Classe só para herança/variações.
- VM recebe **port** (interface em `application/ports/*-port.ts`); a implementação real fica em `infrastructure/*-tauri.ts`. VM nunca chama `invoke`.
- Estado compartilhado: VM instanciado no composition root (`src/app/providers/`) e injetado por **contexto Svelte tipado** — não `export const x = createX()` global. VM escopado a uma subárvore vai por prop.

## Atomic (classificação da View)

| Nível | Conhece domínio/VM? | Onde mora |
|-------|---------------------|-----------|
| **Atom** | Não (primitivo puro) | `shared/components/atoms` |
| **Molecule** | Não (composição de atoms) | `shared/components/molecules` |
| **Organism** | Sim (recebe VM por prop/contexto) | `modules/<m>/presentation/components/<assunto>` |
| **Page** | Sim (obtém VMs, compõe organisms) | `modules/<m>/presentation/pages` |

Atom/molecule **nunca** importa `$modules`/`$studio`/`domain`/`invoke`, e usa só tokens semânticos de CSS (sem hex, sem `--color-*` cru). Precisou de tipo de domínio ou VM → é organism/page.

## Aliases (obrigatórios)
`$app`, `$shared`, `$modules`, `$studio`, `$atoms`, `$molecules`, `$organisms`. Proibido import relativo entre módulos (`../../modules/...`).

## Casing
Dir de módulo e arquivos `.ts`/`.svelte.ts`: `kebab-case`. Componente `.svelte`: `PascalCase`. Identificador TS: `camelCase`/`PascalCase`. VM sempre `*-view-model.svelte.ts`; port `*-port.ts`; infra `*-<tech>.ts` (ex.: `*-tauri.ts`).

## Como verificar
```bash
# View (.svelte) sem IO direto
rg -l "invoke\(|@tauri-apps/api|listen\(" -g '*.svelte' src                              # vazio
# ViewModel não importa View nem fala infra direto
rg -ln "from '.*\.svelte'" src --glob '*view-model.svelte.ts'                            # vazio
rg -l "invoke\(|infrastructure/" src --glob '*view-model.svelte.ts'                      # vazio
# Atoms/molecules domain-free e sem cor crua
rg -l "\$modules|\$studio|/domain/|invoke\(" src/lib/components/atoms src/lib/components/molecules    # vazio
rg -n "#[0-9a-fA-F]{3,6}|var\(--color-" src/lib/components/atoms src/lib/components/molecules          # vazio
# Sem VM singleton global
rg -n "^export const \w+ = create\w+(Store|ViewModel)\(" src                             # vazio (alvo)
```
(Durante migrações, `$shared/components` pode apontar para `src/lib/components`
via alias; os greps acima usam o path físico atual.)

## Exceções aceitas
- Migração incremental: shims de re-export em paths antigos durante a transição (removidos na fase final). Devem ter comentário `Remover na Fase 5`.
- Arquivos legados ainda não migrados não violam a regra retroativamente; migram
  conforme o plano de frontend do projeto.
- Singleton global de VM é tolerado **apenas** em código ainda não invertido para contexto; novo código não cria singleton.
- **Button/Select duplicados** (`$atoms/Button` + `panels/settings/atoms/Button`; `$molecules/Select` + `panels/settings/atoms/Select`): mantidos separados por ora porque são **estilos visuais distintos**, não só tokens — o do settings é *filled* (fundo accent sólido, variantes `ghost`/`soft`, `iconLeft`/`iconRight`) e o de `$atoms` é *outlined* (variantes `secondary`, `size`, `fullWidth`). Unificá-los exige uma **decisão de design** (qual estilo é o canônico) + verificação visual — não é refactor mecânico. Consolidar quando essa decisão for tomada; até lá, novo código prefere `$atoms/Button`.

## Localização do design system compartilhado (decisão — fecha a "Fase 5")

Quando o projeto prever primitivos compartilhados em `src/shared/components/`,
**a fronteira é o ALIAS, não o path físico**: `$atoms`/`$molecules`/`$organisms`
(+ `$shared/components` durante a migração) são o contrato. Se os componentes
ficarem em `src/lib/components/{atoms,molecules,organisms}`, consuma-os só pelos
aliases — isso já entrega a fronteira arquitetural.

O move físico para `src/shared/` foi **deliberadamente adiado** porque:
- **Organisms têm domínio** (importam `toast_store`, `db_connection_list`, `routes/dados`, etc. — por design, conforme o README). `src/shared/` é zero-domínio → organisms com feature **não pertencem** a `shared`; o destino correto deles é dentro dos **módulos de feature**, não `shared`. Mover para `shared` seria *errado*.
- atoms/molecules têm alguns imports relativos a helpers de lib (`theme`, `focus_trap`, `toast_store`, `search_tabs`) que precisariam virar aliases/mover junto antes do move físico.

**Regra:** novo primitivo de UI sem domínio → `$atoms`/`$molecules`. Novo bloco com domínio → organism dentro do módulo de feature (`modules/<m>/presentation/components`), **não** em `lib/components/organisms`. A relocação física `lib → shared` + extração dos organisms-de-domínio para os módulos é refinamento futuro, sem impacto no contrato de alias.
