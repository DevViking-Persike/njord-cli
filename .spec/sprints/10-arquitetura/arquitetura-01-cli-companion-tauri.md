# Arquitetura 01 - CLI companion do njord-tauri

## Entrada

Discovery aprovado em `.spec/sprints/00-discovery/discovery-01-cli-companion-tauri.md`.

## Plano tecnico

- Criar pacote `internal/app/tauri` com logica pura e testavel:
  - resolver root do `njord-tauri`;
  - validar marcadores de projeto (`package.json`, `src-tauri/Cargo.toml`);
  - ler metadados do `package.json`;
  - escolher script de dev preferindo `dev:clean` quando existir;
  - montar status textual.
- Manter `cmd/njord/main.go` como wiring Cobra fino.
- `tauri status` imprime status em stdout.
- `tauri dev` executa `npm run <script>` no root resolvido, com `--dry-run` para
  validacao sem processo longo.

## Camadas e contratos

| Camada | Responsabilidade |
|---|---|
| `cmd/njord` | registrar comandos Cobra, flags e IO |
| `internal/app/tauri` | regras de descoberta, validacao e comando operacional |
| sistema operacional | execucao real de `npm` apenas em `tauri dev` |

Contratos:

- `--path` tem maior precedencia.
- `NJORD_TAURI_PATH` vem depois.
- Diretorio irmao `../njord-tauri` e fallback.
- Erros devem indicar o path analisado.
- O comando raiz sem subcomando nao muda.

## Riscos

- Processo longo em `tauri dev`; mitigado com `--dry-run` e testes em logica pura.
- Dependencia de `npm`; a CLI apenas delega e propaga erro.
- `package.json` pode ser invalido; parser retorna erro claro.

## ADR

Nao necessario neste incremento. A decisao e pequena, local e reversivel.

## Veredito design

Aprovado. A fatia respeita as camadas atuais do Go e evita acoplamento ao banco
interno do `njord-tauri`.

## Veredito review

PASS.

## Review do diff

- Camadas: PASS. Regras de descoberta/validacao/comando ficam em
  `internal/app/tauri`; `cmd/njord/tauri.go` faz apenas wiring Cobra e IO.
- Contrato de stdout/stderr: PASS. `tauri status` e `tauri dev --dry-run`
  escrevem saida operacional em stdout; a TUI raiz permanece isolada em stderr.
- Tamanho de arquivo: PASS apos correcao. `cmd/njord/main.go` ficou com 259
  linhas e `cmd/njord/tauri.go` com 82 linhas.
- Seguranca: PASS. Nao le secrets, nao acessa SurrealDB, nao imprime tokens.
- Criterios de aceitacao: PASS. Path explicito, env var, dry-run, erro invalido
  e preservacao do comando raiz foram cobertos por testes ou smoke local.

Debito registrado: integracao mais profunda com o backend Rust/Tauri fica para
incrementos futuros.
