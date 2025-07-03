// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"
	"path/filepath"

	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FetchSourceRepositoriesUseCase handles fetching repositories from source providers.
// This ports the sourceRepositories functionality from main branch to hexagonal architecture.
type FetchSourceRepositoriesUseCase struct {
	repositoryProvider ports.RepositoryProvider
	gitOperations      ports.GitOperations
	logger             ports.Logger
}

// NewFetchSourceRepositoriesUseCase creates a new fetch source repositories use case.
func NewFetchSourceRepositoriesUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOps ports.GitOperations,
	logger ports.Logger,
) FetchSourceRepositoriesUseCase {
	return FetchSourceRepositoriesUseCase{
		repositoryProvider: repositoryProvider,
		gitOperations:      gitOps,
		logger:             logger,
	}
}

// FetchSourceRequest represents the input for fetching source repositories.
type FetchSourceRequest struct {
	ProviderConfig ports.ProviderConfig
	DryRun         bool
	IncludeForks   bool
	Filters        ports.FilterOptions
}

// FetchSourceResponse represents the result of fetching source repositories.
type FetchSourceResponse struct {
	Repositories   []entities.Repository
	ClonedRepos    []ports.GitRepository
	Success        bool
	ProcessedCount int
	SkippedCount   int
	Errors         []error
}

// Execute fetches repositories from the source provider.
// This implements the core sourceRepositories logic from main branch.
func (uc FetchSourceRepositoriesUseCase) Execute(
	ctx context.Context,
	request FetchSourceRequest,
) (FetchSourceResponse, error) {
	uc.logger.Info(ctx, "Fetching source repositories", map[string]interface{}{
		"provider":      request.ProviderConfig.ProviderType,
		"domain":        request.ProviderConfig.Domain,
		"owner":         request.ProviderConfig.Owner,
		"dry_run":       request.DryRun,
		"include_forks": request.IncludeForks,
	})

	// Step 1: Fetch repository metadata from provider (equivalent to FetchProjectInfos)
	repositories, err := uc.repositoryProvider.ListRepositories(ctx, request.ProviderConfig)
	if err != nil {
		return FetchSourceResponse{}, fmt.Errorf("failed to fetch repositories from provider: %w", err)
	}

	response := FetchSourceResponse{
		Repositories:   repositories,
		Success:        true,
		ProcessedCount: len(repositories),
	}

	// Apply filters (equivalent to repository filtering in main branch)
	filteredRepos := uc.applyFilters(repositories, request.Filters, request.IncludeForks)
	response.Repositories = filteredRepos
	response.SkippedCount = len(repositories) - len(filteredRepos)

	uc.logger.Info(ctx, "Repository metadata fetched", map[string]interface{}{
		"total_found":   len(repositories),
		"after_filters": len(filteredRepos),
		"skipped":       response.SkippedCount,
	})

	// Step 2: If dry run, log and return without cloning
	if request.DryRun {
		uc.logger.Info(ctx, "Dry run enabled, skipping repository cloning", nil)

		for _, repo := range filteredRepos {
			uc.logger.Debug(ctx, "Would clone repository", map[string]interface{}{
				"name":        repo.Name(),
				"clone_url":   repo.HTTPSURL(),
				"visibility":  repo.Visibility(),
				"description": repo.Description(),
			})
		}

		return response, nil
	}

	// Step 3: Clone repositories (equivalent to provider.Clone)
	clonedRepos, cloneErrors := uc.cloneRepositories(ctx, filteredRepos, request.ProviderConfig)
	response.ClonedRepos = clonedRepos
	response.Errors = cloneErrors

	if len(cloneErrors) > 0 {
		response.Success = false

		uc.logger.Error(ctx, "Some repositories failed to clone", map[string]interface{}{
			"successful_clones": len(clonedRepos),
			"failed_clones":     len(cloneErrors),
		})
	}

	uc.logger.Info(ctx, "Source repository fetching completed", map[string]interface{}{
		"successful_clones": len(clonedRepos),
		"failed_clones":     len(cloneErrors),
		"success":           response.Success,
	})

	return response, nil
}

