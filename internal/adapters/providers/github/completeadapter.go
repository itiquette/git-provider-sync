// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v71/github"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// CompleteAdapter provides a GitHub adapter with all services.
type CompleteAdapter struct {
	*Adapter // Embed the basic adapter

	// Service layers
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *FilterService

	client *github.Client
	logger ports.Logger
}

// NewCompleteAdapter creates a new complete GitHub adapter with all services.
func NewCompleteAdapter(ctx context.Context, config Config, logger ports.Logger) *CompleteAdapter {
	// Create basic adapter
	basicAdapter := NewWithConfig(ctx, config)

	// Extract client from the basic adapter (we need access to it)
	var client *github.Client
	if config.HTTPClient != nil {
		client = github.NewClient(config.HTTPClient)
	} else {
		client = github.NewClient(nil)
	}

	// Set BaseURL if provided (important for testing with mock servers)
	if config.BaseURL != "" {
		var err error

		client, err = client.WithEnterpriseURLs(config.BaseURL, config.UploadURL)
		if err != nil {
			// Log error but continue with default client
			logger.Error(ctx, "Failed to set GitHub enterprise URLs", map[string]any{
				"baseURL":   config.BaseURL,
				"uploadURL": config.UploadURL,
				"error":     err.Error(),
			})
		}
	}

	if config.Token != "" {
		client = client.WithAuthToken(config.Token)
	}

	// Create service layers
	projectService := NewProjectService(client, logger)
	protectionService := NewProtectionService(client, logger)
	filterService := NewFilterService(logger)

	return &CompleteAdapter{
		Adapter:           basicAdapter,
		projectService:    projectService,
		protectionService: protectionService,
		filterService:     filterService,
		client:            client,
		logger:            logger,
	}
}

// CreateRepositoryWithAdvancedOptions creates a repository with options.
func (ca *CompleteAdapter) CreateRepositoryWithAdvancedOptions(
	ctx context.Context,
	config ports.ProviderConfig,
	options CreateProjectRequest,
) (*entities.Repository, error) {
	ca.logger.Info(ctx, "Creating GitHub repository", map[string]any{
		"owner": config.Owner,
		"name":  options.Name,
	})

	// Validate repository name
	if err := ca.projectService.ValidateProjectName(options.Name); err != nil {
		return nil, fmt.Errorf("invalid repository name: %w", err)
	}

	// Create repository using project service
	repository, err := ca.projectService.CreateProject(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	ca.logger.Info(ctx, "GitHub repository created successfully", map[string]any{
		"owner": config.Owner,
		"name":  options.Name,
		"url":   repository.HTTPSURL(),
	})

	return repository, nil
}

// ApplyRepositoryProtection applies protection to a repository.
func (ca *CompleteAdapter) ApplyRepositoryProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryName string,
	protectionOptions ProtectRepositoryRequest,
) error {
	ca.logger.Info(ctx, "Applying repository protection", map[string]any{
		"owner":      config.Owner,
		"repository": repositoryName,
	})

	protectionOptions.Owner = config.Owner
	protectionOptions.RepositoryName = repositoryName

	return ca.protectionService.ProtectRepository(ctx, protectionOptions)
}

// RemoveRepositoryProtection removes protection from a repository.
func (ca *CompleteAdapter) RemoveRepositoryProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryName string,
) error {
	ca.logger.Info(ctx, "Removing repository protection", map[string]any{
		"owner":      config.Owner,
		"repository": repositoryName,
	})

	return ca.protectionService.UnprotectRepository(ctx, config.Owner, repositoryName)
}

// FilterRepositoriesWithAdvancedCriteria filters repositories using criteria.
func (ca *CompleteAdapter) FilterRepositoriesWithAdvancedCriteria(
	ctx context.Context,
	repositories []entities.Repository,
	filterOptions FilterRepositoriesRequest,
) ([]entities.Repository, error) {
	ca.logger.Debug(ctx, "Filtering repositories", map[string]any{
		"total_count": len(repositories),
	})

	return ca.filterService.FilterRepositories(ctx, filterOptions)
}

