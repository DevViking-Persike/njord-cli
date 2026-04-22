package ui

import (
	"fmt"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/app"
	"github.com/DevViking-Persike/njord-cli/internal/config"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/DevViking-Persike/njord-cli/internal/ui/components"
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
	settingsEditGroupProject
	settingsEditGroupSelect
	settingsEditGroupCustom
	settingsEditGitLabToken
	settingsDone
)

// Chrome lines used for scroll calculations in settings.
const settingsChromeLines = 9

var mainMenuOptions = []string{"Editar Categorias", "Editar Locais (paths)", "Editar Grupos", "Deletar Projeto", "GitLab Token"}

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

	scroll components.ScrollState
	width  int
}

func NewSettingsModel(cfg *config.Config, configPath string) SettingsModel {
	return SettingsModel{
		cfg:        cfg,
		configPath: configPath,
		screen:     settingsMenu,
		options:    append([]string{}, mainMenuOptions...),
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
		case settingsEditCatName, settingsEditCatSub, settingsEditPathInput, settingsEditGroupCustom, settingsEditGitLabToken:
			return m.handleTextInput(msg)
		case settingsEditPaths:
			return m.handleEditPaths(msg)
		case settingsDeleteProject:
			return m.handleDeleteProject(msg)
		case settingsDeleteConfirm:
			return m.handleDeleteConfirm(msg)
		case settingsEditGroupProject:
			return m.handleEditGroupProject(msg)
		case settingsEditGroupSelect:
			return m.handleEditGroupSelect(msg)
		case settingsDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				m.screen = settingsMenu
				m.cursor = 0
				m.message = ""
				m.options = []string{"Editar Categorias", "Editar Locais (paths)", "Editar Grupos", "Deletar Projeto"}
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

	selectedFn := func(s string) string { return theme.TitleSelectedStyle.Render(s) }
	normalFn := func(s string) string { return theme.TextStyle.Render(s) }

	switch m.screen {
	case settingsMenu:
		b.WriteString("  " + theme.TextStyle.Render("O que deseja configurar?") + "\n\n")
		b.WriteString(components.RenderMenuOptions(m.options, m.cursor, selectedFn, normalFn))
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))

	case settingsEditCategories:
		b.WriteString("  " + theme.TextStyle.Render("Selecione a categoria para editar:") + "\n\n")
		b.WriteString(components.RenderMenuOptions(m.options, m.cursor, selectedFn, normalFn))
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter editar  esc voltar"))

	case settingsEditCatName:
		cat := m.cfg.Categories[m.selectedCatIdx]
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+cat.Name) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Novo nome da categoria:") + "\n\n")
		b.WriteString(components.RenderTextInput(m.inputBuf, ""))
		b.WriteString(components.RenderError(m.message))
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsEditCatSub:
		cat := m.cfg.Categories[m.selectedCatIdx]
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+cat.Name) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Nova descrição da categoria:") + "\n\n")
		b.WriteString(components.RenderTextInput(m.inputBuf, ""))
		b.WriteString(components.RenderError(m.message))
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsEditPaths:
		b.WriteString("  " + theme.TextStyle.Render("Selecione o path para editar:") + "\n\n")
		b.WriteString(components.RenderMenuOptions(m.options, m.cursor, selectedFn, normalFn))
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter editar  esc voltar"))

	case settingsEditPathInput:
		label := "Base projetos"
		if m.editingField == "personal_base" {
			label = "Base pessoal"
		}
		b.WriteString("  " + theme.DimStyle.Render("Editando: "+label) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Novo path:") + "\n\n")
		b.WriteString(components.RenderTextInput(m.inputBuf, ""))
		b.WriteString(components.RenderError(m.message))
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsDeleteProject:
		b.WriteString("  " + theme.TextStyle.Render("Selecione o projeto para deletar:") + "\n\n")
		if len(m.options) == 0 {
			b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto cadastrado") + "\n")
		} else {
			start, end := m.scroll.Bounds(len(m.options), settingsChromeLines)
			components.RenderScrollUp(&b, start)
			for i := start; i < end; i++ {
				opt := m.options[i]
				if i == m.cursor {
					b.WriteString("  " + theme.ErrorStyle.Render("▶ "+opt) + "\n")
				} else {
					b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
				}
			}
			components.RenderScrollDown(&b, end, len(m.options))
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter selecionar  esc voltar"))

	case settingsDeleteConfirm:
		ref := m.allProjects[m.selectedProjIdx]
		b.WriteString("  " + theme.WarningStyle.Render("Tem certeza que deseja deletar?") + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render(ref.display) + "\n\n")
		confirmOpts := []string{"Sim, deletar", "Não, cancelar"}
		b.WriteString(components.RenderMenuOptions(confirmOpts, m.cursor, selectedFn, normalFn))
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter confirmar  esc cancelar"))

	case settingsEditGroupProject:
		b.WriteString("  " + theme.TextStyle.Render("Selecione o projeto para editar grupo:") + "\n\n")
		if len(m.options) == 0 {
			b.WriteString("  " + theme.DimStyle.Render("Nenhum projeto cadastrado") + "\n")
		} else {
			start, end := m.scroll.Bounds(len(m.options), settingsChromeLines)
			components.RenderScrollUp(&b, start)
			for i := start; i < end; i++ {
				opt := m.options[i]
				if i == m.cursor {
					b.WriteString("  " + theme.TitleSelectedStyle.Render("▶ "+opt) + "\n")
				} else {
					b.WriteString("  " + theme.TextStyle.Render("  "+opt) + "\n")
				}
			}
			components.RenderScrollDown(&b, end, len(m.options))
		}
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter selecionar  esc voltar"))

	case settingsEditGroupSelect:
		ref := m.allProjects[m.selectedProjIdx]
		b.WriteString("  " + theme.DimStyle.Render("Projeto: "+ref.display) + "\n\n")
		b.WriteString("  " + theme.TextStyle.Render("Selecione o grupo:") + "\n\n")
		b.WriteString(components.RenderMenuOptions(m.options, m.cursor, selectedFn, normalFn))
		b.WriteString("\n" + theme.HelpStyle.Render("  ↑↓ navigate  enter select  esc voltar"))

	case settingsEditGroupCustom:
		b.WriteString("  " + theme.TextStyle.Render("Nome do novo grupo:") + "\n\n")
		b.WriteString(components.RenderTextInput(m.inputBuf, ""))
		b.WriteString(components.RenderError(m.message))
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsEditGitLabToken:
		b.WriteString("  " + theme.TextStyle.Render("GitLab Personal Access Token:") + "\n\n")
		masked := m.inputBuf
		if len(masked) > 8 {
			masked = masked[:4] + strings.Repeat("*", len(masked)-8) + masked[len(masked)-4:]
		}
		b.WriteString(components.RenderTextInput(masked, ""))
		b.WriteString(components.RenderError(m.message))
		b.WriteString("\n  " + theme.DimStyle.Render("Scope necessário: api") + "\n")
		b.WriteString("  " + theme.DimStyle.Render("Crie em: gitlab.com → Settings → Access Tokens") + "\n")
		b.WriteString("\n" + theme.HelpStyle.Render("  enter confirmar  esc cancelar"))

	case settingsDone:
		b.WriteString(components.RenderMessage(m.message, m.messageType))
		b.WriteString("\n" + theme.HelpStyle.Render("  enter voltar"))
	}

	return b.String()
}

