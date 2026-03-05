package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addStep int

const (
	stepGitURL addStep = iota
	stepDestination
	stepClone
	stepAlias
	stepDescription
	stepCategory
	stepCustomCatName
	stepCustomCatSub
	stepGroup
	stepCustomGroup
	stepConfirm
	stepDone
)

type cloneDoneMsg struct {
	err error
}

type AddProjectModel struct {
	cfg        *config.Config
	configPath string
	step       addStep
	goBack     bool

	// Input state
	gitURL      string
	destination string
	clonePath   string
	alias       string
	description string
	categoryID  string
	newCatName  string
	newCatSub   string
	group       string

	// UI state
	inputBuf    string
	cursor      int
	options     []string
	message     string
	messageType string
	cloning     bool
	width       int
	height      int
}

func NewAddProjectModel(cfg *config.Config, configPath string) AddProjectModel {
	return AddProjectModel{
		cfg:        cfg,
		configPath: configPath,
		step:       stepGitURL,
	}
}

func (m AddProjectModel) Init() tea.Cmd {
	return nil
}

func (m AddProjectModel) Update(msg tea.Msg) (AddProjectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case cloneDoneMsg:
		m.cloning = false
		if msg.err != nil {
			m.message = fmt.Sprintf("Erro no clone: %s", msg.err)
			m.messageType = "error"
			m.step = stepGitURL
		} else {
			m.message = "Clone concluído!"
			m.messageType = "ok"
			m.alias = suggestAlias(m.gitURL)
			m.inputBuf = m.alias
			m.step = stepAlias
		}
		return m, nil

	case tea.KeyMsg:
		if m.cloning {
			return m, nil
		}

		switch m.step {
		case stepGitURL:
			return m.handleTextInput(msg)
		case stepDestination:
			return m.handleDestinationSelect(msg)
		case stepClone:
			return m, nil
		case stepAlias:
			return m.handleTextInput(msg)
		case stepDescription:
			return m.handleTextInput(msg)
		case stepCategory:
			return m.handleCategorySelect(msg)
		case stepCustomCatName:
			return m.handleTextInput(msg)
		case stepCustomCatSub:
			return m.handleTextInput(msg)
		case stepGroup:
			return m.handleGroupSelect(msg)
		case stepCustomGroup:
			return m.handleTextInput(msg)
		case stepConfirm:
			return m.handleConfirm(msg)
		case stepDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.goBack = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m AddProjectModel) View() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(theme.AddGreen).Render("  + Adicionar Projeto")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	// Progress
	type progressStep struct {
		name string
		step addStep
	}
	progressSteps := []progressStep{
		{"URL", stepGitURL},
		{"Destino", stepDestination},
		{"Clone", stepClone},
		{"Alias", stepAlias},
		{"Desc", stepDescription},
		{"Categoria", stepCategory},
		{"Grupo", stepGroup},
		{"Confirmar", stepConfirm},
	}
	// Map custom cat steps to the Categoria step for progress display
	displayStep := m.step
	if displayStep == stepCustomCatName || displayStep == stepCustomCatSub {
		displayStep = stepCategory
	}
	if displayStep == stepCustomGroup {
		displayStep = stepGroup
	}
	var progress string
	for _, ps := range progressSteps {
		if ps.step == displayStep {
			progress += theme.TitleSelectedStyle.Render("["+ps.name+"]") + " "
		} else if ps.step < displayStep {
			progress += theme.SuccessStyle.Render("✓"+ps.name) + " "
		} else {
			progress += theme.DimStyle.Render(ps.name) + " "
		}
	}
	b.WriteString("  " + progress + "\n\n")

	switch m.step {
	case stepGitURL:
		b.WriteString("  " + theme.TextStyle.Render("Git URL (SSH ou HTTPS):") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case stepDestination:
		b.WriteString("  " + theme.TextStyle.Render("Destino do clone:") + "\n\n")
		options := []string{"~/Avita (Projetos)", "~/Persike (Pessoal)", "Custom path"}
		for i, opt := range options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))

	case stepClone:
		b.WriteString("  " + theme.TextStyle.Render("Clonando repositório...") + "\n\n")
		b.WriteString("  " + theme.DimStyle.Render(m.gitURL) + "\n")
		b.WriteString("  → " + theme.DimStyle.Render(m.clonePath) + "\n")

	case stepAlias:
		b.WriteString("  " + theme.TextStyle.Render("Alias do projeto:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case stepDescription:
		b.WriteString("  " + theme.TextStyle.Render("Descrição curta:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case stepCategory:
		b.WriteString("  " + theme.TextStyle.Render("Categoria:") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))

	case stepCustomCatName:
		b.WriteString("  " + theme.TextStyle.Render("Nome da nova categoria:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case stepCustomCatSub:
		b.WriteString("  " + theme.TextStyle.Render("Descrição da categoria:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case stepGroup:
		b.WriteString("  " + theme.TextStyle.Render("Grupo (subdivisão):") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc back"))

	case stepCustomGroup:
		b.WriteString("  " + theme.TextStyle.Render("Nome do novo grupo:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc back"))

	case stepConfirm:
		b.WriteString("  " + theme.TextStyle.Render("Confirmar adição:") + "\n\n")
		b.WriteString("  " + theme.DimStyle.Render("URL:     ") + theme.TextStyle.Render(m.gitURL) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Path:    ") + theme.TextStyle.Render(m.clonePath) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Alias:   ") + theme.TextStyle.Render(m.alias) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Desc:    ") + theme.TextStyle.Render(m.description) + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Cat:     ") + theme.TextStyle.Render(m.categoryID) + "\n")
		if m.group != "" {
			b.WriteString("  " + theme.DimStyle.Render("Grupo:   ") + theme.TextStyle.Render(m.group) + "\n")
		}
		b.WriteString("\n")

		options := []string{"Confirmar", "Cancelar"}
		for i, opt := range options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select"))

	case stepDone:
		if m.messageType == "ok" {
			b.WriteString("  " + theme.SuccessStyle.Render("✓ Projeto adicionado com sucesso!") + "\n\n")
			b.WriteString("  " + theme.DimStyle.Render("Alias: "+m.alias) + "\n")
			b.WriteString("  " + theme.DimStyle.Render("Config salva em: "+m.configPath) + "\n")
		} else {
			b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	}

	return b.String()
}

func (m *AddProjectModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *AddProjectModel) GoBack() bool { return m.goBack }

func (m AddProjectModel) handleTextInput(msg tea.KeyMsg) (AddProjectModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.step == stepGitURL {
			m.goBack = true
		} else if m.step == stepCustomCatName {
			m.step = stepCategory
			m.message = ""
		} else if m.step == stepCustomCatSub {
			m.step = stepCustomCatName
			m.inputBuf = m.newCatName
			m.message = ""
		} else if m.step == stepCustomGroup {
			m.buildGroupOptions()
			m.step = stepGroup
			m.message = ""
		} else {
			m.step--
			m.message = ""
		}
		return m, nil
	case "enter":
		return m.submitTextInput()
	case "backspace":
		if len(m.inputBuf) > 0 {
			m.inputBuf = m.inputBuf[:len(m.inputBuf)-1]
		}
	case "ctrl+u":
		m.inputBuf = ""
	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.inputBuf += string(msg.Runes)
		}
	}
	return m, nil
}

func (m AddProjectModel) submitTextInput() (AddProjectModel, tea.Cmd) {
	switch m.step {
	case stepGitURL:
		url := strings.TrimSpace(m.inputBuf)
		if url == "" {
			m.message = "URL não pode ser vazia"
			return m, nil
		}
		if !strings.HasPrefix(url, "git@") && !strings.HasPrefix(url, "https://") {
			m.message = "URL deve começar com git@ ou https://"
			return m, nil
		}
		m.gitURL = url
		m.inputBuf = ""
		m.message = ""
		m.cursor = 0
		m.step = stepDestination
		return m, nil

	case stepAlias:
		alias := strings.TrimSpace(m.inputBuf)
		if alias == "" {
			m.message = "Alias não pode ser vazio"
			return m, nil
		}
		for _, p := range m.cfg.AllProjects() {
			if p.Alias == alias {
				m.message = fmt.Sprintf("Alias '%s' já existe", alias)
				return m, nil
			}
		}
		m.alias = alias
		m.inputBuf = extractRepoName(m.gitURL)
		m.message = ""
		m.step = stepDescription
		return m, nil

	case stepDescription:
		desc := strings.TrimSpace(m.inputBuf)
		if desc == "" {
			desc = m.alias
		}
		m.description = desc
		m.inputBuf = ""
		m.options = nil
		for _, cat := range m.cfg.Categories {
			m.options = append(m.options, cat.Name+" ("+cat.ID+")")
		}
		m.options = append(m.options, "+ Nova categoria")
		m.cursor = 0
		m.step = stepCategory
		return m, nil

	case stepCustomCatName:
		name := strings.TrimSpace(m.inputBuf)
		if name == "" {
			m.message = "Nome não pode ser vazio"
			return m, nil
		}
		m.newCatName = name
		m.categoryID = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		m.inputBuf = ""
		m.message = ""
		m.step = stepCustomCatSub
		return m, nil

	case stepCustomCatSub:
		sub := strings.TrimSpace(m.inputBuf)
		if sub == "" {
			m.message = "Descrição não pode ser vazia"
			return m, nil
		}
		m.newCatSub = sub
		m.inputBuf = ""
		m.message = ""
		m.buildGroupOptions()
		m.step = stepGroup
		return m, nil

	case stepCustomGroup:
		g := strings.TrimSpace(m.inputBuf)
		if g == "" {
			m.message = "Nome do grupo não pode ser vazio"
			return m, nil
		}
		m.group = g
		m.inputBuf = ""
		m.message = ""
		m.cursor = 0
		m.step = stepConfirm
		return m, nil
	}
	return m, nil
}

func (m AddProjectModel) handleDestinationSelect(msg tea.KeyMsg) (AddProjectModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 2 {
			m.cursor++
		}
	case "esc":
		m.step = stepGitURL
		m.inputBuf = m.gitURL
	case "enter":
		home, _ := os.UserHomeDir()
		repoName := extractRepoName(m.gitURL)

		switch m.cursor {
		case 0:
			m.destination = "avita"
			m.clonePath = filepath.Join(home, "Avita", repoName)
		case 1:
			m.destination = "pessoal"
			m.clonePath = filepath.Join(home, "Persike", repoName)
		case 2:
			m.destination = "custom"
			m.clonePath = filepath.Join(home, repoName)
		}

		if _, err := os.Stat(m.clonePath); err == nil {
			m.message = "Diretório já existe, pulando clone"
			m.messageType = "info"
			m.alias = suggestAlias(m.gitURL)
			m.inputBuf = m.alias
			m.step = stepAlias
			return m, nil
		}

		m.step = stepClone
		m.cloning = true
		return m, m.doClone()
	}
	return m, nil
}

func (m AddProjectModel) handleCategorySelect(msg tea.KeyMsg) (AddProjectModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "esc":
		m.step = stepDescription
		m.inputBuf = m.description
	case "enter":
		if m.cursor < len(m.cfg.Categories) {
			m.categoryID = m.cfg.Categories[m.cursor].ID
			m.buildGroupOptions()
			m.step = stepGroup
		} else {
			m.inputBuf = ""
			m.message = ""
			m.step = stepCustomCatName
		}
	}
	return m, nil
}

func (m *AddProjectModel) buildGroupOptions() {
	// Collect existing groups for the selected category
	seen := make(map[string]bool)
	var existing []string
	for _, cat := range m.cfg.Categories {
		if cat.ID == m.categoryID {
			for _, p := range cat.Projects {
				if p.Group != "" && !seen[p.Group] {
					seen[p.Group] = true
					existing = append(existing, p.Group)
				}
			}
			break
		}
	}
	m.options = nil
	m.options = append(m.options, "Sem grupo")
	m.options = append(m.options, existing...)
	m.options = append(m.options, "+ Novo grupo")
	m.cursor = 0
}

func (m AddProjectModel) handleGroupSelect(msg tea.KeyMsg) (AddProjectModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "esc":
		// Go back to category
		m.options = nil
		for _, cat := range m.cfg.Categories {
			m.options = append(m.options, cat.Name+" ("+cat.ID+")")
		}
		m.options = append(m.options, "+ Nova categoria")
		m.cursor = 0
		m.step = stepCategory
	case "enter":
		if m.cursor == 0 {
			// Sem grupo
			m.group = ""
			m.cursor = 0
			m.step = stepConfirm
		} else if m.cursor == len(m.options)-1 {
			// Novo grupo - use text input
			m.inputBuf = ""
			m.message = ""
			m.step = stepCustomGroup
		} else {
			// Grupo existente
			m.group = m.options[m.cursor]
			m.cursor = 0
			m.step = stepConfirm
		}
	}
	return m, nil
}

func (m AddProjectModel) handleConfirm(msg tea.KeyMsg) (AddProjectModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 1 {
			m.cursor++
		}
	case "esc":
		m.buildGroupOptions()
		m.step = stepGroup
	case "enter":
		if m.cursor == 0 {
			if err := m.saveProject(); err != nil {
				m.message = fmt.Sprintf("Erro ao salvar: %s", err)
				m.messageType = "error"
			} else {
				m.message = "Projeto salvo!"
				m.messageType = "ok"
			}
			m.step = stepDone
			return m, nil
		}
		m.goBack = true
	}
	return m, nil
}

