package gitlabclient

import (
	"time"
)

type MergeRequestInfo struct {
	IID       int64
	Title     string
	State     string // opened, merged, closed
	Author    string
	Branch    string
	Target    string
	CreatedAt time.Time
	Pipeline  string // status da pipeline do MR
	URL       string
	ProjectID int64  // usado por ListMyOpenMRs
}

type PipelineInfo struct {
	ID        int64
	Status    string // running, success, failed, pending, canceled
	Ref       string
	CreatedAt time.Time
	URL       string
}

type BranchInfo struct {
	Name       string
	CommitDate time.Time
	Merged     bool
	Protected  bool
	Default    bool
	MRApproval *MRApprovalInfo // nil if no open MR for this branch
}

type MRApprovalInfo struct {
	MRIID             int64
	Title             string
	ApprovalsRequired int
	ApprovalsGiven    int
	Approved          bool
	RuleName          string // e.g. "Code Review B1"
}

type RecentPush struct {
	ProjectID int64
	Ref       string
	CreatedAt time.Time
}

// ProjectInfo é a projeção usada quando listamos todos os projetos acessíveis
// pelo token — serve de input pro fluxo "Clonar novo" na TUI.
type ProjectInfo struct {
	PathWithNamespace string // "grupo/sub/repo"
	Description       string
	SSHURLToRepo      string // git@gitlab.com:grupo/sub/repo.git
	WebURL            string // https://gitlab.com/grupo/sub/repo
}

// GroupInfo é a projeção de um grupo/subgrupo acessível pelo token.
// Usada na TUI pra navegar por camadas antes de listar os repos (fetch leve).
type GroupInfo struct {
	ID          int64
	Name        string // "bill"
	FullPath    string // "avitaseg/bill" (identificador estável pro HTTP)
	FullName    string // "avitaseg / bill" (pra display)
	Description string
	WebURL      string
}
