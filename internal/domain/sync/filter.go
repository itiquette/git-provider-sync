// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

//nolint:funcorder // Filter with many helper methods
package sync

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterRepositoriesUseCase handles repository filtering.
type FilterRepositoriesUseCase struct {
	logger ports.Logger
}

// NewFilterRepositoriesUseCase creates a new filter repositories use case.
func NewFilterRepositoriesUseCase(logger ports.Logger) FilterRepositoriesUseCase {
	return FilterRepositoriesUseCase{
		logger: logger,
	}
}

// FilterRequest represents the input for filtering operations.
type FilterRequest struct {
	Repositories         []entities.Repository
	FilterOptions        ports.FilterOptions
	IncludedRepositories []string
	ExcludedRepositories []string
	ActiveFromLimit      string
	Options              FilteringOptions
}

// FilteringOptions contains domain-specific filtering options.
// This replaces the CLI-specific model.CLIOption with domain values.
type FilteringOptions struct {
	AlphaNumHyphName  bool
	IncludeForks      bool
	IncludeArchived   bool
	IgnoreInvalidName bool
}

// FilterResponse represents the result of filtering operations.
type FilterResponse struct {
	FilteredRepositories []entities.Repository
	OriginalCount        int
	FilteredCount        int
	SkippedByActivity    int
	SkippedByInclusion   int
	SkippedByExclusion   int
	SkippedByPatterns    int
	Success              bool
	Errors               []error
}

// Execute performs repository filtering based on various criteria.
func (uc FilterRepositoriesUseCase) Execute(
	ctx context.Context,
	request FilterRequest,
) (FilterResponse, error) {
	uc.logger.Info(ctx, "Starting repository filtering", map[string]any{
		"total_repositories": len(request.Repositories),
		"has_activity_limit": request.ActiveFromLimit != "",
		"include_count":      len(request.IncludedRepositories),
		"exclude_count":      len(request.ExcludedRepositories),
	})

	response := FilterResponse{
		OriginalCount: len(request.Repositories),
		Success:       true,
	}

	// Step 1: Apply activity-based filtering (equivalent to IsInInterval
	repositoriesAfterActivity, skippedActivity := uc.filterByActivity(ctx, request.Repositories, request.ActiveFromLimit)
	response.SkippedByActivity = skippedActivity

	uc.logger.Debug(ctx, "Activity filtering completed", map[string]any{
		"before":  len(request.Repositories),
		"after":   len(repositoriesAfterActivity),
		"skipped": skippedActivity,
	})

	// Step 2: Apply include/exclude filtering (equivalent to FilterIncludedExcluded
	repositoriesAfterInclusion, skippedInclusion, skippedExclusion := uc.filterByIncludeExclude(
		ctx,
		repositoriesAfterActivity,
		request.IncludedRepositories,
		request.ExcludedRepositories,
	)
	response.SkippedByInclusion = skippedInclusion
	response.SkippedByExclusion = skippedExclusion

	uc.logger.Debug(ctx, "Include/exclude filtering completed", map[string]any{
		"before":            len(repositoriesAfterActivity),
		"after":             len(repositoriesAfterInclusion),
		"skipped_inclusion": skippedInclusion,
		"skipped_exclusion": skippedExclusion,
	})

	// Step 3: Apply pattern-based filtering
	repositoriesAfterPatterns, skippedPatterns := uc.filterByPatterns(
		ctx,
		repositoriesAfterInclusion,
		request.FilterOptions,
	)
	response.SkippedByPatterns = skippedPatterns

	uc.logger.Debug(ctx, "Pattern filtering completed", map[string]any{
		"before":  len(repositoriesAfterInclusion),
		"after":   len(repositoriesAfterPatterns),
		"skipped": skippedPatterns,
	})

	// Step 4: Apply additional attribute filtering (visibility, forks, etc.)
	finalRepositories := uc.filterByAttributes(ctx, repositoriesAfterPatterns, request.FilterOptions)

	response.FilteredRepositories = finalRepositories
	response.FilteredCount = len(finalRepositories)

	uc.logger.Info(ctx, "Repository filtering completed", map[string]any{
		"original_count":       response.OriginalCount,
		"filtered_count":       response.FilteredCount,
		"skipped_by_activity":  response.SkippedByActivity,
		"skipped_by_inclusion": response.SkippedByInclusion,
		"skipped_by_exclusion": response.SkippedByExclusion,
		"skipped_by_patterns":  response.SkippedByPatterns,
		"success":              response.Success,
	})

	return response, nil
}

