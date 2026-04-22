package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/DevViking-Persike/njord-cli/internal/gitlabclient"
	"github.com/DevViking-Persike/njord-cli/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func mrStateStyle(state string) lipgloss.Style {
	switch state {
	case "opened":
		return theme.MROpenStyle
	case "merged":
		return theme.MRMergedStyle
	case "closed":
		return theme.MRClosedStyle
	default:
		return theme.DimStyle
	}
}

func pipelineStateStyle(status string) lipgloss.Style {
	switch status {
	case "success":
		return theme.PipelineSuccessStyle
	case "failed":
		return theme.PipelineFailedStyle
	case "running":
		return theme.PipelineRunningStyle
	case "pending":
		return theme.PipelinePendingStyle
	case "canceled":
		return theme.DimStyle
	default:
		return theme.DimStyle
	}
}

func (m GitLabActionsModel) renderBranchLine(i int) string {
	branch := m.branches[i]
	prefix := "  "
	if i == m.cursor {
		prefix = "▶ "
	}

	var namePart string
	if i == m.cursor {
		namePart = theme.GitLabTitleSelectedStyle.Render(prefix + branch.Name)
	} else {
		namePart = theme.TextStyle.Render(prefix + branch.Name)
	}

	tags := ""
	if branch.Default {
		tags += " " + theme.DimStyle.Render("[default]")
	}
	if branch.Protected {
		tags += " " + theme.DimStyle.Render("[protected]")
	}

	approvalTag := ""
	if branch.MRApproval != nil {
		approvalTag = " " + renderApprovalTag(branch.MRApproval)
	}

	ago := ""
	if !branch.CommitDate.IsZero() {
		ago = " " + theme.DimStyle.Render(timeAgo(branch.CommitDate))
	}

	return "  " + namePart + tags + approvalTag + ago + "\n"
}

func renderApprovalTag(a *gitlabclient.MRApprovalInfo) string {
	if a.Approved {
		return theme.PipelineSuccessStyle.Render("✓ aprovado")
	}
	left := a.ApprovalsRequired - a.ApprovalsGiven
	if left <= 0 {
		left = 1
	}
	label := fmt.Sprintf("⏳ %d/%d", a.ApprovalsGiven, a.ApprovalsRequired)
	if a.RuleName != "" {
		label += " " + a.RuleName
	}
	return theme.PipelinePendingStyle.Render(label)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "agora"
	case d < time.Hour:
		return fmt.Sprintf("%dm atrás", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh atrás", int(d.Hours()))
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "ontem"
		}
		return fmt.Sprintf("%dd atrás", days)
	}
}
