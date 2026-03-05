package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// git@gitlab.com:grupo/repo.git
	sshRemoteRe = regexp.MustCompile(`git@[^:]+:(.+?)(?:\.git)?$`)
	// https://gitlab.com/grupo/repo.git
	httpsRemoteRe = regexp.MustCompile(`https?://[^/]+/(.+?)(?:\.git)?$`)
)

// ParseGitLabPath extracts the GitLab project path (e.g. "grupo/repo") from the
// git origin remote URL found in the repository at repoDir.
func ParseGitLabPath(repoDir string) (string, error) {
	gitConfig := filepath.Join(repoDir, ".git", "config")
	data, err := os.ReadFile(gitConfig)
	if err != nil {
		return "", fmt.Errorf("reading git config: %w", err)
	}

	originURL := extractOriginURL(string(data))
	if originURL == "" {
		return "", fmt.Errorf("no origin remote found in %s", gitConfig)
	}

	path := parseURL(originURL)
	if path == "" {
		return "", fmt.Errorf("could not parse gitlab path from URL: %s", originURL)
	}
	return path, nil
}

func extractOriginURL(gitConfig string) string {
	lines := strings.Split(gitConfig, "\n")
	inOrigin := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == `[remote "origin"]` {
			inOrigin = true
			continue
		}
		if strings.HasPrefix(trimmed, "[") {
			inOrigin = false
			continue
		}
		if inOrigin && strings.HasPrefix(trimmed, "url = ") {
			return strings.TrimPrefix(trimmed, "url = ")
		}
	}
	return ""
}

func parseURL(url string) string {
	if m := sshRemoteRe.FindStringSubmatch(url); len(m) > 1 {
		return m[1]
	}
	if m := httpsRemoteRe.FindStringSubmatch(url); len(m) > 1 {
		return m[1]
	}
	return ""
}
