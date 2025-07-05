// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"regexp"

	"itiquette/git-provider-sync/internal/domain/constants"
)

var (
	// validNameRegex defines the allowed characters in a GitHub repository name.
	// It allows alphanumeric characters, hyphens, and underscores.
	validNameRegex = regexp.MustCompile(`^[A-Za-z0-9-_]+$`)

	// invalidNames is a map of repository names that are not allowed by GitHub.
	// Currently, it includes "." and ".." which are reserved names.
	invalidNames = map[string]bool{
		".":       true,
		"..":      true,
		".git":    true,
		".github": true,
	}

	// maxNameLength is the maximum allowed length for a GitHub repository name.
	// GitHub imposes a limit of 100 characters for repository names.
	maxNameLength = 100
)

// IsValidGitHubRepositoryName checks if the given name is a valid GitHub repository name.
// It applies several rules based on GitHub's repository naming conventions:
//  1. The name must not be in the list of invalid names (e.g., "." or "..").
//  2. The name must only contain alphanumeric characters, hyphens, or underscores.
//  3. The name must not exceed the maximum allowed length (100 characters).
//
// Parameters:
//   - name: The repository name to validate.
//
// Returns:
//   - bool: true if the name is valid, false otherwise.
//
// Usage:
//
//	if github.IsValidGitHubRepositoryName("my-repo") {
//	    fmt.Println("Valid repository name")
//	} else {
//	    fmt.Println("Invalid repository name")
//	}
func IsValidGitHubRepositoryName(name string) bool {
	return !invalidNames[name] &&
		validNameRegex.MatchString(name) &&
		len(name) <= maxNameLength
}

// ValidateAndCleanName validates a name and returns a cleaned version if needed.
// This function provides more detailed validation feedback and cleaning options.
//
// Parameters:
//   - name: The repository name to validate and clean.
//
// Returns:
//   - cleanName: A cleaned version of the name that should be valid.
//   - isValid: Whether the original name was valid.
//   - issues: A slice of issues found with the original name.
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

// cleanInvalidCharacters removes or replaces invalid characters in a repository name.
func cleanInvalidCharacters(name string) string {
	result := ""

	for _, char := range name {
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' {
			result += string(char)
		} else if char == ' ' || char == '.' {
			result += "-"
		}
	}

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