func (m *SettingsModel) SetSize(w, h int) {
	m.width = w
	m.scroll.Height = h
}

func (m *SettingsModel) GoBack() bool { return m.goBack }

// --- Menu handler ---

func (m SettingsModel) handleMenu(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
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
		case 2: // Editar Grupos
			m.allProjects = nil
			m.options = nil
			for ci, cat := range m.cfg.Categories {
				for pi, proj := range cat.Projects {
					groupTag := ""
					if proj.Group != "" {
						groupTag = " [" + proj.Group + "]"
					}
					display := proj.Alias + groupTag + " — " + proj.Desc + " (" + cat.Name + ")"
					m.allProjects = append(m.allProjects, settingsProjectRef{
						catIdx:  ci,
						projIdx: pi,
						display: display,
					})
					m.options = append(m.options, display)
				}
			}
			m.cursor = 0
			m.scroll.Offset = 0
			m.screen = settingsEditGroupProject
		case 3: // Deletar Projeto
			m.allProjects = nil
			m.options = nil
			for ci, cat := range m.cfg.Categories {
				for pi, proj := range cat.Projects {
					display := proj.Alias + " — " + proj.Desc + " (" + cat.Name + ")"
					if proj.Group != "" {
						display = proj.Alias + " [" + proj.Group + "] — " + proj.Desc + " (" + cat.Name + ")"
					}
					m.allProjects = append(m.allProjects, settingsProjectRef{
						catIdx:  ci,
						projIdx: pi,
						display: display,
					})
					m.options = append(m.options, display)
				}
			}
			m.cursor = 0
			m.scroll.Offset = 0
			m.screen = settingsDeleteProject
		case 4: // GitLab Token
			m.inputBuf = m.cfg.GitLab.Token
			m.message = ""
			m.screen = settingsEditGitLabToken
		}
	}
	return m, nil
}

