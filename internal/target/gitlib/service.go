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
)

type Service struct {
	authService *authService
	Ops         operation
	metadata    MetadataHandler
}

func NewService() *Service {
	ops := NewOperation()
	auth := NewAuthService()
	metadata := NewMetadataHandler()

	return &Service{
		Ops:         *ops,
		authService: auth,
		metadata:    *metadata,
	}
}

func (s *Service) Clone(ctx context.Context, opt model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Clone")
	opt.DebugLog(logger).Msg("GitService:Clone")

	auth, err := s.authService.GetAuthMethod(ctx, opt.Git, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	var fileSys billy.Filesystem
	if opt.NonBareRepo {
		fileSys = memfs.New()
	}

	cloneOpt := s.buildCloneOptions(opt.URL, opt.Mirror, auth)

	repo, err := git.Clone(memory.NewStorage(), fileSys, cloneOpt)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	return model.NewRepository(repo) //nolint
}

func (s *Service) Pull(ctx context.Context, opt model.PullOption, targetDir string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Pull")
	opt.DebugLog(logger).Str("targetDir", targetDir).Msg("GitService:Pull")

	repo, worktree, err := s.prepareRepository(ctx, targetDir)
	if err != nil {
		return err
	}

	auth, err := s.authService.GetAuthMethod(ctx, opt.GitOption, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pullOpts := s.buildPullOptions(gpsconfig.ORIGIN, targetDir, auth)
	if err := s.performPull(ctx, worktree, pullOpts, targetDir); err != nil {
		return err
	}

	return s.Ops.FetchBranches(ctx, repo, auth, filepath.Dir(targetDir))
}

func (s *Service) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption, gitOpt gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitService:Push")
	opt.DebugLog(logger).Str("gitOpt", gitOpt.String()).Msg("GitService:Push")

	auth, err := s.authService.GetAuthMethod(ctx, gitOpt, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pushOpts := s.buildPushOptions(opt.Target, opt.RefSpecs, opt.Prune, auth)

	if err := repo.GoGitRepository().Push(&pushOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Str("targetDir", opt.Target).Msg("repository already up-to-date")
			s.metadata.UpdateSyncMetadata(ctx, "uptodate", opt.Target)

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	return nil
}

func (s *Service) prepareRepository(ctx context.Context, targetDir string) (*git.Repository, *git.Worktree, error) {
	repo, err := s.Ops.Open(ctx, targetDir)
	if err != nil {
		return nil, nil, err
	}

	worktree, err := s.Ops.GetWorktree(ctx, repo)
	if err != nil {
		return nil, nil, err
	}

	if err := s.Ops.WorktreeStatus(ctx, worktree); err != nil {
		return nil, nil, fmt.Errorf("%w: %s", err, targetDir)
	}

	return repo, worktree, nil
}

func (s *Service) performPull(ctx context.Context, worktree *git.Worktree, opts *git.PullOptions, targetDir string) error {
	if err := worktree.Pull(opts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			s.metadata.UpdateSyncMetadata(ctx, "uptodate", targetDir)

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return nil
}
