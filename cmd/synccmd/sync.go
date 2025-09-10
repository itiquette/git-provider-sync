// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/composition"
	"itiquette/git-provider-sync/internal/adapters/configuration/dto"
	"itiquette/git-provider-sync/internal/adapters/filesystem"
	"itiquette/git-provider-sync/internal/adapters/log"
	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// PerformSync executes the sync operation using domain use cases.
func performSync(ctx context.Context, cfg *dto.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Starting sync")
	cfg.DebugLog(logger)

	// Check for cancellation early
	if ctx.Err() != nil {
		return fmt.Errorf("sync cancelled: %w", ctx.Err())
	}

	// Create temporary directory with timestamp for uniqueness
	tmpPrefix := fmt.Sprintf("gitprovidersync-%d", time.Now().Unix())

	ctx, err := filesystem.CreateTmpDir(ctx, "", tmpPrefix)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// NO CLEANUP - /tmp is managed by OS, exit immediately on Ctrl-C

	// 1. Create dependency injection container (skip config loading since we already have it)
	container, err := createContainerWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize application services: %w", err)
	}

	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			logger.Error().Err(closeErr).Msg("Failed to close container")
		}
	}()

	// Initialize results tracking
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	syncResults := sync.NewResults(cliConfig.DryRun())

	// 2. Execute sync for each environment using proper use cases
	// Create the formatter based on output format
	formatter := createSyncFormatter(ctx)

	// Execute with the formatter (handles all output)
	if err := executeAllEnvironmentSyncsWithFormatter(ctx, logger, container, cfg, syncResults, formatter); err != nil {
		return err
	}

	syncResults.Complete()

	// Show summary with formatter
	formatter.SyncCompleted(convertSyncResults(syncResults))

	// Success is shown by the formatter, no need for additional log
	return nil
}

// CreateContainerWithConfig builds the dependency injection container
// with already-loaded configuration, avoiding duplicate loading.
func createContainerWithConfig(ctx context.Context, _ *dto.AppConfiguration) (*composition.Container, error) {
	// Extract CLI options for container configuration
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	// Extract log level from context (set by initLogger)
	logLevel := "info" // Default

	if lvl := ctx.Value(logLevelKey); lvl != nil {
		if lvlStr, ok := lvl.(string); ok {
			logLevel = lvlStr
		}
	}

	// Use the SAME config file path that print command uses
	configPath := cliConfig.ConfigFilePath()
	if configPath == "" {
		configPath = "gitprovidersync.yaml" // Same default as print command
	}

	containerConfig := composition.ContainerConfig{
		ConfigPath:     configPath, // Use the same path as DefaultConfigLoader
		Environment:    "runtime",
		LogLevel:       logLevel,                 // Use the actual log level from CLI
		OutputFormat:   cliConfig.OutputFormat(), // Pass the output format for logger configuration
		DryRun:         cliConfig.DryRun(),
		SkipTLSVerify:  false, // Set to false for security by default
		MaxConcurrency: 5,     // Default concurrency limit
	}

	// Container requires file path for config loading
	container, err := composition.NewContainer(ctx, containerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize application services: %w", err)
	}

	return container, nil
}

// ExecuteSyncConfigurationWithResults executes sync for a single source-to-mirrors configuration
// ExecuteSyncConfigurationWithResults fetches from source provider and syncs to all configured mirrors
//
// FetchSourceRepositoriesWithGitRepos fetches and clones repositories returning full response.
func fetchSourceRepositoriesWithGitRepos(
	ctx context.Context,
	container *composition.Container,
	syncCfg dto.SyncConfig,
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

	// 4. Create and execute FetchSourceRepositoriesUseCase

	fetchUseCase := sync.NewFetchSourceRepositoriesUseCase(repositoryProvider, gitOperations)

	// Get CLI configuration for dry run setting
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	request := sync.FetchSourceRequest{
		ProviderConfig: providerConfig,
		DryRun:         cliConfig.DryRun(),
		IncludeForks:   true,
		Filters:        convertRepositoryFilters(syncCfg.Repositories),
	}

	response, err := fetchUseCase.Execute(ctx, request)
	if err != nil {
		return sync.FetchSourceResponse{}, fmt.Errorf("failed to execute fetch use case: %w", err)
	}

	return response, nil
}

