// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"fmt"
	"regexp"
	"strings"
)

// Regular expression for valid GitLab repository name characters.
// It allows names that start with a letter, number, or underscore,
// followed by any number of letters, numbers, underscores, dots, plus signs, hyphens, or spaces.
var gitlabNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.+\- ]*$`)

// invalidGitLabNames is a comprehensive set of repository names that are not allowed in GitLab.
// These names are reserved for GitLab's internal use or have special meanings.
var invalidGitLabNames = map[string]bool{
	// Basic reserved names
	"-":  true,
	".":  true,
	"..": true,

	// GitLab system paths
	"badges":             true,
	"blame":              true,
	"blob":               true,
	"builds":             true,
	"commits":            true,
	"create":             true,
	"create_dir":         true,
	"edit":               true,
	"environments":       true,
	"files":              true,
	"find_file":          true,
	"gitlab-lfs":         true,
	"groups":             true,
	"hooks":              true,
	"issues":             true,
	"merge_requests":     true,
	"new":                true,
	"notes":              true,
	"notable":            true,
	"pipelines":          true,
	"protected_branches": true,
	"raw":                true,
	"repository":         true,
	"snippets":           true,
	"tree":               true,
	"update":             true,
	"wikis":              true,

	// Admin and API paths
	"admin":        true,
	"api":          true,
	"autodeploy":   true,
	"explore":      true,
	"health":       true,
	"import":       true,
	"jwt":          true,
	"koding":       true,
	"help":         true,
	"s":            true,
	"unsubscribes": true,
	"users":        true,
	"v2":           true,

	// Additional system reserved names
	"dashboard":   true,
	"projects":    true,
	"public":      true,
	"sitemap":     true,
	"robots.txt":  true,
	"favicon.ico": true,
	"assets":      true,
	"uploads":     true,
}

// ValidateAndCleanGitLabName validates and cleans a GitLab repository name.
// It returns the cleaned name, whether it's valid, and any validation issues.
func ValidateAndCleanGitLabName(name string) (string, bool, []string) {
	var issues []string

	// Trim whitespace
	cleanName := strings.TrimSpace(name)

	// Check for empty name
	if cleanName == "" {
		issues = append(issues, "repository name cannot be empty")
		return cleanName, false, issues
	}

	// Check if the name matches the allowed pattern
	if !gitlabNameRegex.MatchString(cleanName) {
		issues = append(issues, "repository name contains invalid characters")
	}

	// Check against reserved names (case insensitive)
	if invalidGitLabNames[strings.ToLower(cleanName)] {
		issues = append(issues, fmt.Sprintf("'%s' is a reserved GitLab name", cleanName))
	}

	// GitLab specific validations
	if len(cleanName) > 255 {
		issues = append(issues, "repository name is too long (max 255 characters)")
	}

	// Cannot start or end with a dot
	if strings.HasPrefix(cleanName, ".") || strings.HasSuffix(cleanName, ".") {
		issues = append(issues, "repository name cannot start or end with a dot")
	}

	// Cannot be just numbers (GitLab specific restriction)
	if regexp.MustCompile(`^\d+$`).MatchString(cleanName) {
		issues = append(issues, "repository name cannot be only numbers")
	}

	return cleanName, len(issues) == 0, issues
}
