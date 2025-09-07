// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
)

// FetchSourceRepositoriesUseCase handles fetching repositories from source providers
// Uses hexagonal architecture with context-based logging (functional pattern).
type FetchSourceRepositoriesUseCase struct {
	repositoryProvider ports.RepositoryProvider
	gitOperations      ports.GitOperations
}

// NewFetchSourceRepositoriesUseCase creates a new fetch source repositories use case
// Uses functional pattern - logger comes from context (idiomatic Go).
func NewFetchSourceRepositoriesUseCase(
	repositoryProvider ports.RepositoryProvider,
	gitOps ports.GitOperations,
) FetchSourceRepositoriesUseCase {
	return FetchSourceRepositoriesUseCase{
		repositoryProvider: repositoryProvider,
		gitOperations:      gitOps,
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
func (uc FetchSourceRepositoriesUseCase) Execute(
	ctx context.Context,
	request FetchSourceRequest,
) (FetchSourceResponse, error) {
	logger := ports.LoggerFromContext(ctx)

	var response FetchSourceResponse

	var err error

	// TRACE: Use case entry point (hexagonal boundary)
	logger.Trace(ctx, "FetchSourceRepositoriesUseCase.Execute entry", map[string]any{
		"provider":      request.ProviderConfig.ProviderType,
		"domain":        request.ProviderConfig.Domain,
		"owner":         request.ProviderConfig.Owner,
		"dry_run":       request.DryRun,
		"include_forks": request.IncludeForks,
	})

	defer func() {
		// TRACE: Use case exit point with outcome
		logger.Trace(ctx, "FetchSourceRepositoriesUseCase.Execute exit", map[string]any{
			"success":      response.Success,
			"processed":    response.ProcessedCount,
			"cloned_count": len(response.ClonedRepos),
			"skipped":      response.SkippedCount,
			"error":        err != nil,
		})
	}()

	// INFO: Immediate feedback before network operation (< 100ms response)
	logger.Info(ctx, "Fetching repository list from provider", map[string]any{
		"provider": request.ProviderConfig.ProviderType,
		"owner":    request.ProviderConfig.Owner,
	})

	// TRACE: Step 1 - Cross to provider adapter
	logger.Trace(ctx, "crossing to provider adapter: ListRepositories", map[string]any{
		"step":     "1_list_repositories",
		"provider": request.ProviderConfig.ProviderType,
	})

	repositories, err := uc.repositoryProvider.ListRepositories(ctx, request.ProviderConfig)
	if err != nil {
		return FetchSourceResponse{}, fmt.Errorf("failed to fetch repositories from provider: %w", err)
	}

	logger.Debug(ctx, "Repository metadata retrieved from provider", map[string]any{
		"total_found": len(repositories),
		"provider":    request.ProviderConfig.ProviderType,
	})

	response = FetchSourceResponse{
		Repositories:   repositories,
		Success:        true,
		ProcessedCount: len(repositories),
		ClonedRepos:    []ports.GitRepository{},
		SkippedCount:   0,
		Errors:         []error{},
	}

	// TRACE: Step 2 - Apply filters
	logger.Trace(ctx, "applying repository filters", map[string]any{
		"step":                "2_apply_filters",
		"include_forks":       request.IncludeForks,
		"has_include_filters": len(request.Filters.IncludePatterns) > 0,
		"has_exclude_filters": len(request.Filters.ExcludePatterns) > 0,
	})

	filteredRepos := uc.applyFilters(ctx, repositories, request.Filters, request.IncludeForks)
	response.Repositories = filteredRepos
	response.SkippedCount = len(repositories) - len(filteredRepos)
	// Update ProcessedCount to reflect the actual repositories that will be processed
	response.ProcessedCount = len(filteredRepos)

	logger.Debug(ctx, "Repository filtering completed", map[string]any{
		"total_found":   len(repositories),
		"after_filters": len(filteredRepos),
		"skipped":       response.SkippedCount,
	})

	// TRACE: Step 3 - Check dry run mode
	if request.DryRun {
		logger.Debug(ctx, "Dry run mode enabled", map[string]any{
			"skipping_clone":     true,
			"repositories_count": len(filteredRepos),
		})

		// TRACE: Dry run repository enumeration
		logger.Trace(ctx, "enumerating repositories for dry run", map[string]any{
			"step": "3_dry_run_enumeration",
		})

		for _, repo := range filteredRepos {
			logger.Debug(ctx, "Would clone repository", map[string]any{
				"name":        repo.Name(),
				"clone_url":   repo.HTTPSURL(),
				"visibility":  repo.Visibility(),
				"description": repo.Description(),
			})
		}

		return response, nil
	}

	// TRACE: Step 4 - Clone repositories (critical git operations)
	logger.Trace(ctx, "cloning repositories from source", map[string]any{
		"step":           "4_clone_repositories",
		"to_clone_count": len(filteredRepos),
	})

	clonedRepos, cloneErrors := uc.cloneRepositories(ctx, filteredRepos, request.ProviderConfig)
	response.ClonedRepos = clonedRepos
	response.Errors = cloneErrors

	logger.Debug(ctx, "Repository cloning completed", map[string]any{
		"successful_clones": len(clonedRepos),
		"failed_clones":     len(cloneErrors),
		"success":           len(cloneErrors) == 0,
	})

	if len(cloneErrors) > 0 {
		response.Success = false

		logger.Debug(ctx, "Some repositories failed to clone", map[string]any{
			"successful_clones": len(clonedRepos),
			"failed_clones":     len(cloneErrors),
			"total_attempted":   len(filteredRepos),
		})
	}

	return response, nil
}

// ApplyFilters applies repository filtering logic with precedence rules
// Filter precedence (performance optimized for sync workflows):
// 1) Fork exclusion - Most repositories are forks, quick elimination saves processing
// 2) Visibility (public/private) - Simple metadata check, eliminates access issues early
// 3) Archive status - Prevents sync attempts on read-only archived repositories
// 4) Custom criteria - Complex business rules applied last to minimize overhead
// Repositories must pass ALL applicable filters to be included in results.
func (uc FetchSourceRepositoriesUseCase) applyFilters(
	ctx context.Context,
	repositories []entities.Repository,
	filters ports.FilterOptions,
	includeForks bool,
) []entities.Repository {
	logger := ports.LoggerFromContext(ctx)

	// TRACE: Internal method entry
	logger.Trace(ctx, "applyFilters entry", map[string]any{
		"total_repositories": len(repositories),
		"include_forks":      includeForks,
	})

	filtered := make([]entities.Repository, 0, len(repositories))

	for _, repo := range repositories {
		if uc.shouldSkipRepo(ctx, repo, filters, includeForks) {
			logger.Debug(ctx, "Repository skipped by filters", map[string]any{
				"name":        repo.Name(),
				"is_fork":     repo.IsFork(),
				"is_private":  repo.IsPrivate(),
				"is_archived": repo.IsArchived(),
			})

			continue
		}

		filtered = append(filtered, repo)
	}

	logger.Debug(ctx, "Repository filtering completed", map[string]any{
		"input_count":  len(repositories),
		"output_count": len(filtered),
		"filtered_out": len(repositories) - len(filtered),
	})

	return filtered
}

// ShouldSkipRepo determines if a repository should be skipped based on filters
//
//nolint:cyclop // Multiple filter conditions create logical complexity
func (uc FetchSourceRepositoriesUseCase) shouldSkipRepo(ctx context.Context, repo entities.Repository, filters ports.FilterOptions, includeForks bool) bool {
	logger := ports.LoggerFromContext(ctx)

	// Check fork exclusion first (performance optimization)
	if repo.IsFork() && !includeForks && !filters.IncludeForks {
		logger.Debug(ctx, "Repository excluded: is fork", map[string]any{
			"name":          repo.Name(),
			"include_forks": includeForks,
		})

		return true
	}

	// Check archive status
	if repo.IsArchived() && !filters.IncludeArchived {
		logger.Debug(ctx, "Repository excluded: is archived", map[string]any{
			"name": repo.Name(),
		})

		return true
	}

	// Check private visibility
	if repo.IsPrivate() && !filters.IncludePrivate {
		logger.Debug(ctx, "Repository excluded: is private", map[string]any{
			"name": repo.Name(),
		})

		return true
	}

	// Check public visibility
	if !repo.IsPrivate() && !filters.IncludePublic {
		logger.Debug(ctx, "Repository excluded: is public", map[string]any{
			"name": repo.Name(),
		})

		return true
	}

	// Check name patterns (most complex check last)
	if !uc.matchesPatterns(ctx, repo.Name(), filters.IncludePatterns, filters.ExcludePatterns) {
		logger.Debug(ctx, "Repository excluded: pattern mismatch", map[string]any{
			"name":             repo.Name(),
			"include_patterns": filters.IncludePatterns,
			"exclude_patterns": filters.ExcludePatterns,
		})

		return true
	}

	return false
}

// MatchesPatterns checks if repository name matches include/exclude patterns.
func (uc FetchSourceRepositoriesUseCase) matchesPatterns(
	ctx context.Context,
	name string,
	includePatterns, excludePatterns []string,
) bool {
	logger := ports.LoggerFromContext(ctx)

	// TRACE: Pattern matching entry
	logger.Trace(ctx, "matchesPatterns entry", map[string]any{
		"name":             name,
		"include_patterns": includePatterns,
		"exclude_patterns": excludePatterns,
	})

	// If no include patterns, assume included
	included := len(includePatterns) == 0

	// Check include patterns
	for _, pattern := range includePatterns {
		if uc.matchPattern(ctx, name, pattern) {
			included = true

			logger.Debug(ctx, "Repository included by pattern", map[string]any{
				"name":    name,
				"pattern": pattern,
			})

			break
		}
	}

	if !included {
		logger.Debug(ctx, "Repository excluded: no include pattern matched", map[string]any{
			"name":             name,
			"include_patterns": includePatterns,
		})

		return false
	}

	// Check exclude patterns
	for _, pattern := range excludePatterns {
		if uc.matchPattern(ctx, name, pattern) {
			logger.Debug(ctx, "Repository excluded by pattern", map[string]any{
				"name":    name,
				"pattern": pattern,
			})

			return false
		}
	}

	return true
}

// MatchPattern performs simple pattern matching (simplified version).
func (uc FetchSourceRepositoriesUseCase) matchPattern(ctx context.Context, name, pattern string) bool {
	logger := ports.LoggerFromContext(ctx)

	// Simple wildcard matching - could be enhanced with proper glob matching
	if pattern == "*" {
		logger.Debug(ctx, "Pattern wildcard match", map[string]any{
			"name":    name,
			"pattern": pattern,
			"result":  true,
		})

		return true
	}

	matches := name == pattern
	logger.Debug(ctx, "Pattern exact match", map[string]any{
		"name":    name,
		"pattern": pattern,
		"result":  matches,
	})

	return matches
}

// CloneRepositories clones the repositories using git operations.
func (uc FetchSourceRepositoriesUseCase) cloneRepositories(
	ctx context.Context,
	repositories []entities.Repository,
	providerConfig ports.ProviderConfig,
) ([]ports.GitRepository, []error) {
	logger := ports.LoggerFromContext(ctx)

	// TRACE: Internal method entry (git operations boundary)
	logger.Trace(ctx, "cloneRepositories entry", map[string]any{
		"repository_count": len(repositories),
		"provider":         providerConfig.ProviderType,
	})

	clonedRepos := make([]ports.GitRepository, 0, len(repositories))
	errors := make([]error, 0)

	for i, repo := range repositories {
		// TRACE: Individual repository clone operation
		logger.Trace(ctx, "cloning individual repository", map[string]any{
			"index":     i + 1,
			"total":     len(repositories),
			"name":      repo.Name(),
			"clone_url": repo.HTTPSURL(),
		})

		logger.Debug(ctx, "Starting repository clone", map[string]any{
			"name":           repo.Name(),
			"clone_url":      repo.HTTPSURL(),
			"default_branch": repo.DefaultBranch(),
			"visibility":     repo.Visibility(),
		})

		// Get temporary directory path for cloning using GitOperations interface
		tmpDir, err := uc.gitOperations.GetTmpDirPath(ctx)
		if err != nil {
			logger.Debug(ctx, "Failed to get temp directory for clone", map[string]any{
				"name":  repo.Name(),
				"error": err.Error(),
			})
			errors = append(errors, fmt.Errorf("failed to get temp directory for %s: %w", repo.Name(), err))

			continue
		}

		clonePath := tmpDir + "/" + repo.Name()
		logger.Debug(ctx, "Prepared clone options", map[string]any{
			"name":       repo.Name(),
			"clone_path": clonePath,
			"tmp_dir":    tmpDir,
		})

		// Create clone options with proper temp directory
		cloneOptions := ports.CloneOptions{
			URL:          repo.HTTPSURL(),
			Path:         clonePath,
			Branch:       repo.DefaultBranch(),
			SingleBranch: false,
			Depth:        0,
			Mirror:       false,
			Bare:         false,
			Progress:     nil,
			Tags:         ports.TagModeAll,
			Timeout:      0,
			Auth: ports.AuthOptions{
				Type:       ports.AuthTypeToken,
				Token:      providerConfig.AuthConfig.Token,
				Username:   providerConfig.AuthConfig.Username,
				Password:   "",
				SSHKeyPath: "",
				SSHKey:     []byte{},
				Passphrase: "",
			},
		}

		// TRACE: Before crossing to git adapter (hexagonal boundary)
		logger.Trace(ctx, "crossing to git adapter: Clone", map[string]any{
			"operation": "Clone",
			"name":      repo.Name(),
			"url":       repo.HTTPSURL(),
		})

		// Check for cancellation before expensive clone operation (idiomatic Go)
		if ctx.Err() != nil {
			errors = append(errors, fmt.Errorf("clone cancelled for %s: %w", repo.Name(), ctx.Err()))

			break
		}

		// Log before network operation
		logger.Info(ctx, "Cloning repository", map[string]any{
			"name": repo.Name(),
			"url":  repo.HTTPSURL(),
		})

		// Clone repository
		gitRepo, err := uc.gitOperations.Clone(ctx, cloneOptions)
		if err != nil {
			logger.Debug(ctx, "Repository clone failed", map[string]any{
				"name":  repo.Name(),
				"url":   repo.HTTPSURL(),
				"error": err.Error(),
			})
			errors = append(errors, fmt.Errorf("failed to clone %s: %w", repo.Name(), err))

			continue
		}

		logger.Debug(ctx, "Repository cloned successfully", map[string]any{
			"name":       repo.Name(),
			"clone_path": clonePath,
		})

		clonedRepos = append(clonedRepos, gitRepo)
	}

	logger.Debug(ctx, "Repository cloning completed", map[string]any{
		"total_attempted":   len(repositories),
		"successful_clones": len(clonedRepos),
		"failed_clones":     len(errors),
		"success_rate":      float64(len(clonedRepos)) / float64(len(repositories)),
	})

	return clonedRepos, errors
}
