// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
)

type Service struct {
	gitHandler GitHandler
	storage    StorageHandler
}

func NewService(git GitHandler, storage StorageHandler) *Service {
	return &Service{
		gitHandler: git,
		storage:    storage,
	}
}

func (serv *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:Push")
	opt.DebugLog(ctx, logger).Msg("Directory:Push")

	targetDir, err := serv.storage.GetTargetPath(ctx, opt.Target, repo.ProjectInfo().Name(ctx))
	if err != nil {
		return fmt.Errorf("%w%w", ErrDirGetPath, err)
	}

	cliOpt := model.CLIOptions(ctx)
	if cliOpt.ForcePush || !serv.storage.DirectoryExists(targetDir) {
		return serv.gitHandler.InitializeRepository(ctx, targetDir, repo)
	}

	return nil
}

func (serv *Service) Pull(ctx context.Context, opt model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:Pull")
	//	logger.Debug().Str("targetPath", targetPath).Msg("Directory:Pull")

	targetDir, err := serv.storage.GetTargetPath(ctx, opt.Path, opt.Name)
	if err != nil {
		return fmt.Errorf("%w %w", ErrDirGetPath, err)
	}

	pullOpt := model.NewPullOption("", "", opt.SyncCfg, opt.SyncCfg.Auth, targetDir, targetDir)
	if err := serv.gitHandler.PullToDir(ctx, pullOpt); err != nil {
		return fmt.Errorf("%w: targetDir: %s: %w", ErrPullRepository, targetDir, err)
	}

	return nil
}
