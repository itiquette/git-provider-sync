// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// filtering.go - Repository filtering utilities restored from main branch using hexagonal architecture
package utilities

import (
	"context"
	"fmt"
	"slices"
	"time"

	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/model"
)

// RepositoryFilter provides sophisticated repository filtering capabilities.
// This restores the advanced filtering functionality from main branch.
type RepositoryFilter struct {
	logger ports.Logger
}

// NewRepositoryFilter creates a new repository filter.
func NewRepositoryFilter(logger ports.Logger) *RepositoryFilter {
	return &RepositoryFilter{
		logger: logger,
	}
}

// IsInInterval checks if the given updatedAt time is within the specified interval.
// This restores the main branch targetfilter.IsInInterval functionality.
func (rf *RepositoryFilter) IsInInterval(ctx context.Context, updatedAt time.Time) (bool, error) {
	rf.logger.Debug(ctx, "Checking if time is in interval", map[string]interface{}{
		"updatedAt": updatedAt.Format(time.RFC3339),
	})

	// If updatedAt is zero, consider it within the interval
	if updatedAt.IsZero() {
		return true, nil
	}

	cliOption := model.CLIOptions(ctx)
	// If no ActiveFromLimit is specified, consider all times within the interval
	if cliOption.ActiveFromLimit == "" {
		return true, nil
	}

	// Parse the duration string from CLI options
	parsedDuration, err := time.ParseDuration(cliOption.ActiveFromLimit)
	if err != nil {
		return false, fmt.Errorf("failed to parse time duration: %w", err)
	}

	// Calculate the threshold time
	then := time.Now().Add(parsedDuration)

	// Check if updatedAt is after or equal to the threshold
	result := updatedAt.After(then) || updatedAt.Equal(then)

	rf.logger.Debug(ctx, "Interval check completed", map[string]interface{}{
		"updatedAt":      updatedAt.Format(time.RFC3339),
		"threshold":      then.Format(time.RFC3339),
		"withinInterval": result,
	})

	return result, nil
}

// FilterIncludedExcluded filters repositories based on inclusion and exclusion lists.
// This restores the main branch targetfilter.FilterIncludedExcludedGen functionality.
func (rf *RepositoryFilter) FilterIncludedExcluded(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	rf.logger.Debug(ctx, "Filtering repositories by include/exclude lists", map[string]interface{}{
		"totalRepositories":    len(projectinfos),
		"includedRepositories": len(opt.IncludedRepositories),
		"excludedRepositories": len(opt.ExcludedRepositories),
	})

	included := opt.IncludedRepositories
	excluded := opt.ExcludedRepositories

	// Use slices.DeleteFunc to efficiently filter the projectinfos slice
	filtered := slices.DeleteFunc(projectinfos, func(m model.ProjectInfo) bool {
		return !rf.shouldIncludeRepo(m.OriginalName, included, excluded)
	})

	rf.logger.Info(ctx, "Repository filtering completed", map[string]interface{}{
		"originalCount": len(projectinfos),
		"filteredCount": len(filtered),
	})

	return filtered, nil
}

// shouldIncludeRepo determines if a repository should be included based on the inclusion and exclusion lists.
// This restores the main branch targetfilter.shouldIncludeRepo functionality.
func (rf *RepositoryFilter) shouldIncludeRepo(repoName string, included, excluded []string) bool {
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

// FilterByActivity filters repositories based on last activity time.
// This restores time-based filtering functionality from main branch.
func (rf *RepositoryFilter) FilterByActivity(ctx context.Context, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	rf.logger.Debug(ctx, "Filtering repositories by activity", map[string]interface{}{
		"totalRepositories": len(projectinfos),
	})

	cliOption := model.CLIOptions(ctx)
	if cliOption.ActiveFromLimit == "" {
		// No activity filtering requested
		return projectinfos, nil
	}

	var filtered []model.ProjectInfo

	var errors []error

	for _, projectInfo := range projectinfos {
		if projectInfo.LastActivityAt == nil {
			// If no activity time, exclude for safety
			continue
		}

		withinInterval, err := rf.IsInInterval(ctx, *projectInfo.LastActivityAt)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to check interval for %s: %w", projectInfo.OriginalName, err))
			continue
		}

		if withinInterval {
			filtered = append(filtered, projectInfo)
		}
	}

	if len(errors) > 0 {
		return filtered, fmt.Errorf("activity filtering encountered %d errors: %v", len(errors), errors[0])
	}

	rf.logger.Info(ctx, "Activity filtering completed", map[string]interface{}{
		"originalCount":   len(projectinfos),
		"filteredCount":   len(filtered),
		"activeFromLimit": cliOption.ActiveFromLimit,
	})

	return filtered, nil
}

