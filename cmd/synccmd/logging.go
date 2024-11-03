// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// logging.go - All logging operations
package synccmd

import (
	"context"
	"strings"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/rs/zerolog"
)

func initTargetSync(ctx context.Context, sourceProvider gpsconfig.ProviderConfig, targetProvider gpsconfig.ProviderConfig, repositories []interfaces.GitRepository) context.Context {
	meta := model.NewSyncRunMetainfo(0, sourceProvider.GetDomain(), targetProvider.ProviderType, len(repositories))
	ctx = context.WithValue(ctx, model.SyncRunMetainfoKey{}, meta)

	logSyncStart(ctx, sourceProvider, targetProvider)

	return ctx
}

func logSyncStart(ctx context.Context, _, target gpsconfig.ProviderConfig) {
	logger := log.Logger(ctx)

	logTargetInfo(logger, target)
}

func logTargetInfo(logger *zerolog.Logger, target gpsconfig.ProviderConfig) {
	switch strings.ToLower(target.ProviderType) {
	case gpsconfig.DIRECTORY:
		logger.Info().Str("directory", target.DirectoryTargetDir()).Msg("Targeting")
	case gpsconfig.ARCHIVE:
		logger.Info().Str("archive directory", target.ArchiveTargetDir()).Msg("Targeting")
	default:
		logger.Info().
			Str("provider", target.ProviderType).
			Str("domain", target.GetDomain()).
			Str("usr/group", formatUserGroup(target.User, target.Group)).
			Msg("Targeting")
	}
}

func formatUserGroup(user, group string) string {
	return strings.Join([]string{user, group}, "/")
}

func summary(ctx context.Context, sourceProvider gpsconfig.ProviderConfig) {
	logger := log.Logger(ctx)
	userGroup := formatUserGroup(sourceProvider.User, sourceProvider.Group)

	syncRunMetaInfo, ok := ctx.Value(model.SyncRunMetainfoKey{}).(*model.SyncRunMetainfo)
	if !ok {
		model.HandleError(ctx, ErrMissingSyncRunMeta)

		return
	}

	logger.Info().
		Str("domain", sourceProvider.GetDomain()).
		Str("usr/group", userGroup).
		Msg("Completed sync run")

	logger.Info().Msgf("Sync request: %d repositories", syncRunMetaInfo.Total)
	logFailures(logger, syncRunMetaInfo)
}

func logFailures(logger *zerolog.Logger, meta *model.SyncRunMetainfo) {
	if len(meta.Fail) == 0 {
		return
	}

	if invalidCount := len(meta.Fail["invalid"]); invalidCount > 0 {
		logger.Info().
			Int("count", invalidCount).
			Strs("repositories", meta.Fail["invalid"]).
			Msg("skipped repositories due to invalid naming")
	}

	if upToDateCount := len(meta.Fail["uptodate"]); upToDateCount > 0 {
		logger.Info().
			Int("count", upToDateCount).
			Strs("repositories", meta.Fail["uptodate"]).
			Msg("ignored up-to-date repositories")
	}
}

func logDryRun(ctx context.Context, cfg gpsconfig.ProviderConfig, metainfo []model.ProjectInfo) {
	logger := log.Logger(ctx)

	logger.Info().
		Str("domain", cfg.GetDomain()).
		Strs("user/group", []string{cfg.User, cfg.Group}).
		Msg("dry-run enabled, skipping local clone")

	for _, meta := range metainfo {
		meta.DebugLog(logger).Msg("fetched repository metadata")
	}
}
