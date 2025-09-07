// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package filter

import (
	"context"

	"itiquette/git-provider-sync/internal/domain/entities"
)

// IncludedExcludedFunc defines a function type for filtering repository metadata
// Based on inclusion/exclusion criteria. This maintains compatibility with legacy
// Filtering while supporting hexagonal architecture patterns.
type IncludedExcludedFunc func(ctx context.Context, opt entities.ProviderOption, projectinfos []entities.ProjectInfo) ([]entities.ProjectInfo, error)

// IsInIntervalFunc defines a function type for filtering repositories based on
// Activity intervals or time-based criteria.
type IsInIntervalFunc func(ctx context.Context, projectinfo entities.ProjectInfo) (bool, error)
