# Regra 4 — Clean Architecture

## Camadas (de dentro pra fora)
1. `src-tauri/src/app/` — regras de negócio puras (Rust, testáveis sem infra nem Tauri)
2. `src-tauri/src/gateways/{docker,gitlab,git,config}/` — gateways/infra (exec, HTTP, FS)
3. `src-tauri/src/commands/` — ponte Tauri (`#[tauri::command]` thin handlers)
4. `src/lib/components/` — componentes Svelte de UI
5. `src/routes/` — composition root do frontend (telas/rotas SvelteKit)

## Regras de dependência
- **Fluxo aponta sempre para dentro.**
  - Backend: `commands → app → (nada externo importa app)`. `gateways → app` apenas via traits definidas em `app`.
  - Frontend: `routes → lib/components → invoke('cmd')`. Nunca importa direto de `src-tauri/`.
- `src-tauri/src/app/` nunca importa `src-tauri/src/commands/` nem `tauri::`.
- `src-tauri/src/gateways/` nunca importa `src-tauri/src/commands/` nem componentes Svelte.
- `src/` (frontend) nunca importa `src-tauri/` — comunicação só via `invoke`/eventos Tauri.
- `src-tauri/src/commands/` é a única camada que cita Tauri E regras de negócio juntas (wiring).

## Onde colocar o quê
- **Regra de negócio** (ex.: "se branch é subtask, disparar pipeline depois do push"): `src-tauri/src/app/`.
- **Chamada ao Docker/GitLab/Git/FS**: gateway correspondente em `src-tauri/src/gateways/`.
- **Renderização e estado de tela**: `src/lib/components/` ou `src/routes/`.
- **Comando Tauri (`#[tauri::command]`)**: `src-tauri/src/commands/`. Mantenha thin — só desserializa input, chama `app`, devolve resultado.
- **Wiring/bootstrap** (`tauri::Builder`, registros, injeção de gateways): `src-tauri/src/main.rs` ou `src-tauri/src/lib.rs`.

## Teste seco
Se `src-tauri/src/app/*.rs` importa `tauri::`, `reqwest::`, `std::process::Command`, ou qualquer SDK externo, é violação — mover a chamada para o gateway.

## Como verificar
```bash
# app puro (sem Tauri, sem IO bruto)
rg -l 'tauri::|reqwest::|std::process::Command' src-tauri/src/app/

# gateways não conhecem UI nem commands
rg -l 'crate::commands' src-tauri/src/gateways/

# frontend não importa backend
rg -l 'src-tauri' src/
```
Saída esperada: vazia.
