// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package target

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

type GitLib struct {
	Op       GitLibOperation
	authProv authProvider
}

func NewGitLib() *GitLib {
	return &GitLib{
		Op:       newGitLibOperation(),
		authProv: newAuthProvider(),
	}
}

func (g GitLib) Clone(ctx context.Context, opt model.CloneOption) (model.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLib:Clone")
	opt.DebugLog(logger).Msg("GitLib:Clone")

	auth, err := g.authProv.getAuthMethod(ctx, opt.Git, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	var fileSys billy.Filesystem
	if opt.NonBareRepo {
		fileSys = memfs.New()
	}

	cloneOpt := newGitLibCloneOption(opt.URL, opt.Mirror, auth)

	repo, err := git.Clone(memory.NewStorage(), fileSys, &cloneOpt)
	if err != nil {
		return model.Repository{}, fmt.Errorf("%w: %w", ErrCloneRepository, err)
	}

	return model.NewRepository(repo) //nolint
}

func (g GitLib) Pull(ctx context.Context, opt model.PullOption, targetDir string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLib:Pull")
	opt.DebugLog(logger).Str("targetDir", targetDir).Msg("GitLib:Pull")

	repo, err := g.Op.Open(ctx, targetDir)
	if err != nil {
		return err //nolint
	}

	worktree, err := g.Op.GetWorktree(ctx, repo)
	if err != nil {
		return err //nolint
	}

	if err := g.Op.WorktreeStatus(ctx, worktree); err != nil {
		return fmt.Errorf("%w: %s", err, targetDir)
	}

	auth, err := g.authProv.getAuthMethod(ctx, opt.GitOption, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pullOpts := newGitLibPullOption(gpsconfig.ORIGIN, targetDir, auth)

	if err := worktree.Pull(&pullOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Str("targetDir", targetDir).Msg("repository already up-to-date")
			g.updateSyncRunMetainfo(ctx, "uptodate", targetDir)

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPullRepository, err)
	}

	return g.Op.FetchBranches(ctx, repo, auth, filepath.Dir(targetDir)) //nolint
}

func (g GitLib) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption, gitOpt gpsconfig.GitOption) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLib:Push")
	opt.DebugLog(logger).Str("gitOpt", gitOpt.String()).Msg("GitLib:Push")

	auth, err := g.authProv.getAuthMethod(ctx, gitOpt, opt.HTTPClient, opt.SSHClient)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
	}

	pushOpts := newGitLibPushOption(opt.Target, opt.RefSpecs, opt.Prune, auth)

	if err := repo.GoGitRepository().Push(&pushOpts); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Str("name", "name").Msg("repository already up-to-date")
			g.updateSyncRunMetainfo(ctx, "uptodate", "name")

			return nil
		}

		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	return nil
}

// func (g GitLib) PushRaw(ctx context.Context, repo *git.Repository, opt model.PushOption, gitOpt gpsconfig.GitOption) error {
// 	logger := log.Logger(ctx)
// 	logger.Trace().Msg("Entering GitLib:PushRaw")
// 	opt.DebugLog(logger).Str("gitOpt", gitOpt.String()).Msg("GitLib:PushRaw")

// 	auth, err := g.authProv.getAuthMethod(ctx, gitOpt, opt.HTTPClient, opt.SSHClient)
// 	if err != nil {
// 		return fmt.Errorf("%w: %w", ErrAuthMethod, err)
// 	}

// 	pushOpts := newGitLibPushOption(opt.Target, opt.RefSpecs, opt.Prune, auth)

// 	if err := repo.Push(&pushOpts); err != nil {
// 		if errors.Is(err, git.NoErrAlreadyUpToDate) {

// 			logger.Debug().Str("name", "name").Msg("repository already up-to-date")
// 			g.updateSyncRunMetainfo(ctx, "uptodate", "name")

// 			return nil
// 		}

// 		return fmt.Errorf("%w: %w", ErrPushRepository, err)
// 	}

// 	return nil
// }

func (g GitLib) updateSyncRunMetainfo(ctx context.Context, key, targetDir string) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLib:updateSyncRunMetainfo")
	logger.Debug().Str("key", key).Str("targetDir", targetDir).Msg("GitLib:updateSyncRunMetainfo")

	if syncRunMeta, ok := ctx.Value(model.SyncRunMetainfoKey{}).(model.SyncRunMetainfo); ok {
		syncRunMeta.Fail[key] = append(syncRunMeta.Fail[key], targetDir)
	}
}
