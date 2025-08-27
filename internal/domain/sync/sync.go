// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/domain"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/log"
)

// RepositoriesUseCase orchestrates the synchronization of repositories from source to mirrors.
// This orchestrates repository synchronization using hexagonal architecture principles.
type RepositoriesUseCase struct {
	configProvider     ports.Configuration
	repositoryProvider ports.RepositoryProvider
	gitOperations      ports.GitOperations
	archiveOperations  ports.ArchiveOperations
	logger             ports.Logger
}

// NewRepositoriesUseCase creates a new sync repositories use case with explicit dependencies.
func NewRepositoriesUseCase(
	configProvider ports.Configuration,
	repositoryProvider ports.RepositoryProvider,
	gitOps ports.GitOperations,
	archiveOps ports.ArchiveOperations,
	logger ports.Logger,
) RepositoriesUseCase {
	return RepositoriesUseCase{
		configProvider:     configProvider,
		repositoryProvider: repositoryProvider,
		gitOperations:      gitOps,
		archiveOperations:  archiveOps,
		logger:             logger,
	}
}

// Request represents the input for sync operation.
type Request struct {
	ConfigPath  string
	Environment string
	DryRun      bool
}

// Response represents the result of sync operation.
type Response struct {
	Success         bool
	ProcessedRepos  int
	SuccessfulSyncs int
	FailedSyncs     int
	Errors          []error
	Duration        string
}

