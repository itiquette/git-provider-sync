// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package interfaces

import (
	"context"
)

// GitProvider defines the interface for interacting with a Git provider service.
// This interface encapsulates operations such as creating repositories,
// fetching repository metadata, and validating repository names.
type GitProvider interface {
	ProjectServicer
	IsValidProjectName(ctx context.Context, name string) bool
	SetDefaultBranch(ctx context.Context, owner string, name string, branch string) error
	Name() string
	ProtectProject(ctx context.Context, owner string, defaultBranch string, projectIDStr string) error
	UnprotectProject(ctx context.Context, defaultBranch string, projectIDStr string) error
}
