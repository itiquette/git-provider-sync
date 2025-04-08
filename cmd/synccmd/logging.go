// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

// logging.go - All logging operations
package synccmd

import (
	"context"
	"errors"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/rs/zerolog"
)

// Package-level sentinel errors.
var (
	ErrMissingSyncRunMeta = errors.New("missing sync run metadata")
)

func initMirrorSync(ctx context.Context, syncCfg gpsconfig.SyncConfig, mirrorCfg gpsconfig.MirrorConfig, repositories []interfaces.GitRepository) context.Context {
	meta := model.NewSyncRunMetainfo(0, syncCfg.GetDomain(), mirrorCfg.ProviderType, len(repositories))
	ctx = context.WithValue(ctx, model.SyncRunMetainfoKey{}, meta)

	logSyncStart(ctx, mirrorCfg)

	return ctx
}

func logSyncStart(ctx context.Context, mirrorCfg gpsconfig.MirrorConfig) {
	logger := log.Logger(ctx)

	logSyncRunInfo(logger, mirrorCfg)
}

func logSyncRunInfo(logger *zerolog.Logger, mirrorCfg gpsconfig.MirrorConfig) {
	switch strings.ToLower(mirrorCfg.ProviderType) {
	case gpsconfig.DIRECTORY:
		logger.Info().Str("directory path", mirrorCfg.Path).Msg("Targeting")
	case gpsconfig.ARCHIVE:
		logger.Info().Str("archive directory path", mirrorCfg.Path).Msg("Targeting")
	default:
		logger.Info().
			Str("ProviderType", mirrorCfg.ProviderType).
			Str("GetDomain()", mirrorCfg.GetDomain()).
			Str("Owner", mirrorCfg.Owner).
			Msg("Targeting")
	}
}

func summary(ctx context.Context, syncCfg gpsconfig.SyncConfig) {
	logger := log.Logger(ctx)

	syncRunMetaInfo, ok := ctx.Value(model.SyncRunMetainfoKey{}).(*model.SyncRunMetainfo)
	if !ok {
		model.HandleError(ctx, ErrMissingSyncRunMeta)

		return
	}

	logger.Info().
		Str("Domain", syncCfg.GetDomain()).
		Str("Owner", syncCfg.Owner).
		Msg("Completed sync run")

	logger.Info().Msgf("Sync request: %d repositories", syncRunMetaInfo.Total)
	logFailures(logger, syncRunMetaInfo)
}

func logFailures(logger *zerolog.Logger, meta *model.SyncRunMetainfo) {
	metaFailPtr := (*meta.Fail)
	if len(metaFailPtr) == 0 {
		return
	}

	if invalidCount := len(metaFailPtr); invalidCount > 0 {
		logger.Info().
			Int("invalidCount", invalidCount).
			Strs("repositories", metaFailPtr["invalid"]).
			Msg("skipped repositories due to invalid naming")
	}

	if upToDateCount := len(metaFailPtr["uptodate"]); upToDateCount > 0 {
		logger.Info().
			Int("upToDateCount", upToDateCount).
			Strs("repositories", metaFailPtr["uptodate"]).
			Msg("ignored up-to-date repositories")
	}
}