// Execute performs the repository synchronization operation.
// This implements the core sync logic using the sourceToMirror workflow pattern.
func (uc RepositoriesUseCase) Execute(ctx context.Context, request Request) (Response, error) {
	logger := log.CreateDomainLogger(ctx)

	var response Response

	var err error

	// Check for cancellation early (idiomatic Go)
	if ctx.Err() != nil {
		return Response{}, fmt.Errorf("sync cancelled before start: %w", ctx.Err())
	}

	// TRACE: Use case entry point (hexagonal boundary)
	logger.Trace(ctx, "RepositoriesUseCase.Execute entry", map[string]interface{}{
		"config_path": request.ConfigPath,
		"environment": request.Environment,
		"dry_run":     request.DryRun,
	})

	defer func() {
		// TRACE: Use case exit point with outcome
		logger.Trace(ctx, "RepositoriesUseCase.Execute exit", map[string]interface{}{
			"success":          response.Success,
			"processed_repos":  response.ProcessedRepos,
			"successful_syncs": response.SuccessfulSyncs,
			"failed_syncs":     response.FailedSyncs,
			"error":            err != nil,
		})
	}()

	logger.Info(ctx, "Starting repository synchronization", map[string]interface{}{
		"config_path": request.ConfigPath,
		"environment": request.Environment,
		"dry_run":     request.DryRun,
	})

	// Create temporary directory for sync operations
	ctx, err = uc.gitOperations.CreateTmpDir(ctx, "", "gitprovidersync")
	if err != nil {
		return Response{}, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Load configuration
	configSource := ports.ConfigurationSource{
		Location: request.ConfigPath,
		Type:     ports.SourceTypeFile,
		Required: true,
	}

	config, err := uc.configProvider.Load(ctx, configSource)
	if err != nil {
		return Response{}, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if any environments are configured before validation
	if len(config.Environments) == 0 {
		return Response{}, domain.ErrNoEnvironmentsConfigured
	}

	// Extract source configuration from first enabled environment for validation
	var sourceConfig ports.ProviderConfig

	for _, env := range config.Environments {
		if env.Enabled {
			sourceConfig = ports.ProviderConfig{
				ProviderType: env.Source.ProviderType,
				Domain:       env.Source.Domain,
				Owner:        env.Source.Owner,
				AuthConfig: ports.AuthenticationConfig{
					Token:      env.Source.Authentication.Token,
					Username:   env.Source.Authentication.Username,
					SSHKeyPath: env.Source.Authentication.SSHKeyPath,
					SSHKey:     env.Source.Authentication.SSHKey,
				},
			}

			break
		}
	}

	// Validate configuration using the existing validation infrastructure
	if err := uc.validateSourceConfig(ctx, sourceConfig); err != nil {
		return Response{}, fmt.Errorf("source configuration validation failed: %w", err)
	}

	// Execute sync for each environment
	response = Response{Success: true}

	// Initialize sync run metadata tracking
	metadata := entities.NewSyncRunMetadata("source", "mirrors", "sync", "default")
	metadata.SetTotalRepositories(len(config.Environments))
	ctx = entities.AddMetadataToContext(ctx, metadata)

	// Implementation of sourceToMirror workflow pattern
	// Core workflow: fetch source repos → sync to each mirror
	if err := uc.executeSourceToMirror(ctx, config, request, &response); err != nil {
		return response, fmt.Errorf("sync execution failed: %w", err)
	}

	logger.Info(ctx, "Repository synchronization completed", map[string]interface{}{
		"processed_repos":  response.ProcessedRepos,
		"successful_syncs": response.SuccessfulSyncs,
		"failed_syncs":     response.FailedSyncs,
		"success":          response.Success,
	})

	return response, nil
}

// executeSourceToMirror implements the core sync logic using hexagonal architecture.
// This follows the sourceToMirror workflow pattern with proper dependency injection.
func (uc RepositoriesUseCase) executeSourceToMirror(
	ctx context.Context,
	config ports.AppConfiguration,
	_ Request,
	response *Response,
) error {
	logger := log.CreateDomainLogger(ctx)

	// TRACE: Internal orchestration method entry
	logger.Trace(ctx, "executeSourceToMirror entry", map[string]interface{}{
		"environments_count": len(config.Environments),
	})

	logger.Info(ctx, "Executing source to mirror synchronization", nil)

	// Step 1: Extract environments from config
	environments := config.Environments
	if len(environments) == 0 {
		return domain.ErrNoEnvironmentsConfigured
	}

	for envName, env := range environments {
		if !env.Enabled {
			logger.Info(ctx, "Skipping disabled environment", map[string]interface{}{
				"environment": envName,
			})

			continue
		}

		// TRACE: Environment processing boundary
		logger.Trace(ctx, "processing environment", map[string]interface{}{
			"environment": envName,
			"enabled":     env.Enabled,
		})

		logger.Info(ctx, "Processing environment", map[string]interface{}{
			"environment": envName,
		})

		err := uc.processEnvironment(ctx, envName, env, response)
		if err != nil {
			return fmt.Errorf("failed to process environment %s: %w", envName, err)
		}
	}

	return nil
}

// processEnvironment processes a single environment (equivalent to sourceToMirror per environment).
func (uc RepositoriesUseCase) processEnvironment(
	ctx context.Context,
	envName string,
	env ports.EnvironmentConfiguration,
	response *Response,
) error {
	logger := log.CreateDomainLogger(ctx)

	// TRACE: Per-environment processing entry
	logger.Trace(ctx, "processEnvironment entry", map[string]interface{}{
		"environment": envName,
		"provider":    env.Source.ProviderType,
		"mirrors":     len(env.Mirrors),
	})
	// Step 1: Create source provider config
	sourceConfig := env.Source

	// Step 2: Convert source config to provider config
	providerConfig := uc.convertSourceToProviderConfig(sourceConfig)

	// Fetch source repositories using our ported use case
	fetchUseCase := NewFetchSourceRepositoriesUseCase(
		uc.repositoryProvider,
		uc.gitOperations,
	)

	fetchRequest := FetchSourceRequest{
		ProviderConfig: providerConfig,
		DryRun:         false, // Repository fetching doesn't support dry run
		IncludeForks:   uc.extractIncludeForks(env),
		Filters:        uc.convertToFilterOptions(env),
	}

	fetchResponse, err := fetchUseCase.Execute(ctx, fetchRequest)
	if err != nil {
		return fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	if len(fetchResponse.ClonedRepos) == 0 && !fetchRequest.DryRun {
		logger.Warn(ctx, "No repositories to sync", map[string]interface{}{
			"environment": envName,
		})

		return nil
	}

	// TRACE: Step boundary - fetch completed, starting sync
	logger.Trace(ctx, "fetch completed, starting sync to mirrors", map[string]interface{}{
		"environment":   envName,
		"cloned_repos":  len(fetchResponse.ClonedRepos),
		"total_mirrors": len(env.Mirrors),
	})

	response.ProcessedRepos += fetchResponse.ProcessedCount

	// Step 3: Convert mirrors to mirror targets
	mirrorTargets := uc.convertMirrorsToTargets(env.Mirrors)

	// Sync to all mirrors using our ported use case
	syncToMirrorsUseCase := NewToMirrorsUseCase(
		uc.repositoryProvider,
		uc.gitOperations,
		uc.archiveOperations,
		uc.logger,
	)

	syncRequest := ToMirrorsRequest{
		SourceRepositories: fetchResponse.ClonedRepos,
		MirrorTargets:      mirrorTargets,
		SourceConfig:       providerConfig,
		DryRun:             fetchRequest.DryRun,
		Options: Options{
			ForcePush:          false, // Default to false for safety (CLI option)
			IgnoreInvalidNames: false, // Default to false (CLI option)
			CreateIfNotExists:  true,
			UpdateDescription:  true,
		},
	}

	syncResponse, err := syncToMirrorsUseCase.Execute(ctx, syncRequest)
	if err != nil {
		return fmt.Errorf("failed to sync to mirrors: %w", err)
	}

	// Update response statistics
	response.SuccessfulSyncs += syncResponse.SuccessfulSyncs
	response.FailedSyncs += syncResponse.FailedSyncs
	response.Errors = append(response.Errors, syncResponse.Errors...)

	if syncResponse.FailedSyncs > 0 {
		response.Success = false
	}

	logger.Info(ctx, "Environment processing completed", map[string]interface{}{
		"environment":      envName,
		"successful_syncs": syncResponse.SuccessfulSyncs,
		"failed_syncs":     syncResponse.FailedSyncs,
	})

	return nil
}

// Adapter functions to convert between different type systems

// convertSourceToProviderConfig converts SourceConfiguration to ProviderConfig.
func (uc RepositoriesUseCase) convertSourceToProviderConfig(source ports.SourceConfiguration) ports.ProviderConfig {
	return ports.ProviderConfig{
		ProviderType: source.ProviderType,
		Domain:       source.Domain,
		Owner:        source.Owner,
		AuthConfig: ports.AuthenticationConfig{
			Token:      source.Authentication.Token,
			Username:   source.Authentication.Username,
			SSHKeyPath: source.Authentication.SSHKeyPath,
			SSHKey:     source.Authentication.SSHKey,
		},
	}
}

// convertToFilterOptions converts environment configuration to FilterOptions.
func (uc RepositoriesUseCase) convertToFilterOptions(env ports.EnvironmentConfiguration) ports.FilterOptions {
	return ports.FilterOptions{
		IncludePatterns: env.Source.Repository.IncludePatterns,
		ExcludePatterns: env.Source.Repository.ExcludePatterns,
		IncludeForks:    env.Source.Repository.IncludeForks,
		IncludeArchived: env.Source.Repository.IncludeArchived,
		IncludePrivate:  env.Source.Repository.IncludePrivate,
		IncludePublic:   true, // Default to true if not specified
		Languages:       env.Source.Filtering.Languages,
		MinSize:         env.Source.Filtering.MinSize,
		MaxSize:         env.Source.Filtering.MaxSize,
		ActiveSince:     env.Source.Filtering.ActiveSince,
		InactiveSince:   env.Source.Filtering.InactiveSince,
	}
}

// convertMirrorsToTargets converts mirror configuration map to MirrorTarget slice.
func (uc RepositoriesUseCase) convertMirrorsToTargets(mirrors map[string]ports.MirrorConfiguration) []entities.MirrorTarget {
	targets := make([]entities.MirrorTarget, 0, len(mirrors))

	for name, mirror := range mirrors {
		if !mirror.Enabled {
			continue
		}

		// Convert provider type string to enum
		providerType := uc.convertProviderType(mirror.ProviderType)

		// Create authentication config
		authConfig := entities.NewAuthenticationConfig(
			entities.AuthTypeToken, // Default to token auth
			mirror.Authentication.Token,
			mirror.Authentication.Username,
			mirror.Authentication.SSHKeyPath,
			mirror.Authentication.SSHKey,
		)

		// Create mirror target
		target := entities.NewMirrorTarget(
			name,
			providerType,
			mirror.Domain,
			mirror.Owner,
			mirror.Path,
			authConfig,
			true, // enabled (already filtered above)
		)

		targets = append(targets, target)
	}

	return targets
}

// validateSourceConfig performs basic validation of source configuration without external calls.
// Basic validation: checks required fields are present but doesn't verify connectivity or credentials.
// For full validation (including network connectivity), use the ValidationService.
func (uc RepositoriesUseCase) validateSourceConfig(ctx context.Context, config ports.ProviderConfig) error {
	logger := log.CreateDomainLogger(ctx)

	// TRACE: Validation boundary entry
	logger.Trace(ctx, "validateSourceConfig entry", map[string]interface{}{
		"provider": config.ProviderType,
		"domain":   config.Domain,
		"owner":    config.Owner,
	})
	// Basic validation that doesn't require external dependencies
	if config.ProviderType == "" {
		return errors.New("provider type is required")
	}

	if config.Owner == "" {
		return errors.New("owner is required")
	}

	if config.Domain == "" {
		return errors.New("domain is required")
	}

	// Check for supported provider types
	validProviders := []string{"github", "gitlab", "gitea"}

	var validProvider bool

	for _, provider := range validProviders {
		if config.ProviderType == provider {
			validProvider = true

			break
		}
	}

	if !validProvider {
		return fmt.Errorf("unsupported provider type: %s (supported: %v)", config.ProviderType, validProviders)
	}

	logger.Debug(ctx, "Source configuration validation passed", map[string]interface{}{
		"provider": config.ProviderType,
		"domain":   config.Domain,
		"owner":    config.Owner,
	})

	return nil
}

// Helper methods to extract values from environment configuration

func (uc RepositoriesUseCase) extractIncludeForks(env ports.EnvironmentConfiguration) bool {
	return env.Source.Repository.IncludeForks
}

func (uc RepositoriesUseCase) convertProviderType(providerType string) entities.ProviderType {
	switch strings.ToLower(providerType) {
	case "github":
		return entities.ProviderTypeGitHub
	case "gitlab":
		return entities.ProviderTypeGitLab
	case "gitea":
		return entities.ProviderTypeGitea
	case "directory":
		return entities.ProviderTypeDirectory
	case "archive":
		return entities.ProviderTypeArchive
	default:
		return entities.ProviderTypeGitHub // Default fallback
	}
}