func (m AddProjectModel) doClone() tea.Cmd {
	gitURL := m.gitURL
	clonePath := m.clonePath
	return func() tea.Msg {
		cmd := exec.Command("git", "clone", gitURL, clonePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return cloneDoneMsg{err: fmt.Errorf("%s: %s", err, string(output))}
		}
		return cloneDoneMsg{err: nil}
	}
}

func (m *AddProjectModel) saveProject() error {
	var relPath string
	home, _ := os.UserHomeDir()
	avitaBase := filepath.Join(home, "Avita")
	persBase := filepath.Join(home, "Persike")

	if strings.HasPrefix(m.clonePath, avitaBase) {
		relPath, _ = filepath.Rel(avitaBase, m.clonePath)
	} else if strings.HasPrefix(m.clonePath, persBase) {
		relPath = "Persike/" + filepath.Base(m.clonePath)
	} else {
		relPath = m.clonePath
	}

	project := config.Project{
		Alias: m.alias,
		Desc:  m.description,
		Path:  relPath,
		Group: m.group,
	}

	found := false
	for i, cat := range m.cfg.Categories {
		if cat.ID == m.categoryID {
			m.cfg.Categories[i].Projects = append(m.cfg.Categories[i].Projects, project)
			found = true
			break
		}
	}

	if !found && m.newCatName != "" {
		m.cfg.Categories = append(m.cfg.Categories, config.Category{
			ID:       m.categoryID,
			Name:     m.newCatName,
			Sub:      m.newCatSub,
			Projects: []config.Project{project},
		})
	}

	return config.Save(m.cfg, m.configPath)
}

func extractRepoName(url string) string {
	if strings.Contains(url, ":") && strings.HasPrefix(url, "git@") {
		parts := strings.SplitN(url, ":", 2)
		if len(parts) == 2 {
			name := filepath.Base(parts[1])
			return strings.TrimSuffix(name, ".git")
		}
	}
	name := filepath.Base(url)
	return strings.TrimSuffix(name, ".git")
}

func suggestAlias(url string) string {
	name := extractRepoName(url)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 20 {
		name = name[:20]
	}
	return name
}
