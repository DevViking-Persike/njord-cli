# Regra 3 — SOLID

## SRP — Single Responsibility
Um arquivo, um motivo para mudar. Separar regras de negócio de I/O, UI de estado.

Anti-exemplo: `internal/ui/add_project.go` com 746 linhas misturando form, validação, clone git, e persistência — cada uma deve estar em arquivo/pacote distinto.

## OCP — Open/Closed
Prefira injetar dependências (interfaces) a importar struct concreta quando o ponto de extensão é previsível.

Exemplo bom: `gitlab.Client` injetado no TUI — fácil mockar, fácil trocar provider.

## LSP — Liskov
Interfaces pequenas; não quebre contratos em implementações. Se uma implementação precisa `panic` em métodos herdados, a interface está errada.

## ISP — Interface Segregation
Uma interface por papel. Evite interfaces "gordas" (`DockerClient` com 15 métodos quando o consumidor usa 2).

Exemplo: criar `StackStarter interface { StartProject(path, name) error }` em vez de passar `*docker.Client` inteiro.

## DIP — Dependency Inversion
Pacotes de alto nível (`internal/ui`, `internal/app`) dependem de abstrações, não de SDKs/`os/exec` diretamente.

Chamada concreta fica em `internal/{docker,gitlab,git}`; consumidor declara a interface de que precisa.

## Como verificar
```bash
rg -l '"os/exec"' internal/ui/ internal/app/   # deve vir vazio ou com exceção justificada
rg -l 'njord-cli/internal/ui' internal/app/    # app NUNCA importa ui
```
