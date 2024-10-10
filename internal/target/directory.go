// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package target

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
	"os"
	"path/filepath"

	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

type Directory struct {
	gitClient Git
}

func NewDirectory(repository interfaces.GitRepository) Directory {
	return Directory{gitClient: NewGit(repository, repository.Metainfo().OriginalName)}
}

func (dir Directory) Push(ctx context.Context, repository interfaces.GitRepository, option model.PushOption, sourceGitOption gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory: Push")
	option.DebugLog(logger).Msg("Directory: Push")

	name := dir.getRepositoryName(ctx)
	targetDirPath := filepath.Join(option.Target, name)
	logger.Debug().Str("path", targetDirPath).Msg("Targeting directory")

	if err := os.MkdirAll(option.Target, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory at %s: %w", option.Target, err)
	}

	cliOption := model.CLIOptions(ctx)
	if cliOption.ForcePush || !directoryExists(targetDirPath) {
		if err := dir.initializeAndPushRepo(ctx, repository, targetDirPath); err != nil {
			return err
		}

		if err := dir.setRemoteAndBranch(repository, targetDirPath); err != nil {
			return err
		}

		return nil
	}

	pullOption := model.NewPullOption("", "", targetDirPath, sourceGitOption.Git, sourceGitOption.HTTPClient)

	if err := dir.gitClient.Pull(ctx, targetDirPath, pullOption); err != nil {
		return fmt.Errorf("failed to pull updates to %s: %w", targetDirPath, err)
	}

	return nil
}

func (dir Directory) initializeAndPushRepo(ctx context.Context, repository interfaces.GitRepository, sourceDirPath string) error {
	if _, err := git.PlainInit(sourceDirPath, false); err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOptions := model.NewPushOption(sourceDirPath, false, true, gpsconfig.HTTPClientOption{})

	err := dir.gitClient.Push(ctx, repository, pushOptions, gpsconfig.ProviderConfig{}, gpsconfig.GitOption{})

	if err != nil {
		return fmt.Errorf("%w: %w", ErrRepoPush, err)
	}

	return nil
}

func (dir Directory) setRemoteAndBranch(repository interfaces.GitRepository, sourceDirPath string) error {
	targetRepo, err := git.PlainOpen(sourceDirPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRepoOpen, sourceDirPath, err)
	}

	remote, err := repository.GoGitRepository().Remote(gpsconfig.ORIGIN)
	if err == nil {
		if _, err := targetRepo.CreateRemote(&config.RemoteConfig{
			Name: gpsconfig.ORIGIN,
			URLs: remote.Config().URLs,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrRemoteCreation, err)
		}
	}

	return setArcDefaultBranch(sourceDirPath, repository.Metainfo().DefaultBranch)
}

func (dir Directory) getRepositoryName(ctx context.Context) string {
	name := dir.gitClient.name
	cliOption := model.CLIOptions(ctx)

	if cliOption.CleanupName {
		name = stringconvert.RemoveNonAlphaNumericChars(ctx, name)
	}

	return name
}

func directoryExists(dir string) bool {
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}
