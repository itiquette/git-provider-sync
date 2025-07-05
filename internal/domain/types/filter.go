// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package types

import (
	"context"

	"itiquette/git-provider-sync/internal/model"
)

// FilterIncludedExcludedFunc defines a function type for filtering repository metadata
// based on inclusion/exclusion criteria. This maintains compatibility with legacy
// filtering while supporting hexagonal architecture patterns.
type FilterIncludedExcludedFunc func(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error)

// IsInIntervalFunc defines a function type for filtering repositories based on
// activity intervals or time-based criteria.
type IsInIntervalFunc func(ctx context.Context, projectinfo model.ProjectInfo) (bool, error)
