package jira

import (
	"fmt"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
)

// CreateIssueRequest é o input do caso de uso "criar card". Aceita o tipo sem
// accountId do assignee — o service injeta o usuário atual automaticamente.
type CreateIssueRequest struct {
	ProjectKey   string
	Summary      string
	Description  string
	Type         string
	ParentKey    string
	TransitionID string // opcional; prioridade sobre TargetCategory
	// TargetCategory permite escolher status por categoria ("indeterminate"
	// = "Em desenvolvimento", "done" = "Concluído") sem saber o ID da transição.
	// O service resolve a transição dinâmica depois de criar a issue.
	TargetCategory string
}

// CreateIssueAsMe cria a issue assumindo o assignee como o usuário atual (sempre)
// e, se TransitionID/TargetCategory vier preenchido, aplica a transição.
func (s *JiraService) CreateIssueAsMe(in CreateIssueRequest) (jiraclient.Issue, error) {
	me, err := s.gw.CurrentUser()
	if err != nil {
		return jiraclient.Issue{}, fmt.Errorf("obtendo usuário atual: %w", err)
	}
	issue, err := s.gw.CreateIssue(jiraclient.CreateIssueInput{
		ProjectKey:      in.ProjectKey,
		Summary:         in.Summary,
		Description:     in.Description,
		Type:            in.Type,
		ParentKey:       in.ParentKey,
		AssigneeAccount: me.AccountID,
	})
	if err != nil {
		return jiraclient.Issue{}, err
	}
	transitionID := in.TransitionID
	if transitionID == "" && in.TargetCategory != "" {
		transitionID = s.pickTransitionByCategory(issue.Key, in.TargetCategory)
	}
	if transitionID != "" {
		if err := s.gw.TransitionIssue(issue.Key, transitionID); err != nil {
			return issue, fmt.Errorf("issue %s criada mas falhou a transição: %w", issue.Key, err)
		}
	}
	return issue, nil
}

// pickTransitionByCategory procura a primeira transição que leva pra uma
// categoria específica ("new", "indeterminate", "done"). Retorna "" se não
// achar — nesse caso a issue fica no status default do projeto.
func (s *JiraService) pickTransitionByCategory(key, category string) string {
	transitions, err := s.gw.ListTransitions(key)
	if err != nil {
		return ""
	}
	for _, t := range transitions {
		if t.StatusCat == category {
			return t.ID
		}
	}
	return ""
}

// UpdateIssueRequest é o input do caso de uso "editar card". Campos vazios são
// deixados como estão; TransitionID opcional aplica mudança de status.
type UpdateIssueRequest struct {
	Key          string
	Summary      string
	Description  string
	TransitionID string
}

// UpdateIssue atualiza summary/desc (o que não vier vazio) e, se TransitionID
// estiver preenchido, aplica a transição de status.
func (s *JiraService) UpdateIssue(in UpdateIssueRequest) error {
	hasFieldUpdate := in.Summary != "" || in.Description != ""
	if hasFieldUpdate {
		if err := s.gw.UpdateIssue(in.Key, jiraclient.UpdateIssueInput{
			Summary:     in.Summary,
			Description: in.Description,
		}); err != nil {
			return err
		}
	}
	if in.TransitionID != "" {
		if err := s.gw.TransitionIssue(in.Key, in.TransitionID); err != nil {
			return fmt.Errorf("aplicando transição: %w", err)
		}
	}
	if !hasFieldUpdate && in.TransitionID == "" {
		return fmt.Errorf("nada pra editar")
	}
	return nil
}

// ListTransitions expõe as opções de mudança de status pra uma issue.
// Delega direto no gateway — fica como passthrough pra testabilidade.
func (s *JiraService) ListTransitions(key string) ([]jiraclient.Transition, error) {
	return s.gw.ListTransitions(key)
}
