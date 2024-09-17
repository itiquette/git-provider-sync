// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/configuration"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider"
	"itiquette/git-provider-sync/internal/target"

	"github.com/spf13/cobra"
)

// ErrTargetRepositoryName is returned when a target repository name is invalid.
var ErrTargetRepositoryName = errors.New("failed target repository name validation")

// newSyncCommand creates and returns a new cobra.Command for the 'sync' subcommand.
func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Mirror repositories from a source Git provider to targets",
		Long: `The 'sync' command mirrors your repositories from a source Git provider to one or more targets.
It allows for various options to control the synchronization process.`,
		Run: runSync,
	}

	cmd.Flags().Bool("force-push", false, "Overwrite any existing target")
	cmd.Flags().Bool("ignore-invalid-name", false, "Ignore repositories with invalid names")
	cmd.Flags().Bool("cleanup-name", false, "Remove non-alphanumeric characters from repository names")
	cmd.Flags().String("active-from-limit", "", "A negative time duration (e.g., '-1h') to consider repositories active from")
	cmd.Flags().Bool("dry-run", false, "Simulate sync run without performing clone and push actions")

	return cmd
}

// runSync executes the sync command logic.
func runSync(cmd *cobra.Command, _ []string) {
	ctx := cmd.Root().Context()
	ctx = addInputOptionsToContext(ctx, cmd)

	withCaller := model.CLIOptions(ctx).VerbosityWithCaller
	outputFormat := model.CLIOptions(ctx).OutputFormat
	ctx = log.InitLogger(ctx, cmd, withCaller, outputFormat)

	var configLoaderInstance configuration.ConfigLoader = configuration.DefaultConfigLoader{}
	config, err := configLoaderInstance.LoadConfiguration(ctx)
	model.HandleError(ctx, err)

	err = sync(ctx, config)
	model.HandleError(ctx, err)
}

// addInputOptionsToContext adds command-line flags to the context.
func addInputOptionsToContext(ctx context.Context, cmd *cobra.Command) context.Context {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering addInputFlagsToContext:")

	ctx = addRootInputOptionsToContext(ctx, cmd)

	forcePush, err := cmd.Flags().GetBool("force-push")
	model.HandleError(ctx, err)

	ignoreInvalidName, err := cmd.Flags().GetBool("ignore-invalid-name")
	model.HandleError(ctx, err)

	cleanupName, err := cmd.Flags().GetBool("cleanup-name")
	model.HandleError(ctx, err)

	activeFromLimit, err := cmd.Flags().GetString("active-from-limit")
	model.HandleError(ctx, err)

	dryRun, err := cmd.Flags().GetBool("dry-run")
	model.HandleError(ctx, err)

	cliOption := model.CLIOptions(ctx)

	cliOption.ForcePush = forcePush
	cliOption.IgnoreInvalidName = ignoreInvalidName
	cliOption.CleanupName = cleanupName
	cliOption.DryRun = dryRun
	cliOption.ActiveFromLimit = activeFromLimit

	return model.WithCLIOption(ctx, cliOption)
}

// sync performs the main synchronization process.
func sync(ctx context.Context, conf *configuration.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sync:")
	conf.DebugLog(logger)

	ctx, err := model.CreateTmpDir(ctx, "", "gitprovidersync")
	if err != nil {
		return fmt.Errorf("failed to create a temporary directory: %w", err)
	}
	defer cleanup(ctx)

	for _, config := range conf.Configurations {
		repositories, err := sourceRepositories(ctx, config.SourceProvider)
		if err != nil {
			return fmt.Errorf("failed to fetch the source repositories: %w", err)
		}

		for _, targetProvider := range config.ProviderTargets {
			if err := toTarget(ctx, config.SourceProvider, targetProvider, repositories); err != nil {
				return fmt.Errorf("failed to complete the toTarget operation: %w", err)
			}
		}
	}

	logger.Info().Msg("All sync configurations completed")

	return nil
}

