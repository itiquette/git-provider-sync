// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"itiquette/git-provider-sync/internal/domain/entities"
)

// FilterFunc is a pure function that determines if a repository should be included.
type FilterFunc func(entities.Repository) bool

// ComposeFilters combines multiple filter predicates with AND logic.
// Returns true only if all filters return true.
func ComposeFilters(filters ...FilterFunc) FilterFunc {
	return func(repo entities.Repository) bool {
		for _, f := range filters {
			if !f(repo) {
				return false
			}
		}

		return true
	}
}

// CreateForkFilter returns a filter for fork inclusion.
func CreateForkFilter(includeForks bool) FilterFunc {
	return func(repo entities.Repository) bool {
		return includeForks || !repo.IsFork()
	}
}

// CreateArchivedFilter returns a filter for archived repositories.
func CreateArchivedFilter(includeArchived bool) FilterFunc {
	return func(repo entities.Repository) bool {
		return includeArchived || !repo.IsArchived()
	}
}

// CreateVisibilityFilter returns a filter for repository visibility.
func CreateVisibilityFilter(includePrivate bool) FilterFunc {
	return func(repo entities.Repository) bool {
		return includePrivate || repo.Visibility() == "public"
	}
}

// CreateNameMatchFilter returns a filter that matches repository names against patterns.
// Supports wildcard patterns with * character.
func CreateNameMatchFilter(patterns []string) FilterFunc {
	if len(patterns) == 0 {
		return func(entities.Repository) bool { return true }
	}

	return func(repo entities.Repository) bool {
		name := repo.Name()
		for _, pattern := range patterns {
			if matchPattern(name, pattern) {
				return true
			}
		}

		return false
	}
}

// ApplyFilters applies a composed filter to a list of repositories.
// Returns only repositories that pass the filter.
func ApplyFilters(repos []entities.Repository, filter FilterFunc) []entities.Repository {
	if filter == nil {
		return repos
	}

	result := make([]entities.Repository, 0, len(repos))
	for _, repo := range repos {
		if filter(repo) {
			result = append(result, repo)
		}
	}

	return result
}

// CountFilteredOut counts how many repositories would be filtered out.
func CountFilteredOut(repos []entities.Repository, filter FilterFunc) int {
	if filter == nil {
		return 0
	}

	count := 0

	for _, repo := range repos {
		if !filter(repo) {
			count++
		}
	}

	return count
}

// matchPattern is a pure helper function for pattern matching.
// Moved from filter.go to be reusable in pure context.
//
//nolint:cyclop // Pattern matching requires multiple branches
func matchPattern(name, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}

	// Simple wildcard matching
	switch {
	case pattern[0] == '*' && pattern[len(pattern)-1] == '*':
		// *substring*
		contains := pattern[1 : len(pattern)-1]

		return len(contains) > 0 && len(name) >= len(contains) &&
			stringContains(name, contains)

	case pattern[0] == '*':
		// *suffix
		suffix := pattern[1:]

		return len(suffix) > 0 && len(name) >= len(suffix) &&
			name[len(name)-len(suffix):] == suffix

	case pattern[len(pattern)-1] == '*':
		// prefix*
		prefix := pattern[:len(pattern)-1]

		return len(prefix) > 0 && len(name) >= len(prefix) &&
			name[:len(prefix)] == prefix

	default:
		// Exact match
		return name == pattern
	}
}

// stringContains is a pure helper for substring checking.
func stringContains(text, substr string) bool {
	if len(substr) == 0 {
		return true
	}

	if len(substr) > len(text) {
		return false
	}

	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
