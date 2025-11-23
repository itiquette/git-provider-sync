// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"regexp"
	"strings"

	"itiquette/git-provider-sync/internal/domain/constants"
)

var (
	// ValidNameRegex defines the allowed characters in a GitHub repository name
	// allows alphanumeric characters, hyphens, and underscores.
	validNameRegex = regexp.MustCompile(`^[A-Za-z0-9-_]+$`)

	// InvalidNames is a map of repository names that are not allowed by GitHub
	// Currently, it includes "." and ".." which are reserved names.
	invalidNames = map[string]bool{ //nolint:gochecknoglobals // GitHub name validation constants
		".":       true,
		"..":      true,
		".git":    true,
		".github": true,
	}

	// MaxNameLength is the maximum allowed length for a GitHub repository name
	// GitHub imposes a limit of 100 characters for repository names.
	maxNameLength = 100 //nolint:gochecknoglobals // GitHub name validation constants
)

// IsValidGitHubRepositoryName validates GitHub repository names against platform constraints
// Enforces GitHub's naming rules: alphanumeric/hyphens/underscores only, max 100 chars, no reserved names.
func IsValidGitHubRepositoryName(name string) bool {
	return !invalidNames[name] &&
		validNameRegex.MatchString(name) &&
		len(name) <= maxNameLength
}

// ValidateAndCleanName validates a name and returns a cleaned version if needed.
func ValidateAndCleanName(name string) (string, bool, []string) {
	cleanName := name
	isValid := true
	issues := []string{}

	// Check for empty name
	if name == "" {
		issues = append(issues, "name cannot be empty")
		isValid = false
		cleanName = constants.DefaultRepositoryName

		return cleanName, isValid, issues
	}

	// Check for invalid names
	if invalidNames[name] {
		issues = append(issues, "name is reserved")
		isValid = false
		cleanName = constants.DefaultRepositoryName
	}

	// Check length
	if len(name) > maxNameLength {
		issues = append(issues, "name exceeds maximum length")
		isValid = false
		cleanName = name[:maxNameLength]
	}

	// Check for invalid characters
	if !validNameRegex.MatchString(name) {
		issues = append(issues, "name contains invalid characters")
		isValid = false
		cleanName = cleanInvalidCharacters(name)
	}

	return cleanName, isValid, issues
}

// CleanInvalidCharacters removes or replaces invalid characters in a repository name.
func cleanInvalidCharacters(name string) string {
	result := ""

	var resultSb82 strings.Builder

	for _, char := range name {
		if isValidChar(char) {
			resultSb82.WriteRune(char)
		} else if isReplaceable(char) {
			result += "-"
		}
	}

	result += resultSb82.String()

	// Ensure it's not empty after cleaning
	if result == "" {
		result = "repository"
	}

	return result
}

// SuggestAlternativeName suggests an alternative name if the given name is invalid.
func SuggestAlternativeName(name string) string {
	cleanName, isValid, _ := ValidateAndCleanName(name)
	if isValid {
		return name
	}

	return cleanName
}

// GetValidationRules returns a description of GitHub repository naming rules.
func GetValidationRules() []string {
	return []string{
		"Must contain only alphanumeric characters, hyphens, and underscores",
		"Cannot be one of the reserved names: ., .., .git, .github",
		"Must not exceed 100 characters in length",
		"Cannot be empty",
	}
}

// IsValidChar checks if a character is valid for GitHub repository names.
func isValidChar(char rune) bool {
	return (char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_'
}

// IsReplaceable checks if a character should be replaced with a hyphen.
func isReplaceable(char rune) bool {
	return char == ' ' || char == '.'
}
