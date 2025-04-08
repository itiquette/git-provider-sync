// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2
package gitlib

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/log"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	gogitconfig "github.com/go-git/go-git/v5/config"
)

type Operation struct {
}

func NewOperation() *Operation {
	return &Operation{}
}

func (h *Operation) Open(ctx context.Context, path string) (*git.Repository, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering open")
	logger.Debug().Str("path", path).Msg("open")

	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpenRepository, err)
	}

	return repo, nil
}

func (h *Operation) GetWorktree(ctx context.Context, repo *git.Repository) (*git.Worktree, error) {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering getWorktree")

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorktree, err)
	}

	return worktree, nil
}

func (h *Operation) WorktreeStatus(ctx context.Context, worktree *git.Worktree) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering worktreeStatus")

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorktreeStatus, err)
	}

	if !status.IsClean() {
		return ErrUncleanWorkspace
	}

	return nil
}

func (h *Operation) FetchBranches(ctx context.Context, name string, repo *git.Repository, auth transport.AuthMethod) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering fetchBranches")
	logger.Debug().Str("name", name).Msg("fetchBranches")

	options := &git.FetchOptions{
		RefSpecs: []gogitconfig.RefSpec{
			"refs/*:refs/*",
			"^refs/pull/*:refs/pull/*",
		},
		Auth: auth,
	}

	if err := repo.Fetch(options); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			logger.Debug().Str("name", name).Msg("repository already up-to-date")

			return nil
		}

		return fmt.Errorf("%w: %w", ErrFetchBranches, err)
	}

	return nil
}

func (h *Operation) SetRemoteAndBranch(ctx context.Context, targetDirPath string, repository interfaces.GitRepository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering setRemoteAndBranch")
	logger.Debug().Str("targetDirPath", targetDirPath).Msg("setRemoteAndBranch")

	repo, err := git.PlainOpen(targetDirPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrOpenRepository, targetDirPath, err)
	}

	remote, err := repository.GoGitRepository().Remote(gpsconfig.ORIGIN)
	if err == nil {
		if _, err := repo.CreateRemote(&gogitconfig.RemoteConfig{
			Name: gpsconfig.ORIGIN,
			URLs: remote.Config().URLs,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrRemoteCreation, err)
		}
	} else {
		logger.Warn().Str("targetDirPath", targetDirPath).Msg("Remote origin not found in repository")
	}

	return nil
}

func (h *Operation) SetDefaultBranchBare(ctx context.Context, branch string, repo *git.Repository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering setDefaultBranchBare")
	logger.Debug().Str("branch", branch).Msg("setDefaultBranchBare")

	branchRef := plumbing.ReferenceName("refs/heads/" + branch)

	_, err := repo.Reference(branchRef, false)
	if err != nil {
		return fmt.Errorf("branch does not exist: %w", err)
	}

	// Set HEAD to point to the branch
	ref := plumbing.NewSymbolicReference(plumbing.HEAD, branchRef)
	if err := repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("%w: %w", ErrHeadSet, err)
	}

	return nil
}

func (h *Operation) SetDefaultBranch(ctx context.Context, branch string, repo *git.Repository) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering setDefaultBranch")
	logger.Debug().Str("branch", branch).Msg("setDefaultBranch")

	branchRef := plumbing.NewBranchReferenceName(branch)

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorktree, err)
	}

	if err = worktree.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	}); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrBranchCheckout, branch, err)
	}

	headRef := plumbing.NewSymbolicReference(plumbing.HEAD, branchRef)

	if err := repo.Storer.SetReference(headRef); err != nil {
		return fmt.Errorf("%w: %w", ErrHeadSet, err)
	}

	return nil
}