// sourceRepositories fetches repositories from the source provider.
func sourceRepositories(ctx context.Context, config configuration.ProviderConfig) ([]interfaces.GitRepository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sourceRepositories:")
	config.DebugLog(logger).Msg("sourceRepositories:")

	cliOption := model.CLIOptions(ctx)

	providerClient, err := createProviderClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize new source gitprovider client: %w", err)
	}

	metainfos, err := provider.FetchMetainfo(ctx, config, providerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get source repository URLs for provider %s: %w", config.Provider, err)
	}

	if cliOption.DryRun {
		logger.Info().
			Str("source domain", config.Domain).
			Strs("source user,group", []string{config.User, config.Group}).
			Msg("Enabled dry-run. Skipping local clone")

		for _, metainfo := range metainfos {
			metainfo.DebugLog(logger).Msg("Fetched metainfo:")
		}

		return nil, nil
	}

	gitOption := config.GitInfo //TODO fix next iteration
	gitOption.ProviderToken = config.Token

	repositories, err := provider.Clone(ctx, target.Git{}, gitOption, metainfos)
	if err != nil {
		return nil, fmt.Errorf("failed to clone source git-provider repositories: %w", err)
	}

	return repositories, nil
}

// toTarget synchronizes repositories to the target provider.
func toTarget(ctx context.Context, sourceProvider, targetProvider configuration.ProviderConfig, repositories []interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering toTarget:")
	targetProvider.DebugLog(logger).Msg("toTarget:")

	ctx = initTargetSync(ctx, sourceProvider, targetProvider, repositories)

	providerClient, err := createProviderClient(ctx, targetProvider)
	if err != nil {
		return fmt.Errorf("failed to initialize a new target provider client: %w", err)
	}

	sourceGitInfo := sourceProvider.GitInfo //TODO
	sourceGitInfo.ProviderToken = sourceProvider.Token

	for _, repository := range repositories {
		if err := processRepository(ctx, targetProvider, providerClient, repository, sourceGitInfo); err != nil {
			return fmt.Errorf("failed to process repositories: %w", err)
		}
	}

	summary(ctx, sourceProvider)

	return nil
}

// processRepository handles the synchronization of a single repository.
func processRepository(ctx context.Context, targetProvider configuration.ProviderConfig, providerClient interfaces.GitProvider, repository interfaces.GitRepository, sourceGitInfo model.GitInfo) error {
	logger := log.Logger(ctx)
	repository.Metainfo().DebugLog(logger).Msg("processRepository:")

	cliOption := model.CLIOptions(ctx)
	name := repository.Metainfo().OriginalName

	if !isValidRepository(ctx, providerClient, repository) {
		if !cliOption.IgnoreInvalidName {
			return fmt.Errorf("%w: %s", ErrTargetRepositoryName, name)
		}

		logger.Debug().
			Str("name", name).
			Bool("ignoreInvalidName", cliOption.IgnoreInvalidName).
			Msg("Invalid repository name, ignoring it")

		return nil
	}

	if err := setupRepository(ctx, targetProvider, repository); err != nil {
		return err
	}

	return pushRepository(ctx, sourceGitInfo, targetProvider, providerClient, repository)
}

// setupRepository prepares the repository for synchronization.
func setupRepository(ctx context.Context, targetProvider configuration.ProviderConfig, repository interfaces.GitRepository) error {
	if targetProvider.Provider != configuration.ARCHIVE {
		if err := provider.SetGPSUpstreamRemoteFromOrigin(ctx, repository); err != nil {
			return fmt.Errorf("failed to create gpsupstream remote for archive target: %w", err)
		}
	}

	return nil
}

// pushRepository pushes the repository to the target provider.
func pushRepository(ctx context.Context, sourceGitInfo model.GitInfo, targetProvider configuration.ProviderConfig, providerClient interfaces.GitProvider, repository interfaces.GitRepository) error {
	var targett interfaces.TargetWriter

	switch strings.ToLower(targetProvider.Provider) {
	case configuration.ARCHIVE:
		targett = target.NewArchive(repository, repository.Metainfo().OriginalName)
	case configuration.DIRECTORY:
		targett = target.NewDirectory(repository, repository.Metainfo().OriginalName)
	default:
		targett = target.NewGit(repository, repository.Metainfo().Name(ctx))
	}

	if err := provider.Push(ctx, targetProvider, providerClient, targett, repository, sourceGitInfo); err != nil {
		return fmt.Errorf("failed to push to target repositories: %w", err)
	}

	if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		syncRunMeta.Total++
	}

	return nil
}

