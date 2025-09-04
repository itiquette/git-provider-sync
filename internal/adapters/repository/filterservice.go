// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package repository

import (
	"context"
	"fmt"
	"slices"
	"time"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FilterService filters repositories based on configured criteria.
type FilterService struct {
	logger ports.Logger
}

// NewFilterService creates a new repository filter service.
func NewFilterService(logger ports.Logger) *FilterService {
	return &FilterService{
		logger: logger,
	}
}

// FilterRepositories filters repositories based on various criteria.
func (fs *FilterService) FilterRepositories(
	ctx context.Context,
	repositories []entities.Repository,
	options FilterOptions,
) ([]entities.Repository, error) {
	fs.logger.Debug(ctx, "Starting repository filtering", map[string]any{
		"total_repositories": len(repositories),
		"include_count":      len(options.IncludeList),
		"exclude_count":      len(options.ExcludeList),
		"active_from_limit":  options.ActiveFromLimit,
	})

	filtered := repositories

	// Apply time-based filtering
	if options.ActiveFromLimit != "" {
		timeFiltered, err := fs.filterByTimeInterval(ctx, filtered, options.ActiveFromLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to apply time-based filtering: %w", err)
		}

		filtered = timeFiltered
	}

	// Apply include/exclude filtering
	filtered = fs.filterByIncludeExclude(ctx, filtered, options.IncludeList, options.ExcludeList)

	fs.logger.Info(ctx, "Repository filtering completed", map[string]any{
		"original_count": len(repositories),
		"filtered_count": len(filtered),
		"removed_count":  len(repositories) - len(filtered),
	})

	return filtered, nil
}

// FilterOptions contains options for repository filtering.
type FilterOptions struct {
	IncludeList     []string // Repositories to include
	ExcludeList     []string // Repositories to exclude
	ActiveFromLimit string   // Time duration for activity filtering (e.g., "30d", "1h")
	IncludeForks    bool     // Whether to include forked repositories
	IncludeArchived bool     // Whether to include archived repositories
}

// GetFilterStatistics returns statistics about the filtering operation.
func (fs *FilterService) GetFilterStatistics(original, filtered []entities.Repository) map[string]any {
	stats := map[string]any{
		"original_count": len(original),
		"filtered_count": len(filtered),
		"removed_count":  len(original) - len(filtered),
	}

	if len(original) > 0 {
		stats["retention_percentage"] = float64(len(filtered)) / float64(len(original)) * 100
	}

	return stats
}

// ValidateFilterOptions validates the filter options.
func (fs *FilterService) ValidateFilterOptions(options FilterOptions) error {
	// Validate ActiveFromLimit if provided
	if options.ActiveFromLimit != "" {
		_, err := time.ParseDuration(options.ActiveFromLimit)
		if err != nil {
			return fmt.Errorf("invalid ActiveFromLimit duration '%s': %w", options.ActiveFromLimit, err)
		}
	}

	// Check for conflicting include/exclude lists
	if len(options.IncludeList) > 0 && len(options.ExcludeList) > 0 {
		for _, include := range options.IncludeList {
			if slices.Contains(options.ExcludeList, include) {
				return fmt.Errorf("%w: '%s'", domain.ErrRepositoryInBothLists, include)
			}
		}
	}

	return nil
}

// FilterByTimeInterval filters repositories based on their last activity time.
func (fs *FilterService) filterByTimeInterval(
	ctx context.Context,
	repositories []entities.Repository,
	activeFromLimit string,
) ([]entities.Repository, error) {
	fs.logger.Debug(ctx, "Applying time-based filtering", map[string]any{
		"active_from_limit": activeFromLimit,
	})

	// Parse the duration string
	parsedDuration, err := time.ParseDuration(activeFromLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time duration '%s': %w", activeFromLimit, err)
	}

	// Calculate the threshold time (negative duration means "go back in time")
	threshold := time.Now().Add(-parsedDuration)

	fs.logger.Debug(ctx, "Time filtering threshold calculated", map[string]any{
		"threshold":     threshold.Format(time.RFC3339),
		"duration_back": parsedDuration.String(),
	})

	// Filter repositories
	var filtered []entities.Repository

	for _, repo := range repositories {
		if fs.isInTimeInterval(repo, threshold) {
			filtered = append(filtered, repo)
		} else {
			fs.logger.Debug(ctx, "Repository filtered out by time", map[string]any{
				"repository":    repo.Name(),
				"last_activity": repo.LastActivityAt().Format(time.RFC3339),
				"threshold":     threshold.Format(time.RFC3339),
			})
		}
	}

	return filtered, nil
}

// IsInTimeInterval checks if a repository's last activity is within the specified interval
//
//	exact IsInInterval logic .
func (fs *FilterService) isInTimeInterval(repo entities.Repository, threshold time.Time) bool {
	lastActivity := repo.LastActivityAt()

	// If last activity is zero, consider it within the interval (like main branch)
	if lastActivity.IsZero() {
		return true
	}

	// Check if last activity is after the threshold (more recent than the limit)
	return lastActivity.After(threshold) || lastActivity.Equal(threshold)
}

// FilterByIncludeExclude filters repositories based on inclusion and exclusion lists.
func (fs *FilterService) filterByIncludeExclude(
	ctx context.Context,
	repositories []entities.Repository,
	includeList, excludeList []string,
) []entities.Repository {
	fs.logger.Debug(ctx, "Applying include/exclude filtering", map[string]any{
		"include_count": len(includeList),
		"exclude_count": len(excludeList),
	})

	// Use the exact logic  shouldIncludeRepo
	return slices.DeleteFunc(repositories, func(repo entities.Repository) bool {
		return !fs.shouldIncludeRepository(repo.Name(), includeList, excludeList)
	})
}

// ShouldIncludeRepository determines if a repository should be included based on the inclusion and exclusion lists
//
//	exact shouldIncludeRepo logic .
func (fs *FilterService) shouldIncludeRepository(repoName string, includeList, excludeList []string) bool {
	switch {
	case len(includeList) == 0 && len(excludeList) == 0:
		// If both lists are empty, include all repositories
		return true
	case len(includeList) > 0:
		// If there's an inclusion list, only include repositories in that list
		return slices.Contains(includeList, repoName)
	default:
		// If there's only an exclusion list, include repositories not in that list
		return !slices.Contains(excludeList, repoName)
	}
}
