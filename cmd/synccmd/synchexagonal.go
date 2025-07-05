// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// sync_hexagonal.go - Proper hexagonal architecture sync implementation
package synccmd

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/adapters/logging"
	"itiquette/git-provider-sync/internal/composition"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// syncHexagonal executes the core sync functionality using proper hexagonal architecture.
// This completely replaces the simplified approach with full domain use cases.
func syncHexagonal(ctx context.Context, cfg *gpsconfig.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Starting hexagonal sync")
	cfg.DebugLog(logger)

	// CRITICAL: Create temporary directory (restored from main branch)
	ctx, err := model.CreateTmpDir(ctx, "", "gitprovidersync")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// NOTE: cleanup is commented out in main branch, so preserving that behavior
	// defer cleanup(ctx)

	// 1. Create dependency injection container
	container, err := createContainer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			logger.Error().Err(closeErr).Msg("Failed to close container")
		}
	}()

	// 2. Execute sync for each environment using proper use cases
	for envName, environments := range cfg.GitProviderSyncConfs {
		for syncCfgName, syncCfg := range environments {
			logger.Info().
				Str("environment", envName).
				Str("syncConfig", syncCfgName).
				Str("provider", syncCfg.ProviderType).
				Msg("Processing sync configuration")

			if err := executeSyncConfiguration(ctx, container, syncCfg); err != nil {
				return fmt.Errorf("failed to sync environment %s, config %s: %w", envName, syncCfgName, err)
			}
		}
	}

	logger.Info().Msg("All hexagonal syncs completed successfully")

	return nil
}

// createContainer builds the dependency injection container with proper configuration.
func createContainer(ctx context.Context, cfg *gpsconfig.AppConfiguration) (*composition.Container, error) {
	// Extract CLI options for container configuration
	cliOpts := model.CLIOptions(ctx)

	containerConfig := composition.ContainerConfig{
		ConfigPath:     "", // Already loaded
		Environment:    "runtime",
		LogLevel:       "info", // Default log level
		DryRun:         cliOpts.DryRun,
		SkipTLSVerify:  false, // TODO: Extract from CLI if needed
		MaxConcurrency: 5,     // TODO: Make configurable
	}

	return composition.NewContainer(ctx, containerConfig)
}

