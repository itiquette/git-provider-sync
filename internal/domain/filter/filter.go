// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

// Package filter provides common type definitions for filtering operations.
package filter

import (
	"context"

	"itiquette/git-provider-sync/internal/model"
)

// IncludedExcludedFunc defines a function type for filtering repository metadata
// based on inclusion/exclusion criteria. This maintains compatibility with legacy
// filtering while supporting hexagonal architecture patterns.
type IncludedExcludedFunc func(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo) ([]model.ProjectInfo, error)

// IsInIntervalFunc defines a function type for filtering repositories based on
// activity intervals or time-based criteria.
type IsInIntervalFunc func(ctx context.Context, projectinfo model.ProjectInfo) (bool, error)