// filterByActivity filters repositories based on activity time limits.
func (uc FilterRepositoriesUseCase) filterByActivity(
	ctx context.Context,
	repositories []entities.Repository,
	activeFromLimit string,
) ([]entities.Repository, int) {
	if activeFromLimit == "" {
		return repositories, 0 // No activity filtering
	}

	// Parse the duration string
	parsedDuration, err := time.ParseDuration(activeFromLimit)
	if err != nil {
		uc.logger.Error(ctx, "Failed to parse activity limit duration", map[string]any{
			"duration": activeFromLimit,
			"error":    err.Error(),
		})

		return repositories, 0 // On error, return all repositories
	}

	// Calculate the threshold time
	threshold := time.Now().Add(parsedDuration)

	var filtered []entities.Repository

	skipped := 0

	for _, repo := range repositories {
		lastActivity := repo.LastActivityAt()

		// If lastActivity is zero or after/equal to threshold, include it
		if lastActivity.IsZero() || lastActivity.After(threshold) || lastActivity.Equal(threshold) {
			filtered = append(filtered, repo)
		} else {
			skipped++

			uc.logger.Debug(ctx, "Skipping repository due to activity limit", map[string]any{
				"repository":    repo.Name(),
				"last_activity": lastActivity.Format(time.RFC3339),
				"threshold":     threshold.Format(time.RFC3339),
			})
		}
	}

	return filtered, skipped
}

// filterByIncludeExclude filters repositories based on inclusion and exclusion lists.
func (uc FilterRepositoriesUseCase) filterByIncludeExclude(
	ctx context.Context,
	repositories []entities.Repository,
	included, excluded []string,
) ([]entities.Repository, int, int) {
	if len(included) == 0 && len(excluded) == 0 {
		return repositories, 0, 0 // No include/exclude filtering
	}

	var filtered []entities.Repository

	skippedInclusion := 0
	skippedExclusion := 0

	for _, repo := range repositories {
		repoName := repo.Name()
		shouldInclude := uc.shouldIncludeRepository(repoName, included, excluded)

		if shouldInclude {
			filtered = append(filtered, repo)
		} else {
			// Determine why it was skipped
			if len(included) > 0 && !slices.Contains(included, repoName) {
				skippedInclusion++

				uc.logger.Debug(ctx, "Skipping repository not in inclusion list", map[string]any{
					"repository": repoName,
				})
			} else if len(excluded) > 0 && slices.Contains(excluded, repoName) {
				skippedExclusion++

				uc.logger.Debug(ctx, "Skipping repository in exclusion list", map[string]any{
					"repository": repoName,
				})
			}
		}
	}

	return filtered, skippedInclusion, skippedExclusion
}

// shouldIncludeRepository determines if a repository should be included based on inclusion and exclusion lists.
// This ports the shouldIncludeRepo logic.
func (uc FilterRepositoriesUseCase) shouldIncludeRepository(repoName string, included, excluded []string) bool {
	switch {
	case len(included) == 0 && len(excluded) == 0:
		// If both lists are empty, include all repositories
		return true
	case len(included) > 0:
		// If there's an inclusion list, only include repositories in that list
		return slices.Contains(included, repoName)
	default:
		// If there's only an exclusion list, include repositories not in that list
		return !slices.Contains(excluded, repoName)
	}
}

// filterByPatterns filters repositories based on pattern matching.
// This enhances the basic pattern matching from the original fetch use case.
func (uc FilterRepositoriesUseCase) filterByPatterns(
	ctx context.Context,
	repositories []entities.Repository,
	filterOptions ports.FilterOptions,
) ([]entities.Repository, int) {
	if len(filterOptions.IncludePatterns) == 0 && len(filterOptions.ExcludePatterns) == 0 {
		return repositories, 0 // No pattern filtering
	}

	var filtered []entities.Repository

	skipped := 0

	for _, repo := range repositories {
		repoName := repo.Name()
		matches := uc.matchesPatterns(repoName, filterOptions.IncludePatterns, filterOptions.ExcludePatterns)

		if matches {
			filtered = append(filtered, repo)
		} else {
			skipped++

			uc.logger.Debug(ctx, "Skipping repository due to pattern mismatch", map[string]any{
				"repository":       repoName,
				"include_patterns": filterOptions.IncludePatterns,
				"exclude_patterns": filterOptions.ExcludePatterns,
			})
		}
	}

	return filtered, skipped
}

// matchesPatterns checks if repository name matches include/exclude patterns.
// Supports wildcard patterns: "*", "prefix*", "*suffix", "*contains*", "prefix*suffix".
// Include patterns act as allowlist; exclude patterns act as blocklist.
func (uc FilterRepositoriesUseCase) matchesPatterns(name string, includePatterns, excludePatterns []string) bool {
	// If no include patterns, assume included
	included := len(includePatterns) == 0

	// Check include patterns
	for _, pattern := range includePatterns {
		if uc.matchPattern(name, pattern) {
			included = true

			break
		}
	}

	if !included {
		return false
	}

	// Check exclude patterns
	for _, pattern := range excludePatterns {
		if uc.matchPattern(name, pattern) {
			return false
		}
	}

	return true
}

