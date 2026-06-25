# Reference - njord-cli

Docs de referencia criadas pela esteira.

## Inventario inicial

- `njord-cli`: Go 1.24, Cobra, Bubbletea, Lipgloss, Koanf, Docker SDK e clientes
  GitLab/GitHub/Jira.
- `njord-tauri`: Tauri 2, SvelteKit/Svelte 5 e Rust, com backend organizado por
  bounded contexts em `src-tauri/src/modules`.
- O primeiro incremento nao acessa diretamente a persistencia interna do
  `njord-tauri`; a CLI atua como companion operacional por path, scripts e
  status de projeto.
