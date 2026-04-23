// Package clone concentra as regras puras do fluxo "listar repos disponíveis
// nas APIs (GitLab/GitHub) pra clonar e cadastrar". Não depende de UI nem
// de client SDK — adapters plugam essas camadas aqui.
package clone

import (
	"sort"
	"strings"

	"github.com/DevViking-Persike/njord-cli/internal/githubclient"
	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
)

// Host identifica de onde um Repo foi listado.
type Host string

const (
	HostGitLab Host = "gitlab"
	HostGitHub Host = "github"
)

// Repo é a projeção comum de um repositório remoto, usada pela TUI de clone.
// Campos normalizados independente da origem (GitLab ou GitHub).
type Repo struct {
	FullName    string // path/namespace (GitLab) ou owner/repo (GitHub)
	Description string
	CloneSSH    string // URL pra git clone
	HTMLURL     string // URL pra abrir no browser
	Host        Host
}

// Group representa um grupo/subgrupo GitLab na camada de negócio. Serve pra
// o scope picker oferecer navegação por camadas antes de listar os repos.
type Group struct {
	ID       int64
	Name     string // "bill"
	FullPath string // "avitaseg/bill" — identificador estável
	FullName string // "avitaseg / bill" — pra display
}

// GroupFromGitLab adapta um GroupInfo do gitlabclient pro Group.
func GroupFromGitLab(g gitlabclient.GroupInfo) Group {
	return Group{
		ID:       g.ID,
		Name:     g.Name,
		FullPath: g.FullPath,
		FullName: g.FullName,
	}
}

// CollapseToTopBuckets reduz a lista de grupos pra um nível só, agrupando
// tudo que compartilha o mesmo "primeiro segmento depois do prefixo comum".
// Ex.: [avitaseg, avitaseg/bill, avitaseg/bill/bibliotecas, avitaseg/gap]
// vira [avitaseg/bill, avitaseg/gap] — os subgrupos são alcançados via
// IncludeSubGroups=true na query de projetos.
//
// Se o colapso resultasse em lista vazia (ex.: só tem o prefixo comum), cai
// pro fallback de devolver a lista original.
func CollapseToTopBuckets(groups []Group) []Group {
	if len(groups) == 0 {
		return groups
	}
	prefix := commonPathPrefix(groups)
	seen := map[string]bool{}
	byPath := map[string]Group{}
	for _, g := range groups {
		byPath[g.FullPath] = g
	}

	var out []Group
	for _, g := range groups {
		remaining := strings.TrimPrefix(g.FullPath, prefix)
		remaining = strings.TrimPrefix(remaining, "/")
		if remaining == "" {
			continue // é o próprio prefixo comum — representado pelos filhos
		}
		firstSeg := strings.SplitN(remaining, "/", 2)[0]
		bucketKey := firstSeg
		if prefix != "" {
			bucketKey = prefix + "/" + firstSeg
		}
		if seen[bucketKey] {
			continue
		}
		seen[bucketKey] = true
		if g2, ok := byPath[bucketKey]; ok {
			out = append(out, g2)
		}
	}
	if len(out) == 0 {
		return groups
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].FullPath) < strings.ToLower(out[j].FullPath)
	})
	return out
}

// commonPathPrefix devolve o maior prefixo compartilhado por todos os grupos,
// quebrado por segmento ("/"). Ex.: [avitaseg/bill, avitaseg/gap] → "avitaseg".
func commonPathPrefix(groups []Group) string {
	if len(groups) == 0 {
		return ""
	}
	segs := make([][]string, len(groups))
	minLen := -1
	for i, g := range groups {
		parts := strings.Split(g.FullPath, "/")
		segs[i] = parts
		if minLen < 0 || len(parts) < minLen {
			minLen = len(parts)
		}
	}
	var prefix []string
	for i := 0; i < minLen; i++ {
		seg := segs[0][i]
		same := true
		for _, s := range segs {
			if s[i] != seg {
				same = false
				break
			}
		}
		if !same {
			break
		}
		prefix = append(prefix, seg)
	}
	return strings.Join(prefix, "/")
}

// FilterGroups devolve os grupos cujo FullPath ou FullName contém a query
// (mesma regra do FilterRepos: tokens AND, case-insensitive).
func FilterGroups(groups []Group, query string) []Group {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return groups
	}
	tokens := strings.Fields(q)
	out := make([]Group, 0, len(groups))
	for _, g := range groups {
		haystack := strings.ToLower(g.FullPath + " " + g.FullName)
		match := true
		for _, tk := range tokens {
			if !strings.Contains(haystack, tk) {
				match = false
				break
			}
		}
		if match {
			out = append(out, g)
		}
	}
	return out
}

// FromGitLab adapta um ProjectInfo do gitlabclient pro Repo unificado.
func FromGitLab(p gitlabclient.ProjectInfo) Repo {
	return Repo{
		FullName:    p.PathWithNamespace,
		Description: p.Description,
		CloneSSH:    p.SSHURLToRepo,
		HTMLURL:     p.WebURL,
		Host:        HostGitLab,
	}
}

// FromGitHub adapta um Repo do githubclient pro Repo unificado.
func FromGitHub(r githubclient.Repo) Repo {
	return Repo{
		FullName:    r.FullName,
		Description: r.Description,
		CloneSSH:    r.SSHURL,
		HTMLURL:     r.HTMLURL,
		Host:        HostGitHub,
	}
}

// FilterRepos devolve os repos cujo FullName ou Description contém query
// (case-insensitive, tokens separados por espaço — todos precisam bater).
// Query vazia retorna a lista original.
func FilterRepos(repos []Repo, query string) []Repo {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return repos
	}
	tokens := strings.Fields(q)
	out := make([]Repo, 0, len(repos))
	for _, r := range repos {
		haystack := strings.ToLower(r.FullName + " " + r.Description)
		match := true
		for _, tk := range tokens {
			if !strings.Contains(haystack, tk) {
				match = false
				break
			}
		}
		if match {
			out = append(out, r)
		}
	}
	return out
}

// SortByName ordena em ordem alfabética estável.
func SortByName(repos []Repo) {
	sort.SliceStable(repos, func(i, j int) bool {
		return strings.ToLower(repos[i].FullName) < strings.ToLower(repos[j].FullName)
	})
}
