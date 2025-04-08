// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

// filterService struct provides methods for filtering GitHub repositories.
// It can be extended in the future to include additional filtering options or state if needed.
type filterService struct{}

func NewFilter() *filterService {
	return &filterService{}
}

// FilterProjectInfos filters repository metadata based on inclusion/exclusion rules and activity date.
// This is the main entry point for applying filters to a list of repositories.
//
// Parameters:
//   - ctx: The context for the operation, which may include deadlines or cancellation signals.
//   - config: The configuration for the provider, which contains filtering criteria.
//   - projectinfos: A slice of repository metadata to be filtered.
//
// Returns:
//   - []model.RepositoryMetainfo: A slice of filtered repository metadata.
//   - error: An error if the filtering process fails.
func (f filterService) FilterProjectInfos(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:FilterProjectInfos")

	return filter(ctx, opt, projectinfos, targetfilter.FilterIncludedExcludedGen())
}

// filter applies both inclusion/exclusion and date-based filtering to the projectinfos.
// This is an internal function that orchestrates the complete filtering process.
//
// Parameters:
//   - ctx: The context for the operation.
//   - config: The provider configuration.
//   - projectinfos: The repository metadata to be filtered.
//   - filterExcludedIncludedFunc: A function to filter based on inclusion/exclusion rules.
//
// Returns:
//   - []model.RepositoryMetainfo: The filtered repository metadata.
//   - error: An error if any part of the filtering process fails.
func filter(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:filter")

	filteredByRules, err := filterExcludedIncludedFunc(ctx, opt, projectinfos)
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
//   - projectinfos: The repository metadata to be filtered.
//
// Returns:
//   - []model.RepositoryMetainfo: The filtered repository metadata.
//   - error: An error if the filtering process fails for any repository.
func filterByDate(ctx context.Context, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:filterByDate")

	var filtered []model.ProjectInfo

	for _, metainfo := range projectinfos {
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
func includeByActivityTime(ctx context.Context, projectInfo model.ProjectInfo) (bool, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:includeByActivityTime")

	if projectInfo.LastActivityAt == nil {
		return false, errors.New("last activity time is nil for repository" + projectInfo.OriginalName)
	}

	inInterval, err := targetfilter.IsInInterval(ctx, *projectInfo.LastActivityAt)
	if err != nil {
		return false, fmt.Errorf("failed to check if is interval %w", err)
	}

	return inInterval, nil
}
