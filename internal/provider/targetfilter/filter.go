// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// Package targetfilter provides functions for filtering repositories based on various criteria.
// It includes functionality to filter repositories based on their update time and
// inclusion/exclusion lists specified in the configuration.
package targetfilter

import (
	"context"
	"fmt"
	"slices"
	"time"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
)

// IsInInterval checks if the given updatedAt time is within the specified interval.
// It uses the ActiveFromLimit from CLI options to determine the time range.
//
// Parameters:
//   - ctx: The context for logging and accessing CLI options.
//   - updatedAt: The time to check against the interval.
//
// Returns:
//   - bool: True if the updatedAt is within the interval or if no interval is specified.
//   - error: An error if there's an issue parsing the time duration.
func IsInInterval(ctx context.Context, updatedAt time.Time) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering IsInInterval")

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
	return updatedAt.After(then) || updatedAt.Equal(then), nil
}

// FilterIncludedExcludedGen returns a function that filters repositories based on inclusion and exclusion lists.
// This generator pattern allows for flexible use of the filtering logic.
//
// Returns:
//   - A function that takes a context, provider configuration, and a slice of RepositoryMetainfo,
//     and returns a filtered slice of RepositoryMetainfo and any error encountered.
func FilterIncludedExcludedGen() func(context.Context, model.ProviderOption, []model.ProjectInfo) ([]model.ProjectInfo, error) {
	return func(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
		logger := log.Logger(ctx)
		logger.Trace().Msg("Entering FilterIncludeExcluded")

		included := opt.IncludedRepositories
		excluded := opt.ExcludedRepositories

		// Use slices.DeleteFunc to efficiently filter the projectinfos slice
		return slices.DeleteFunc(projectinfos, func(m model.ProjectInfo) bool {
			return !shouldIncludeRepo(m.OriginalName, included, excluded)
		}), nil
	}
}

// shouldIncludeRepo determines if a repository should be included based on the inclusion and exclusion lists.
// This function encapsulates the logic for deciding whether a repository should be included in the final list.
//
// Parameters:
//   - repoName: The name of the repository to check.
//   - included: A slice of repository names that should be included.
//   - excluded: A slice of repository names that should be excluded.
//
// Returns:
//   - bool: True if the repository should be included, false otherwise.
func shouldIncludeRepo(repoName string, included, excluded []string) bool {
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