// isValidRepository checks if the repository name is valid for the target provider.
func isValidRepository(ctx context.Context, provider interfaces.GitProvider, repository interfaces.GitRepository) bool {
	logger := log.Logger(ctx)
	logger.Trace().Msg("isValidRepository:")

	name := repository.Metainfo().Name(ctx)
	logger.Debug().Str("name", name).Msg("isValidRepository:")

	if !provider.IsValidRepositoryName(ctx, name) {
		if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
			syncRunMeta.Fail["invalid"] = append(syncRunMeta.Fail["invalid"], name)
		}

		return false
	}

	return true
}

// initTargetSync initializes the synchronization process for a target.
func initTargetSync(ctx context.Context, sourceProvider configuration.ProviderConfig, targetProvider configuration.ProviderConfig, repositories []interfaces.GitRepository) context.Context {
	logger := log.Logger(ctx)

	syncRunMeta := model.NewSyncRunMetainfo(0, sourceProvider.Domain, targetProvider.Provider, len(repositories))
	ctx = context.WithValue(ctx, model.SyncRunMetainfoKey{}, syncRunMeta)

	userGrpString := strings.Trim(fmt.Sprint([]string{targetProvider.User, targetProvider.Group}), "[] ")
	logger.Info().Str("domain", sourceProvider.Domain).Str("usr/group", userGrpString).Msg("Syncing from")

	switch strings.ToLower(targetProvider.Provider) {
	case strings.ToLower(configuration.DIRECTORY):
		logger.Info().Str("directory", targetProvider.DirectoryTargetDir()).Msg("Targeting")
	case strings.ToLower(configuration.ARCHIVE):
		logger.Info().Str("archive directory", targetProvider.ArchiveTargetDir()).Msg("Targeting")
	default:
		logger.Info().
			Str("provider", targetProvider.Provider).
			Str("domain", targetProvider.Domain).
			Str("usr/group", userGrpString).
			Msgf("Targeting")
	}

	return ctx
}

// summary logs a summary of the synchronization process.
func summary(ctx context.Context, sourceProvider configuration.ProviderConfig) {
	logger := log.Logger(ctx)
	userGrpString := strings.Join([]string{sourceProvider.User, sourceProvider.Group}, "/")

	syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(*model.SyncRunMetainfo)
	if !ok {
		model.HandleError(ctx, errors.New("failed to get sync run meta"))
	}

	logger.Info().Str("domain", sourceProvider.Domain).Str("usr/group", userGrpString).Msg("Completed sync run from")
	logger.Info().Msgf("Sync request: %d repositories.", syncRunMeta.Total)

	if len(syncRunMeta.Fail) > 0 {
		if invalidCount := len(syncRunMeta.Fail["invalid"]); invalidCount > 0 {
			logger.Info().Msgf("Skipped repositories due to invalid naming : %d", invalidCount)
			logger.Info().Msgf("	- %v", syncRunMeta.Fail["invalid"])
		}

		if upToDateCount := len(syncRunMeta.Fail["uptodate"]); upToDateCount > 0 {
			logger.Info().Msgf("Ignored repositories due to being up-to-date : %d", upToDateCount)
			logger.Info().Msgf("	- Was up-to-date: %v", syncRunMeta.Fail["uptodate"])
		}
	}
}

// createProviderClient creates a new Git provider client.
func createProviderClient(ctx context.Context, providerConfig configuration.ProviderConfig) (interfaces.GitProvider, error) {
	option := model.GitProviderClientOption{
		Provider: providerConfig.Provider,
		Token:    providerConfig.Token,
		Domain:   providerConfig.Domain,
		Scheme:   providerConfig.Scheme,
	}

	client, err := provider.NewGitProviderClient(ctx, option)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider client: %w", err)
	}

	return client, nil
}

// cleanup removes temporary directories created during the sync process.
func cleanup(ctx context.Context) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering cleanup:")

	if err := model.DeleteTmpDir(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to delete tmpdir")
	}
}
