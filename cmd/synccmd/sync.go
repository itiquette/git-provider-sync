// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/logging"
	"itiquette/git-provider-sync/internal/adapters/repository/archive"
	"itiquette/git-provider-sync/internal/adapters/terminal"
	"itiquette/git-provider-sync/internal/composition"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
	"itiquette/git-provider-sync/internal/log"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// PerformSync executes the sync operation using domain use cases.
func performSync(ctx context.Context, cfg *gpsconfig.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Starting sync")
	cfg.DebugLog(logger)

	// Check for cancellation early
	if ctx.Err() != nil {
		return fmt.Errorf("sync cancelled: %w", ctx.Err())
	}

	// Create temporary directory with timestamp for uniqueness
	tmpPrefix := fmt.Sprintf("gitprovidersync-%d", time.Now().Unix())

	ctx, err := entities.CreateTmpDir(ctx, "", tmpPrefix)
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
	if err := executeAllEnvironmentSyncs(ctx, logger, container, cfg, syncResults); err != nil {
		return err
	}

	syncResults.Complete()

	// Output results using the formatter

	if err := outputSyncResults(ctx, syncResults); err != nil {
		logger.Error().Err(err).Msg("Failed to output sync results")
	}

	// Show summary and suggestions
	showSyncSummary(syncResults)

	logger.Info().Msg("All syncs completed successfully")

	return nil
}

