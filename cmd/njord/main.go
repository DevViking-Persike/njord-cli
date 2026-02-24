package main

import (
	"fmt"
	"os"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/docker"
	"github.com/DevViking-Persike/njord-cli/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	version    = "0.1.0"
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "njord-cli",
		Short:   "Njord - Project & Docker Manager TUI",
		Version: version,
		RunE:    runTUI,
	}

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path (default: ~/.config/njord/njord.yaml)")

	migrateCmd := &cobra.Command{
		Use:   "migrate [data.sh path]",
		Short: "Migrate data.sh to njord.yaml",
		Args:  cobra.ExactArgs(1),
		RunE:  runMigrate,
	}

	rootCmd.AddCommand(migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	if configPath == "" {
		configPath = config.DefaultConfigPath()
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		// If config doesn't exist, create a default one
		fmt.Fprintf(os.Stderr, "Config not found at %s, creating default...\n", configPath)
		cfg = defaultConfig()
		if saveErr := config.Save(cfg, configPath); saveErr != nil {
			return fmt.Errorf("creating default config: %w", saveErr)
		}
		fmt.Fprintf(os.Stderr, "Default config created at %s\n", configPath)
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Docker not available: %s\n", err)
		dockerClient = nil
	}
	if dockerClient != nil {
		defer dockerClient.Close()
	}

	app := ui.NewApp(cfg, dockerClient, configPath)

	// Run TUI on stderr (alternate screen), keep stdout clean for shell commands
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithOutput(os.Stderr))

	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	// Output shell command to stdout if a project was selected
	if appModel, ok := model.(ui.AppModel); ok {
		if result := appModel.GetResult(); result != nil && result.Command != "" {
			fmt.Print(result.Command)
		}
	}

	return nil
}

