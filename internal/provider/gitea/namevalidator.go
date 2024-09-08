// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"regexp"
)

var (
	// validNameRegex is a compiled regex for valid Gitea repository name characters.
	// It allows alphanumeric characters, hyphens, and dots.
	validNameRegex = regexp.MustCompile(`^[A-Za-z0-9-\.]+$`)

	// invalidNames contains repository names that are not allowed.
	// These are typically reserved names or names with special meanings in file systems.
	invalidNames = map[string]bool{
		".":  true,
		"..": true,
	}

	// maxNameLength is the maximum allowed length for a Gitea repository name.
	// Gitea imposes this limit to prevent excessively long repository names.
	maxNameLength = 100
)

// IsValidGiteaRepositoryName checks if the given name is a valid Gitea repository name.
// It performs two checks:
// 1. Ensures the name is not in the list of invalid names.
// 2. Validates the characters and length of the name.
//
// Parameters:
//   - name: The repository name to validate.
//
// Returns:
//   - bool: true if the name is valid, false otherwise.
//
// Example usage:
//
//	if gitea.IsValidGiteaRepositoryName("my-repo") {
//	    fmt.Println("Valid repository name")
//	} else {
//	    fmt.Println("Invalid repository name")
//	}
func IsValidGiteaRepositoryName(name string) bool {
	return !invalidNames[name] && isValidGiteaRepositoryNameCharacters(name)
}

// isValidGiteaRepositoryNameCharacters checks if the repository name contains only valid characters
// and is within the allowed length.
//
// The function performs two checks:
// 1. Ensures the name length does not exceed maxNameLength.
// 2. Matches the name against the validNameRegex to check for valid characters.
//
// Parameters:
//   - name: The repository name to validate.
//
// Returns:
//   - bool: true if the name contains only valid characters and is within the length limit, false otherwise.
//
// Note: This is an internal function and should not be used outside of this package.
func isValidGiteaRepositoryNameCharacters(name string) bool {
	return len(name) <= maxNameLength && validNameRegex.MatchString(name)
}

// AddInvalidName allows adding custom invalid names to the invalidNames map.
// This can be useful for organizations that want to restrict certain repository names.
//
// Parameters:
//   - name: The name to be added to the list of invalid names.
//
// Example usage:
//
//	gitea.AddInvalidName("restricted-repo-name")
func AddInvalidName(name string) {
	invalidNames[name] = true
}

// SetMaxNameLength allows changing the maximum allowed length for repository names.
// This can be useful if your Gitea instance has different length restrictions.
//
// Parameters:
//   - length: The new maximum length for repository names.
//
// Example usage:
//
//	gitea.SetMaxNameLength(150)
func SetMaxNameLength(length int) {
	if length > 0 {
		maxNameLength = length
	}
}
