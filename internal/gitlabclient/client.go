package gitlabclient

import (
	"fmt"
	"sort"
	"sync"
	"time"

	gl "gitlab.com/gitlab-org/api/client-go"
)

type Client struct {
	gl           *gl.Client
	username     string
	projectPaths map[int64]string
	pathsMu      sync.Mutex
}

func NewClient(token, url string) (*Client, error) {
	opts := []gl.ClientOptionFunc{}
	if url != "" {
		opts = append(opts, gl.WithBaseURL(url))
	}
	client, err := gl.NewClient(token, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating gitlab client: %w", err)
	}

	c := &Client{gl: client, projectPaths: make(map[int64]string)}

	// Fetch current user to store username for filtering
	user, _, err := client.Users.CurrentUser()
	if err == nil && user != nil {
		c.username = user.Username
	}

	return c, nil
}

func (c *Client) Username() string {
	return c.username
}

func (c *Client) Close() {
	// noop, for consistency with Docker client
}

func (c *Client) ListMyOpenMRs() ([]MergeRequestInfo, error) {
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	opts := &gl.ListMergeRequestsOptions{
		Scope:        gl.Ptr("created_by_me"),
		State:        gl.Ptr("opened"),
		OrderBy:      gl.Ptr("created_at"),
		Sort:         gl.Ptr("desc"),
		CreatedAfter: &threeDaysAgo,
		ListOptions: gl.ListOptions{
			PerPage: int64(10),
		},
	}

	mrs, _, err := c.gl.MergeRequests.ListMergeRequests(opts)
	if err != nil {
		return nil, fmt.Errorf("listing my open MRs: %w", err)
	}

	var result []MergeRequestInfo
	for _, mr := range mrs {
		info := MergeRequestInfo{
			IID:       mr.IID,
			Title:     mr.Title,
			State:     mr.State,
			Branch:    mr.SourceBranch,
			Target:    mr.TargetBranch,
			CreatedAt: *mr.CreatedAt,
			URL:       mr.WebURL,
			ProjectID: mr.ProjectID,
		}
		if mr.Author != nil {
			info.Author = mr.Author.Username
		}
		result = append(result, info)
	}
	return result, nil
}

func (c *Client) ListMergeRequests(projectPath string, state string) ([]MergeRequestInfo, error) {
	opts := &gl.ListProjectMergeRequestsOptions{
		State:   gl.Ptr(state),
		OrderBy: gl.Ptr("created_at"),
		Sort:    gl.Ptr("desc"),
		ListOptions: gl.ListOptions{
			PerPage: int64(20),
		},
	}

	mrs, _, err := c.gl.MergeRequests.ListProjectMergeRequests(projectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("listing merge requests: %w", err)
	}

	var result []MergeRequestInfo
	for _, mr := range mrs {
		info := MergeRequestInfo{
			IID:       mr.IID,
			Title:     mr.Title,
			State:     mr.State,
			Author:    mr.Author.Username,
			Branch:    mr.SourceBranch,
			Target:    mr.TargetBranch,
			CreatedAt: *mr.CreatedAt,
			URL:       mr.WebURL,
		}
		// Pipeline status is not available on BasicMergeRequest from list endpoint
		result = append(result, info)
	}
	return result, nil
}

func (c *Client) ListPipelines(projectPath string, limit int) ([]PipelineInfo, error) {
	opts := &gl.ListProjectPipelinesOptions{
		Sort: gl.Ptr("desc"),
		ListOptions: gl.ListOptions{
			PerPage: int64(limit),
		},
	}
	if c.username != "" {
		opts.Username = gl.Ptr(c.username)
	}

	pipelines, _, err := c.gl.Pipelines.ListProjectPipelines(projectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("listing pipelines: %w", err)
	}

	var result []PipelineInfo
	for _, p := range pipelines {
		result = append(result, PipelineInfo{
			ID:        p.ID,
			Status:    p.Status,
			Ref:       p.Ref,
			CreatedAt: *p.CreatedAt,
			URL:       p.WebURL,
		})
	}
	return result, nil
}