// SyncToMirrorWithGitReposAndResults syncs repositories to mirrors using domain use cases.
func syncToMirrorWithGitReposAndResults(
	ctx context.Context,
	container *composition.Container,
	fetchResponse sync.FetchSourceResponse,
	mirrorCfg dto.MirrorConfig,
	srcCfg dto.SyncConfig,
	envName string,
	sourceName string,
	mirrorName string,
	results *sync.Results,
) error {
	// Use container's logger which respects suppression settings
	// Don't use log.Logger(ctx) as it bypasses the container's configuration
	logger := log.Logger(ctx) // Container logger integration pending
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

	// 4. Use container's logger which respects suppression settings
	// Don't create a new logger adapter - use the one from container
	loggerAdapter := container.Logger()

	// 5. Convert to mirror targets
	tmpSyncCfg := dto.SyncConfig{Mirrors: map[string]dto.MirrorConfig{"target": mirrorCfg}}
	mirrorTargets := convertMirrorConfigToMirrorTargets(tmpSyncCfg)

	if len(mirrorTargets) == 0 {
		logger.Warn().Msg("No valid mirror targets found")

		return nil
	}

	// 6. Use the actual cloned GitRepositories from fetch response
	if len(fetchResponse.ClonedRepos) == 0 {
		logger.Debug().Msg("No cloned repositories to sync - likely dry run or no repositories found")

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

	// 8. Create and execute ToMirrorsUseCase

	// Create archive operations adapter
	archiveOps := archive.NewOperations(loggerAdapter, os.TempDir(), mirrorCfg.Path)

	// Get file system from container
	fileSystem := container.FileSystem()

	// Get string utils from container
	stringUtils := container.StringUtils()

	syncUseCase := sync.NewToMirrorsUseCase(targetProvider, gitOperations, archiveOps, fileSystem, loggerAdapter, stringUtils)

	// Get CLI configuration for sync options
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	request := sync.ToMirrorsRequest{
		SourceRepositories: fetchResponse.ClonedRepos, // Use actual cloned repos
		MirrorTargets:      mirrorTargets,
		SourceConfig:       sourceConfig, // Correct source config
		DryRun:             cliConfig.DryRun(),
		Options: sync.Options{
			ForcePush:          cliConfig.ForcePush(),
			IgnoreInvalidNames: cliConfig.IgnoreInvalidName(),
			CreateIfNotExists:  true,
			UpdateDescription:  true,
			SyncProtection:     false, // Branch protection sync disabled by default
		},
	}

	response, err := syncUseCase.Execute(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to execute sync to mirrors use case: %w", err)
	}

	// Add results for each repository
	for _, repo := range fetchResponse.Repositories {
		startTime := time.Now()
		result := sync.Result{
			Environment:     envName,
			Source:          sourceName,
			SourceProvider:  srcCfg.ProviderType,
			Repository:      repo.Name(),
			Mirror:          mirrorName,
			MirrorProvider:  mirrorCfg.ProviderType,
			Status:          "SUCCESS",
			Action:          "UPDATED",
			StartTime:       startTime,
			EndTime:         time.Now(),
			DurationSeconds: time.Since(startTime).Seconds(),
		}

		if !response.Success {
			result.Status = "FAILED"
			result.Error = "Sync operation failed"
		}

		results.AddResult(result)

		results.TotalRepositories++
	}

	logger.Debug().
		Int("totalRepos", response.TotalRepositories).
		Int("successfulSyncs", response.SuccessfulSyncs).
		Int("failedSyncs", response.FailedSyncs).
		Int("skippedSyncs", response.SkippedSyncs).
		Bool("success", response.Success).
		Msg("Mirror sync completed")

	return nil
}

// ConvertRepositoryFilters converts GPS config filters to domain filters.
func convertRepositoryFilters(repoConfig dto.RepositoriesOption) ports.FilterOptions {
	return ports.FilterOptions{
		IncludePatterns: repoConfig.Include,
		ExcludePatterns: repoConfig.Exclude,
		IncludeForks:    true,  // Default to include forks
		IncludeArchived: false, // Default to exclude archived
		IncludePrivate:  true,  // Default to include private
		IncludePublic:   true,  // Default to include public
	}
}

// ConvertMirrorConfigToMirrorTargets converts GPS mirror config to domain mirror targets.
func convertMirrorConfigToMirrorTargets(syncCfg dto.SyncConfig) []entities.MirrorTarget {
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

// CreateLoggerAdapter creates a domain logger adapter from zerolog.Logger.

// OutputSyncResults outputs sync results using the configured formatter.
func outputSyncResults(ctx context.Context, results *sync.Results) error {
	// Get CLI configuration
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	// Create output formatter with color mode from CLI config
	colorMode := terminal.ColorMode(cliConfig.ColorMode())
	formatter := cli.NewOutputFormatterWithColorMode(colorMode)

	// Output to stdout (data) and stderr (progress)
	if err := formatter.FormatSyncResults(results, cliConfig.OutputFormat(), os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("failed to format sync results: %w", err)
	}

	return nil
}

// ShowSyncSummary displays sync summary and suggestions.
func showSyncSummary(syncResults *sync.Results) {
	// Add simple summary and suggestion
	if syncResults.SuccessfulSyncs > 0 {
		fmt.Fprintf(os.Stderr, "✓ Successfully synced %d repositories\n", syncResults.SuccessfulSyncs)
	}

	if syncResults.FailedSyncs > 0 {
		fmt.Fprintf(os.Stderr, "✗ %d repositories failed\n", syncResults.FailedSyncs)
	}

	if syncResults.SkippedSyncs > 0 {
		fmt.Fprintf(os.Stderr, "- %d repositories skipped\n", syncResults.SkippedSyncs)
	}

	// Store last sync info (simple file-based approach)
	if !syncResults.DryRun {
		saveLastSyncInfo(syncResults)
	}

	// Simple command suggestion
	if !syncResults.DryRun {
		fmt.Fprintf(os.Stderr, "\nNext: gitprovidersync status\n")
	} else {
		fmt.Fprintf(os.Stderr, "\nNext: gitprovidersync sync (without --dry-run)\n")
	}
}

// SaveLastSyncInfo saves simple sync info to a file for status command.
// Uses the default filesystem writer with os.TempDir for backward compatibility.
func saveLastSyncInfo(results *sync.Results) {
	writer := filesystem.NewSyncInfoWriter(os.TempDir())
	saveLastSyncInfoWithWriter(results, writer)
}

// saveLastSyncInfoWithWriter saves sync info using the provided writer.
// This follows dependency injection pattern for better testability.
func saveLastSyncInfoWithWriter(results *sync.Results, writer ports.SyncInfoWriter) {
	info := ports.SyncInfo{
		Timestamp:         time.Now().Unix(),
		TotalRepositories: results.TotalRepositories,
		SuccessfulSyncs:   results.SuccessfulSyncs,
		FailedSyncs:       results.FailedSyncs,
		SkippedSyncs:      results.SkippedSyncs,
	}

	// Write to file (non-critical, so we don't fail the sync)
	if err := writer.WriteSyncInfo(info); err != nil {
		// Log the error but don't fail the sync operation
		fmt.Fprintf(os.Stderr, "warning: failed to write last sync info: %v\n", err)
	}
}
