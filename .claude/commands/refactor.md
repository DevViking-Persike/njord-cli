---
description: Refatorar um arquivo seguindo as regras do CLAUDE.md
argument-hint: <caminho/do/arquivo.go>
---

Refatore o arquivo `$ARGUMENTS` seguindo as **Regras de Engenharia** em `CLAUDE.md`.

## Passos obrigatórios

1. **Leia o arquivo inteiro** antes de propor qualquer mudança.

2. **Diagnóstico** — antes de editar, produza um plano listando:
   - Linhas totais (atual vs alvo ≤ 300)
   - Responsabilidades distintas presentes (cada uma candidata a um novo arquivo/pacote)
   - Dependências para fora da camada (ex.: `internal/ui` chamando `os/exec` → deve mover pra gateway)
   - Funções > 60 linhas — candidatas a split
   - Cobertura atual do pacote (`go test -cover ./<pkg>/`)

3. **Peça confirmação** do plano ao usuário antes de editar. Se o usuário aprovar, execute na ordem:

   a. **Garantir rede de segurança**
      - Se a função-alvo não tem teste, escreva teste caracterizando o comportamento atual primeiro.
      - `go test ./<pkg>/` deve passar antes de qualquer mudança estrutural.

   b. **Split por responsabilidade**
      - Novos arquivos no mesmo pacote seguindo o padrão `<feature>_<subresponsibility>.go` (ver `internal/ui/gitlab_actions_branch.go` como referência existente).
      - Se a responsabilidade pertence a outra camada (regra 4), mover para `internal/app/`, `internal/docker/`, `internal/gitlab/` etc.
      - Preservar a API pública — consumidores não devem precisar mudar, a menos que a mudança de API seja parte do escopo aprovado.

   c. **Injeção de dependências**
      - Substituir chamadas concretas a SDKs/exec por interfaces definidas no consumidor.
      - Struct concreta fica no gateway (`internal/docker/*`), mock fica em `_test.go`.

   d. **Simplificação**
      - Remover wrappers/flags que só repassam parâmetros.
      - Inlinar funções usadas em um único lugar se isso reduz carga cognitiva.
      - Deletar código morto (não comentá-lo — deletar).

4. **Validação final** — sempre, no fim:
   - `go test ./...`
   - `go build ./cmd/njord/`
   - `go vet ./...`
   - `wc -l <arquivos_mexidos>` — confirmar ≤ 300
   - Se pacote alvo era `internal/app` ou `internal/docker`, rodar `gremlins unleash ./<pkg>/` e reportar variação de eficácia.

5. **Relatório final** em resposta ao usuário:
   - Antes/depois: linhas, cobertura, eficácia (se aplicável)
   - Lista de arquivos novos/removidos/modificados
   - Mensagem de commit sugerida (não commitar sem pedido explícito)

## Regras de comportamento

- **Nunca** remova testes existentes para "simplificar".
- **Nunca** use `--no-verify` ou pule hooks.
- Se a refatoração revelar bug pré-existente, **não conserte no mesmo commit** — reportar e perguntar se deve criar commit separado.
- Se `$ARGUMENTS` estiver vazio ou apontar para arquivo inexistente, pedir confirmação do alvo.