// applyFilters applies repository filtering logic.
// This ports the filtering functionality from main branch.
func (uc FetchSourceRepositoriesUseCase) applyFilters(
	repositories []entities.Repository,
	filters ports.FilterOptions,
	includeForks bool,
) []entities.Repository {
	filtered := make([]entities.Repository, 0, len(repositories))

	for _, repo := range repositories {
		// Skip forks if not included
		if repo.IsFork() && !includeForks && !filters.IncludeForks {
			continue
		}

		// Skip archived if not included
		if repo.IsArchived() && !filters.IncludeArchived {
			continue
		}

		// Apply visibility filters
		if repo.IsPrivate() && !filters.IncludePrivate {
			continue
		}

		if !repo.IsPrivate() && !filters.IncludePublic {
			continue
		}

		// Language filtering - skip for now as Repository entity doesn't have Language field
		// TODO: Add language field to Repository entity if needed

		// Size filtering - skip for now as Repository entity doesn't have Size field
		// TODO: Add size field to Repository entity if needed

		// Apply pattern filters
		if !uc.matchesPatterns(repo.Name(), filters.IncludePatterns, filters.ExcludePatterns) {
			continue
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// matchesPatterns checks if repository name matches include/exclude patterns.
func (uc FetchSourceRepositoriesUseCase) matchesPatterns(
	name string,
	includePatterns, excludePatterns []string,
) bool {
	// If no include patterns, assume included
	included := len(includePatterns) == 0

	// Check include patterns
	for _, pattern := range includePatterns {
		if uc.matchPattern(name, pattern) {
			included = true

			break
		}
	}

	if !included {
		return false
	}

	// Check exclude patterns
	for _, pattern := range excludePatterns {
		if uc.matchPattern(name, pattern) {
			return false
		}
	}

	return true
}

// matchPattern performs simple pattern matching (simplified version).
func (uc FetchSourceRepositoriesUseCase) matchPattern(name, pattern string) bool {
	// Simple wildcard matching - could be enhanced with proper glob matching
	if pattern == "*" {
		return true
	}

	return name == pattern
}

// cloneRepositories clones the repositories using git operations.
// This ports the provider.Clone functionality to hexagonal architecture.
func (uc FetchSourceRepositoriesUseCase) cloneRepositories(
	ctx context.Context,
	repositories []entities.Repository,
	providerConfig ports.ProviderConfig,
) ([]ports.GitRepository, []error) {
	clonedRepos := make([]ports.GitRepository, 0, len(repositories))

	errors := make([]error, 0)

	for _, repo := range repositories {
		uc.logger.Debug(ctx, "Cloning repository", map[string]interface{}{
			"name":      repo.Name(),
			"clone_url": repo.HTTPSURL(),
		})

		// Get temporary directory path for cloning (from main branch functionality)
		tmpDir, err := filesystem.GetTmpDirPath(ctx)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to get temp directory for %s: %w", repo.Name(), err))

			continue
		}

		// Create clone options with proper temp directory
		cloneOptions := ports.CloneOptions{
			URL:    repo.HTTPSURL(),
			Path:   filepath.Join(tmpDir, repo.Name()),
			Branch: repo.DefaultBranch(),
			Auth: ports.AuthOptions{
				Type:     ports.AuthTypeToken,
				Token:    providerConfig.AuthConfig.Token,
				Username: providerConfig.AuthConfig.Username,
			},
		}

		// Clone repository
		gitRepo, err := uc.gitOperations.Clone(ctx, cloneOptions)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to clone %s: %w", repo.Name(), err))

			continue
		}

		clonedRepos = append(clonedRepos, gitRepo)
	}

	return clonedRepos, errors
}
