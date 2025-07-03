// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"regexp"
	"strings"

	"itiquette/git-provider-sync/internal/domain/constants"
)

// Regular expression for valid GitLab repository name characters.
// It allows names that start with a letter, number, or underscore,
// followed by any number of letters, numbers, underscores, dots, plus signs, hyphens, or spaces.
var gitlabNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.+\- ]*$`)

// invalidGitLabNames is a comprehensive set of repository names that are not allowed in GitLab.
// These names are reserved for GitLab's internal use or have special meanings.
// This restores the extensive reserved names list from main branch.
var invalidGitLabNames = map[string]bool{
	// Basic reserved names
	"-":  true,
	".":  true,
	"..": true,

	// GitLab system paths
	"badges":               true,
	"blame":                true,
	"blob":                 true,
	"builds":               true,
	"commits":              true,
	"create":               true,
	"create_dir":           true,
	"edit":                 true,
	"environments/folders": true,
	"files":                true,
	"find_file":            true,
	"gitlab-lfs/objects":   true,
	"info/lfs/objects":     true,
	"new":                  true,
	"preview":              true,
	"raw":                  true,
	"refs":                 true,
	"tree":                 true,
	"update":               true,
	"wikis":                true,

	// Additional GitLab reserved paths
	"admin":         true,
	"api":           true,
	"assets":        true,
	"autocomplete":  true,
	"dashboard":     true,
	"explore":       true,
	"groups":        true,
	"help":          true,
	"import":        true,
	"notifications": true,
	"profile":       true,
	"projects":      true,
	"public":        true,
	"search":        true,
	"snippets":      true,
	"unsubscribes":  true,
	"users":         true,

	// CI/CD and DevOps paths
	"ci":           true,
	"runners":      true,
	"pipelines":    true,
	"jobs":         true,
	"deploy_keys":  true,
	"environments": true,
	"releases":     true,
	"tags":         true,
	"branches":     true,

	// Security and compliance
	"security":        true,
	"compliance":      true,
	"audit":           true,
	"vulnerabilities": true,
	"dependencies":    true,

	// Package registry
	"packages":           true,
	"container_registry": true,
	"npm":                true,
	"maven":              true,
	"pypi":               true,
	"nuget":              true,
	"rubygems":           true,

	// Monitoring and analytics
	"analytics":  true,
	"insights":   true,
	"metrics":    true,
	"prometheus": true,
	"grafana":    true,

	// Infrastructure
	"infrastructure": true,
	"terraform":      true,
	"kubernetes":     true,
	"clusters":       true,

	// Version control
	"git":              true,
	"git-upload-pack":  true,
	"git-receive-pack": true,
	"info":             true,
	"objects":          true,

	// Web and API endpoints
	"oauth":       true,
	"auth":        true,
	"login":       true,
	"logout":      true,
	"register":    true,
	"session":     true,
	"settings":    true,
	"preferences": true,

	// Documentation and wiki
	"docs":          true,
	"documentation": true,
	"wiki":          true,
	"pages":         true,

	// Issue tracking
	"issues":         true,
	"merge_requests": true,
	"milestones":     true,
	"labels":         true,
	"boards":         true,

	// Project management
	"project":      true,
	"group":        true,
	"organization": true,
	"org":          true,

	// System administration
	"administrator": true,
	"root":          true,
	"system":        true,
	"support":       true,
	"postmaster":    true,
	"webmaster":     true,
	"hostmaster":    true,
	"abuse":         true,
	"noreply":       true,
	"no-reply":      true,

	// Common service names
	"www":   true,
	"ftp":   true,
	"mail":  true,
	"email": true,
	"smtp":  true,
	"pop":   true,
	"imap":  true,
	"dns":   true,
	"ns":    true,
	"ns1":   true,
	"ns2":   true,
	"ns3":   true,
	"ns4":   true,
	"mx":    true,
	"mx1":   true,
	"mx2":   true,

	// Development and testing
	"test":        true,
	"testing":     true,
	"dev":         true,
	"development": true,
	"staging":     true,
	"production":  true,
	"prod":        true,
	"demo":        true,
	"example":     true,
	"sample":      true,
	"tmp":         true,
	"temp":        true,
	"backup":      true,
}

// IsValidGitLabName checks if the given name is a valid GitLab repository name.
// It returns true if the name:
//  1. Contains only valid characters (defined by gitlabNameRegex)
//  2. Is not in the list of invalid names (defined in invalidGitLabNames)
//  3. Does not exceed the maximum length of 256 characters
//
// The check is case-insensitive for invalid names.
//
// Parameters:
//   - name: The repository name to validate
//
// Returns:
//   - bool: true if the name is valid, false otherwise
func IsValidGitLabName(name string) bool {
	return isValidGitLabNameCharacters(name) && !isInvalidGitLabRepositoryName(name)
}

// isValidGitLabNameCharacters checks if the given name contains only
// characters that are allowed in GitLab repository names.
//
// Parameters:
//   - name: The repository name to check
//
// Returns:
//   - bool: true if the name contains only valid characters, false otherwise
func isValidGitLabNameCharacters(name string) bool {
	return gitlabNameRegex.MatchString(name)
}

// isInvalidGitLabRepositoryName checks if the given name is in the list of
// invalid GitLab repository names.
//
// The check is case-insensitive and includes length validation.
//
// Parameters:
//   - name: The repository name to check
//
// Returns:
//   - bool: true if the name is invalid, false otherwise
func isInvalidGitLabRepositoryName(name string) bool {
	// Check maximum length (GitLab limit)
	if len(name) > 256 {
		return true
	}

	// Check minimum length
	if len(name) == 0 {
		return true
	}

	// Check against reserved names (case-insensitive)
	return invalidGitLabNames[strings.ToLower(name)]
}

// ValidateAndCleanGitLabName validates a GitLab name and returns a cleaned version if needed.
// This function provides more detailed validation feedback and cleaning options.
//
// Parameters:
//   - name: The repository name to validate and clean.
//
// Returns:
//   - cleanName: A cleaned version of the name that should be valid.
//   - isValid: Whether the original name was valid.
//   - issues: A slice of issues found with the original name.
func ValidateAndCleanGitLabName(name string) (cleanName string, isValid bool, issues []string) {
	cleanName = name
	isValid = true
	issues = []string{}

	// Check for empty name
	if name == "" {
		issues = append(issues, "name cannot be empty")
		isValid = false
		cleanName = constants.DefaultProjectName

		return cleanName, isValid, issues
	}

	// Check for invalid reserved names
	if invalidGitLabNames[strings.ToLower(name)] {
		issues = append(issues, "name is reserved by GitLab")
		isValid = false
		cleanName = constants.DefaultProjectName
	}

	// Check length
	if len(name) > 256 {
		issues = append(issues, "name exceeds maximum length of 256 characters")
		isValid = false
		cleanName = name[:256]
	}

	// Check for invalid characters
	if !gitlabNameRegex.MatchString(name) {
		issues = append(issues, "name contains invalid characters")
		isValid = false
		cleanName = cleanInvalidGitLabCharacters(name)
	}

	return cleanName, isValid, issues
}

// cleanInvalidGitLabCharacters removes or replaces invalid characters in a GitLab repository name.
func cleanInvalidGitLabCharacters(name string) string {
	result := ""

	// Ensure the first character is valid (letter, number, or underscore)
	if len(name) > 0 {
		firstChar := rune(name[0])
		if (firstChar < 'A' || firstChar > 'Z') &&
			(firstChar < 'a' || firstChar > 'z') &&
			(firstChar < '0' || firstChar > '9') &&
			firstChar != '_' {
			result = "p" // Start with 'p' for 'project'
		} else {
			result = string(firstChar)
		}
	}

	// Process remaining characters
	for i := 1; i < len(name); i++ {
		char := rune(name[i])
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' || char == '+' || char == ' ' {
			result += string(char)
		} else {
			result += "-"
		}
	}

	// Ensure it's not empty after cleaning
	if result == "" {
		result = "project"
	}

	return result
}

// SuggestAlternativeGitLabName suggests an alternative name if the given name is invalid.
func SuggestAlternativeGitLabName(name string) string {
	cleanName, isValid, _ := ValidateAndCleanGitLabName(name)
	if isValid {
		return name
	}

	return cleanName
}

// GetGitLabValidationRules returns a description of GitLab repository naming rules.
func GetGitLabValidationRules() []string {
	return []string{
		"Must start with a letter, number, or underscore",
		"May contain letters, numbers, underscores, dots, plus signs, hyphens, or spaces",
		"Cannot be one of the 100+ reserved names used by GitLab",
		"Must not exceed 256 characters in length",
		"Cannot be empty",
		"Reserved names include: admin, api, badges, builds, commits, groups, projects, etc.",
	}
}

// GetReservedNamesCount returns the number of reserved GitLab names.
func GetReservedNamesCount() int {
	return len(invalidGitLabNames)
}

// IsReservedGitLabName checks if a specific name is reserved.
func IsReservedGitLabName(name string) bool {
	return invalidGitLabNames[strings.ToLower(name)]
}
