// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"fmt"

	"gitlab.com/gitlab-org/api/client-go"

	"itiquette/git-provider-sync/internal/domain/constants"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// CompleteAdapter provides a comprehensive GitLab adapter with all services.
// This integrates all the sophisticated functionality from the main branch.
type CompleteAdapter struct {
	*Adapter // Embed the basic adapter

	// Sophisticated service layers
	projectService    *ProjectService
	protectionService *ProtectionService
	filterService     *FilterService
	optionsBuilder    *AdvancedProjectOptionsBuilder

	client *gitlab.Client
	logger ports.Logger
}

// NewCompleteAdapter creates a new complete GitLab adapter with all services.
func NewCompleteAdapter(ctx context.Context, config Config, logger ports.Logger) (*CompleteAdapter, error) {
	// Create basic adapter
	basicAdapter, err := NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create basic GitLab adapter: %w", err)
	}

	// Create client using same configuration
	var options []gitlab.ClientOptionFunc

	if config.BaseURL != "" {
		options = append(options, gitlab.WithBaseURL(config.BaseURL))
	}

	if config.HTTPClient != nil {
		options = append(options, gitlab.WithHTTPClient(config.HTTPClient))
	}

	client, err := gitlab.NewClient(config.Token, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Create service layers
	projectService := NewProjectService(client, logger)
	protectionService := NewProtectionService(client, logger)
	filterService := NewFilterService(logger)
	optionsBuilder := NewAdvancedProjectOptionsBuilder()

	return &CompleteAdapter{
		Adapter:           basicAdapter,
		projectService:    projectService,
		protectionService: protectionService,
		filterService:     filterService,
		optionsBuilder:    optionsBuilder,
		client:            client,
		logger:            logger,
	}, nil
}

// CreateRepositoryWithAdvancedOptions creates a repository with sophisticated GitLab options.
func (ca *CompleteAdapter) CreateRepositoryWithAdvancedOptions(
	ctx context.Context,
	config ports.ProviderConfig,
	options CreateProjectRequest,
) (*entities.Repository, error) {
	ca.logger.Info(ctx, "Creating GitLab repository with advanced options", map[string]interface{}{
		"owner":      config.Owner,
		"name":       options.Name,
		"visibility": options.Visibility,
		"is_org":     options.IsOrganization,
	})

	// Validate repository name using advanced validation
	cleanName, isValid, issues := ValidateAndCleanGitLabName(options.Name)
	if !isValid {
		ca.logger.Warn(ctx, "Repository name validation issues", map[string]interface{}{
			"original_name": options.Name,
			"clean_name":    cleanName,
			"issues":        issues,
		})

		// Use the cleaned name
		options.Name = cleanName
	}

	// Create repository using project service
	repository, err := ca.projectService.CreateProject(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	ca.logger.Info(ctx, "GitLab repository created successfully", map[string]interface{}{
		"owner": config.Owner,
		"name":  options.Name,
		"url":   repository.HTTPSURL(),
	})

	return repository, nil
}

// ApplyRepositoryProtection applies comprehensive protection to a GitLab repository.
func (ca *CompleteAdapter) ApplyRepositoryProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryName string,
	protectionOptions ProtectRepositoryRequest,
) error {
	ca.logger.Info(ctx, "Applying GitLab repository protection", map[string]interface{}{
		"owner":      config.Owner,
		"repository": repositoryName,
	})

	protectionOptions.Owner = config.Owner
	protectionOptions.RepositoryName = repositoryName

	return ca.protectionService.ProtectRepository(ctx, protectionOptions)
}

// RemoveRepositoryProtection removes protection from a GitLab repository.
func (ca *CompleteAdapter) RemoveRepositoryProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryName string,
) error {
	ca.logger.Info(ctx, "Removing GitLab repository protection", map[string]interface{}{
		"owner":      config.Owner,
		"repository": repositoryName,
	})

	return ca.protectionService.UnprotectRepository(ctx, config.Owner, repositoryName)
}

