// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"regexp"
	"strings"
)

// Regular expression for valid GitLab repository name characters.
// It allows names that start with a letter, number, or underscore,
// followed by any number of letters, numbers, underscores, dots, plus signs, hyphens, or spaces.
var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.+\- ]*$`)

// invalidNames is a set of repository names that are not allowed in GitLab.
// These names are reserved for GitLab's internal use or have special meanings.
var invalidNames = map[string]bool{
	"-":                    true,
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
}

// IsValidGitLabName checks if the given name is a valid GitLab repository name.
// It returns true if the name:
//  1. Contains only valid characters (defined by nameRegex)
//  2. Is not in the list of invalid names (defined in invalidNames)
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
	return nameRegex.MatchString(name)
}

// isInvalidGitLabRepositoryName checks if the given name is in the list of
// invalid GitLab repository names.
//
// The check is case-insensitive.
//
// Parameters:
//   - name: The repository name to check
//
// Returns:
//   - bool: true if the name is invalid, false otherwise
func isInvalidGitLabRepositoryName(name string) bool {
	return invalidNames[strings.ToLower(name)]
}
