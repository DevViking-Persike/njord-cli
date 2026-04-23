package jiraclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CreateIssue manda POST /rest/api/3/issue e devolve a Issue recém-criada
// (só o key + summary conhecidos — status/type é o default do projeto).
//
// Para Subtask, passe ParentKey.
func (c *Client) CreateIssue(in CreateIssueInput) (Issue, error) {
	if err := validateCreate(in); err != nil {
		return Issue{}, err
	}

	payload := map[string]any{
		"fields": buildCreateFields(in),
	}

	var resp struct {
		Key string `json:"key"`
	}
	if err := c.sendJSON(http.MethodPost, "/rest/api/3/issue", payload, &resp); err != nil {
		return Issue{}, err
	}
	return Issue{
		Key:     resp.Key,
		Summary: in.Summary,
		Type:    in.Type,
	}, nil
}

// UpdateIssue manda PUT /rest/api/3/issue/{key}. Campos vazios ficam de fora.
func (c *Client) UpdateIssue(key string, in UpdateIssueInput) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("jira: issue key é obrigatória")
	}
	fields := map[string]any{}
	if strings.TrimSpace(in.Summary) != "" {
		fields["summary"] = in.Summary
	}
	if strings.TrimSpace(in.Description) != "" {
		fields["description"] = adfDescription(in.Description)
	}
	if len(fields) == 0 {
		return fmt.Errorf("jira: nada pra atualizar")
	}
	return c.sendJSON(http.MethodPut, "/rest/api/3/issue/"+key, map[string]any{"fields": fields}, nil)
}

// ListTransitions devolve as transições disponíveis pra issue — depende do
// workflow do projeto e do status atual dela.
func (c *Client) ListTransitions(key string) ([]Transition, error) {
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("jira: issue key é obrigatória")
	}
	var raw struct {
		Transitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			To   struct {
				Name           string `json:"name"`
				StatusCategory struct {
					Key string `json:"key"`
				} `json:"statusCategory"`
			} `json:"to"`
		} `json:"transitions"`
	}
	if err := c.getJSON("/rest/api/3/issue/"+key+"/transitions", nil, &raw); err != nil {
		return nil, err
	}
	out := make([]Transition, 0, len(raw.Transitions))
	for _, t := range raw.Transitions {
		out = append(out, Transition{
			ID:        t.ID,
			Name:      t.Name,
			ToStatus:  t.To.Name,
			StatusCat: t.To.StatusCategory.Key,
		})
	}
	return out, nil
}

// TransitionIssue aplica a transição identificada pelo id.
func (c *Client) TransitionIssue(key, transitionID string) error {
	if strings.TrimSpace(key) == "" || strings.TrimSpace(transitionID) == "" {
		return fmt.Errorf("jira: key e transitionID são obrigatórios")
	}
	payload := map[string]any{
		"transition": map[string]string{"id": transitionID},
	}
	return c.sendJSON(http.MethodPost, "/rest/api/3/issue/"+key+"/transitions", payload, nil)
}

// --- helpers ---

func validateCreate(in CreateIssueInput) error {
	if strings.TrimSpace(in.ProjectKey) == "" {
		return fmt.Errorf("jira: project key é obrigatório")
	}
	if strings.TrimSpace(in.Summary) == "" {
		return fmt.Errorf("jira: summary é obrigatório")
	}
	if strings.TrimSpace(in.Type) == "" {
		return fmt.Errorf("jira: issue type é obrigatório")
	}
	if isSubtaskType(in.Type) && strings.TrimSpace(in.ParentKey) == "" {
		return fmt.Errorf("jira: subtask exige parent key")
	}
	return nil
}

// isSubtaskType aceita "Subtask" e "Sub-task" (Jira usa os dois nomes no
// mesmo workspace dependendo da versão do template do projeto).
func isSubtaskType(t string) bool {
	low := strings.ToLower(strings.TrimSpace(t))
	return low == "subtask" || low == "sub-task"
}

func buildCreateFields(in CreateIssueInput) map[string]any {
	fields := map[string]any{
		"project":   map[string]string{"key": in.ProjectKey},
		"summary":   in.Summary,
		"issuetype": map[string]string{"name": in.Type},
	}
	if strings.TrimSpace(in.Description) != "" {
		fields["description"] = adfDescription(in.Description)
	}
	if strings.TrimSpace(in.AssigneeAccount) != "" {
		fields["assignee"] = map[string]string{"accountId": in.AssigneeAccount}
	}
	if isSubtaskType(in.Type) && strings.TrimSpace(in.ParentKey) != "" {
		fields["parent"] = map[string]string{"key": in.ParentKey}
	}
	return fields
}

// adfDescription embrulha texto puro no formato Atlassian Document Format
// (ADF) que a API v3 exige no campo description. É o mínimo pra texto simples.
func adfDescription(text string) map[string]any {
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]string{"type": "text", "text": text},
				},
			},
		},
	}
}

// sendJSON manda um request com body JSON e decoda out se resp for 2xx. Se
// out for nil, ignora o corpo (usado pra operações que só retornam status).
func (c *Client) sendJSON(method, path string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("jira: encode payload: %w", err)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("jira: build request: %w", err)
	}
	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("jira: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jira: %s %s: HTTP %d: %s", method, path, resp.StatusCode, truncate(string(respBody), 200))
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("jira: decode response: %w", err)
	}
	return nil
}
