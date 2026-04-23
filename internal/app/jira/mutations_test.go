package jira

import (
	"errors"
	"strings"
	"testing"

	"github.com/DevViking-Persike/njord-cli/internal/jiraclient"
)

func TestCreateIssueAsMe_ForcesAssignee(t *testing.T) {
	gw := &fakeJiraGW{
		user:         jiraclient.User{AccountID: "acc-42"},
		createResult: jiraclient.Issue{Key: "GAP-100"},
	}
	svc := NewJiraService(gw)

	issue, err := svc.CreateIssueAsMe(CreateIssueRequest{
		ProjectKey: "GAP", Summary: "x", Type: "Task",
	})
	if err != nil {
		t.Fatal(err)
	}
	if issue.Key != "GAP-100" {
		t.Fatalf("Key = %q", issue.Key)
	}
	if gw.createdInput.AssigneeAccount != "acc-42" {
		t.Fatalf("AssigneeAccount = %q, expected from CurrentUser()", gw.createdInput.AssigneeAccount)
	}
}

func TestCreateIssueAsMe_AppliesTransition(t *testing.T) {
	gw := &fakeJiraGW{createResult: jiraclient.Issue{Key: "GAP-7"}}
	svc := NewJiraService(gw)

	_, err := svc.CreateIssueAsMe(CreateIssueRequest{
		ProjectKey: "GAP", Summary: "x", Type: "Task", TransitionID: "31",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gw.transitionKey != "GAP-7" || gw.transitionID != "31" {
		t.Fatalf("transition not applied: key=%q id=%q", gw.transitionKey, gw.transitionID)
	}
}

func TestCreateIssueAsMe_UserErr(t *testing.T) {
	gw := &fakeJiraGW{userErr: errors.New("401")}
	svc := NewJiraService(gw)
	_, err := svc.CreateIssueAsMe(CreateIssueRequest{ProjectKey: "GAP", Summary: "x", Type: "Task"})
	if err == nil {
		t.Fatal("expected err propagated")
	}
}

func TestUpdateIssue_OnlyStatus(t *testing.T) {
	gw := &fakeJiraGW{}
	svc := NewJiraService(gw)
	err := svc.UpdateIssue(UpdateIssueRequest{Key: "GAP-1", TransitionID: "31"})
	if err != nil {
		t.Fatal(err)
	}
	if gw.updateKey != "" {
		t.Fatalf("não era pra chamar UpdateIssue quando só tem transição")
	}
	if gw.transitionID != "31" {
		t.Fatalf("transição não aplicada")
	}
}

func TestUpdateIssue_OnlySummary(t *testing.T) {
	gw := &fakeJiraGW{}
	svc := NewJiraService(gw)
	err := svc.UpdateIssue(UpdateIssueRequest{Key: "GAP-1", Summary: "novo"})
	if err != nil {
		t.Fatal(err)
	}
	if gw.updateKey != "GAP-1" || gw.updateInput.Summary != "novo" {
		t.Fatalf("update não propagado: %+v", gw)
	}
}

func TestUpdateIssue_Empty(t *testing.T) {
	err := NewJiraService(&fakeJiraGW{}).UpdateIssue(UpdateIssueRequest{Key: "GAP-1"})
	if err == nil || !strings.Contains(err.Error(), "nada pra editar") {
		t.Fatalf("expected 'nada pra editar', got %v", err)
	}
}

func TestCreateIssueAsMe_TargetCategoryResolvesTransition(t *testing.T) {
	gw := &fakeJiraGW{
		createResult: jiraclient.Issue{Key: "GAP-5"},
		transitions: []jiraclient.Transition{
			{ID: "10", Name: "A fazer", StatusCat: "new"},
			{ID: "11", Name: "Em desenvolvimento", StatusCat: "indeterminate"},
			{ID: "12", Name: "Concluído", StatusCat: "done"},
		},
	}
	svc := NewJiraService(gw)
	_, err := svc.CreateIssueAsMe(CreateIssueRequest{
		ProjectKey: "GAP", Summary: "x", Type: "Task", TargetCategory: "indeterminate",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gw.transitionID != "11" {
		t.Fatalf("expected transition 11 (indeterminate), got %q", gw.transitionID)
	}
}

func TestCreateIssueAsMe_TargetCategoryNotFound(t *testing.T) {
	// Nenhuma transição bate com a categoria — não deve retornar erro, só ignora.
	gw := &fakeJiraGW{
		createResult: jiraclient.Issue{Key: "GAP-6"},
		transitions:  []jiraclient.Transition{{ID: "10", StatusCat: "new"}},
	}
	svc := NewJiraService(gw)
	_, err := svc.CreateIssueAsMe(CreateIssueRequest{
		ProjectKey: "GAP", Summary: "x", Type: "Task", TargetCategory: "done",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gw.transitionID != "" {
		t.Fatalf("não deveria ter aplicado transição, got %q", gw.transitionID)
	}
}

func TestListTransitions_Passthrough(t *testing.T) {
	gw := &fakeJiraGW{transitions: []jiraclient.Transition{{ID: "11", Name: "Em desenvolvimento"}}}
	ts, err := NewJiraService(gw).ListTransitions("GAP-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ts) != 1 || ts[0].ID != "11" {
		t.Fatalf("unexpected: %+v", ts)
	}
}
