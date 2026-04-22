---
description: Refatorar um arquivo aplicando as regras em .claude/rules/
argument-hint: <caminho/do/arquivo.go>
---

Refatore `$ARGUMENTS` seguindo as regras em `.claude/rules/`.

Leia antes de começar:
- `.claude/rules/01-file-size.md` (alvo ≤ 300 linhas)
- `.claude/rules/03-solid.md` (SRP, DIP)
- `.claude/rules/04-clean-architecture.md` (camadas)
- `.claude/rules/05-simplicity.md` (anti-patterns)
- `.claude/rules/06-continuous-refactoring.md` (ordem de trabalho)

## Fluxo

### 1. Leia o arquivo inteiro antes de propor mudança.

### 2. Diagnóstico (antes de editar)
- Linhas atuais vs alvo
- Responsabilidades distintas (cada uma candidata a novo arquivo/pacote)
- Imports violando camada (ex.: `os/exec` em `internal/ui/`)
- Funções > 60 linhas
- Cobertura atual: `go test -cover ./<pkg>/`

### 3. Peça confirmação do plano ao usuário.

### 4. Execução (na ordem)

**a. Rede de segurança**
- Função alvo sem teste → escrever teste de caracterização primeiro (regra 6).
- `go test ./<pkg>/` deve passar antes de qualquer mudança estrutural.

**b. Split por responsabilidade**
- Novos arquivos no mesmo pacote: `<feature>_<subresponsibility>.go` (referência: `internal/ui/gitlab_actions_branch.go`).
- Responsabilidade de outra camada → mover para pacote correto (regra 4).
- Preservar API pública, salvo escopo aprovado.

**c. Injeção de dependências (regra 3 DIP)**
- Substituir chamadas concretas a SDK/exec por interfaces do consumidor.
- Struct concreta permanece no gateway; mock em `_test.go`.

**d. Simplificação (regra 5)**
- Remover wrappers/flags que só repassam.
- Inlinar funções usadas em 1 lugar.
- Deletar código morto (não comentar).

### 5. Validação final
```bash
go test ./...
go build ./cmd/njord/
go vet ./...
wc -l <arquivos_mexidos>
```
Se alvo era `internal/app` ou `internal/docker`: `gremlins unleash ./<pkg>/` e reportar variação.

### 6. Relatório final
- Antes/depois: linhas, cobertura, eficácia
- Arquivos criados/removidos/modificados
- Mensagem de commit sugerida (não commitar sem pedido)

## Regras de comportamento
- **Nunca** remova testes para "simplificar".
- **Nunca** use `--no-verify` ou pule hooks.
- Bug pré-existente descoberto → parar e perguntar (regra 6).
- `$ARGUMENTS` vazio/inexistente → pedir confirmação do alvo.