// ValidateAndTransformRepositoryName validates and transforms a repository name according to transformation rules.
func (ca *CompleteAdapter) ValidateAndTransformRepositoryName(
	name string,
	options ports.NameTransformOptions,
) (string, error) {
	// First validate the original name
	if err := ca.projectService.ValidateProjectName(name); err != nil {
		// If invalid, try to transform it
		transformed := ca.projectService.TransformProjectName(name, options)

		// Validate the transformed name
		if err := ca.projectService.ValidateProjectName(transformed); err != nil {
			return "", fmt.Errorf("name validation failed even after transformation: %w", err)
		}

		return transformed, nil
	}

	// If already valid, apply transformations if requested
	if options.ToLowercase || options.ToUppercase || len(options.Replacements) > 0 ||
		options.Prefix != "" || options.Suffix != "" || options.AlphaNumericOnly ||
		options.MaxLength > 0 {
		return ca.projectService.TransformProjectName(name, options), nil
	}

	return name, nil
}

// GetRepositoryStatistics returns detailed statistics about repositories.
func (ca *CompleteAdapter) GetRepositoryStatistics(
	_ context.Context,
	repositories []entities.Repository,
) map[string]any {
	stats := map[string]any{
		"total_count": len(repositories),
	}

	// Count by visibility
	publicCount := 0
	privateCount := 0
	archivedCount := 0
	forkCount := 0

	for _, repo := range repositories {
		if repo.Visibility() == constants.VisibilityPrivate {
			privateCount++
		} else {
			publicCount++
		}

		if repo.IsArchived() {
			archivedCount++
		}

		if repo.IsFork() {
			forkCount++
		}
	}

	stats["public_count"] = publicCount
	stats["private_count"] = privateCount
	stats["archived_count"] = archivedCount
	stats["fork_count"] = forkCount
	stats["original_count"] = len(repositories) - forkCount

	return stats
}

// BulkApplyProtection applies protection to multiple repositories.
func (ca *CompleteAdapter) BulkApplyProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryNames []string,
	protectionOptions ProtectRepositoryRequest,
) error {
	ca.logger.Info(ctx, "Applying bulk repository protection", map[string]any{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	var errors []error

	for _, repoName := range repositoryNames {
		if err := ca.ApplyRepositoryProtection(ctx, config, repoName, protectionOptions); err != nil {
			ca.logger.Error(ctx, "Failed to apply protection to repository", map[string]any{
				"owner":      config.Owner,
				"repository": repoName,
				"error":      err.Error(),
			})

			errors = append(errors, fmt.Errorf("failed to protect %s: %w", repoName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %d repositories failed: %v", domain.ErrBulkProtectionFailed, len(errors), errors)
	}

	ca.logger.Info(ctx, "Bulk repository protection completed successfully", map[string]any{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	return nil
}

// BulkRemoveProtection removes protection from multiple repositories.
func (ca *CompleteAdapter) BulkRemoveProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryNames []string,
) error {
	ca.logger.Info(ctx, "Removing bulk repository protection", map[string]any{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	var errors []error

	for _, repoName := range repositoryNames {
		if err := ca.RemoveRepositoryProtection(ctx, config, repoName); err != nil {
			ca.logger.Error(ctx, "Failed to remove protection from repository", map[string]any{
				"owner":      config.Owner,
				"repository": repoName,
				"error":      err.Error(),
			})

			errors = append(errors, fmt.Errorf("failed to unprotect %s: %w", repoName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %d repositories failed: %v", domain.ErrBulkRemovalFailed, len(errors), errors)
	}

	ca.logger.Info(ctx, "Bulk repository protection removal completed successfully", map[string]any{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	return nil
}

// GetFilterStatistics returns filtering statistics.
func (ca *CompleteAdapter) GetFilterStatistics(original, filtered []entities.Repository) map[string]any {
	return ca.filterService.GetFilterStatistics(original, filtered)
}
