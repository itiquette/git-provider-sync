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
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

type Service struct {
	git     GitHandler
	storage StorageHandler
}

func NewService(git GitHandler, storage StorageHandler) *Service {
	return &Service{
		git:     git,
		storage: storage,
	}
}

func (serv *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:Push")
	opt.DebugLog(logger).Msg("Directory:Push")

	targetDir, err := serv.storage.GetTargetPath(ctx, opt.Target, repo.ProjectInfo().Name(ctx))
	if err != nil {
		return fmt.Errorf("%w%w", ErrDirGetPath, err)
	}

	cliOpt := model.CLIOptions(ctx)
	if cliOpt.ForcePush || !serv.storage.DirectoryExists(targetDir) {
		return serv.git.InitializeRepository(ctx, targetDir, repo)
	}

	return nil
}

func (serv *Service) Pull(ctx context.Context, syncCfg gpsconfig.SyncConfig, targetPath string, repo interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:Pull")
	logger.Debug().Str("targetPath", targetPath).Msg("Directory:Pull")

	targetDir, err := serv.storage.GetTargetPath(ctx, targetPath, repo.ProjectInfo().Name(ctx))
	if err != nil {
		return fmt.Errorf("%w %w", ErrDirGetPath, err)
	}

	pullOpt := model.NewPullOption("", "", syncCfg, syncCfg.Auth)
	if err := serv.git.Pull(ctx, pullOpt, targetDir); err != nil {
		return fmt.Errorf("%w: targetDir: %s: %w", ErrPullRepository, targetDir, err)
	}

	return nil
}
