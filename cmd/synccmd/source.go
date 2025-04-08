// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// source - source operations and validation
package synccmd

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/mirror/gitbinary"
	"itiquette/git-provider-sync/internal/mirror/gitlib"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider"
)

func sourceRepositories(ctx context.Context, syncCfg gpsconfig.SyncConfig) ([]interfaces.GitRepository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sourceRepositories")

	providerClient, err := createProviderClient(ctx, syncCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create source provider client: %w", err)
	}

	projectInfos, err := provider.FetchProjectInfos(ctx, syncCfg, providerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project infos for %s: %w", syncCfg.ProviderType, err)
	}

	if model.CLIOptions(ctx).DryRun {
		for _, meta := range projectInfos {
			meta.DebugLog(logger).Msg("fetched repository meta data")
		}

		logger.Info().
			Str("Domain", syncCfg.Domain).
			Str("Owner", syncCfg.Owner).
			Str("OwnerType", syncCfg.OwnerType).
			Msg("option dry-run enabled, skipping local clone")

		return nil, nil
	}

	reader, err := getSourceReader(ctx, syncCfg)
	if err != nil {
		return nil, fmt.Errorf("get source reader: %w", err)
	}

	repositories, err := provider.Clone(ctx, reader, syncCfg, projectInfos)
	if err != nil {
		return nil, fmt.Errorf("clone repositories: %w", err)
	}

	return repositories, nil
}

func getSourceReader(ctx context.Context, syncCfg gpsconfig.SyncConfig) (interfaces.SourceReader, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering getSourceReader")

	if !syncCfg.UseGitBinary {
		logger.Debug().Msg("Initialized go-git SourceReader")

		return gitlib.NewService(), nil
	}

	reader, err := gitbinary.NewService()
	if err != nil {
		return nil, fmt.Errorf("failed to create git binary SourceReader: %w", err)
	}

	logger.Debug().Msg("Initialized Git Binary SourceReader")

	return reader, nil
}

func createProviderClient(ctx context.Context, syncCfg gpsconfig.SyncConfig) (interfaces.GitProvider, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering createProviderClient")

	client, err := provider.NewGitProviderClient(ctx, model.GitProviderClientOption{
		ProviderType: syncCfg.ProviderType,
		AuthCfg:      syncCfg.Auth,
		Domain:       syncCfg.Domain,
		Repositories: syncCfg.Repositories,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider client: %w", err)
	}

	return client, nil
}
