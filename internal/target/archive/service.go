// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

type Service struct {
	git      GitHandler
	storage  StorageHandler
	archiver Handlerer
}

func NewService(git GitHandler, storage StorageHandler, archiver Handlerer) *Service {
	return &Service{
		git:      git,
		storage:  storage,
		archiver: archiver,
	}
}

func (serv *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Archive:Push")
	opt.DebugLog(logger).Msg("Archive:Push")

	storagePath, err := serv.storage.GetStoragePath(ctx, opt)
	if err != nil {
		return err
	}

	if err := serv.git.InitializeRepository(ctx, storagePath, repo); err != nil {
		return fmt.Errorf("failed to initialize target repository: %w", err)
	}

	return serv.archiver.CreateArchive(ctx, storagePath, opt.Target, repo.ProjectInfo().Name(ctx)) //nolint
}