// executeSyncConfiguration executes sync for a single source-to-mirrors configuration.
// This implements the core sourceToMirror workflow from main branch using hexagonal architecture.
func executeSyncConfiguration(
	ctx context.Context,
	container *composition.Container,
	syncCfg gpsconfig.SyncConfig,
) error {
	logger := log.Logger(ctx)
	logger.Debug().
		Str("provider", syncCfg.ProviderType).
		Str("domain", syncCfg.Domain).
		Str("owner", syncCfg.Owner).
		Msg("Executing sync configuration")

	// Execute hexagonal sync using domain use cases
	logger.Info().
		Str("provider", syncCfg.ProviderType).
		Str("domain", syncCfg.Domain).
		Str("owner", syncCfg.Owner).
		Msg("Starting hexagonal sync execution")

	// Use the hexagonal fetch and sync approach
	fetchResponse, err := fetchSourceRepositoriesWithGitRepos(ctx, container, syncCfg)
	if err != nil {
		return fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	// Process each mirror configuration
	for mirrorName, mirrorCfg := range syncCfg.Mirrors {
		logger.Info().
			Str("mirror", mirrorName).
			Str("provider", mirrorCfg.ProviderType).
			Msg("Processing mirror configuration")

		if err := syncToMirrorWithGitRepos(ctx, container, fetchResponse, mirrorCfg, syncCfg); err != nil {
			return fmt.Errorf("failed to sync to mirror %s: %w", mirrorName, err)
		}
	}

	logger.Info().Msg("Sync configuration completed successfully")

	return nil
}

// fetchSourceRepositoriesWithGitRepos fetches and clones repositories returning full response.
func fetchSourceRepositoriesWithGitRepos(
	ctx context.Context,
	container *composition.Container,
	syncCfg gpsconfig.SyncConfig,
) (sync.FetchSourceResponse, error) {
	// 1. Create provider configuration
	providerConfig := ports.ProviderConfig{
		ProviderType: syncCfg.ProviderType,
		Domain:       syncCfg.Domain,
		Owner:        syncCfg.Owner,
		AuthConfig: ports.AuthenticationConfig{
			Token: syncCfg.Auth.Token,
		},
	}

	// 2. Create repository provider using factory
	repositoryProvider, err := container.CreateProvider(ctx, providerConfig)
	if err != nil {
		return sync.FetchSourceResponse{}, fmt.Errorf("failed to create provider: %w", err)
	}

	// 3. Create git operations for cloning - respect UseGitBinary configuration
	implementation := "go-git"
	if syncCfg.UseGitBinary {
		implementation = "git-binary"
	}

	gitConfig := ports.GitConfig{
		PreferredImplementation: implementation,
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	gitOperations, err := container.CreateGitOperations(gitConfig)
	if err != nil {
		return sync.FetchSourceResponse{}, fmt.Errorf("failed to create git operations: %w", err)
	}

	// 4. Create logger adapter for use case
	logger := log.Logger(ctx)
	loggerAdapter := createLoggerAdapter(*logger)

	// 5. Create and execute FetchSourceRepositoriesUseCase
	fetchUseCase := sync.NewFetchSourceRepositoriesUseCase(repositoryProvider, gitOperations, loggerAdapter)

	request := sync.FetchSourceRequest{
		ProviderConfig: providerConfig,
		DryRun:         model.CLIOptions(ctx).DryRun,
		IncludeForks:   true,
		Filters:        convertRepositoryFilters(syncCfg.Repositories),
	}

	response, err := fetchUseCase.Execute(ctx, request)
	if err != nil {
		return sync.FetchSourceResponse{}, fmt.Errorf("failed to execute fetch use case: %w", err)
	}

	return response, nil
}

// syncToMirrorWithGitRepos uses the proper hexagonal use case with cloned repositories.
func syncToMirrorWithGitRepos(
	ctx context.Context,
	container *composition.Container,
	fetchResponse sync.FetchSourceResponse,
	mirrorCfg gpsconfig.MirrorConfig,
	srcCfg gpsconfig.SyncConfig,
) error {
	logger := log.Logger(ctx)
	logger.Debug().
		Str("provider", mirrorCfg.ProviderType).
		Str("domain", mirrorCfg.Domain).
		Str("owner", mirrorCfg.Owner).
		Msg("Syncing to mirror target")

	// 1. Create target provider configuration
	targetConfig := ports.ProviderConfig{
		ProviderType: mirrorCfg.ProviderType,
		Domain:       mirrorCfg.Domain,
		Owner:        mirrorCfg.Owner,
		AuthConfig: ports.AuthenticationConfig{
			Token: mirrorCfg.Auth.Token,
		},
	}

	// 2. Create target repository provider
	targetProvider, err := container.CreateProvider(ctx, targetConfig)
	if err != nil {
		return fmt.Errorf("failed to create target provider: %w", err)
	}

	// 3. Create git operations for repository syncing - respect UseGitBinary configuration
	implementation := "go-git"
	if srcCfg.UseGitBinary {
		implementation = "git-binary"
	}

	gitConfig := ports.GitConfig{
		PreferredImplementation: implementation,
		UserName:                "git-provider-sync",
		UserEmail:               "sync@git-provider-sync.local",
		MaxConcurrent:           5,
		VerifySSL:               true,
		Debug:                   false,
	}

	gitOperations, err := container.CreateGitOperations(gitConfig)
	if err != nil {
		return fmt.Errorf("failed to create git operations: %w", err)
	}

	// 4. Create logger adapter for use case
	loggerAdapter := createLoggerAdapter(*logger)

	// 5. Convert to mirror targets
	tmpSyncCfg := gpsconfig.SyncConfig{Mirrors: map[string]gpsconfig.MirrorConfig{"target": mirrorCfg}}
	mirrorTargets := convertMirrorConfigToMirrorTargets(tmpSyncCfg)

	if len(mirrorTargets) == 0 {
		logger.Warn().Msg("No valid mirror targets found")
		return nil
	}

	// 6. Use the actual cloned GitRepositories from fetch response
	if len(fetchResponse.ClonedRepos) == 0 {
		logger.Info().Msg("No cloned repositories to sync - likely dry run or no repositories found")
		return nil
	}

	// 7. Create source provider config
	sourceConfig := ports.ProviderConfig{
		ProviderType: srcCfg.ProviderType,
		Domain:       srcCfg.Domain,
		Owner:        srcCfg.Owner,
		AuthConfig: ports.AuthenticationConfig{
			Token: srcCfg.Auth.Token,
		},
	}

	// 8. Create and execute SyncToMirrorsUseCase
	syncUseCase := sync.NewSyncToMirrorsUseCase(targetProvider, gitOperations, loggerAdapter)

	cliOpts := model.CLIOptions(ctx)
	request := sync.SyncToMirrorsRequest{
		SourceRepositories: fetchResponse.ClonedRepos, // Use actual cloned repos
		MirrorTargets:      mirrorTargets,
		SourceConfig:       sourceConfig, // Correct source config
		DryRun:             cliOpts.DryRun,
		Options: sync.SyncOptions{
			ForcePush:          cliOpts.ForcePush,
			IgnoreInvalidNames: cliOpts.IgnoreInvalidName,
			CreateIfNotExists:  true,
			UpdateDescription:  true,
			SyncProtection:     false, // TODO: Extract from settings
		},
	}

	response, err := syncUseCase.Execute(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to execute sync to mirrors use case: %w", err)
	}

	logger.Info().
		Int("totalRepos", response.TotalRepositories).
		Int("successfulSyncs", response.SuccessfulSyncs).
		Int("failedSyncs", response.FailedSyncs).
		Int("skippedSyncs", response.SkippedSyncs).
		Bool("success", response.Success).
		Msg("Mirror sync completed")

	return nil
}

// convertRepositoryFilters converts GPS config filters to domain filters.
func convertRepositoryFilters(repoConfig gpsconfig.RepositoriesOption) ports.FilterOptions {
	return ports.FilterOptions{
		IncludePatterns: repoConfig.Include,
		ExcludePatterns: repoConfig.Exclude,
		IncludeForks:    true,  // Default to include forks
		IncludeArchived: false, // Default to exclude archived
		IncludePrivate:  true,  // Default to include private
		IncludePublic:   true,  // Default to include public
	}
}

// convertMirrorConfigToMirrorTargets converts GPS mirror config to hexagonal mirror targets.
func convertMirrorConfigToMirrorTargets(syncCfg gpsconfig.SyncConfig) []entities.MirrorTarget {
	targets := make([]entities.MirrorTarget, 0, len(syncCfg.Mirrors))

	for mirrorName, mirrorCfg := range syncCfg.Mirrors {
		// Create auth config from mirror config
		authConfig := entities.NewAuthConfigWithToken(mirrorCfg.Auth.Token, "")

		// Parse provider type
		_, err := entities.ParseProviderType(mirrorCfg.ProviderType)
		if err != nil {
			// Skip invalid provider types
			continue
		}

		// Create mirror target using builder pattern
		builder := entities.NewMirrorTargetBuilder()

		builder, err = builder.WithName(mirrorName)
		if err != nil {
			continue
		}

		builder, err = builder.WithProvider(mirrorCfg.ProviderType)
		if err != nil {
			continue
		}

		builder = builder.WithDomain(mirrorCfg.Domain)

		builder, err = builder.WithOwner(mirrorCfg.Owner)
		if err != nil {
			continue
		}

		builder, err = builder.WithPath(mirrorCfg.Path)
		if err != nil {
			continue
		}

		builder = builder.WithAuth(authConfig)

		// Set options based on mirror settings
		options := entities.MirrorOptions{}
		builder = builder.WithOptions(options)

		target, err := builder.Build()
		if err != nil {
			// Skip invalid targets
			continue
		}

		targets = append(targets, target)
	}

	return targets
}

// createLoggerAdapter creates a domain logger adapter from zerolog.Logger.
func createLoggerAdapter(logger zerolog.Logger) ports.Logger {
	return logging.NewZerologAdapter(&logger)
}
