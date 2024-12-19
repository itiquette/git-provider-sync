// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// target.go - Target-specific operations and writers
package synccmd

import (
	"context"
	"fmt"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider"
	"itiquette/git-provider-sync/internal/target/archive"
	"itiquette/git-provider-sync/internal/target/directory"
	"itiquette/git-provider-sync/internal/target/gitbinary"
	"itiquette/git-provider-sync/internal/target/gitlib"
)

func toTarget(ctx context.Context, sourceCfg, targetCfg gpsconfig.ProviderConfig, repositories []interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering toTarget")
	targetCfg.DebugLog(logger)

	ctx = initTargetSync(ctx, sourceCfg, targetCfg, repositories)

	client, err := createProviderClient(ctx, targetCfg)
	if err != nil {
		return fmt.Errorf("create target provider client: %w", err)
	}

	for _, repo := range repositories {
		repo.ProjectInfo().Name(ctx)

		if err := processRepository(ctx, targetCfg, client, repo, sourceCfg); err != nil {
			return fmt.Errorf("process repository: %w", err)
		}
	}

	summary(ctx, sourceCfg)

	return nil
}

func pushRepository(ctx context.Context, sourceCfg, targetCfg gpsconfig.ProviderConfig, client interfaces.GitProvider, repo interfaces.GitRepository) error {
	writer, err := getTargetWriter(targetCfg)
	if err != nil {
		return fmt.Errorf("get target writer: %w", err)
	}

	if err := provider.Push(ctx, targetCfg, client, writer, repo, sourceCfg); err != nil {
		return fmt.Errorf("push to target: %w", err)
	}

	incrementSyncCount(ctx)

	return nil
}

func getTargetWriter(cfg gpsconfig.ProviderConfig) (interfaces.TargetWriter, error) {
	switch strings.ToLower(cfg.ProviderType) {
	case gpsconfig.ARCHIVE:
		gitHandler := archive.NewGitHandler(gitlib.NewService())
		storageHandler := archive.NewStorageHandler()
		archiverHandler := archive.NewHandler()

		return archive.NewService(*gitHandler, storageHandler, archiverHandler), nil
	case gpsconfig.DIRECTORY:
		gitHandler := directory.NewGitHandler(gitlib.NewService())
		storageHandler := directory.NewStorageHandler()

		return directory.NewService(gitHandler, storageHandler), nil
	default:
		if cfg.Git.UseGitBinary {
			writer, err := gitbinary.NewService()
			if err != nil {
				return nil, fmt.Errorf("create git binary writer: %w", err)
			}

			return writer, nil
		}

		return gitlib.NewService(), nil
	}
}

func incrementSyncCount(ctx context.Context) {
	if meta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		meta.Total++
	}
}
