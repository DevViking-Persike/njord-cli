# ADR-001: Pacote components para helpers compartilhados

## Status
Aceito

## Contexto
Os arquivos `settings.go` (771 linhas) e `gitlab_actions.go` (1052 linhas) acumularam codigo duplicado: navegacao de lista (up/down + bounds), text input (backspace, ctrl+u, runes), scroll indicators e menu rendering apareciam repetidos entre 3 e 12 vezes cada.

## Decisao
Extrair helpers de navegacao, scroll e renderizacao em `internal/ui/components/` com tres arquivos:
- `nav.go` — `ListNav`, `TextInput`, `DigitsOnly`
- `scroll.go` — `ScrollState` com `VisibleRows`, `EnsureVisible`, `Bounds`
- `render.go` — `RenderMenuOptions`, `RenderTextInput`, `RenderScrollUp/Down`, `RenderMessage`, `RenderError`, `SaveConfig`

## Alternativa rejeitada
Interfaces e generics — adicionariam complexidade desnecessaria para helpers simples que sao funcoes puras ou structs com poucos metodos.

## Consequencias
- `settings.go` reduziu de 771 para ~608 linhas (~21%)
- `gitlab_actions.go` reduziu de 1052 para ~915 linhas (~13%)
- Novos arquivos totalizam ~150 linhas
- Qualquer nova tela pode reutilizar os mesmos helpers
- Zero mudancas visuais ou funcionais
