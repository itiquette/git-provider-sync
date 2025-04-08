// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"regexp"
)

var (
	// validNameRegex defines the allowed characters in a GitHub repository name.
	// It allows alphanumeric characters, hyphens, and underscores.
	validNameRegex = regexp.MustCompile(`^[A-Za-z0-9-_]+$`)

	// invalidNames is a map of repository names that are not allowed by GitHub.
	// Currently, it includes "." and ".." which are reserved names.
	invalidNames = map[string]bool{
		".":  true,
		"..": true,
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
