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

// ProjectOptionsBuilder provides Gitea repository creation options.
type ProjectOptionsBuilder struct {
	opts   *gitea.CreateRepoOption
	client *gitea.Client
	logger ports.Logger
}

// NewProjectOptionsBuilder creates a new Gitea project options builder.
func NewProjectOptionsBuilder(client *gitea.Client, logger ports.Logger) *ProjectOptionsBuilder {
	return &ProjectOptionsBuilder{
		opts:   &gitea.CreateRepoOption{},
		client: client,
		logger: logger,
	}
}

// BasicOpts sets the fundamental repository options.
func (b *ProjectOptionsBuilder) BasicOpts(visibility, name, description, defaultBranch string) {
	b.opts.Name = name
	b.opts.Description = description
	b.opts.DefaultBranch = defaultBranch
	b.opts.Private = visibility == visibilityPrivate
}

// WithVisibility sets the repository visibility.
func (b *ProjectOptionsBuilder) WithVisibility(visibility string) {
	b.opts.Private = visibility == visibilityPrivate
}

// WithInitialization configures repository initialization options.
func (b *ProjectOptionsBuilder) WithInitialization(autoInit bool, readme, gitignore, license string) {
	b.opts.AutoInit = autoInit

	if readme != "" {
		b.opts.Readme = readme
	}

	if gitignore != "" {
		b.opts.Gitignores = gitignore
	}

	if license != "" {
		b.opts.License = license
	}
}

// WithTemplate configures template repository settings.
func (b *ProjectOptionsBuilder) WithTemplate(isTemplate bool) {
	b.opts.Template = isTemplate
}

// WithIssueLabels configures issue labels.
func (b *ProjectOptionsBuilder) WithIssueLabels(issueLabels string) {
	if issueLabels != "" {
		b.opts.IssueLabels = issueLabels
	}
}

// WithTrustModel configures the trust model for the repository.
func (b *ProjectOptionsBuilder) WithTrustModel(trustModel string) {
	if trustModel != "" {
		// Convert string to TrustModel enum if needed
		switch trustModel {
		case "default":
			b.opts.TrustModel = gitea.TrustModelDefault
		case "collaborator":
			b.opts.TrustModel = gitea.TrustModelCollaborator
		case "committer":
			b.opts.TrustModel = gitea.TrustModelCommitter
		case "collaboratorcommitter":
			b.opts.TrustModel = gitea.TrustModelCollaboratorCommitter
		}
	}
}

// Build returns the configured repository options.
func (b *ProjectOptionsBuilder) Build() *gitea.CreateRepoOption {
	return b.opts
}

// Reset clears all options and starts fresh.
func (b *ProjectOptionsBuilder) Reset() {
	b.opts = &gitea.CreateRepoOption{}
}

// ApplyDisabledSettings applies feature disabling after repository creation.
func (b *ProjectOptionsBuilder) ApplyDisabledSettings(ctx context.Context, owner, projectName string) error {
	b.logger.Debug(ctx, "Applying disabled settings to Gitea repository", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	// These settings can only be applied after repository creation
	editOpts := gitea.EditRepoOption{
		// Core repository features
		HasIssues:       gitea.OptionalBool(false),
		HasWiki:         gitea.OptionalBool(false),
		HasProjects:     gitea.OptionalBool(false),
		HasPullRequests: gitea.OptionalBool(false),
		HasReleases:     gitea.OptionalBool(false),

		// CI/CD and automation
		HasActions: gitea.OptionalBool(false),

		// Additional features that may be available
		// Note: Some of these might not be available in all Gitea versions
	}

	_, _, err := b.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to edit repository settings: owner: %s, repo: %s, err: %w", owner, projectName, err)
	}

	b.logger.Info(ctx, "Disabled settings applied successfully", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	return nil
}

// ApplyEnabledSettings enables all repository features.
func (b *ProjectOptionsBuilder) ApplyEnabledSettings(ctx context.Context, owner, projectName string) error {
	b.logger.Debug(ctx, "Applying enabled settings to Gitea repository", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	editOpts := gitea.EditRepoOption{
		HasIssues:       gitea.OptionalBool(true),
		HasWiki:         gitea.OptionalBool(true),
		HasProjects:     gitea.OptionalBool(true),
		HasPullRequests: gitea.OptionalBool(true),
		HasReleases:     gitea.OptionalBool(true),
		HasActions:      gitea.OptionalBool(true),
	}

	_, _, err := b.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to edit repository settings: owner: %s, repo: %s, err: %w", owner, projectName, err)
	}

	b.logger.Info(ctx, "Enabled settings applied successfully", map[string]any{
		"owner": owner,
		"repo":  projectName,
	})

	return nil
}

// ConfigureCollaborationFeatures configures collaboration-related features.
func (b *ProjectOptionsBuilder) ConfigureCollaborationFeatures(_ context.Context, owner, projectName string, config CollaborationConfig) error {
	editOpts := gitea.EditRepoOption{
		HasIssues:       gitea.OptionalBool(config.EnableIssues),
		HasPullRequests: gitea.OptionalBool(config.EnablePullRequests),
		HasWiki:         gitea.OptionalBool(config.EnableWiki),
		HasProjects:     gitea.OptionalBool(config.EnableProjects),
	}

	if config.DefaultBranch != "" {
		editOpts.DefaultBranch = &config.DefaultBranch
	}

	_, _, err := b.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to configure collaboration features: %w", err)
	}

	return nil
}

// ConfigureCIFeatures configures CI/CD-related features.
func (b *ProjectOptionsBuilder) ConfigureCIFeatures(_ context.Context, owner, projectName string, config CIConfig) error {
	editOpts := gitea.EditRepoOption{
		HasActions:  gitea.OptionalBool(config.EnableActions),
		HasReleases: gitea.OptionalBool(config.EnableReleases),
	}

	_, _, err := b.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to configure CI features: %w", err)
	}

	return nil
}

// ConfigureRepositorySettings configures general repository settings.
func (b *ProjectOptionsBuilder) ConfigureRepositorySettings(_ context.Context, owner, projectName string, config RepositoryConfig) error {
	editOpts := gitea.EditRepoOption{}

	if config.Description != "" {
		editOpts.Description = &config.Description
	}

	if config.Website != "" {
		editOpts.Website = &config.Website
	}

	if config.Archived != nil {
		editOpts.Archived = config.Archived
	}

	if config.Private != nil {
		editOpts.Private = config.Private
	}

	_, _, err := b.client.EditRepo(owner, projectName, editOpts)
	if err != nil {
		return fmt.Errorf("failed to configure repository settings: %w", err)
	}

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

// GetFeatureDefaults returns default feature settings for different repository types.
func GetFeatureDefaults(repoType string) (CollaborationConfig, CIConfig) {
	switch repoType {
	case "minimal":
		return CollaborationConfig{
				EnableIssues:       false,
				EnablePullRequests: false,
				EnableWiki:         false,
				EnableProjects:     false,
			}, CIConfig{
				EnableActions:  false,
				EnableReleases: false,
			}
	case "development":
		return CollaborationConfig{
				EnableIssues:       true,
				EnablePullRequests: true,
				EnableWiki:         true,
				EnableProjects:     true,
			}, CIConfig{
				EnableActions:  true,
				EnableReleases: true,
			}
	default: // "standard"
		return CollaborationConfig{
				EnableIssues:       true,
				EnablePullRequests: true,
				EnableWiki:         false,
				EnableProjects:     false,
			}, CIConfig{
				EnableActions:  false,
				EnableReleases: true,
			}
	}
}
