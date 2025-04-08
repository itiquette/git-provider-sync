// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/functiondefinition"
	"itiquette/git-provider-sync/internal/model"
	"time"
)

type FilterServicer interface {
	FilterProjectinfos(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc, isInInterval IsInIntervalFunc) ([]model.ProjectInfo, error)
}

type IsInIntervalFunc func(context.Context, time.Time) (bool, error)
