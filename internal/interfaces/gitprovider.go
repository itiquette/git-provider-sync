// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// GitProvider defines the interface for interacting with a Git hosting service.
// This interface encapsulates operations such as creating repositories,
// fetching repository metadata, and validating repository names.
type GitProvider interface {
	Create(ctx context.Context, config config.ProviderConfig, option model.CreateOption) (string, error)
	DefaultBranch(ctx context.Context, owner string, projectname string, branch string) error
	IsValidRepositoryName(ctx context.Context, name string) bool
	Name() string
	ProjectInfos(ctx context.Context, config config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error)
	Protect(ctx context.Context, owner string, defaultBranch string, projectID string) error
	Unprotect(ctx context.Context, defaultBranch string, projectID string) error
}