// CreateContainerWithConfig builds the dependency injection container
// with already-loaded configuration, avoiding duplicate loading.
func createContainerWithConfig(ctx context.Context, _ *gpsconfig.AppConfiguration) (*composition.Container, error) {
	// Extract CLI options for container configuration
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	// Use the SAME config file path that print command uses
	configPath := cliConfig.ConfigFilePath()
	if configPath == "" {
		configPath = "gitprovidersync.yaml" // Same default as print command
	}

	containerConfig := composition.ContainerConfig{
		ConfigPath:     configPath, // Use the same path as DefaultConfigLoader
		Environment:    "runtime",
		LogLevel:       "info", // Default log level
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
//nolint:cyclop,nestif // Multiple validation and error paths required
func executeSyncConfigurationWithResults(
	ctx context.Context,
	container *composition.Container,
	syncCfg gpsconfig.SyncConfig,
	envName string,
	sourceName string,
	results *sync.Results,
) error {
	logger := log.Logger(ctx)
	logger.Debug().
		Str("provider", syncCfg.ProviderType).
		Str("domain", syncCfg.Domain).
		Str("owner", syncCfg.Owner).
		Msg("Executing sync configuration")

	// Default to auto mode - honors NO_COLOR environment variable
	symbols := cli.GetSymbols(terminal.ColorAuto)

	// Show what we're syncing
	fmt.Fprintf(os.Stderr, "%s Syncing %s/%s (%s)\n",
		symbols.Info, syncCfg.Domain, syncCfg.Owner, syncCfg.ProviderType)

	// Fetch source repositories and sync to mirrors
	fetchResponse, err := fetchSourceRepositoriesWithGitRepos(ctx, container, syncCfg)
	if err != nil {
		return fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	// Show fetch results
	if fetchResponse.ProcessedCount > 0 {
		fmt.Fprintf(os.Stderr, "  %s Fetched %d repositories\n",
			symbols.Check, fetchResponse.ProcessedCount)
	}

	// Display any fetch errors in grouped format
	if len(fetchResponse.Errors) > 0 {
		errorGroup := cli.NewErrorGroup("Clone")

		for i, err := range fetchResponse.Errors {
			// Extract repo name from error if possible
			repoName := fmt.Sprintf("repository-%d", i+1)

			if errStr := err.Error(); strings.Contains(errStr, "failed to clone ") {
				parts := strings.Split(errStr, "failed to clone ")
				if len(parts) > 1 {
					if colonIdx := strings.Index(parts[1], ":"); colonIdx > 0 {
						repoName = parts[1][:colonIdx]
					}
				}
			}

			errorGroup.Add(repoName, err)
		}

		// Display grouped errors
		if errorGroup.HasErrors() {
			fmt.Fprint(os.Stderr, errorGroup.Format(symbols))
		}
	}

	// Process each mirror configuration
	for mirrorName, mirrorCfg := range syncCfg.Mirrors {
		fmt.Fprintf(os.Stderr, "  %s Syncing to %s (%s)\n",
			symbols.Arrow, mirrorName, mirrorCfg.ProviderType)

		if err := syncToMirrorWithGitReposAndResults(ctx, container, fetchResponse, mirrorCfg, syncCfg, envName, sourceName, mirrorName, results); err != nil {
			return fmt.Errorf("failed to sync to mirror %s: %w", mirrorName, err)
		}
	}

	logger.Info().Msg("Sync configuration completed successfully")

	return nil
}

// FetchSourceRepositoriesWithGitRepos fetches and clones repositories returning full response.
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
	mirrorCfg gpsconfig.MirrorConfig,
	srcCfg gpsconfig.SyncConfig,
	envName string,
	sourceName string,
	mirrorName string,
	results *sync.Results,
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

	// 8. Create and execute ToMirrorsUseCase

	// Create archive operations adapter
	archiveOps := archive.NewOperations(loggerAdapter, os.TempDir(), mirrorCfg.Path)

	// Get file system from container
	fileSystem := container.FileSystem()

	syncUseCase := sync.NewToMirrorsUseCase(targetProvider, gitOperations, archiveOps, fileSystem, loggerAdapter)

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

	logger.Info().
		Int("totalRepos", response.TotalRepositories).
		Int("successfulSyncs", response.SuccessfulSyncs).
		Int("failedSyncs", response.FailedSyncs).
		Int("skippedSyncs", response.SkippedSyncs).
		Bool("success", response.Success).
		Msg("Mirror sync completed")

	return nil
}

// ConvertRepositoryFilters converts GPS config filters to domain filters.
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

// ConvertMirrorConfigToMirrorTargets converts GPS mirror config to domain mirror targets.
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

// CreateLoggerAdapter creates a domain logger adapter from zerolog.Logger.
func createLoggerAdapter(logger zerolog.Logger) ports.Logger { //nolint:ireturn // Factory function returning interface
	return logging.NewZerologAdapter(&logger)
}

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

// ExecuteAllEnvironmentSyncs executes sync for all environments and configurations.
func executeAllEnvironmentSyncs(ctx context.Context, logger *zerolog.Logger, container *composition.Container, cfg *gpsconfig.AppConfiguration, syncResults *sync.Results) error {
	for envName, environments := range cfg.GitProviderSyncConfs {
		syncResults.TotalSources++
		for syncCfgName, syncCfg := range environments {
			syncResults.TotalMirrors += len(syncCfg.Mirrors)

			// Progress feedback
			fmt.Fprintf(os.Stderr, "Syncing %s/%s (%s)...", envName, syncCfgName, syncCfg.ProviderType)

			logger.Info().
				Str("environment", envName).
				Str("syncConfig", syncCfgName).
				Str("provider", syncCfg.ProviderType).
				Msg("Processing sync configuration")

			if err := executeSyncConfigurationWithResults(ctx, container, syncCfg, envName, syncCfgName, syncResults); err != nil {
				fmt.Fprintf(os.Stderr, "failed\n")

				return fmt.Errorf("failed to sync environment %s, config %s: %w", envName, syncCfgName, err)
			}

			fmt.Fprintf(os.Stderr, "done\n")
		}
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

// GetLastSyncFilePath returns the path to the last sync state file in temp directory.
func getLastSyncFilePath() string {
	return filepath.Join(os.TempDir(), ".gitprovidersync-last-sync")
}

// SaveLastSyncInfo saves simple sync info to a file for status command.
func saveLastSyncInfo(results *sync.Results) {
	// Write to temp directory instead of current working directory
	content := fmt.Sprintf("timestamp=%d\nrepos=%d\nsuccessful=%d\nfailed=%d\nskipped=%d\n",
		time.Now().Unix(),
		results.TotalRepositories,
		results.SuccessfulSyncs,
		results.FailedSyncs,
		results.SkippedSyncs)

	// Write to file (non-critical, so we don't fail the sync)
	if err := os.WriteFile(getLastSyncFilePath(), []byte(content), 0600); err != nil {
		// Log the error but don't fail the sync operation
		fmt.Fprintf(os.Stderr, "warning: failed to write last sync info: %v\n", err)
	}
}