// FilterRepositoriesWithAdvancedCriteria filters repositories using sophisticated GitLab criteria.
func (ca *CompleteAdapter) FilterRepositoriesWithAdvancedCriteria(
	ctx context.Context,
	repositories []entities.Repository,
	filterOptions FilterRepositoriesRequest,
) ([]entities.Repository, error) {
	ca.logger.Debug(ctx, "Filtering GitLab repositories with advanced criteria", map[string]interface{}{
		"total_count":    len(repositories),
		"visibility":     filterOptions.VisibilityFilter,
		"topics":         filterOptions.TopicFilter,
		"archive_filter": filterOptions.ArchiveFilter != nil,
	})

	return ca.filterService.FilterRepositories(ctx, filterOptions)
}

// ValidateAndTransformRepositoryName validates and transforms a GitLab repository name.
func (ca *CompleteAdapter) ValidateAndTransformRepositoryName(
	name string,
	options ports.NameTransformOptions,
) (string, error) {
	// First validate the original name using advanced validation
	cleanName, isValid, issues := ValidateAndCleanGitLabName(name)

	if !isValid {
		ca.logger.Info(context.Background(), "Repository name validation failed", map[string]interface{}{
			"original_name": name,
			"clean_name":    cleanName,
			"issues":        issues,
		})

		// Use the cleaned name as starting point
		name = cleanName
	}

	// Apply additional transformations using project service
	transformed := ca.projectService.TransformProjectName(name, options)

	// Validate the final transformed name
	if err := ca.projectService.ValidateProjectName(transformed); err != nil {
		return "", fmt.Errorf("name validation failed even after transformation: %w", err)
	}

	return transformed, nil
}

