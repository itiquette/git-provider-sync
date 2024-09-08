// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

// Filter represents a filter for Gitea repositories.
// It provides methods to filter repository metadata based on configured rules.
type Filter struct{}

// FilterMetainfo filters repository metadata based on configured rules.
// It applies both inclusion/exclusion rules and date-based filtering.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing request-scoped values.
// - config: The provider configuration, which includes settings for filtering.
// - metainfos: A slice of RepositoryMetainfo to be filtered.
//
// Returns:
// - A slice of filtered RepositoryMetainfo.
// - An error if any part of the filtering process fails.
func (Filter) FilterMetainfo(ctx context.Context, config configuration.ProviderConfig, metainfos []model.RepositoryMetainfo) ([]model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering FilterMetainfo: Starting repository filtering process")

	// Apply inclusion/exclusion rules
	includedMetainfos, err := targetfilter.FilterIncludedExcludedGen()(ctx, config, metainfos)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repositories by include/exclude rules: %w", err)
	}

	// Apply date-based filtering
	return filterByDate(ctx, config, includedMetainfos)
}

// filterByDate filters repositories based on their last activity date.
// It uses the includeByActivityTime function to determine if each repository
// should be included based on its last activity time.
//
// Parameters:
// - ctx: The context for the operation.
// - config: The provider configuration.
// - metainfos: A slice of RepositoryMetainfo to be filtered by date.
//
// Returns:
// - A slice of RepositoryMetainfo that passed the date filter.
// - An error if the filtering process fails for any repository.
func filterByDate(ctx context.Context, config configuration.ProviderConfig, metainfos []model.RepositoryMetainfo) ([]model.RepositoryMetainfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering filterByDate: Filtering repositories by last activity date")

	var filtered []model.RepositoryMetainfo

	for _, metainfo := range metainfos {
		include, err := includeByActivityTime(ctx, config, metainfo)
		if err != nil {
			return nil, fmt.Errorf("failed to check activity time for %s: %w", metainfo.OriginalName, err)
		}

		if include {
			filtered = append(filtered, metainfo)
		}
	}

	logger.Debug().Msgf("filterByDate: Filtered %d repositories out of %d", len(filtered), len(metainfos))

	return filtered, nil
}

// includeByActivityTime checks if a repository's last activity is within the configured time interval.
// It uses the targetfilter.IsInInterval function to determine if the last activity time
// falls within the specified interval.
//
// Parameters:
// - ctx: The context for the operation.
// - config: The provider configuration, used for logging.
// - metainfo: The RepositoryMetainfo to check.
//
// Returns:
// - A boolean indicating whether the repository should be included based on its activity time.
// - An error if the check fails, e.g., if the last activity time is nil.
func includeByActivityTime(ctx context.Context, config configuration.ProviderConfig, metainfo model.RepositoryMetainfo) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering includeByActivityTime: Checking repository activity time")
	config.DebugLog(logger).Msg("includeByActivityTime: Using configuration")

	if metainfo.LastActivityAt == nil {
		return false, fmt.Errorf("last activity time is nil for repository %s", metainfo.OriginalName)
	}

	isInInterval, err := targetfilter.IsInInterval(ctx, *metainfo.LastActivityAt)
	if err != nil {
		return false, fmt.Errorf("failed to check if activity time is in interval for %s: %w", metainfo.OriginalName, err)
	}

	logger.Debug().
		Str("repository", metainfo.OriginalName).
		Time("lastActivity", *metainfo.LastActivityAt).
		Bool("included", isInInterval).
		Msg("includeByActivityTime: Repository activity time check result")

	return isInInterval, nil
}
