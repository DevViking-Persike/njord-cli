// Package clone é a TUI que lista os repos disponíveis nas APIs do GitLab e do
// GitHub pro usuário escolher um pra clonar + adicionar à config. A execução
// do clone em si acontece no fluxo de add_project, pra onde a gente entrega
// o Repo selecionado.
package clone

import (
	"context"
	"fmt"
	"time"

	cloneapp "github.com/DevViking-Persike/njord-cli/internal/app/clone"
	"github.com/DevViking-Persike/njord-cli/internal/githubclient"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	tea "github.com/charmbracelet/bubbletea"
)

// Scope descreve o conjunto de repos a buscar: um grupo GitLab específico
// (Host=GitLab + Group set) ou todos do GitHub (Host=GitHub + Group nil).
type Scope struct {
	Host  cloneapp.Host
	Group *cloneapp.Group // nil quando Host=GitHub ou quando GitLab "todos"
}

type loadedMsg struct {
	repos []cloneapp.Repo
	err   error
}

// Model é a tela de listagem de repos pra um Scope fixo. O toggle de fonte
// vive na GroupsModel — aqui a fonte já foi decidida.
type Model struct {
	glClient *gitlabclient.Client
	ghClient *githubclient.Client

	scope Scope

	repos    []cloneapp.Repo
	loading  bool
	loadErr  string
	search   string
	filtered []cloneapp.Repo
	cursor   int
	offset   int
	width    int
	height   int
	selected *cloneapp.Repo
	goBack   bool
}

// NewModel recebe os clients e o escopo. Clients podem ser nil — a tela
// mostra erro se a fonte correspondente precisar deles.
func NewModel(gl *gitlabclient.Client, gh *githubclient.Client, scope Scope) Model {
	return Model{
		glClient: gl,
		ghClient: gh,
		scope:    scope,
		loading:  true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.fetchCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			return m, nil
		}
		m.loadErr = ""
		cloneapp.SortByName(msg.repos)
		m.repos = msg.repos
		m.recomputeFiltered()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if m.search != "" {
			m.search = ""
			m.recomputeFiltered()
			m.offset = 0
			return m, nil
		}
		m.goBack = true
		return m, nil
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
			m.ensureVisible()
		}
		return m, nil
	case tea.KeyDown:
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.ensureVisible()
		}
		return m, nil
	case tea.KeyEnter:
		if m.cursor < len(m.filtered) {
			r := m.filtered[m.cursor]
			m.selected = &r
		}
		return m, nil
	case tea.KeyBackspace:
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.recomputeFiltered()
			m.offset = 0
			m.cursor = 0
		}
		return m, nil
	case tea.KeyRunes, tea.KeySpace:
		m.search += string(msg.Runes)
		m.recomputeFiltered()
		m.offset = 0
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

func (m Model) fetchCmd() tea.Cmd {
	gl := m.glClient
	gh := m.ghClient
	scope := m.scope
	return func() tea.Msg {
		switch scope.Host {
		case cloneapp.HostGitLab:
			if gl == nil {
				return loadedMsg{err: fmt.Errorf("GitLab token não configurado")}
			}
			var items []gitlabclient.ProjectInfo
			var err error
			if scope.Group != nil {
				items, err = gl.ListGroupProjects(scope.Group.ID, true)
			} else {
				items, err = gl.ListMyAccessibleProjects()
			}
			if err != nil {
				return loadedMsg{err: err}
			}
			repos := make([]cloneapp.Repo, 0, len(items))
			for _, it := range items {
				repos = append(repos, cloneapp.FromGitLab(it))
			}
			return loadedMsg{repos: repos}
		case cloneapp.HostGitHub:
			if gh == nil {
				return loadedMsg{err: fmt.Errorf("GitHub token não configurado")}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			items, err := gh.ListMyRepos(ctx)
			if err != nil {
				return loadedMsg{err: err}
			}
			repos := make([]cloneapp.Repo, 0, len(items))
			for _, it := range items {
				repos = append(repos, cloneapp.FromGitHub(it))
			}
			return loadedMsg{repos: repos}
		}
		return loadedMsg{err: fmt.Errorf("escopo inválido")}
	}
}

func (m *Model) recomputeFiltered() {
	m.filtered = cloneapp.FilterRepos(m.repos, m.search)
	if m.cursor >= len(m.filtered) {
		m.cursor = 0
		m.offset = 0
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.ensureVisible()
}

func (m *Model) GoBack() bool             { return m.goBack }
func (m *Model) Selected() *cloneapp.Repo { return m.selected }
func (m *Model) ClearSelection()          { m.selected = nil }

func (m Model) visibleRows() int {
	v := m.height - 11 // chrome: njord title + header + scope + busca + divider + help
	if v < 3 {
		return 3
	}
	return v
}

func (m *Model) ensureVisible() {
	visible := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}
