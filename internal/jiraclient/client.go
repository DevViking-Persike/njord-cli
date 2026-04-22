package jiraclient

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to the Jira Cloud REST API v3 using Basic auth (email:token).
type Client struct {
	baseURL string
	auth    string // pre-encoded "Basic <base64(email:token)>"
	http    *http.Client
}

type clientOptions struct {
	httpClient *http.Client
}

// Option customises the Client.
type Option func(*clientOptions)

// WithHTTPClient injects a custom *http.Client (useful for tests).
func WithHTTPClient(h *http.Client) Option {
	return func(o *clientOptions) { o.httpClient = h }
}

// NewClient creates a Jira API client. baseURL is the workspace URL
// (e.g. https://foo.atlassian.net). email and token come from the local config.
func NewClient(baseURL, email, token string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("jira: baseURL is required")
	}
	if email == "" || token == "" {
		return nil, fmt.Errorf("jira: email and token are required")
	}

	o := clientOptions{httpClient: &http.Client{Timeout: 15 * time.Second}}
	for _, apply := range opts {
		apply(&o)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		auth:    "Basic " + creds,
		http:    o.httpClient,
	}, nil
}

// CurrentUser fetches the authenticated user — useful as a connectivity check.
func (c *Client) CurrentUser() (User, error) {
	var raw struct {
		AccountID    string `json:"accountId"`
		DisplayName  string `json:"displayName"`
		EmailAddress string `json:"emailAddress"`
	}
	if err := c.getJSON("/rest/api/3/myself", nil, &raw); err != nil {
		return User{}, err
	}
	return User(raw), nil
}

// ListProjects returns projects the authenticated user is related to — i.e.
// projects they've recently accessed (viewed, commented, transitioned, etc.).
// Uses /project/recent, which Jira ranks per-user. For a full catalogue, use a
// different endpoint (not currently exposed).
func (c *Client) ListProjects() ([]Project, error) {
	q := url.Values{}
	q.Set("maxResults", "50")

	var raw []projectRecent
	if err := c.getJSON("/rest/api/3/project/recent", q, &raw); err != nil {
		return nil, err
	}
	projects := make([]Project, 0, len(raw))
	for _, p := range raw {
		projects = append(projects, Project{Key: p.Key, Name: p.Name, ID: p.ID})
	}
	return projects, nil
}

// SearchIssues runs a JQL query and returns the first page (up to 50 issues).
func (c *Client) SearchIssues(jql string) (SearchResult, error) {
	q := url.Values{}
	q.Set("jql", jql)
	q.Set("maxResults", "50")
	q.Set("fields", "summary,status,issuetype,assignee,parent,customfield_10014")

	var raw searchResponse
	if err := c.getJSON("/rest/api/3/search/jql", q, &raw); err != nil {
		return SearchResult{}, err
	}

	issues := make([]Issue, 0, len(raw.Issues))
	for _, r := range raw.Issues {
		issues = append(issues, parseIssue(r))
	}
	return SearchResult{
		Issues:     issues,
		Total:      raw.Total,
		IsLast:     raw.IsLast,
		NextCursor: raw.NextPageToken,
	}, nil
}

// --- internals ---

type projectRecent struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type searchResponse struct {
	Issues        []rawIssue `json:"issues"`
	Total         int        `json:"total"`
	IsLast        bool       `json:"isLast"`
	NextPageToken string     `json:"nextPageToken"`
}

type rawIssue struct {
	Key    string       `json:"key"`
	Fields rawIssueFlds `json:"fields"`
}

type rawIssueFlds struct {
	Summary   string          `json:"summary"`
	Status    rawNamedField   `json:"status"`
	IssueType rawNamedField   `json:"issuetype"`
	Assignee  *rawAssigneeFld `json:"assignee"`
	Parent    *rawParent      `json:"parent"`
	EpicLink  string          `json:"customfield_10014"`
}

type rawNamedField struct {
	Name string `json:"name"`
}

type rawAssigneeFld struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type rawParent struct {
	Key    string         `json:"key"`
	Fields rawParentFlds  `json:"fields"`
}

type rawParentFlds struct {
	IssueType rawNamedField `json:"issuetype"`
}

func parseIssue(r rawIssue) Issue {
	issue := Issue{
		Key:     r.Key,
		Summary: r.Fields.Summary,
		Status:  r.Fields.Status.Name,
		Type:    r.Fields.IssueType.Name,
	}
	if r.Fields.Assignee != nil {
		issue.Assignee = r.Fields.Assignee.DisplayName
	}
	issue.EpicKey = resolveEpicKey(r.Fields)
	return issue
}

func resolveEpicKey(f rawIssueFlds) string {
	if f.EpicLink != "" {
		return f.EpicLink
	}
	if f.Parent != nil && strings.EqualFold(f.Parent.Fields.IssueType.Name, "Epic") {
		return f.Parent.Key
	}
	return ""
}

func (c *Client) getJSON(path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("jira: build request: %w", err)
	}
	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("jira: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("jira: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jira: %s %s: HTTP %d: %s", http.MethodGet, path, resp.StatusCode, truncate(string(body), 200))
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("jira: decode response: %w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