// --- Edit Categories handlers ---

func (m SettingsModel) handleEditCategories(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = settingsMenu
		m.cursor = 0
		m.options = append([]string{}, mainMenuOptions...)
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
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.screen = settingsMenu
		m.cursor = 0
		m.options = append([]string{}, mainMenuOptions...)
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
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, settingsChromeLines)
		}
	case "esc":
		m.screen = settingsMenu
		m.cursor = 0
		m.options = append([]string{}, mainMenuOptions...)
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
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, 2); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.cursor = m.selectedProjIdx
		m.screen = settingsDeleteProject
	case "enter":
		if m.cursor == 0 {
			ref := m.allProjects[m.selectedProjIdx]
			if err := app.RemoveProject(m.cfg, m.configPath, ref.catIdx, ref.projIdx); err != nil {
				m.message = fmt.Sprintf("Erro ao remover projeto: %s", err)
				m.messageType = "error"
				m.screen = settingsDone
			} else {
				m.message = "Projeto removido com sucesso!"
				m.messageType = "ok"
				m.screen = settingsDone
			}
		} else {
			m.cursor = m.selectedProjIdx
			m.screen = settingsDeleteProject
		}
	}
	return m, nil
}

// --- Edit Group handlers ---

func (m SettingsModel) handleEditGroupProject(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
			m.scroll.EnsureVisible(m.cursor, settingsChromeLines)
		}
	case "esc":
		m.screen = settingsMenu
		m.cursor = 0
		m.options = append([]string{}, mainMenuOptions...)
	case "enter":
		if len(m.allProjects) > 0 && m.cursor < len(m.allProjects) {
			m.selectedProjIdx = m.cursor
			m.buildGroupSelectOptions()
			m.screen = settingsEditGroupSelect
		}
	}
	return m, nil
}

func (m *SettingsModel) buildGroupSelectOptions() {
	ref := m.allProjects[m.selectedProjIdx]
	cat := m.cfg.Categories[ref.catIdx]

	seen := make(map[string]bool)
	var existing []string
	for _, p := range cat.Projects {
		if p.Group != "" && !seen[p.Group] {
			seen[p.Group] = true
			existing = append(existing, p.Group)
		}
	}

	m.options = nil
	m.options = append(m.options, "Sem grupo")
	m.options = append(m.options, existing...)
	m.options = append(m.options, "+ Novo grupo")
	m.cursor = 0
}

