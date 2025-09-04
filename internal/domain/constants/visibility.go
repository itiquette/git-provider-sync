// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package constants

// Repository visibility constants used across providers.
const (
	// VisibilityPublic represents a public repository.
	VisibilityPublic = "public"

	// VisibilityPrivate represents a private repository.
	VisibilityPrivate = "private"

	// VisibilityInternal represents an internal repository (GitLab specific).
	VisibilityInternal = "internal"
)

// Default names for repositories.
const (
	// DefaultRepositoryName is the default name for renamed repositories.
	DefaultRepositoryName = "repository"

	// DefaultProjectName is the default name for renamed projects.
	DefaultProjectName = "project"
)

// Git branch names.
const (
	// DefaultBranch is the default branch name.
	DefaultBranch = "main"
)