func (c *Client) TriggerPipeline(projectPath, ref string) (*PipelineInfo, error) {
	pipeline, _, err := c.gl.Pipelines.CreatePipeline(projectPath, &gl.CreatePipelineOptions{
		Ref: gl.Ptr(ref),
	})
	if err != nil {
		return nil, fmt.Errorf("triggering pipeline: %w", err)
	}

	return &PipelineInfo{
		ID:        pipeline.ID,
		Status:    pipeline.Status,
		Ref:       pipeline.Ref,
		CreatedAt: *pipeline.CreatedAt,
		URL:       pipeline.WebURL,
	}, nil
}

func (c *Client) CreateBranch(projectPath, branchName, ref string) error {
	_, _, err := c.gl.Branches.CreateBranch(projectPath, &gl.CreateBranchOptions{
		Branch: gl.Ptr(branchName),
		Ref:    gl.Ptr(ref),
	})
	if err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}
	return nil
}

func (c *Client) ListRecentPushes(days int) ([]RecentPush, error) {
	after := gl.ISOTime(time.Now().AddDate(0, 0, -days))
	opts := &gl.ListContributionEventsOptions{
		Action: gl.Ptr(gl.PushedEventType),
		After:  &after,
		Sort:   gl.Ptr("desc"),
		ListOptions: gl.ListOptions{
			PerPage: int64(100),
		},
	}

	events, _, err := c.gl.Events.ListCurrentUserContributionEvents(opts)
	if err != nil {
		return nil, fmt.Errorf("listing push events: %w", err)
	}

	// Deduplicate by project ID, keep most recent
	seen := make(map[int64]bool)
	var result []RecentPush
	for _, e := range events {
		if seen[e.ProjectID] {
			continue
		}
		seen[e.ProjectID] = true
		result = append(result, RecentPush{
			ProjectID: e.ProjectID,
			Ref:       e.PushData.Ref,
			CreatedAt: *e.CreatedAt,
		})
	}
	return result, nil
}

// ResolveProjectPath returns the web path for a project ID (cached).
func (c *Client) ResolveProjectPath(projectID int64) (string, error) {
	c.pathsMu.Lock()
	if path, ok := c.projectPaths[projectID]; ok {
		c.pathsMu.Unlock()
		return path, nil
	}
	c.pathsMu.Unlock()

	project, _, err := c.gl.Projects.GetProject(projectID, nil)
	if err != nil {
		return "", err
	}

	c.pathsMu.Lock()
	c.projectPaths[projectID] = project.PathWithNamespace
	c.pathsMu.Unlock()

	return project.PathWithNamespace, nil
}

func (c *Client) ListBranches(projectPath string) ([]string, error) {
	opts := &gl.ListBranchesOptions{
		ListOptions: gl.ListOptions{
			PerPage: int64(100),
		},
	}

	branches, _, err := c.gl.Branches.ListBranches(projectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}

	var result []string
	for _, b := range branches {
		result = append(result, b.Name)
	}
	return result, nil
}

func (c *Client) ListBranchesDetailed(projectPath string) ([]BranchInfo, error) {
	opts := &gl.ListBranchesOptions{
		ListOptions: gl.ListOptions{
			PerPage: int64(100),
		},
	}

	branches, _, err := c.gl.Branches.ListBranches(projectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("listing branches: %w", err)
	}

	var result []BranchInfo
	for _, b := range branches {
		info := BranchInfo{
			Name:      b.Name,
			Merged:    b.Merged,
			Protected: b.Protected,
			Default:   b.Default,
		}
		if b.Commit != nil {
			if b.Commit.CommittedDate != nil {
				info.CommitDate = *b.Commit.CommittedDate
			} else if b.Commit.CreatedAt != nil {
				info.CommitDate = *b.Commit.CreatedAt
			} else if b.Commit.AuthoredDate != nil {
				info.CommitDate = *b.Commit.AuthoredDate
			}
		}
		result = append(result, info)
	}

	// Sort by commit date descending (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CommitDate.After(result[j].CommitDate)
	})

	// Fetch open MRs and enrich with approval info
	mrs, err := c.ListMergeRequests(projectPath, "opened")
	if err == nil && len(mrs) > 0 {
		// Map branch name → MR
		branchToMR := make(map[string]MergeRequestInfo)
		for _, mr := range mrs {
			branchToMR[mr.Branch] = mr
		}

		for i, bi := range result {
			if mr, ok := branchToMR[bi.Name]; ok {
				approval := c.fetchApprovalInfo(projectPath, mr.IID, mr.Title)
				result[i].MRApproval = approval
			}
		}
	}

	return result, nil
}

