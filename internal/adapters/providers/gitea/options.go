// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"
	"itiquette/git-provider-sync/internal/domain/ports"
)

const (
	visibilityPrivate = "private"
)

// ProjectOption is a functional option for configuring Gitea project options.
type ProjectOption func(*gitea.CreateRepoOption)

// EditOption is a functional option for configuring Gitea repository edit options.
type EditOption func(*gitea.EditRepoOption)

// WithVisibility sets the repository visibility.
func WithVisibility(isPrivate bool) ProjectOption {
	return func(opts *gitea.CreateRepoOption) {
		opts.Private = isPrivate
	}
}

// WithInitialization configures repository initialization options.
func WithInitialization(autoInit bool, readme, gitignore, license string) ProjectOption {
	return func(opts *gitea.CreateRepoOption) {
		opts.AutoInit = autoInit
		if readme != "" {
			opts.Readme = readme
		}

		if gitignore != "" {
			opts.Gitignores = gitignore
		}

		if license != "" {
			opts.License = license
		}
	}
}

// WithTemplate configures template repository settings.
func WithTemplate(isTemplate bool) ProjectOption {
	return func(opts *gitea.CreateRepoOption) {
		opts.Template = isTemplate
	}
}

// WithIssueLabels configures issue labels.
func WithIssueLabels(issueLabels string) ProjectOption {
	return func(opts *gitea.CreateRepoOption) {
		if issueLabels != "" {
			opts.IssueLabels = issueLabels
		}
	}
}

// WithTrustModel configures the trust model for the repository.
func WithTrustModel(trustModel string) ProjectOption {
	return func(opts *gitea.CreateRepoOption) {
		switch trustModel {
		case "default":
			opts.TrustModel = gitea.TrustModelDefault
		case "collaborator":
			opts.TrustModel = gitea.TrustModelCollaborator
		case "committer":
			opts.TrustModel = gitea.TrustModelCommitter
		case "collaboratorcommitter":
			opts.TrustModel = gitea.TrustModelCollaboratorCommitter
		}
	}
}

// BuildGiteaProject creates a Gitea project configuration using functional options.
// This is a pure function - creates new object each time, no mutation.
func BuildGiteaProject(visibility, name, description, defaultBranch string, opts ...ProjectOption) *gitea.CreateRepoOption {
	// Create base project options
	project := &gitea.CreateRepoOption{
		Name:          name,
		Description:   description,
		DefaultBranch: defaultBranch,
		Private:       visibility == visibilityPrivate,
	}

	// Apply functional options
	for _, opt := range opts {
		opt(project)
	}

	return project
}

// WithCollaborationFeatures returns an EditOption that configures collaboration features.
func WithCollaborationFeatures(config CollaborationConfig) EditOption {
	return func(opts *gitea.EditRepoOption) {
		opts.HasIssues = gitea.OptionalBool(config.EnableIssues)
		opts.HasPullRequests = gitea.OptionalBool(config.EnablePullRequests)
		opts.HasWiki = gitea.OptionalBool(config.EnableWiki)

		opts.HasProjects = gitea.OptionalBool(config.EnableProjects)
		if config.DefaultBranch != "" {
			opts.DefaultBranch = &config.DefaultBranch
		}
	}
}

// WithCIFeatures returns an EditOption that configures CI/CD features.
func WithCIFeatures(config CIConfig) EditOption {
	return func(opts *gitea.EditRepoOption) {
		opts.HasActions = gitea.OptionalBool(config.EnableActions)
		opts.HasReleases = gitea.OptionalBool(config.EnableReleases)
	}
}

// WithRepositorySettings returns an EditOption that configures general repository settings.
func WithRepositorySettings(config RepositoryConfig) EditOption {
	return func(opts *gitea.EditRepoOption) {
		if config.Description != "" {
			opts.Description = &config.Description
		}

		if config.Website != "" {
			opts.Website = &config.Website
		}

		if config.Archived != nil {
			opts.Archived = config.Archived
		}

		if config.Private != nil {
			opts.Private = config.Private
		}
	}
}

// WithAllFeaturesDisabled returns an EditOption that disables all features.
func WithAllFeaturesDisabled() EditOption {
	return func(opts *gitea.EditRepoOption) {
		opts.HasIssues = gitea.OptionalBool(false)
		opts.HasWiki = gitea.OptionalBool(false)
		opts.HasProjects = gitea.OptionalBool(false)
		opts.HasPullRequests = gitea.OptionalBool(false)
		opts.HasReleases = gitea.OptionalBool(false)
		opts.HasActions = gitea.OptionalBool(false)
	}
}

// WithAllFeaturesEnabled returns an EditOption that enables all features.
func WithAllFeaturesEnabled() EditOption {
	return func(opts *gitea.EditRepoOption) {
		opts.HasIssues = gitea.OptionalBool(true)
		opts.HasWiki = gitea.OptionalBool(true)
		opts.HasProjects = gitea.OptionalBool(true)
		opts.HasPullRequests = gitea.OptionalBool(true)
		opts.HasReleases = gitea.OptionalBool(true)
		opts.HasActions = gitea.OptionalBool(true)
	}
}

// BuildGiteaEditOptions creates Gitea repository edit options using functional options.
// This is a pure function - creates new object each time, no mutation.
func BuildGiteaEditOptions(opts ...EditOption) *gitea.EditRepoOption {
	editOpts := &gitea.EditRepoOption{}

	// Apply functional options
	for _, opt := range opts {
		opt(editOpts)
	}

	return editOpts
}

// ApplyRepositorySettings is a pure function that applies settings to a repository.
// It's a wrapper that uses the functional options pattern instead of stateful builder.
func ApplyRepositorySettings(ctx context.Context, client *gitea.Client, logger ports.Logger, owner, projectName string, opts ...EditOption) error {
	logger.Debug(ctx, "Applying repository settings", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	// Build edit options using functional pattern
	editOpts := BuildGiteaEditOptions(opts...)

	// Apply the settings
	_, _, err := client.EditRepo(owner, projectName, *editOpts)
	if err != nil {
		return fmt.Errorf("failed to edit repository settings: owner: %s, repo: %s, err: %w", owner, projectName, err)
	}

	logger.Info(ctx, "Repository settings applied successfully", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	return nil
}

// CollaborationConfig contains collaboration feature settings.
type CollaborationConfig struct {
	EnableIssues       bool
	EnablePullRequests bool
	EnableWiki         bool
	EnableProjects     bool
	DefaultBranch      string
}

// CIConfig contains CI/CD feature settings.
type CIConfig struct {
	EnableActions  bool
	EnableReleases bool
}

// RepositoryConfig contains general repository settings.
type RepositoryConfig struct {
	Description string
	Website     string
	Private     *bool
	Archived    *bool
}