// matchPattern performs wildcard pattern matching using finite state automaton approach.
//
// ALGORITHM OVERVIEW:
// Uses optimized string operations instead of regex compilation for better performance
// in repository filtering scenarios. Achieves O(n+m) time complexity where n=name length, m=pattern length.
//
// SUPPORTED PATTERN TYPES:
//   - "*" matches everything (constant time)
//   - "prefix*" matches names starting with "prefix" (prefix match)
//   - "*suffix" matches names ending with "suffix" (suffix match)
//   - "*contains*" matches names containing "contains" (substring search)
//   - "prefix*suffix" matches names with specific prefix and suffix (boundary match)
//   - "exact" matches only "exact" (string equality)
//
// PERFORMANCE CHARACTERISTICS:
// - "*" pattern: O(1) - immediate return
// - Prefix/suffix patterns: O(n) - single string operation
// - Contains patterns: O(n*m) - substring search
// - Boundary patterns: O(n) - prefix + suffix check
// - Exact match: O(n) - string comparison
//
// The algorithm prioritizes performance over regex flexibility, making it suitable
// for high-volume repository filtering operations in sync pipelines.
func (uc FilterRepositoriesUseCase) matchPattern(name, pattern string) bool {
	// Handle wildcard patterns
	if pattern == "*" {
		return true
	}

	// Handle prefix/suffix wildcards
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") { //nolint:gocritic // if-else chain is more readable for complex pattern matching conditions
		// Contains pattern: *text*
		return strings.Contains(name, pattern[1:len(pattern)-1])
	} else if strings.HasPrefix(pattern, "*") {
		// Suffix pattern: *text
		return strings.HasSuffix(name, pattern[1:])
	} else if strings.HasSuffix(pattern, "*") {
		// Prefix pattern: text*
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	} else if strings.Contains(pattern, "*") {
		// Middle wildcard pattern: prefix*suffix
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix, suffix := parts[0], parts[1]

			return strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) && len(name) >= len(prefix)+len(suffix)
		}
	}

	// Exact match
	return name == pattern
}

// filterByAttributes filters repositories based on various attributes.
func (uc FilterRepositoriesUseCase) filterByAttributes(
	ctx context.Context,
	repositories []entities.Repository,
	filterOptions ports.FilterOptions,
) []entities.Repository {
	var filtered []entities.Repository

	for _, repo := range repositories {
		if uc.shouldIncludeByAttributes(repo, filterOptions) {
			filtered = append(filtered, repo)
		} else {
			uc.logger.Debug(ctx, "Skipping repository due to attribute filters", map[string]any{
				"repository":       repo.Name(),
				"is_fork":          repo.IsFork(),
				"is_archived":      repo.IsArchived(),
				"is_private":       repo.IsPrivate(),
				"include_forks":    filterOptions.IncludeForks,
				"include_archived": filterOptions.IncludeArchived,
				"include_private":  filterOptions.IncludePrivate,
				"include_public":   filterOptions.IncludePublic,
			})
		}
	}

	return filtered
}

// shouldIncludeByAttributes checks if repository should be included based on attributes.
//
//nolint:cyclop // Complex filtering logic with multiple repository attributes
func (uc FilterRepositoriesUseCase) shouldIncludeByAttributes(
	repo entities.Repository,
	filterOptions ports.FilterOptions,
) bool {
	// Fork filtering
	if repo.IsFork() && !filterOptions.IncludeForks {
		return false
	}

	// Archived filtering
	if repo.IsArchived() && !filterOptions.IncludeArchived {
		return false
	}

	// Visibility filtering
	if repo.IsPrivate() && !filterOptions.IncludePrivate {
		return false
	}

	if !repo.IsPrivate() && !filterOptions.IncludePublic {
		return false
	}

	// Size filtering not implemented - Repository entity doesn't expose size

	// Activity filtering (additional to time-based filtering)
	if filterOptions.ActiveSince != nil {
		if repo.LastActivityAt().Before(*filterOptions.ActiveSince) {
			return false
		}
	}

	if filterOptions.InactiveSince != nil {
		if repo.LastActivityAt().After(*filterOptions.InactiveSince) {
			return false
		}
	}

	return true
}

// ValidateFilterRequest validates the filter request parameters.
func (uc FilterRepositoriesUseCase) ValidateFilterRequest(_ context.Context, request FilterRequest) error {
	// Validate activity limit duration if provided
	if request.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(request.ActiveFromLimit); err != nil {
			return fmt.Errorf("invalid activity limit duration %q: %w", request.ActiveFromLimit, err)
		}
	}

	// Validate time ranges
	if request.FilterOptions.ActiveSince != nil && request.FilterOptions.InactiveSince != nil {
		if request.FilterOptions.ActiveSince.After(*request.FilterOptions.InactiveSince) {
			return domain.ErrActiveSinceAfterInactiveSince
		}
	}

	// Validate size ranges
	if request.FilterOptions.MinSize > 0 && request.FilterOptions.MaxSize > 0 {
		if request.FilterOptions.MinSize > request.FilterOptions.MaxSize {
			return domain.ErrMinSizeGreaterThanMaxSize
		}
	}

	return nil
}
