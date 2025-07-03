// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterRepositoriesUseCase handles advanced repository filtering.
// This ports the filtering functionality from main branch to hexagonal architecture.
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
// This implements the filtering logic from main branch: activity filters, include/exclude lists, patterns.
func (uc FilterRepositoriesUseCase) Execute(
	ctx context.Context,
	request FilterRequest,
) (FilterResponse, error) {
	uc.logger.Info(ctx, "Starting repository filtering", map[string]interface{}{
		"total_repositories": len(request.Repositories),
		"has_activity_limit": request.ActiveFromLimit != "",
		"include_count":      len(request.IncludedRepositories),
		"exclude_count":      len(request.ExcludedRepositories),
	})

	response := FilterResponse{
		OriginalCount: len(request.Repositories),
		Success:       true,
	}

	// Step 1: Apply activity-based filtering (equivalent to IsInInterval from main branch)
	repositoriesAfterActivity, skippedActivity := uc.filterByActivity(ctx, request.Repositories, request.ActiveFromLimit)
	response.SkippedByActivity = skippedActivity

	uc.logger.Debug(ctx, "Activity filtering completed", map[string]interface{}{
		"before":  len(request.Repositories),
		"after":   len(repositoriesAfterActivity),
		"skipped": skippedActivity,
	})

	// Step 2: Apply include/exclude filtering (equivalent to FilterIncludedExcluded from main branch)
	repositoriesAfterInclusion, skippedInclusion, skippedExclusion := uc.filterByIncludeExclude(
		ctx,
		repositoriesAfterActivity,
		request.IncludedRepositories,
		request.ExcludedRepositories,
	)
	response.SkippedByInclusion = skippedInclusion
	response.SkippedByExclusion = skippedExclusion

	uc.logger.Debug(ctx, "Include/exclude filtering completed", map[string]interface{}{
		"before":            len(repositoriesAfterActivity),
		"after":             len(repositoriesAfterInclusion),
		"skipped_inclusion": skippedInclusion,
		"skipped_exclusion": skippedExclusion,
	})

	// Step 3: Apply pattern-based filtering (enhanced from basic pattern matching)
	repositoriesAfterPatterns, skippedPatterns := uc.filterByPatterns(
		ctx,
		repositoriesAfterInclusion,
		request.FilterOptions,
	)
	response.SkippedByPatterns = skippedPatterns

	uc.logger.Debug(ctx, "Pattern filtering completed", map[string]interface{}{
		"before":  len(repositoriesAfterInclusion),
		"after":   len(repositoriesAfterPatterns),
		"skipped": skippedPatterns,
	})

	// Step 4: Apply additional attribute filtering (visibility, forks, etc.)
	finalRepositories := uc.filterByAttributes(ctx, repositoriesAfterPatterns, request.FilterOptions)

	response.FilteredRepositories = finalRepositories
	response.FilteredCount = len(finalRepositories)

	uc.logger.Info(ctx, "Repository filtering completed", map[string]interface{}{
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
// This ports the IsInInterval functionality from main branch.
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
		uc.logger.Error(ctx, "Failed to parse activity limit duration", map[string]interface{}{
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

			uc.logger.Debug(ctx, "Skipping repository due to activity limit", map[string]interface{}{
				"repository":    repo.Name(),
				"last_activity": lastActivity.Format(time.RFC3339),
				"threshold":     threshold.Format(time.RFC3339),
			})
		}
	}

	return filtered, skipped
}

// filterByIncludeExclude filters repositories based on inclusion and exclusion lists.
// This ports the FilterIncludedExcluded functionality from main branch.
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

				uc.logger.Debug(ctx, "Skipping repository not in inclusion list", map[string]interface{}{
					"repository": repoName,
				})
			} else if len(excluded) > 0 && slices.Contains(excluded, repoName) {
				skippedExclusion++

				uc.logger.Debug(ctx, "Skipping repository in exclusion list", map[string]interface{}{
					"repository": repoName,
				})
			}
		}
	}

	return filtered, skippedInclusion, skippedExclusion
}

// shouldIncludeRepository determines if a repository should be included based on inclusion and exclusion lists.
// This ports the shouldIncludeRepo logic from main branch.
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

			uc.logger.Debug(ctx, "Skipping repository due to pattern mismatch", map[string]interface{}{
				"repository":       repoName,
				"include_patterns": filterOptions.IncludePatterns,
				"exclude_patterns": filterOptions.ExcludePatterns,
			})
		}
	}

	return filtered, skipped
}

// matchesPatterns checks if repository name matches include/exclude patterns.
// This enhances the pattern matching with support for wildcards and regex-like patterns.
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

// matchPattern performs pattern matching with support for wildcards.
// This enhances the basic pattern matching from the original implementation.
func (uc FilterRepositoriesUseCase) matchPattern(name, pattern string) bool {
	// Handle wildcard patterns
	if pattern == "*" {
		return true
	}

	// Handle prefix/suffix wildcards
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// Contains pattern: *text*
		return strings.Contains(name, pattern[1:len(pattern)-1])
	} else if strings.HasPrefix(pattern, "*") {
		// Suffix pattern: *text
		return strings.HasSuffix(name, pattern[1:])
	} else if strings.HasSuffix(pattern, "*") {
		// Prefix pattern: text*
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}

	// Exact match
	return name == pattern
}

// filterByAttributes filters repositories based on various attributes.
// This provides additional filtering beyond the main branch implementation.
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
			uc.logger.Debug(ctx, "Skipping repository due to attribute filters", map[string]interface{}{
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

	// Size filtering (if repository has size information)
	// Note: Repository entity would need size field for this to work
	// TODO: Add size filtering when Repository entity has size information

	// Activity filtering (additional to time-based filtering)
	// Note: Could add more sophisticated activity filters here
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
func (uc FilterRepositoriesUseCase) ValidateFilterRequest(ctx context.Context, request FilterRequest) error {
	// Validate activity limit duration if provided
	if request.ActiveFromLimit != "" {
		if _, err := time.ParseDuration(request.ActiveFromLimit); err != nil {
			return fmt.Errorf("invalid activity limit duration %q: %w", request.ActiveFromLimit, err)
		}
	}

	// Validate time ranges
	if request.FilterOptions.ActiveSince != nil && request.FilterOptions.InactiveSince != nil {
		if request.FilterOptions.ActiveSince.After(*request.FilterOptions.InactiveSince) {
			return errors.New("active since time cannot be after inactive since time")
		}
	}

	// Validate size ranges
	if request.FilterOptions.MinSize > 0 && request.FilterOptions.MaxSize > 0 {
		if request.FilterOptions.MinSize > request.FilterOptions.MaxSize {
			return errors.New("minimum size cannot be greater than maximum size")
		}
	}

	return nil
}
