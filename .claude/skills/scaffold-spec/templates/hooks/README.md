# Hooks recomendados para Claude Code (opt-in)

Hooks que ajudam a **entregar o projeto claro e funcionando** — validam a base
`.spec/` automaticamente. **Não são auto-aplicados** (mexer no `settings.json` é
intrusivo); instale conscientemente.

## O que fazem
- **Stop** — ao terminar um turno, roda `spec-check.sh` (estrutura + links). Se
  falhar, imprime um aviso (não bloqueia).
- **PostToolUse (Write|Edit)** — ao editar um arquivo sob `.spec/`, revalida os
  links/estrutura na hora. Requer `jq`.

## Instalar
Mescle o conteúdo de `settings.hooks.json` no `.claude/settings.json` do projeto
(some o objeto `hooks`; **não** substitua o arquivo). No projeto há a skill
`update-config` para isso — ou edite à mão:

```bash
# pré-requisito: a tool instalada em .claude/tools/spec-check.sh
ls .claude/tools/spec-check.sh
# depois, adicione o bloco "hooks" de settings.hooks.json ao seu .claude/settings.json
```

> Só o bloco **Stop** já cobre o essencial (valida no fim de cada turno). O
> **PostToolUse** é mais imediato mas mais verboso — opcional.
