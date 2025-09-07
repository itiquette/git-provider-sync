// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package synccmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"

	"itiquette/git-provider-sync/internal/adapters/cli"
	"itiquette/git-provider-sync/internal/adapters/composition"
	"itiquette/git-provider-sync/internal/application/dto"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/domain/sync"
)

// executeAllEnvironmentSyncsWithFormatter executes all syncs using the new formatter.
func executeAllEnvironmentSyncsWithFormatter(
	ctx context.Context,
	logger *zerolog.Logger,
	container *composition.Container,
	cfg *dto.AppConfiguration,
	syncResults *sync.Results,
	formatter ports.SyncOutputFormatter,
) error {
	for envName, environments := range cfg.GitProviderSyncConfs {
		formatter.StartEnvironment(envName)
		syncResults.TotalSources++

		for syncCfgName, syncCfg := range environments {
			syncResults.TotalMirrors += len(syncCfg.Mirrors)

			formatter.StartSync(envName, syncCfgName, syncCfg.ProviderType)

			// Progress feedback for source
			formatter.SourceFetching(syncCfg.ProviderType, syncCfg.Domain, syncCfg.Owner)

			// Only log at debug level to avoid duplicate output
			logger.Debug().
				Str("environment", envName).
				Str("syncConfig", syncCfgName).
				Str("provider", syncCfg.ProviderType).
				Msg("Processing sync configuration")

			startTime := time.Now()

			// Execute the sync
			repoCount, err := executeSyncConfigurationWithResultsAndFormatter(
				ctx, container, syncCfg, envName, syncCfgName, syncResults, formatter,
			)
			if err != nil {
				formatter.Error("Failed to sync configuration", err)
				return fmt.Errorf("failed to sync environment %s, config %s: %w", envName, syncCfgName, err)
			}

			// Report source fetched with the actual count from this sync
			formatter.SourceFetched(repoCount, time.Since(startTime))
		}
	}

	return nil
}

// executeSyncConfigurationWithResultsAndFormatter executes a single sync config with formatter.
// Returns the number of repositories processed and any error.
func executeSyncConfigurationWithResultsAndFormatter(
	ctx context.Context,
	container *composition.Container,
	syncCfg dto.SyncConfig,
	envName string,
	sourceName string,
	results *sync.Results,
	formatter ports.SyncOutputFormatter,
) (int, error) {
	// Properly implement without calling the old function to avoid duplicate output
	logger := zerolog.Ctx(ctx)
	logger.Debug().
		Str("provider", syncCfg.ProviderType).
		Str("domain", syncCfg.Domain).
		Str("owner", syncCfg.Owner).
		Msg("Executing sync configuration")

	// Fetch source repositories
	fetchResponse, err := fetchSourceRepositoriesWithGitRepos(ctx, container, syncCfg)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	// Update results
	results.TotalRepositories += fetchResponse.ProcessedCount

	// Process each mirror
	for mirrorName, mirrorCfg := range syncCfg.Mirrors {
		formatter.MirrorSyncing(mirrorName, mirrorCfg.ProviderType, mirrorCfg.Domain, mirrorCfg.Owner)

		startTime := time.Now()

		// Sync to this mirror
		err := syncToMirrorWithGitReposAndResults(
			ctx, container, fetchResponse, mirrorCfg, syncCfg,
			envName, sourceName, mirrorName, results,
		)

		if err != nil {
			formatter.Error(fmt.Sprintf("Failed to sync to mirror %s", mirrorName), err)
			return 0, fmt.Errorf("failed to sync to mirror %s: %w", mirrorName, err)
		}

		// Report completion (simplified - would track actual stats per mirror)
		formatter.MirrorSynced(mirrorName, 0, 0, 0, time.Since(startTime))
	}

	logger.Debug().Msg("Sync configuration completed successfully")

	return fetchResponse.ProcessedCount, nil
}

// convertSyncResults converts internal sync.Results to formatter-compatible format.
func convertSyncResults(results *sync.Results) ports.SyncResults {
	repos := make([]ports.RepositoryResult, 0)

	// Convert repository results if available
	for _, result := range results.Results {
		repos = append(repos, ports.RepositoryResult{
			Name:         result.Repository,
			Success:      result.Status == "SUCCESS",
			Skipped:      result.Status == "SKIPPED",
			ErrorMessage: result.Error,
			// LastUpdated and Size would need to be tracked in sync.Result
		})
	}

	// Calculate duration
	duration := time.Duration(results.DurationSeconds * float64(time.Second))
	if results.EndTime.IsZero() {
		// If not completed yet, calculate from start time
		duration = time.Since(results.StartTime)
	}

	return ports.SyncResults{
		TotalSources:      results.TotalSources,
		TotalMirrors:      results.TotalMirrors,
		TotalRepositories: results.TotalRepositories,
		SuccessfulSyncs:   results.SuccessfulSyncs,
		FailedSyncs:       results.FailedSyncs,
		SkippedSyncs:      results.SkippedSyncs,
		Duration:          duration,
		DryRun:            results.DryRun,
		Repositories:      repos,
	}
}

// createSyncFormatter creates the appropriate formatter based on configuration.
func createSyncFormatter(ctx context.Context) ports.SyncOutputFormatter {
	// Extract CLI config
	cliConfig, ok := cli.ConfigFromContext(ctx)
	if !ok {
		cliConfig = entities.NewCLIConfigBuilder().Build()
	}

	// Determine log level from logger
	logLevel := "brief"
	if logger := zerolog.Ctx(ctx); logger != nil {
		switch logger.GetLevel() {
		case zerolog.TraceLevel:
			logLevel = "trace"
		case zerolog.DebugLevel:
			logLevel = "debug"
		case zerolog.InfoLevel:
			logLevel = "verbose"
		}
	}

	// Create formatter - using stdout for user-facing output
	factory := cli.NewFormatterFactory()
	return factory.CreateFormatter(cliConfig, logLevel, os.Stdout)
}
