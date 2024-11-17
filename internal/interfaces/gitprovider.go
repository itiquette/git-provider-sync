// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
	"itiquette/git-provider-sync/internal/model"
	config "itiquette/git-provider-sync/internal/model/configuration"
)

// GitProvider defines the interface for interacting with a Git provider service.
// This interface encapsulates operations such as creating repositories,
// fetching repository metadata, and validating repository names.
type GitProvider interface {
	CreateProject(ctx context.Context, cfg config.ProviderConfig, opt model.CreateProjectOption) (string, error)
	IsValidProjectName(ctx context.Context, name string) bool
	Name() string
	ProjectInfos(ctx context.Context, cfg config.ProviderConfig, filtering bool) ([]model.ProjectInfo, error)
	ProtectProject(ctx context.Context, owner string, defaultBranch string, projectIDStr string) error
	SetDefaultBranch(ctx context.Context, owner string, name string, branch string) error
	UnprotectProject(ctx context.Context, defaultBranch string, projectIDStr string) error
}
