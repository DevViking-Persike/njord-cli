// Package githubclient é um cliente REST minimalista pra API do GitHub.
// Usa só net/http + encoding/json — sem SDK externo, porque a superfície que
// a gente consome hoje é pequena.
package githubclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.github.com"
	perPage        = 100
	maxPages       = 1000 // cap razoável: 1000 repos
)

// Client guarda as credenciais e o http.Client reutilizado.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewClient valida o token (não-vazio) e retorna um Client pronto.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("github token vazio")
	}
	return &Client{
		token:   token,
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Repo é a projeção enxuta do /user/repos usada pela TUI.
type Repo struct {
	FullName    string // "owner/repo"
	Description string
	SSHURL      string // git@github.com:owner/repo.git
	HTMLURL     string // https://github.com/owner/repo
	Private     bool
	Fork        bool
}

type apiRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	SSHURL      string `json:"ssh_url"`
	HTMLURL     string `json:"html_url"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
}

// ListMyRepos devolve todos os repositórios que o token alcança (affiliation=
// owner,collaborator,organization_member). Pagina até maxPages ou até a API
// parar de mandar itens.
func (c *Client) ListMyRepos(ctx context.Context) ([]Repo, error) {
	var out []Repo
	for page := 1; page <= maxPages; page++ {
		batch, err := c.fetchReposPage(ctx, page)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}
		out = append(out, batch...)
		if len(batch) < perPage {
			break
		}
	}
	return out, nil
}

func (c *Client) fetchReposPage(ctx context.Context, page int) ([]Repo, error) {
	q := url.Values{}
	q.Set("per_page", strconv.Itoa(perPage))
	q.Set("page", strconv.Itoa(page))
	q.Set("sort", "updated")
	q.Set("affiliation", "owner,collaborator,organization_member")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/user/repos?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var raw []apiRepo
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	out := make([]Repo, 0, len(raw))
	for _, r := range raw {
		out = append(out, Repo{
			FullName:    r.FullName,
			Description: r.Description,
			SSHURL:      r.SSHURL,
			HTMLURL:     r.HTMLURL,
			Private:     r.Private,
			Fork:        r.Fork,
		})
	}
	return out, nil
}
