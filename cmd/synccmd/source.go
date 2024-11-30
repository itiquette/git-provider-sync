// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// source - source operations and validation
package synccmd

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider"
	"itiquette/git-provider-sync/internal/target/directory"
	"itiquette/git-provider-sync/internal/target/gitbinary"
	"itiquette/git-provider-sync/internal/target/gitlib"
)

func sourceRepositories(ctx context.Context, sourceCfg gpsconfig.ProviderConfig) ([]interfaces.GitRepository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sourceRepositories")

	providerClient, err := createProviderClient(ctx, sourceCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider client: %w", err)
	}

	metainfo, err := provider.FetchProjectInfo(ctx, sourceCfg, providerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository metainfo for %s: %w", sourceCfg.ProviderType, err)
	}

	if model.CLIOptions(ctx).DryRun {
		logDryRun(ctx, sourceCfg, metainfo)

		return nil, nil
	}

	reader, err := getSourceReader(sourceCfg)
	if err != nil {
		return nil, fmt.Errorf("get source reader: %w", err)
	}

	repositories, err := provider.Clone(ctx, reader, sourceCfg, metainfo)
	if err != nil {
		return nil, fmt.Errorf("clone repositories: %w", err)
	}

	return repositories, nil
}

func getSourceReader(cfg gpsconfig.ProviderConfig) (interfaces.SourceReader, error) {
	if !cfg.Git.UseGitBinary {
		return gitlib.NewService(), nil
	}

	reader, err := gitbinary.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create git binary source reader: %w", err)
	}

	return reader, nil
}

func processRepository(ctx context.Context, targetCfg gpsconfig.ProviderConfig, client interfaces.GitProvider, repo interfaces.GitRepository, sourceCfg gpsconfig.ProviderConfig) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering processRepository")
	repo.ProjectInfo().DebugLog(logger).Msg("processRepository")

	if repo.ProjectInfo().OriginalName == "" {
		return ErrEmptyMetainfo
	}

	if err := validateRepository(ctx, client, repo, targetCfg); err != nil {
		return err
	}

	if err := prepareRepository(ctx, targetCfg, repo); err != nil {
		return fmt.Errorf("failed to prepare repository: %w", err)
	}

	if err := pushRepository(ctx, sourceCfg, targetCfg, client, repo); err != nil {
		return fmt.Errorf("failed to push repository: %w", err)
	}

	if targetCfg.ProviderType == gpsconfig.DIRECTORY {
		gitHandler := directory.NewGitHandler(gitlib.NewService())
		storageHandler := directory.NewStorageHandler()
		dirService := directory.NewService(gitHandler, storageHandler)

		if err := dirService.Pull(ctx, sourceCfg, targetCfg.DirectoryTargetDir(), repo); err != nil {
			return fmt.Errorf("failed to pull repository for directory target: %w", err)
		}
	}

	return nil
}

func validateRepository(ctx context.Context, client interfaces.GitProvider, repo interfaces.GitRepository, targetCfg gpsconfig.ProviderConfig) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering validateRepository")

	if client.IsValidProjectName(ctx, repo.ProjectInfo().Name(ctx)) {
		return nil
	}

	name := repo.ProjectInfo().OriginalName
	markRepositoryInvalid(ctx, name)

	opts := model.CLIOptions(ctx)
	if !opts.IgnoreInvalidName && !targetCfg.SyncRun.IgnoreInvalidName {
		return fmt.Errorf("%w: %s", ErrInvalidRepoName, name)
	}

	log.Logger(ctx).Debug().
		Str("name", name).
		Bool("ignoreInvalidName", opts.IgnoreInvalidName).
		Msg("invalid repository name, ignoring")

	return nil
}

func markRepositoryInvalid(ctx context.Context, repoName string) {
	if meta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		meta.Fail["invalid"] = append(meta.Fail["invalid"], repoName)
	}
}

func prepareRepository(ctx context.Context, targetCfg gpsconfig.ProviderConfig, repo interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering prepareRepository")

	if targetCfg.ProviderType == gpsconfig.ARCHIVE {
		return nil
	}

	if err := provider.SetGPSUpstreamRemoteFromOrigin(ctx, repo); err != nil {
		return fmt.Errorf("create gpsupstream remote: %w", err)
	}

	return nil
}

func createProviderClient(ctx context.Context, cfg gpsconfig.ProviderConfig) (interfaces.GitProvider, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering createProviderClient")

	client, err := provider.NewGitProviderClient(ctx, model.GitProviderClientOption{
		ProviderType: cfg.ProviderType,
		HTTPClient:   cfg.HTTPClient,
		Domain:       cfg.GetDomain(),
		Repositories: cfg.Repositories,
		UploadURL:    cfg.GitHubUploadURL(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider client: %w", err)
	}

	return client, nil
}
