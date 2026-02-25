package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

type settingsScreen int

const (
	settingsMenu settingsScreen = iota
	settingsEditCategories
	settingsEditCatName
	settingsEditCatSub
	settingsEditPaths
	settingsEditPathInput
	settingsDeleteProject
	settingsDeleteConfirm
	settingsDone
)

type settingsProjectRef struct {
	catIdx  int
	projIdx int
	display string
}

type SettingsModel struct {
	cfg        *config.Config
	configPath string
	screen     settingsScreen
	goBack     bool

	cursor      int
	options     []string
	inputBuf    string
	message     string
	messageType string

	// Context
	selectedCatIdx  int
	selectedProjIdx int
	editingField    string
	allProjects     []settingsProjectRef

	offset        int // scroll offset for long lists
	width, height int
}

func NewSettingsModel(cfg *config.Config, configPath string) SettingsModel {
	return SettingsModel{
		cfg:        cfg,
		configPath: configPath,
		screen:     settingsMenu,
		options:    []string{"Editar Categorias", "Editar Locais (paths)", "Deletar Projeto"},
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return nil
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case settingsMenu:
			return m.handleMenu(msg)
		case settingsEditCategories:
			return m.handleEditCategories(msg)
		case settingsEditCatName:
			return m.handleTextInput(msg)
		case settingsEditCatSub:
			return m.handleTextInput(msg)
		case settingsEditPaths:
			return m.handleEditPaths(msg)
		case settingsEditPathInput:
			return m.handleTextInput(msg)
		case settingsDeleteProject:
			return m.handleDeleteProject(msg)
		case settingsDeleteConfirm:
			return m.handleDeleteConfirm(msg)
		case settingsDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.screen = settingsMenu
				m.cursor = 0
				m.message = ""
				m.options = []string{"Editar Categorias", "Editar Locais (paths)", "Deletar Projeto"}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m SettingsModel) View() string {
	var b strings.Builder

	header := theme.SettingsTitleSelectedStyle.Render("  ⚙ Configurações")
	divider := theme.DimStyle.Render("  " + strings.Repeat("─", 50))
	b.WriteString("\n" + header + "\n" + divider + "\n\n")

	switch m.screen {
	case settingsMenu:
		b.WriteString("  " + theme.TextStyle.Render("O que deseja configurar?") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))

	case settingsEditCategories:
		b.WriteString("  " + theme.TextStyle.Render("Selecione a categoria para editar:") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter editar  esc voltar"))

	case settingsEditCatName:
		cat := m.cfg.Categories[m.selectedCatIdx]
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+cat.Name) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Novo nome da categoria:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsEditCatSub:
		cat := m.cfg.Categories[m.selectedCatIdx]
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+cat.Name) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Nova descrição da categoria:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsEditPaths:
		b.WriteString("  " + theme.TextStyle.Render("Selecione o path para editar:") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter editar  esc voltar"))

	case settingsEditPathInput:
		label := "Base projetos"
		if m.editingField == "personal_base" {
			label = "Base pessoal"
		}
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+label) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Novo path:") + "\n\n")
		b.WriteString("  " + theme.TitleStyle.Render("> ") + m.inputBuf + theme.DimStyle.Render("█") + "\n")
		if m.message != "" {
			b.WriteString("\n  " + theme.ErrorStyle.Render(m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsDeleteProject:
		b.WriteString("  " + theme.TextStyle.Render("Selecione o projeto para deletar:") + "\n\n")
		if len(m.options) == 0 {
			b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto cadastrado") + "\n")
		} else {
			visible := m.visibleListRows()
			start := m.offset
			end := start + visible
			if end > len(m.options) {
				end = len(m.options)
			}
			if start > 0 {
				b.WriteString("  " + theme.DimStyle.Render("  ↑ mais projetos...") + "\n")
			}
			for i := start; i < end; i++ {
				opt := m.options[i]
				if i == m.cursor {
					b.WriteString("  " + theme.ErrorStyle.Render("▶ "+opt) + "\n")
				} else {
					b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
				}
			}
			if end < len(m.options) {
				b.WriteString("  " + theme.DimStyle.Render("  ↓ mais projetos...") + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter selecionar  esc voltar"))

	case settingsDeleteConfirm:
		ref := m.allProjects[m.selectedProjIdx]
		b.WriteString("  " + theme.WarningStyle.Render("Tem certeza que deseja deletar?") + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render(ref.display) + "\n\n")
		confirmOpts := []string{"Sim, deletar", "Não, cancelar"}
		for i, opt := range confirmOpts {
			if i == m.cursor {
				b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
			} else {
				b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
			}
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter confirmar  esc cancelar"))

	case settingsDone:
		if m.messageType == "ok" {
			b.WriteString("  " + theme.SuccessStyle.Render("✓ "+m.message) + "\n")
		} else {
			b.WriteString("  " + theme.ErrorStyle.Render("✗ "+m.message) + "\n")
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	}

	return b.String()
}

func (m *SettingsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *SettingsModel) GoBack() bool { return m.goBack }

func (m SettingsModel) visibleListRows() int {
	// header(3) + title+blank(2) + help(2) + app help(2) = ~9 lines of chrome
	available := m.height - 9
	if available < 3 {
		return 3
	}
	return available
}

func (m *SettingsModel) ensureVisible() {
	visible := m.visibleListRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

// --- Menu handler ---

func (m SettingsModel) handleMenu(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
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
		m.goBack = true
	case "enter":
		switch m.cursor {
		case 0: // Editar Categorias
			m.options = nil
			for _, cat := range m.cfg.Categories {
				m.options = append(m.options, cat.Name+" — "+cat.Sub)
			}
			m.cursor = 0
			m.screen = settingsEditCategories
		case 1: // Editar Locais
			m.options = []string{
				"Base projetos: " + m.cfg.Settings.ProjectsBase,
				"Base pessoal: " + m.cfg.Settings.PersonalBase,
			}
			m.cursor = 0
			m.screen = settingsEditPaths
		case 2: // Deletar Projeto
			m.allProjects = nil
			m.options = nil
			for ci, cat := range m.cfg.Categories {
				for pi, proj := range cat.Projects {
					display := proj.Alias + " — " + proj.Desc + " (" + cat.Name + ")"
					m.allProjects = append(m.allProjects, settingsProjectRef{
						catIdx:  ci,
						projIdx: pi,
						display: display,
					})
					m.options = append(m.options, display)
				}
			}
			m.cursor = 0
			m.offset = 0
			m.screen = settingsDeleteProject
		}
	}
	return m, nil
}

// --- Edit Categories handlers ---

func (m SettingsModel) handleEditCategories(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
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
		m.screen = settingsMenu
		m.cursor = 0
		m.options = []string{"Editar Categorias", "Editar Locais (paths)", "Deletar Projeto"}
	case "enter":
		if m.cursor < len(m.cfg.Categories) {
			m.selectedCatIdx = m.cursor
			m.inputBuf = m.cfg.Categories[m.cursor].Name
			m.message = ""
			m.screen = settingsEditCatName
		}
	}
	return m, nil
}

// --- Edit Paths handler ---

func (m SettingsModel) handleEditPaths(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
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
		m.screen = settingsMenu
		m.cursor = 0
		m.options = []string{"Editar Categorias", "Editar Locais (paths)", "Deletar Projeto"}
	case "enter":
		switch m.cursor {
		case 0:
			m.editingField = "projects_base"
			m.inputBuf = m.cfg.Settings.ProjectsBase
		case 1:
			m.editingField = "personal_base"
			m.inputBuf = m.cfg.Settings.PersonalBase
		}
		m.message = ""
		m.screen = settingsEditPathInput
	}
	return m, nil
}

// --- Delete Project handlers ---

func (m SettingsModel) handleDeleteProject(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
			m.ensureVisible()
		}
	case "esc":
		m.screen = settingsMenu
		m.cursor = 0
		m.options = []string{"Editar Categorias", "Editar Locais (paths)", "Deletar Projeto"}
	case "enter":
		if len(m.allProjects) > 0 && m.cursor < len(m.allProjects) {
			m.selectedProjIdx = m.cursor
			m.cursor = 0
			m.screen = settingsDeleteConfirm
		}
	}
	return m, nil
}

func (m SettingsModel) handleDeleteConfirm(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
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
		m.cursor = m.selectedProjIdx
		m.screen = settingsDeleteProject
	case "enter":
		if m.cursor == 0 {
			// Confirm delete
			ref := m.allProjects[m.selectedProjIdx]
			cat := &m.cfg.Categories[ref.catIdx]
			cat.Projects = append(cat.Projects[:ref.projIdx], cat.Projects[ref.projIdx+1:]...)

			if err := config.Save(m.cfg, m.configPath); err != nil {
				m.message = fmt.Sprintf("Erro ao salvar: %s", err)
				m.messageType = "error"
			} else {
				m.message = "Projeto removido com sucesso!"
				m.messageType = "ok"
			}
			m.screen = settingsDone
		} else {
			// Cancel
			m.cursor = m.selectedProjIdx
			m.screen = settingsDeleteProject
		}
	}
	return m, nil
}

// --- Text input handler ---

func (m SettingsModel) handleTextInput(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.message = ""
		switch m.screen {
		case settingsEditCatName:
			m.screen = settingsEditCategories
			m.cursor = m.selectedCatIdx
			m.options = nil
			for _, cat := range m.cfg.Categories {
				m.options = append(m.options, cat.Name+" — "+cat.Sub)
			}
		case settingsEditCatSub:
			m.screen = settingsEditCatName
			m.inputBuf = m.cfg.Categories[m.selectedCatIdx].Name
		case settingsEditPathInput:
			m.screen = settingsEditPaths
			m.options = []string{
				"Base projetos: " + m.cfg.Settings.ProjectsBase,
				"Base pessoal: " + m.cfg.Settings.PersonalBase,
			}
		}
		return m, nil
	case "enter":
		return m.submitSettingsInput()
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

func (m SettingsModel) submitSettingsInput() (SettingsModel, tea.Cmd) {
	val := strings.TrimSpace(m.inputBuf)

	switch m.screen {
	case settingsEditCatName:
		if val == "" {
			m.message = "Nome não pode ser vazio"
			return m, nil
		}
		m.cfg.Categories[m.selectedCatIdx].Name = val
		m.inputBuf = m.cfg.Categories[m.selectedCatIdx].Sub
		m.message = ""
		m.screen = settingsEditCatSub
		return m, nil

	case settingsEditCatSub:
		if val == "" {
			m.message = "Descrição não pode ser vazia"
			return m, nil
		}
		m.cfg.Categories[m.selectedCatIdx].Sub = val
		if err := config.Save(m.cfg, m.configPath); err != nil {
			m.message = fmt.Sprintf("Erro ao salvar: %s", err)
			m.messageType = "error"
		} else {
			m.message = "Categoria atualizada com sucesso!"
			m.messageType = "ok"
		}
		m.screen = settingsDone
		return m, nil

	case settingsEditPathInput:
		if val == "" {
			m.message = "Path não pode ser vazio"
			return m, nil
		}
		switch m.editingField {
		case "projects_base":
			m.cfg.Settings.ProjectsBase = val
		case "personal_base":
			m.cfg.Settings.PersonalBase = val
		}
		if err := config.Save(m.cfg, m.configPath); err != nil {
			m.message = fmt.Sprintf("Erro ao salvar: %s", err)
			m.messageType = "error"
		} else {
			m.message = "Path atualizado com sucesso!"
			m.messageType = "ok"
		}
		m.screen = settingsDone
		return m, nil
	}

	return m, nil
}
