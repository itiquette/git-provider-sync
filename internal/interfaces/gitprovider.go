// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
)

// GitProvider defines the interface for interacting with a Git provider service.
// This interface encapsulates operations such as creating repositories,
// fetching repository metadata, and validating repository names.
type GitProvider interface {
	ProjectServicer
	ProtectionServicer
	IsValidProjectName(ctx context.Context, name string) bool
	SetDefaultBranch(ctx context.Context, owner string, name string, branch string) error
	Name() string
}

type ProjectServicer interface {
	CreateProject(ctx context.Context, opt model.CreateProjectOption) (string, error)
	GetProjectInfos(ctx context.Context, providerOpt model.ProviderOption, filtering bool) ([]model.ProjectInfo, error)
	ProjectExists(ctx context.Context, owner, repo string) (bool, string, error)
	SetDefaultBranch(ctx context.Context, owner, projectName, branch string) error
}

type ProtectionServicer interface {
	Protect(ctx context.Context, owner string, defaultBranch string, projectIDstr string) error
	Unprotect(ctx context.Context, defaultBranch string, projectIDStr string) error
}
