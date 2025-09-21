// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"strings"

	"github.com/google/go-github/v71/github"
)

// RepositoryOptionsFunc is a functional option for configuring repository options.
type RepositoryOptionsFunc func(*github.Repository)

// WithDisabledFeatures disables all GitHub features for a minimalist repository.
func WithDisabledFeatures() RepositoryOptionsFunc {
	return func(repo *github.Repository) {
		repo.HasIssues = github.Ptr(false)
		repo.HasWiki = github.Ptr(false)
		repo.HasPages = github.Ptr(false)
		repo.HasProjects = github.Ptr(false)
		repo.HasDownloads = github.Ptr(false)
		repo.HasDiscussions = github.Ptr(false)
		repo.AllowAutoMerge = github.Ptr(false)
		repo.AllowForking = github.Ptr(false)
		repo.AllowMergeCommit = github.Ptr(false)
		repo.AllowRebaseMerge = github.Ptr(false)
		repo.AllowSquashMerge = github.Ptr(false)
		repo.DeleteBranchOnMerge = github.Ptr(false)
	}
}

// WithAutoInit enables repository auto-initialization.
func WithAutoInit() RepositoryOptionsFunc {
	return func(repo *github.Repository) {
		repo.AutoInit = github.Ptr(true)
	}
}

// WithGitIgnoreTemplate sets the gitignore template.
func WithGitIgnoreTemplate(template string) RepositoryOptionsFunc {
	return func(repo *github.Repository) {
		if template != "" {
			repo.GitignoreTemplate = github.Ptr(template)
		}
	}
}

// WithLicenseTemplate sets the license template.
func WithLicenseTemplate(template string) RepositoryOptionsFunc {
	return func(repo *github.Repository) {
		if template != "" {
			repo.License = &github.License{Key: github.Ptr(template)}
		}
	}
}

// BuildGitHubRepository creates a GitHub repository configuration using functional options.
// This is a pure function - creates new object each time, no mutation.
func BuildGitHubRepository(visibility, name, description, defaultBranch string, opts ...RepositoryOptionsFunc) *github.Repository {
	isPrivate := strings.EqualFold(visibility, "private")

	// Create base repository with sensible defaults
	repo := &github.Repository{
		Name:                github.Ptr(name),
		Description:         github.Ptr(description),
		Private:             github.Ptr(isPrivate),
		DefaultBranch:       github.Ptr(defaultBranch),
		HasIssues:           github.Ptr(true),
		HasWiki:             github.Ptr(true),
		HasPages:            github.Ptr(false),
		HasProjects:         github.Ptr(true),
		HasDownloads:        github.Ptr(true),
		HasDiscussions:      github.Ptr(false),
		AllowAutoMerge:      github.Ptr(true),
		AllowForking:        github.Ptr(true),
		AllowMergeCommit:    github.Ptr(true),
		AllowRebaseMerge:    github.Ptr(true),
		AllowSquashMerge:    github.Ptr(true),
		AllowUpdateBranch:   github.Ptr(true),
		DeleteBranchOnMerge: github.Ptr(true),
		AutoInit:            github.Ptr(false),
	}

	// Apply functional options
	for _, opt := range opts {
		opt(repo)
	}

	return repo
}
