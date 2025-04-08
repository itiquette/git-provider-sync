// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package directory

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/mirror/gitlib"
	"itiquette/git-provider-sync/internal/model"

	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"

	"github.com/go-git/go-git/v5"
)

type GitHandler struct {
	client interfaces.GitInterface
	ops    gitlib.Operation
}

func NewGitHandler(client interfaces.GitInterface) GitHandler {
	return GitHandler{client: client, ops: *gitlib.NewOperation()}
}

func (h *GitHandler) InitializeRepository(ctx context.Context, targetDir string, repo interfaces.GitRepository) error {
	initializedRepo, err := git.PlainInit(targetDir, false)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOpt := model.NewPushOption(targetDir, false, true, gpsconfig.AuthConfig{})
	if err := h.client.Push(ctx, repo, pushOpt); err != nil {
		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	if err := h.ops.SetRemoteAndBranch(ctx, targetDir, repo); err != nil {
		return fmt.Errorf("failed to set remote and branch: %w", err)
	}

	if err := h.ops.SetDefaultBranch(ctx, repo.ProjectInfo().DefaultBranch, initializedRepo); err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}

func (h *GitHandler) PullToDir(ctx context.Context, opt model.PullOption) error {
	return h.client.Pull(ctx, opt) //nolint
}
