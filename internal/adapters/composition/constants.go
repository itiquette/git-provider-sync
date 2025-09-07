// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package composition

// Provider type constants.
const (
	ProviderTypeGitHub    = "github"
	ProviderTypeGitLab    = "gitlab"
	ProviderTypeGitea     = "gitea"
	ProviderTypeGoGit     = "go-git"
	ProviderTypeGitBinary = "git-binary"
	ProviderTypeDirectory = "directory"
	ProviderTypeArchive   = "archive"
)

// Default domain constants.
const (
	DomainGitHub = "github.com"
	DomainGitLab = "gitlab.com"
	DomainGitea  = "gitea.com"
)

// Visibility constants.
const (
	VisibilityPrivate  = "private"
	VisibilityPublic   = "public"
	VisibilityInternal = "internal"
)