// GetProjectLatestMRApproval returns the approval info for the most recent open MR.
func (c *Client) GetProjectLatestMRApproval(projectPath string) (*MRApprovalInfo, error) {
	mrs, err := c.ListMergeRequests(projectPath, "opened")
	if err != nil {
		return nil, err
	}
	if len(mrs) == 0 {
		return nil, nil
	}
	mr := mrs[0]
	return c.fetchApprovalInfo(projectPath, mr.IID, mr.Title), nil
}

// GetMRApproval returns the approval info for a specific MR.
func (c *Client) GetMRApproval(projectPath string, mrIID int64, mrTitle string) *MRApprovalInfo {
	return c.fetchApprovalInfo(projectPath, mrIID, mrTitle)
}

// TriggerMRPipeline finds the open MR for a branch and creates a merge request pipeline.
func (c *Client) TriggerMRPipeline(projectPath, branch string) (*PipelineInfo, error) {
	// Busca MR aberto para esta branch
	opts := &gl.ListProjectMergeRequestsOptions{
		State:        gl.Ptr("opened"),
		SourceBranch: gl.Ptr(branch),
		ListOptions:  gl.ListOptions{PerPage: int64(1)},
	}
	mrs, _, err := c.gl.MergeRequests.ListProjectMergeRequests(projectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("buscando MR para branch %s: %w", branch, err)
	}
	if len(mrs) == 0 {
		return nil, fmt.Errorf("nenhum MR aberto encontrado para branch %s", branch)
	}

	mr := mrs[0]
	pipeline, _, err := c.gl.MergeRequests.CreateMergeRequestPipeline(projectPath, mr.IID)
	if err != nil {
		return nil, fmt.Errorf("criando pipeline do MR !%d: %w", mr.IID, err)
	}

	return &PipelineInfo{
		ID:        pipeline.ID,
		Status:    pipeline.Status,
		Ref:       pipeline.Ref,
		CreatedAt: *pipeline.CreatedAt,
		URL:       pipeline.WebURL,
	}, nil
}

func (c *Client) fetchApprovalInfo(projectPath string, mrIID int64, mrTitle string) *MRApprovalInfo {
	rules, _, err := c.gl.MergeRequestApprovals.GetApprovalRules(projectPath, mrIID)
	if err != nil || len(rules) == 0 {
		// Fallback: try GetConfiguration for basic approval info
		config, _, err := c.gl.MergeRequestApprovals.GetConfiguration(projectPath, mrIID)
		if err != nil {
			return &MRApprovalInfo{MRIID: mrIID, Title: mrTitle}
		}
		return &MRApprovalInfo{
			MRIID:             mrIID,
			Title:             mrTitle,
			ApprovalsRequired: int(config.ApprovalsRequired),
			ApprovalsGiven:    len(config.ApprovedBy),
			Approved:          config.Approved,
		}
	}

	// Aggregate: find the first non-approved rule, or summarize all
	totalRequired := 0
	totalGiven := 0
	allApproved := true
	ruleName := ""
	for _, r := range rules {
		totalRequired += int(r.ApprovalsRequired)
		totalGiven += len(r.ApprovedBy)
		if !r.Approved {
			allApproved = false
			if ruleName == "" {
				ruleName = r.Name
			}
		}
	}

	return &MRApprovalInfo{
		MRIID:             mrIID,
		Title:             mrTitle,
		ApprovalsRequired: totalRequired,
		ApprovalsGiven:    totalGiven,
		Approved:          allApproved,
		RuleName:          ruleName,
	}
}
