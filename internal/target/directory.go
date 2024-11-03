// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

// Package target handles operations related to target directories for git repositories.
package target

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

// Common errors.
var (
	ErrDirCreate  = errors.New("failed to create directory")
	ErrDirGetPath = errors.New("failed to get directory path")
)

// Directory represents a target directory for git operations.
type Directory struct {
	gitClient GitLib
}

// Push performs a push operation on the target directory.
func (dir Directory) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption, sourceConfig gpsconfig.ProviderConfig, _ gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:Push")
	opt.DebugLog(logger).Msg("Directory:Push")

	targetDir, err := getTargetDirPath(ctx, opt.Target, repo.ProjectInfo().Name(ctx))
	if err != nil {
		return fmt.Errorf("%w %w", ErrDirGetPath, err)
	}

	cliOpt := model.CLIOptions(ctx)
	if cliOpt.ForcePush || !directoryExists(targetDir) {
		return dir.initializeTargetRepository(ctx, repo, targetDir)
	}

	return updateRepository(ctx, sourceConfig, targetDir, dir)
}

func updateRepository(ctx context.Context, sourceCfg gpsconfig.ProviderConfig, targetDir string, dir Directory) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:updateRepository")

	pullOpt := model.NewPullOption("", "", sourceCfg.Git, sourceCfg.HTTPClient, sourceCfg.SSHClient)

	if err := dir.gitClient.Pull(ctx, targetDir, pullOpt); err != nil {
		return fmt.Errorf("%w: targetDir: %s: %w", ErrRepoPull, targetDir, err)
	}

	return nil
}

func getTargetDirPath(ctx context.Context, targetDir, name string) (string, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:getTargetDirPath")

	fullPath := filepath.Join(targetDir, name)
	logger.Debug().Str("path", fullPath).Msg("Targeting directory")

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("%w: %s: %w", ErrDirCreate, targetDir, err)
	}

	return fullPath, nil
}

func (dir Directory) initializeTargetRepository(ctx context.Context, repo interfaces.GitRepository, targetDir string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Directory:initializeTargetRepository")

	if _, err := git.PlainInit(targetDir, false); err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOpt := model.NewPushOption(targetDir, false, true, gpsconfig.HTTPClientOption{})
	if err := dir.gitClient.Push(ctx, repo, pushOpt, gpsconfig.ProviderConfig{}, gpsconfig.GitOption{}); err != nil {
		return fmt.Errorf("%w: %w", ErrRepoPush, err)
	}

	if err := setRemoteAndBranch(repo, targetDir); err != nil {
		return err
	}

	if err := setDefaultBranch(targetDir, repo.ProjectInfo().DefaultBranch); err != nil {
		return err
	}

	return nil
}

func directoryExists(dir string) bool {
	_, err := os.Stat(dir)

	return !os.IsNotExist(err)
}

// NewDirectory creates a new Directory Target instance.
func NewDirectory() Directory {
	return Directory{gitClient: NewGitLib()}
}
