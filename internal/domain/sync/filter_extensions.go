// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/filter"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// EnhancedFilterUseCase filters repositories using composable predicates.
type EnhancedFilterUseCase struct {
	logger ports.Logger
}

// NewEnhancedFilterUseCase creates a new filter use case.
func NewEnhancedFilterUseCase(logger ports.Logger) EnhancedFilterUseCase {
	return EnhancedFilterUseCase{
		logger: logger,
	}
}

// BuildPredicate constructs a composite predicate from filter options.
func BuildPredicate(request FilterRequest) filter.Predicate {
	var predicates []filter.Predicate

	// Activity filter
	if request.ActiveFromLimit != "" {
		if duration, err := time.ParseDuration(request.ActiveFromLimit); err == nil {
			predicates = append(predicates, filter.ActiveWithin(duration))
		}
	}

	// Include/Exclude lists
	if len(request.IncludedRepositories) > 0 {
		predicates = append(predicates, filter.NameIn(request.IncludedRepositories))
	}

	if len(request.ExcludedRepositories) > 0 {
		predicates = append(predicates, filter.NameNotIn(request.ExcludedRepositories))
	}

	// Fork filter
	if !request.Options.IncludeForks {
		predicates = append(predicates, filter.NotFork)
	}

	// Archive filter
	if !request.Options.IncludeArchived {
		predicates = append(predicates, filter.NotArchived)
	}

	// Combine all predicates with AND logic
	if len(predicates) == 0 {
		// No filters, include everything
		return func(entities.Repository) bool { return true }
	}

	return filter.All(predicates...)
}

// ExecuteFunctional performs filtering using composable predicates.
func (uc EnhancedFilterUseCase) ExecuteFunctional(
	ctx context.Context,
	request FilterRequest,
) (FilterResponse, error) {
	uc.logger.Info(ctx, "Starting functional repository filtering", map[string]any{
		"total_repositories": len(request.Repositories),
	})

	// Build composite predicate (pure function)
	predicate := BuildPredicate(request)

	// Apply filter (pure function)
	filtered := filter.Filter(request.Repositories, predicate)

	// Calculate statistics (pure functions)
	response := FilterResponse{
		FilteredRepositories: filtered,
		OriginalCount:        len(request.Repositories),
		FilteredCount:        len(filtered),
		Success:              true,
	}

	// Calculate what was filtered out
	skipped := response.OriginalCount - response.FilteredCount

	uc.logger.Info(ctx, "Functional filtering completed", map[string]any{
		"original": response.OriginalCount,
		"filtered": response.FilteredCount,
		"skipped":  skipped,
	})

	return response, nil
}

// CreateActivityPredicate creates a pure predicate for activity filtering
// Extracted as a pure function for testing and reuse.
func CreateActivityPredicate(activeFromLimit string) (filter.Predicate, error) {
	if activeFromLimit == "" {
		return func(entities.Repository) bool { return true }, nil
	}

	duration, err := time.ParseDuration(activeFromLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	return filter.ActiveWithin(duration), nil
}

// CreateInclusionPredicate creates a predicate for inclusion/exclusion filtering.
func CreateInclusionPredicate(included, excluded []string) filter.Predicate {
	var predicates []filter.Predicate

	if len(included) > 0 {
		predicates = append(predicates, filter.NameIn(included))
	}

	if len(excluded) > 0 {
		predicates = append(predicates, filter.NameNotIn(excluded))
	}

	if len(predicates) == 0 {
		return func(entities.Repository) bool { return true }
	}

	return filter.All(predicates...)
}

// CreateOptionsPredicate creates a predicate from filtering options (fork, archive, etc).
func CreateOptionsPredicate(options FilteringOptions) filter.Predicate {
	var predicates []filter.Predicate

	if !options.IncludeForks {
		predicates = append(predicates, filter.NotFork)
	}

	if !options.IncludeArchived {
		predicates = append(predicates, filter.NotArchived)
	}

	if len(predicates) == 0 {
		return func(entities.Repository) bool { return true }
	}

	return filter.All(predicates...)
}

// ApplyPredicateWithStats applies a predicate and returns both filtered results and statistics.
func ApplyPredicateWithStats(
	repos []entities.Repository,
	predicate filter.Predicate,
) ([]entities.Repository, FilterStats) {
	filtered := filter.Filter(repos, predicate)

	stats := FilterStats{
		Original: len(repos),
		Filtered: len(filtered),
		Skipped:  len(repos) - len(filtered),
	}

	return filtered, stats
}

// FilterStats holds statistics about a filtering operation.
type FilterStats struct {
	Original int
	Filtered int
	Skipped  int
}

// CombineStats combines multiple FilterStats into one.
func CombineStats(stats ...FilterStats) FilterStats {
	result := FilterStats{}
	for _, s := range stats {
		result.Original += s.Original
		result.Filtered += s.Filtered
		result.Skipped += s.Skipped
	}

	return result
}
