// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/provider/stringconvert"
)

type Service struct {
	authService AuthService
	Ops         Operation
	metadata    MetadataHandlerer
}

func NewService() *Service {
	ops := NewOperation()
	auth := NewAuthService()
	metadata := NewMetadataHandler()

	return &Service{
		Ops:         *ops,
		authService: auth,
		metadata:    metadata,
	}
}

func (serv *Service) Clone(ctx context.Context, opt model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Clone")
	opt.DebugLog(ctx, logger).Msg("GitService:Clone")

	auth, err := serv.authService.GetAuthMethod(ctx, opt.AuthCfg)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	var fileSys billy.Filesystem
	if opt.NonBareRepo {
		fileSys = memfs.New()
	}

	cloneOpt := serv.buildCloneOptions(opt.URL, opt.Mirror, auth)

	repo, err := git.Clone(memory.NewStorage(), fileSys, cloneOpt)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	return model.NewRepository(repo) //nolint
}

func (serv *Service) Pull(ctx context.Context, opt model.PullOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Pull")
	opt.DebugLog(logger).Str("targetDir", opt.TargetDir).Msg("GitService:Pull")

	repo, worktree, err := serv.prepareRepository(ctx, opt.TargetDir)
	if err != nil {
		return err
	}

	auth, err := serv.authService.GetAuthMethod(ctx, opt.AuthCfg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pullOpts := serv.buildPullOptions(gpsconfig.ORIGIN, opt.TargetDir, auth)
	if err := serv.performPull(ctx, worktree, pullOpts, opt.TargetDir); err != nil {
		return err
	}

	return serv.Ops.FetchBranches(ctx, filepath.Dir(opt.TargetDir), repo, auth)
}

func (serv *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Push")
	//:opt.DebugLog(logger).Str("sourceCfg", sourceCfg.String()).Msg("GitService:Push")

	auth, err := serv.authService.GetAuthMethod(ctx, opt.AuthCfg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pushOpts := serv.buildPushOptions(opt.Target, opt.RefSpecs, opt.Prune, auth)

	if err := repo.GoGitRepository().Push(&pushOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Str("targetDir", stringconvert.RemoveBasicAuthFromURL(ctx, opt.Target, false)).Msg("repository already up-to-date")
			serv.metadata.UpdateSyncMetadata(ctx, "uptodate", opt.Target)

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	return nil
}

func (serv *Service) prepareRepository(ctx context.Context, targetDir string) (*git.Repository, *git.Worktree, error) {
	repo, err := serv.Ops.Open(ctx, targetDir)
	if err != nil {
		return nil, nil, err
	}

	worktree, err := serv.Ops.GetWorktree(ctx, repo)
	if err != nil {
		return nil, nil, err
	}

	if err := serv.Ops.WorktreeStatus(ctx, worktree); err != nil {
		return nil, nil, fmt.Errorf("%w: %s", err, targetDir)
	}

	return repo, worktree, nil
}

func (serv *Service) performPull(ctx context.Context, worktree *git.Worktree, opts *git.PullOptions, targetDir string) error {
	if err := worktree.Pull(opts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			serv.metadata.UpdateSyncMetadata(ctx, "uptodate", targetDir)

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return nil
}
