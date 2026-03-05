# ADR-002: ScrollState centralizado

## Status
Aceito

## Contexto
O padrao offset+visibleRows+ensureVisible se repete em 5 telas (settings, gitlab_actions, gitlab, grid, docker_actions), cada uma com implementacao quase identica: calculo de linhas visiveis baseado em height menos chrome, ajuste de offset para manter cursor visivel, e calculo de bounds (start, end) para slice de itens.

## Decisao
Usar struct `ScrollState` com campos `Offset` e `Height`, e metodos:
- `VisibleRows(chromeLines)` — calcula linhas disponiveis
- `EnsureVisible(cursor, chromeLines)` — ajusta offset para cursor
- `Bounds(total, chromeLines)` — retorna (start, end)

Cada tela define sua constante de `chromeLines` (ex: `settingsChromeLines = 9`, `glActionsChromeLines = 10`).

## Alternativa rejeitada
Manter os campos `offset` e `height` espalhados e metodos `visibleRows()`/`ensureVisible()` em cada model — funciona mas resulta em duplicacao exata de logica.

## Consequencias
- Uma unica implementacao de scroll testavel
- Cada model substitui dois campos (`offset`, `height`) por um (`scroll ScrollState`)
- Constante de chrome lines torna explicito o overhead de cada tela
