# ADR-003: SaveConfig wrapper

## Status
Aceito

## Contexto
O padrao `config.Save` + formatacao de erro/sucesso aparece 7 vezes so no `settings.go`:

```go
if err := config.Save(m.cfg, m.configPath); err != nil {
    m.message = fmt.Sprintf("Erro ao salvar: %s", err)
    m.messageType = "error"
} else {
    m.message = "Mensagem de sucesso"
    m.messageType = "ok"
}
```

## Decisao
Wrapper `SaveConfig(cfg, path, successMsg) (message, msgType string)` em `components/render.go` que encapsula o padrao save+error em uma unica chamada.

## Alternativa rejeitada
Metodo no model — nao seria compartilhavel entre diferentes models sem embedding ou interfaces.

## Consequencias
- 7 blocos de 6 linhas substituidos por uma linha cada
- Mensagem de erro padronizada ("Erro ao salvar: ...") em todos os locais
- Facilmente reutilizavel por qualquer nova tela que precise salvar config
