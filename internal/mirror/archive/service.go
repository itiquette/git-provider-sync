// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"errors"
	"fmt"
	"os"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
)

type Service struct {
	git      GitHandler
	storage  StorageHandler
	archiver Handlerer
}

// Pull implements interfaces.MirrorWriter.
func (serv *Service) Pull(_ context.Context, _ model.PullOption) error {
	return nil
}

func NewService(git GitHandler, storage StorageHandler, archiver Handlerer) *Service {
	return &Service{
		git:      git,
		storage:  storage,
		archiver: archiver,
	}
}

func (serv *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Archive:Push")
	opt.DebugLog(ctx, logger).Msg("Archive:Push")

	storagePath, err := serv.storage.GetStoragePath(ctx, opt)
	if err != nil {
		return err
	}

	if err := serv.git.InitializeRepository(ctx, storagePath, repo); err != nil {
		return fmt.Errorf("failed to initialize target repository: %w", err)
	}

	err = serv.archiver.CreateArchive(ctx, storagePath, opt.Target, repo.ProjectInfo().Name(ctx))
	if err != nil {
		return errors.New("failed to create archive ")
	}

	err = os.RemoveAll(storagePath)
	if err != nil {
		return fmt.Errorf("failed to remove dir %s. err: %w ", storagePath, err)
	}

	logger.Trace().Str("storagePath", storagePath).Msg("Removed")

	return nil
}