func (m SettingsModel) handleEditGroupSelect(msg tea.KeyMsg) (SettingsModel, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "down", "j":
		if newCursor, moved := components.ListNav(msg, m.cursor, len(m.options)); moved {
			m.cursor = newCursor
		}
	case "esc":
		m.options = nil
		for _, ref := range m.allProjects {
			m.options = append(m.options, ref.display)
		}
		m.cursor = m.selectedProjIdx
		m.scroll.Offset = 0
		m.scroll.EnsureVisible(m.cursor, settingsChromeLines)
		m.screen = settingsEditGroupProject
	case "enter":
		ref := m.allProjects[m.selectedProjIdx]
		if m.cursor == 0 {
			// Sem grupo
			if err := app.UpdateProjectGroup(m.cfg, m.configPath, ref.catIdx, ref.projIdx, ""); err != nil {
				m.message = fmt.Sprintf("Erro ao atualizar grupo: %s", err)
				m.messageType = "error"
				m.screen = settingsDone
			} else {
				m.message = "Grupo removido com sucesso!"
				m.messageType = "ok"
				m.screen = settingsDone
			}
		} else if m.cursor == len(m.options)-1 {
			// Novo grupo
			m.inputBuf = ""
			m.message = ""
			m.screen = settingsEditGroupCustom
		} else {
			// Grupo existente
			group := m.options[m.cursor]
			if err := app.UpdateProjectGroup(m.cfg, m.configPath, ref.catIdx, ref.projIdx, group); err != nil {
				m.message = fmt.Sprintf("Erro ao atualizar grupo: %s", err)
				m.messageType = "error"
				m.screen = settingsDone
			} else {
				m.message = fmt.Sprintf("Grupo '%s' definido com sucesso!", group)
				m.messageType = "ok"
				m.screen = settingsDone
			}
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
		case settingsEditGroupCustom:
			m.buildGroupSelectOptions()
			m.screen = settingsEditGroupSelect
		case settingsEditGitLabToken:
			m.screen = settingsMenu
			m.cursor = 0
			m.options = append([]string{}, mainMenuOptions...)
		}
		return m, nil
	case "enter":
		return m.submitSettingsInput()
	default:
		if newBuf, handled := components.TextInput(msg, m.inputBuf, nil); handled {
			m.inputBuf = newBuf
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
		if err := app.UpdateCategory(m.cfg, m.configPath, m.selectedCatIdx, m.cfg.Categories[m.selectedCatIdx].Name, val); err != nil {
			m.message = fmt.Sprintf("Erro ao atualizar categoria: %s", err)
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
		var err error
		switch m.editingField {
		case "projects_base":
			err = app.UpdateProjectsBase(m.cfg, m.configPath, val)
		case "personal_base":
			err = app.UpdatePersonalBase(m.cfg, m.configPath, val)
		}
		if err != nil {
			m.message = fmt.Sprintf("Erro ao atualizar path: %s", err)
			m.messageType = "error"
		} else {
			m.message = "Path atualizado com sucesso!"
			m.messageType = "ok"
		}
		m.screen = settingsDone
		return m, nil

	case settingsEditGroupCustom:
		if val == "" {
			m.message = "Nome do grupo não pode ser vazio"
			return m, nil
		}
		ref := m.allProjects[m.selectedProjIdx]
		if err := app.UpdateProjectGroup(m.cfg, m.configPath, ref.catIdx, ref.projIdx, val); err != nil {
			m.message = fmt.Sprintf("Erro ao atualizar grupo: %s", err)
			m.messageType = "error"
		} else {
			m.message = fmt.Sprintf("Grupo '%s' definido com sucesso!", val)
			m.messageType = "ok"
		}
		m.screen = settingsDone
		return m, nil

	case settingsEditGitLabToken:
		if err := app.UpdateGitLabToken(m.cfg, m.configPath, val); err != nil {
			m.message = fmt.Sprintf("Erro ao salvar token GitLab: %s", err)
			m.messageType = "error"
		} else {
			m.message = "Token GitLab salvo com sucesso!"
			if val == "" {
				m.message = "Token GitLab removido!"
			}
			m.messageType = "ok"
		}
		m.screen = settingsDone
		return m, nil
	}

	return m, nil
}
