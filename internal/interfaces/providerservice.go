// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/functiondefinition"
	"itiquette/git-provider-sync/internal/model"
	"time"
)

type ProjectServicer interface {
	CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error)
	Exists(ctx context.Context, owner, repo string) (bool, string, error)
	GetProjectInfos(ctx context.Context, opt model.ProviderOption) ([]model.ProjectInfo, error)
	SetDefaultBranch(ctx context.Context, owner, projectName, branch string) error
}

type ProtectionServicer interface {
	Protect(ctx context.Context, defaultBranch string, projectIDstr string) error
	Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error
}

type FilterServicer interface {
	FilterProjectinfos(ctx context.Context, opt model.ProviderOption, projectinfos []model.ProjectInfo, filterExcludedIncludedFunc functiondefinition.FilterIncludedExcludedFunc, isInInterval IsInIntervalFunc) ([]model.ProjectInfo, error)
}

type IsInIntervalFunc func(context.Context, time.Time) (bool, error)
