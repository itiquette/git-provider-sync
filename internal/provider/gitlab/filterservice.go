// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/functiondefinition"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

type filterService struct {
	isInInterval interfaces.IsInIntervalFunc
}

func NewFilter() filterService { //nolint
	isInInterval := targetfilter.IsInInterval

	return filterService{isInInterval: isInInterval}
}

func (filterService) FilterProjectinfos(ctx context.Context, cfg config.ProviderConfig, projectinfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc, isInInterval interfaces.IsInIntervalFunc) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:FilterProjectinfos")

	filteredURLs, err := filterExcludedIncludedFunc(ctx, cfg, projectinfos)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repository URLs by include/exclude: %w", err)
	}

	return filterByDate(ctx, filteredURLs, isInInterval)
}

func filterByDate(ctx context.Context, projectinfos []model.ProjectInfo, isInInterval interfaces.IsInIntervalFunc) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:filterByDate")

	filteredProjectinfos := make([]model.ProjectInfo, 0)

	for _, metainfo := range projectinfos {
		if metainfo.LastActivityAt == nil {
			continue
		}

		include, err := isInInterval(ctx, *metainfo.LastActivityAt)
		if err != nil {
			return nil, fmt.Errorf("failed to filter include by activity time: %w", err)
		}

		if include {
			filteredProjectinfos = append(filteredProjectinfos, metainfo)
		}
	}

	return filteredProjectinfos, nil
}