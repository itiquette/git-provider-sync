// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"fmt"
	"time"

	"itiquette/git-provider-sync/internal/functiondefinition"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/targetfilter"
)

type IsInIntervalFunc func(context.Context, time.Time) (bool, error)

type Filter struct {
	isInInterval IsInIntervalFunc
}

func NewFilter(isInInterval IsInIntervalFunc) *Filter {
	if isInInterval == nil {
		isInInterval = targetfilter.IsInInterval
	}

	return &Filter{isInInterval: isInInterval}
}

func (Filter) FilterMetainfo(ctx context.Context, config config.ProviderConfig, metainfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc, isInInterval IsInIntervalFunc) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering FilterMetainfo: starting")

	filteredURLs, err := filterExcludedIncludedFunc(ctx, config, metainfos)
	if err != nil {
		return nil, fmt.Errorf("failed to filter repository URLs by include/exclude: %w", err)
	}

	return filterByDate(ctx, filteredURLs, isInInterval)
}

func filterByDate(ctx context.Context, metainfos []model.ProjectInfo, isInInterval IsInIntervalFunc) ([]model.ProjectInfo, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering filterByDate: starting")

	filteredMetainfos := make([]model.ProjectInfo, 0)

	for _, metainfo := range metainfos {
		if metainfo.LastActivityAt == nil {
			continue
		}

		include, err := isInInterval(ctx, *metainfo.LastActivityAt)
		if err != nil {
			return nil, fmt.Errorf("failed to filter include by activity time: %w", err)
		}

		if include {
			filteredMetainfos = append(filteredMetainfos, metainfo)
		}
	}

	return filteredMetainfos, nil
}