// GetRepositoryStatistics returns detailed statistics about GitLab repositories.
func (ca *CompleteAdapter) GetRepositoryStatistics(
	ctx context.Context,
	repositories []entities.Repository,
) map[string]interface{} {
	stats := map[string]interface{}{
		"total_count": len(repositories),
	}

	// Count by visibility (GitLab has 3 visibility levels)
	publicCount := 0
	privateCount := 0
	internalCount := 0
	archivedCount := 0
	forkCount := 0

	for _, repo := range repositories {
		switch repo.Visibility() {
		case constants.VisibilityPrivate:
			privateCount++
		case constants.VisibilityInternal:
			internalCount++
		default:
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
	stats["internal_count"] = internalCount
	stats["archived_count"] = archivedCount
	stats["fork_count"] = forkCount
	stats["original_count"] = len(repositories) - forkCount

	return stats
}

// BulkApplyProtection applies protection to multiple GitLab repositories.
func (ca *CompleteAdapter) BulkApplyProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryNames []string,
	protectionOptions ProtectRepositoryRequest,
) error {
	ca.logger.Info(ctx, "Applying bulk GitLab repository protection", map[string]interface{}{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	var errors []error

	for _, repoName := range repositoryNames {
		if err := ca.ApplyRepositoryProtection(ctx, config, repoName, protectionOptions); err != nil {
			ca.logger.Error(ctx, "Failed to apply protection to repository", map[string]interface{}{
				"owner":      config.Owner,
				"repository": repoName,
				"error":      err.Error(),
			})

			errors = append(errors, fmt.Errorf("failed to protect %s: %w", repoName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("protection failed for %d repositories: %v", len(errors), errors)
	}

	ca.logger.Info(ctx, "Bulk GitLab repository protection completed successfully", map[string]interface{}{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	return nil
}

// BulkRemoveProtection removes protection from multiple GitLab repositories.
func (ca *CompleteAdapter) BulkRemoveProtection(
	ctx context.Context,
	config ports.ProviderConfig,
	repositoryNames []string,
) error {
	ca.logger.Info(ctx, "Removing bulk GitLab repository protection", map[string]interface{}{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	var errors []error

	for _, repoName := range repositoryNames {
		if err := ca.RemoveRepositoryProtection(ctx, config, repoName); err != nil {
			ca.logger.Error(ctx, "Failed to remove protection from repository", map[string]interface{}{
				"owner":      config.Owner,
				"repository": repoName,
				"error":      err.Error(),
			})

			errors = append(errors, fmt.Errorf("failed to unprotect %s: %w", repoName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("protection removal failed for %d repositories: %v", len(errors), errors)
	}

	ca.logger.Info(ctx, "Bulk GitLab repository protection removal completed successfully", map[string]interface{}{
		"owner":            config.Owner,
		"repository_count": len(repositoryNames),
	})

	return nil
}

// GetFilterStatistics returns GitLab filtering statistics.
func (ca *CompleteAdapter) GetFilterStatistics(original, filtered []entities.Repository) map[string]interface{} {
	return ca.filterService.GetFilterStatistics(original, filtered)
}

// GetProjectInfos retrieves detailed project information with GitLab-specific features.
func (ca *CompleteAdapter) GetProjectInfos(
	ctx context.Context,
	owner string,
	isOrganization bool,
	includeForks bool,
) ([]*entities.Repository, error) {
	ca.logger.Debug(ctx, "Fetching GitLab project infos", map[string]interface{}{
		"owner":          owner,
		"isOrganization": isOrganization,
		"includeForks":   includeForks,
	})

	// Use the basic adapter to get repositories
	config := ports.ProviderConfig{Owner: owner}

	repos, err := ca.ListRepositories(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	// Filter out forks if not requested
	var result []*entities.Repository

	for _, repo := range repos {
		if !includeForks && repo.IsFork() {
			continue
		}

		// Make a copy of the repository
		repoCopy := repo
		result = append(result, &repoCopy)
	}

	ca.logger.Info(ctx, "Successfully fetched GitLab project infos", map[string]interface{}{
		"owner":             owner,
		"totalRepositories": len(repos),
		"filteredCount":     len(result),
		"includeForks":      includeForks,
	})

	return result, nil
}

// ResolveNamespace resolves a GitLab namespace ID from owner name.
func (ca *CompleteAdapter) ResolveNamespace(ctx context.Context, owner string) (int, error) {
	ca.logger.Debug(ctx, "Resolving GitLab namespace", map[string]interface{}{
		"owner": owner,
	})

	// Search for namespace
	namespaces, _, err := ca.client.Namespaces.ListNamespaces(&gitlab.ListNamespacesOptions{
		Search: gitlab.Ptr(owner),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return 0, fmt.Errorf("failed to search namespaces: %w", err)
	}

	// Find exact match
	for _, ns := range namespaces {
		if ns.Path == owner {
			ca.logger.Debug(ctx, "Namespace resolved successfully", map[string]interface{}{
				"owner":        owner,
				"namespace_id": ns.ID,
				"namespace":    ns.FullPath,
			})

			return ns.ID, nil
		}
	}

	return 0, fmt.Errorf("namespace not found: %s", owner)
}

// GetAdvancedProjectOptions creates sophisticated project options for repository creation.
func (ca *CompleteAdapter) GetAdvancedProjectOptions(
	visibility, name, description, defaultBranch string,
	namespaceID int,
	disableFeatures bool,
) *gitlab.CreateProjectOptions {
	ca.optionsBuilder.Reset()
	ca.optionsBuilder.WithBasicOpts(visibility, name, description, defaultBranch, namespaceID)

	if disableFeatures {
		ca.optionsBuilder.WithDisabledFeatures()
	} else {
		ca.optionsBuilder.WithEnabledFeatures()
	}

	return ca.optionsBuilder.Build()
}

// FilterRepositoriesByNamespace filters repositories by GitLab namespace.
func (ca *CompleteAdapter) FilterRepositoriesByNamespace(
	ctx context.Context,
	repos []entities.Repository,
	namespaces []string,
) []entities.Repository {
	return ca.filterService.FilterRepositoriesByNamespace(ctx, repos, namespaces)
}

// FilterRepositoriesByAccessLevel filters repositories by user access level.
func (ca *CompleteAdapter) FilterRepositoriesByAccessLevel(
	ctx context.Context,
	repos []entities.Repository,
	minAccessLevel string,
) []entities.Repository {
	return ca.filterService.FilterRepositoriesByAccessLevel(ctx, repos, minAccessLevel)
}

// FilterRepositoriesByCompliance filters repositories by compliance frameworks.
func (ca *CompleteAdapter) FilterRepositoriesByCompliance(
	ctx context.Context,
	repos []entities.Repository,
	frameworks []string,
) []entities.Repository {
	return ca.filterService.FilterRepositoriesByCompliance(ctx, repos, frameworks)
}
