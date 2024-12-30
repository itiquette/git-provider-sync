// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// sync.go - Core sync orchestration
package synccmd

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

func sync(ctx context.Context, cfg *gpsconfig.AppConfiguration) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sync")
	cfg.DebugLog(logger)

	ctx, err := model.CreateTmpDir(ctx, "", "gitprovidersync")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	//defer cleanup(ctx)

	for envName, environments := range cfg.GitProviderSyncConfs {
		for syncCfgName, syncCfg := range environments {
			if err := sourceToMirror(ctx, syncCfg); err != nil {
				return fmt.Errorf("failed to mirror environment: %s, syncCfg: %s, %w", envName, syncCfgName, err)
			}
		}
	}

	logger.Info().Msg("All syncs completed")

	return nil
}

func sourceToMirror(ctx context.Context, syncCfg gpsconfig.SyncConfig) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sourceToMirror")

	repositories, err := sourceRepositories(ctx, syncCfg)
	if err != nil {
		return fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	for _, mirrorCfg := range syncCfg.Mirrors {
		if err := toMirror(ctx, syncCfg, mirrorCfg, repositories); err != nil {
			return fmt.Errorf("failed to sync to mirror: %w", err)
		}
	}

	return nil
}

// func cleanup(ctx context.Context) {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering cleanup")

// 	if err := model.DeleteTmpDir(ctx); err != nil {
// 		log.Logger(ctx).Error().Err(err).Msg("failed to delete temporary directory")
// 	}
// }
