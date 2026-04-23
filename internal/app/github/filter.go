package github

import "github.com/DevViking-Persike/njord-cli/internal/config"

// PersonalCategoryID é o id da categoria que marca projetos pessoais (GitHub).
// Heurística: todo projeto da categoria "pessoal" é considerado GitHub, mesmo
// que ainda não tenha github_path preenchido (o usuário preenche gradualmente).
const PersonalCategoryID = "pessoal"

// ProjectRef guarda o projeto junto dos índices originais na config, pra permitir
// que telas subsequentes editem a YAML (ex.: preencher github_path que falta).
type ProjectRef struct {
	CatIdx  int
	ProjIdx int
	CatID   string
	CatName string
	Project config.Project
}

// AllProjectRefs devolve todos os projetos achatados preservando os índices.
func AllProjectRefs(cfg *config.Config) []ProjectRef {
	if cfg == nil {
		return nil
	}
	var out []ProjectRef
	for ci, cat := range cfg.Categories {
		for pi, p := range cat.Projects {
			out = append(out, ProjectRef{
				CatIdx:  ci,
				ProjIdx: pi,
				CatID:   cat.ID,
				CatName: cat.Name,
				Project: p,
			})
		}
	}
	return out
}

// FilterGitLab devolve os projetos que têm gitlab_path configurado.
func FilterGitLab(cfg *config.Config) []ProjectRef {
	var out []ProjectRef
	for _, ref := range AllProjectRefs(cfg) {
		if ref.Project.GitLabPath != "" {
			out = append(out, ref)
		}
	}
	return out
}

// FilterGitHub devolve os projetos considerados GitHub: ou têm github_path,
// ou estão na categoria pessoal (heurística do usuário).
func FilterGitHub(cfg *config.Config) []ProjectRef {
	var out []ProjectRef
	for _, ref := range AllProjectRefs(cfg) {
		if ref.Project.GitHubPath != "" || ref.CatID == PersonalCategoryID {
			out = append(out, ref)
		}
	}
	return out
}

// FilterLocal devolve os projetos cuja pasta existe no disco, independente
// do host remoto. Útil pra "abrir o que já tá no PC".
func FilterLocal(cfg *config.Config) []ProjectRef {
	var out []ProjectRef
	for _, ref := range AllProjectRefs(cfg) {
		if LocalExists(cfg, ref.Project) {
			out = append(out, ref)
		}
	}
	return out
}