func runMigrate(cmd *cobra.Command, args []string) error {
	dataShPath := args[0]

	cfg, err := config.MigrateFromDataSh(dataShPath)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	outPath := configPath
	if outPath == "" {
		outPath = config.DefaultConfigPath()
	}

	if err := config.Save(cfg, outPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Migration complete! Config saved to %s\n", outPath)
	fmt.Fprintf(os.Stderr, "Categories: %d\n", len(cfg.Categories))

	total := 0
	for _, cat := range cfg.Categories {
		total += len(cat.Projects)
		fmt.Fprintf(os.Stderr, "  - %s: %d projects\n", cat.Name, len(cat.Projects))
	}
	fmt.Fprintf(os.Stderr, "Total projects: %d\n", total)
	fmt.Fprintf(os.Stderr, "Docker stacks: %d\n", len(cfg.DockerStacks))

	return nil
}

func defaultConfig() *config.Config {
	return &config.Config{
		Settings: config.Settings{
			Editor:       "code",
			ProjectsBase: "~/Avita",
			PersonalBase: "~/Persike",
		},
		Categories: []config.Category{
			{
				ID:   "alfandega",
				Name: "Alfandega",
				Sub:  "Frontend + API",
				Projects: []config.Project{
					{Alias: "avita-alfa", Desc: "Alfandega Frontend Angular", Path: "sgi-modulo-alfandega-angular-typescript"},
					{Alias: "avita-api", Desc: "Alfandega API .NET", Path: "sgi-alfandega-gestao-api-dotnet"},
				},
			},
			{
				ID:   "jobs",
				Name: "Jobs",
				Sub:  "Processamento",
				Projects: []config.Project{
					{Alias: "avita-proc", Desc: "Alfandega Processamento Job", Path: "sgi-alfandega-processamento-arquivo-job-dotnet"},
					{Alias: "avita-valid", Desc: "Alfandega Validacao Job", Path: "sgi-alfandega-validacao-job-dotnet"},
					{Alias: "avita-import", Desc: "Alfandega Importacao Job", Path: "sgi-alfandega-importacao-job-dotnet"},
					{Alias: "avita-col", Desc: "Pre-Alfandega Coleta Job", Path: "sgi-pre-alfandega-coleta-job-dotnet"},
					{Alias: "avita-pre-proc", Desc: "Pre-Alfandega Processamento Job", Path: "sgi-pre-alfandega-processamento-job-dotnet"},
				},
			},
			{
				ID:   "financeiro",
				Name: "Financeiro",
				Sub:  "Modulos financeiros",
				Projects: []config.Project{
					{Alias: "avita-fin", Desc: "SGA Modulo Financeiro Frontend", Path: "sga-modulo-financeiro-angular-typescript"},
					{Alias: "avita-fin-gs", Desc: "FIN Gestao Financeira Frontend", Path: "fin-modulo-gestao-financeira-angular-typescript"},
					{Alias: "avita-fin-api", Desc: "FIN Financeiro API .NET", Path: "fin-financeiro-api-dotnet"},
					{Alias: "avita-fin-gs-api", Desc: "FIN Gestao Financeira API .NET", Path: "fin-gestao-financeira-api-dotnet"},
					{Alias: "avita-fin-domain-lib", Desc: "Domain do Repo Fin", Path: "fin-domain-library-dotnet"},
					{Alias: "avita-fin-repository", Desc: "Repo Fin", Path: "fin-repository-library-dotnet"},
				},
			},
			{
				ID:   "core",
				Name: "Core / APIs",
				Sub:  "APIs e bibliotecas",
				Projects: []config.Project{
					{Alias: "avita-sgo", Desc: "SGO Backoffice API .NET", Path: "sgo-backoffice-api-dotnet"},
					{Alias: "avita-core", Desc: "SGA API Core", Path: "sga-api-core"},
					{Alias: "avita-apol", Desc: "SGA Apolice API", Path: "sga-apolice-api-dotnet"},
					{Alias: "avita-canc", Desc: "SGA Cancelamento Manual API", Path: "sga-cancelamento-manual-apolices-api-dotnet"},
					{Alias: "avita-pan", Desc: "PAN Notificacao API", Path: "pan-notificacao-api-dotnet"},
					{Alias: "avita-orq", Desc: "PAN Orquestrador API", Path: "pan-orquestrador-api-dotnet"},
					{Alias: "avita-docs", Desc: "Core Documents API", Path: "core-documents-api-dotnet"},
					{Alias: "avita-repo", Desc: "SGA Core Repository Library", Path: "sga-core-repository-library-dotnet"},
					{Alias: "avita-domain", Desc: "SGA Domain Library .NET", Path: "sga-domain-library-dotnet"},
					{Alias: "avita-plat", Desc: "Plataforma Portal Admin", Path: "plat-portal-admin-angular-typescript"},
				},
			},
			{
				ID:   "banco",
				Name: "Banco de Dados",
				Sub:  "DDL + DML",
				Projects: []config.Project{
					{Alias: "avita-ddl", Desc: "SGA DDL Liquibase", Path: "sga-ddl-db-liquibase-sql"},
					{Alias: "avita-dml", Desc: "SGA DML Liquibase", Path: "sga-dml-db-liquibase-sql"},
					{Alias: "avita-fin-db-relator", Desc: "DB Liquid Financeiro", Path: "fin-db-relatorio-ddl-db-liquibase-sql"},
				},
			},
			{
				ID:   "infra",
				Name: "Infra",
				Sub:  "Stack dev",
				Projects: []config.Project{
					{Alias: "avita-gap", Desc: "GAP Stack Desenvolvimento", Path: "gap-stack-desenvolvimento"},
				},
			},
			{
				ID:   "env",
				Name: "ENV",
				Sub:  "Ambientes",
				Projects: []config.Project{
					{Alias: "env-canc", Desc: "SGA Cancelamento Manual API (ENV)", Path: "env/sga-cancelamento-manual-apolices-api-dotnet"},
				},
			},
			{
				ID:   "pessoal",
				Name: "Pessoal",
				Sub:  "Projetos pessoais",
				Projects: []config.Project{
					{Alias: "uni-front", Desc: "App Unificado Frontend", Path: "Persike/ProjetoUnificado/AppUnificado"},
					{Alias: "uni-back", Desc: "App Unificado Backend", Path: "Persike/ProjetoUnificado/AppUnificadoBack"},
					{Alias: "RPAs", Desc: "99frela - RPAs", Path: "Persike/99frela"},
					{Alias: "ares-llm", Desc: "vm local llm", Path: "Persike/ares-llm"},
				},
			},
			{
				ID:   "vps",
				Name: "VPS",
				Sub:  "Acesso remoto",
				Projects: []config.Project{
					{Alias: "tron", Desc: "VPS Remoto via RDP (Cloudflare Tunnel)", Path: "@rdp"},
				},
			},
		},
		DockerStacks: []config.DockerStack{
			{Name: "GAP Stack", Desc: "MySQL Database (porta 3306)", Path: "gap-stack-desenvolvimento"},
			{Name: "Cancelamento", Desc: "APIs + MongoDB", Path: "sga-cancelamento-manual-apolices-api-dotnet"},
			{Name: "Alfandega Jobs", Desc: "Jobs Alfandega", Path: "sgi-modulo-alfandega-angular-typescript"},
		},
	}
}
