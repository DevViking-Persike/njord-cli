package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// MigrateFromDataSh reads the shell-based data.sh file and produces a Config.
// This is a one-time migration tool.
func MigrateFromDataSh(dataShPath string) (*Config, error) {
	f, err := os.Open(dataShPath)
	if err != nil {
		return nil, fmt.Errorf("opening data.sh: %w", err)
	}
	defer f.Close()

	cfg := &Config{
		Settings: Settings{
			Editor:       "code",
			ProjectsBase: "~/Avita",
			PersonalBase: "~/Persike",
		},
	}

	scanner := bufio.NewScanner(f)

	var currentArrayName string
	var currentEntries []string

	var catNames []string
	var catSubs []string
	var catIDs []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip function definitions
		if strings.HasPrefix(line, "njord_") || strings.HasPrefix(line, "case") ||
			strings.HasPrefix(line, "esac") || line == ";;" ||
			strings.HasPrefix(line, "echo") || strings.HasPrefix(line, "printf") {
			continue
		}

		// Handle single-line arrays: VAR=("val1" "val2" "val3")
		if strings.HasPrefix(line, "NJORD_") && strings.Contains(line, "=(") && strings.HasSuffix(line, ")") {
			varName := strings.Split(line, "=")[0]
			// Extract content between ( and )
			start := strings.Index(line, "(")
			end := strings.LastIndex(line, ")")
			if start >= 0 && end > start {
				content := line[start+1 : end]
				entries := parseShellQuotedArray(content)

				switch varName {
				case "NJORD_CAT_NAMES":
					catNames = entries
				case "NJORD_CAT_SUBS":
					catSubs = entries
				case "NJORD_CAT_IDS":
					catIDs = entries
				}
			}
			continue
		}

		// Handle multi-line arrays: VAR=(
		if strings.Contains(line, "=(") && strings.HasPrefix(line, "NJORD_") {
			currentArrayName = strings.Split(line, "=")[0]
			currentEntries = nil
			continue
		}

		// Close of multi-line array
		if line == ")" && currentArrayName != "" {
			switch {
			case strings.HasPrefix(currentArrayName, "NJORD_PROJ_"):
				catKey := strings.ToLower(strings.TrimPrefix(currentArrayName, "NJORD_PROJ_"))
				projects := parseProjectEntries(currentEntries)
				cfg.Categories = append(cfg.Categories, Category{
					ID:       catKey,
					Projects: projects,
				})

			case currentArrayName == "NJORD_DOCKER_STACKS":
				cfg.DockerStacks = parseDockerEntries(currentEntries)
			}

			currentArrayName = ""
			continue
		}

		// Collect entries inside multi-line array
		if currentArrayName != "" {
			entry := strings.Trim(line, `"`)
			currentEntries = append(currentEntries, entry)
		}
	}

	// Map category names and subs using IDs
	catIDToName := map[string]string{}
	catIDToSub := map[string]string{}
	for i, id := range catIDs {
		if i < len(catNames) {
			catIDToName[id] = catNames[i]
		}
		if i < len(catSubs) {
			catIDToSub[id] = catSubs[i]
		}
	}

	// Apply names and subs to categories
	for i := range cfg.Categories {
		cat := &cfg.Categories[i]
		if name, ok := catIDToName[cat.ID]; ok {
			cat.Name = name
		} else {
			cat.Name = strings.ToUpper(cat.ID[:1]) + cat.ID[1:]
		}
		if sub, ok := catIDToSub[cat.ID]; ok {
			cat.Sub = sub
		}
	}

	return cfg, scanner.Err()
}

// parseShellQuotedArray extracts quoted strings from shell array content.
// e.g. `"Todos" "Alfandega" "Jobs"` -> ["Todos", "Alfandega", "Jobs"]
func parseShellQuotedArray(content string) []string {
	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindAllStringSubmatch(content, -1)
	var result []string
	for _, m := range matches {
		if len(m) >= 2 {
			result = append(result, m[1])
		}
	}
	return result
}

func parseProjectEntries(entries []string) []Project {
	var projects []Project
	for _, entry := range entries {
		parts := strings.SplitN(entry, "|", 3)
		if len(parts) != 3 {
			continue
		}
		projects = append(projects, Project{
			Alias: strings.TrimSpace(parts[0]),
			Desc:  strings.TrimSpace(parts[1]),
			Path:  strings.TrimSpace(parts[2]),
		})
	}
	return projects
}

func parseDockerEntries(entries []string) []DockerStack {
	var stacks []DockerStack
	for _, entry := range entries {
		parts := strings.SplitN(entry, "|", 3)
		if len(parts) != 3 {
			continue
		}
		stacks = append(stacks, DockerStack{
			Name: strings.TrimSpace(parts[0]),
			Desc: strings.TrimSpace(parts[1]),
			Path: strings.TrimSpace(parts[2]),
		})
	}
	return stacks
}
