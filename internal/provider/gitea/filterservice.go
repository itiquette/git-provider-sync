// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

// FilterService represents a filter for Gitea repositories.
// It provides methods to filter repository metadata based on configured rules.
type FilterService struct{}

func NewFilter() *FilterService {
	return &FilterService{}
}

// FilterProjectinfos filters repository metadata based on configured rules.
// It applies both inclusion/exclusion rules and date-based filtering.
//
// Parameters:
// - ctx: The context for the operation, which can be used for cancellation and passing request-scoped values.
// - config: The provider configuration, which includes settings for filtering.
// - projectinfos: A slice of RepositoryMetainfo to be filtered.
//
// Returns:
// - A slice of filtered RepositoryMetainfo.
// - An error if any part of the filtering process fails.
func (FilterService) FilterProjectinfos(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:FilterProjectinfos")

	// Apply inclusion/exclusion rules
	includedProjectinfos, err := targetfilter.FilterIncludedExcludedGen()(ctx, opt, projectinfos)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repositories by include/exclude rules: %w", err)
	}

	// Apply date-based filtering
	return filterByDate(ctx, includedProjectinfos)
}

// filterByDate filters repositories based on their last activity date.
// It uses the includeByActivityTime function to determine if each repository
// should be included based on its last activity time.
//
// Parameters:
// - ctx: The context for the operation.
// - config: The provider configuration.
// - projectinfos: A slice of RepositoryMetainfo to be filtered by date.
//
// Returns:
// - A slice of RepositoryMetainfo that passed the date filter.
// - An error if the filtering process fails for any repository.
func filterByDate(ctx context.Context, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:filterByDate")

	var filtered []model.ProjectInfo

	for _, metainfo := range projectinfos {
		include, err := includeByActivityTime(ctx, metainfo)
		if err != nil {
			return nil, fmt.Errorf("failed to check activity time for %s: %w", metainfo.OriginalName, err)
		}

		if include {
			filtered = append(filtered, metainfo)
		}
	}

	logger.Debug().Msgf("filterByDate: Filtered %d repositories out of %d", len(filtered), len(projectinfos))

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
func includeByActivityTime(ctx context.Context, metainfo model.ProjectInfo) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:includeByActivityTime")
	//	config.DebugLog(logger).Msg("includeByActivityTime: Using configuration")

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
