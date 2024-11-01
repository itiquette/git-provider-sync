// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"errors"
	"fmt"
	"itiquette/git-provider-sync/internal/functiondefinition"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

// Filter struct provides methods for filtering GitHub repositories.
// It can be extended in the future to include additional filtering options or state if needed.
type Filter struct{}

// FilterMetainfo filters repository metadata based on inclusion/exclusion rules and activity date.
// This is the main entry point for applying filters to a list of repositories.
//
// Parameters:
//   - ctx: The context for the operation, which may include deadlines or cancellation signals.
//   - config: The configuration for the provider, which contains filtering criteria.
//   - metainfos: A slice of repository metadata to be filtered.
//
// Returns:
//   - []model.RepositoryMetainfo: A slice of filtered repository metadata.
//   - error: An error if the filtering process fails.
func (f Filter) FilterMetainfo(ctx context.Context, config config.ProviderConfig, metainfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering FilterMetainfo: starting")

	return filter(ctx, config, metainfos, targetfilter.FilterIncludedExcludedGen())
}

// filter applies both inclusion/exclusion and date-based filtering to the metainfos.
// This is an internal function that orchestrates the complete filtering process.
//
// Parameters:
//   - ctx: The context for the operation.
//   - config: The provider configuration.
//   - metainfos: The repository metadata to be filtered.
//   - filterExcludedIncludedFunc: A function to filter based on inclusion/exclusion rules.
//
// Returns:
//   - []model.RepositoryMetainfo: The filtered repository metadata.
//   - error: An error if any part of the filtering process fails.
func filter(ctx context.Context, config config.ProviderConfig, metainfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc) ([]model.ProjectInfo, error) {
	filteredByRules, err := filterExcludedIncludedFunc(ctx, config, metainfos)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repositories by inclusion/exclusion rules: %w", err)
	}

	filteredByDate, err := filterByDate(ctx, filteredByRules)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repositories by date: %w", err)
	}

	return filteredByDate, nil
}

// filterByDate filters repositories based on their last activity date.
// It uses the includeByActivityTime function to determine if each repository should be included.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metainfos: The repository metadata to be filtered.
//
// Returns:
//   - []model.RepositoryMetainfo: The filtered repository metadata.
//   - error: An error if the filtering process fails for any repository.
func filterByDate(ctx context.Context, metainfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering filterByDate: starting")

	var filtered []model.ProjectInfo

	for _, metainfo := range metainfos {
		include, err := includeByActivityTime(ctx, metainfo)
		if err != nil {
			return nil, fmt.Errorf("failed to check activity time for repository %s: %w", metainfo.OriginalName, err)
		}

		if include {
			filtered = append(filtered, metainfo)
		}
	}

	return filtered, nil
}

// includeByActivityTime checks if a repository's last activity is within the specified interval.
// It uses the targetfilter.IsInInterval function to perform the actual time check.
//
// Parameters:
//   - ctx: The context for the operation.
//   - metainfo: The metadata of the repository being evaluated.
//
// Returns:
//   - bool: true if the repository's last activity is within the specified interval, false otherwise.
//   - error: An error if the evaluation process fails or if the last activity time is nil.
func includeByActivityTime(ctx context.Context, metainfo model.ProjectInfo) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering includeByActivityTime: checking")

	if metainfo.LastActivityAt == nil {
		return false, errors.New("last activity time is nil for repository" + metainfo.OriginalName)
	}

	inInterval, err := targetfilter.IsInInterval(ctx, *metainfo.LastActivityAt)
	if err != nil {
		return false, fmt.Errorf("failed to check if is interval %w", err)
	}

	return inInterval, nil
}