// FilterByForks filters repositories based on fork inclusion preference.
// This restores fork filtering functionality from main branch.
func (rf *RepositoryFilter) FilterByForks(ctx context.Context, projectinfos []model.ProjectInfo, includeForks bool) []model.ProjectInfo {
	if includeForks {
		// Include all repositories
		return projectinfos
	}

	// Fork filtering is typically done at provider level during fetching
	// For now, return all repositories as fork info is not available in ProjectInfo
	filtered := projectinfos

	rf.logger.Debug(ctx, "Fork filtering completed", map[string]interface{}{
		"originalCount": len(projectinfos),
		"filteredCount": len(filtered),
		"includeForks":  includeForks,
	})

	return filtered
}

// ApplyAllFilters applies all configured filters to repository list.
// This combines multiple filtering operations from main branch.
func (rf *RepositoryFilter) ApplyAllFilters(ctx context.Context, projectinfos []model.ProjectInfo, opt model.ProviderOption) ([]model.ProjectInfo, error) {
	rf.logger.Info(ctx, "Applying all repository filters", map[string]interface{}{
		"initialCount": len(projectinfos),
	})

	// Step 1: Filter by include/exclude lists
	filtered, err := rf.FilterIncludedExcluded(ctx, opt, projectinfos)
	if err != nil {
		return nil, fmt.Errorf("include/exclude filtering failed: %w", err)
	}

	// Step 2: Filter by activity time
	filtered, err = rf.FilterByActivity(ctx, filtered)
	if err != nil {
		return nil, fmt.Errorf("activity filtering failed: %w", err)
	}

	// Step 3: Filter by forks
	filtered = rf.FilterByForks(ctx, filtered, opt.IncludeForks)

	rf.logger.Info(ctx, "All filters applied successfully", map[string]interface{}{
		"initialCount": len(projectinfos),
		"finalCount":   len(filtered),
	})

	return filtered, nil
}

// CreateFilterFunction returns a function that filters repositories based on criteria.
// This restores the main branch FilterIncludedExcludedGen pattern.
func (rf *RepositoryFilter) CreateFilterFunction() func(context.Context, model.ProviderOption, []model.ProjectInfo) ([]model.ProjectInfo, error) {
	return func(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
		return rf.ApplyAllFilters(ctx, projectinfos, opt)
	}
}

// ValidateFilterCriteria validates that filter criteria are valid.
func (rf *RepositoryFilter) ValidateFilterCriteria(ctx context.Context, opt model.ProviderOption) error {
	// Validate ActiveFromLimit duration if specified
	cliOption := model.CLIOptions(ctx)
	if cliOption.ActiveFromLimit != "" {
		_, err := time.ParseDuration(cliOption.ActiveFromLimit)
		if err != nil {
			return fmt.Errorf("invalid ActiveFromLimit duration '%s': %w", cliOption.ActiveFromLimit, err)
		}
	}

	// Validate include/exclude patterns
	for _, pattern := range opt.IncludedRepositories {
		if pattern == "" {
			return fmt.Errorf("empty include pattern not allowed")
		}
	}

	for _, pattern := range opt.ExcludedRepositories {
		if pattern == "" {
			return fmt.Errorf("empty exclude pattern not allowed")
		}
	}

	rf.logger.Debug(ctx, "Filter criteria validation passed", nil)

	return nil
}
