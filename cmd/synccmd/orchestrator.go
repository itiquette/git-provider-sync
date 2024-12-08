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

	for _, config := range cfg.GitProviderSyncConfs {
		if err := sourceToTarget(ctx, config); err != nil {
			return fmt.Errorf("failed source to target: %w", err)
		}
	}

	logger.Info().Msg("All syncs completed")

	return nil
}

func sourceToTarget(ctx context.Context, config gpsconfig.ProvidersConfig) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering sourceToTarget")

	repositories, err := sourceRepositories(ctx, config.SourceProvider)
	if err != nil {
		return fmt.Errorf("failed to fetch source repositories: %w", err)
	}

	for _, targetProvider := range config.ProviderTargets {
		if err := toTarget(ctx, config.SourceProvider, targetProvider, repositories); err != nil {
			return fmt.Errorf("failed to sync to target: %w", err)
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
