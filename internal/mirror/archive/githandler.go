// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/mirror/gitlib"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
)

type GitHandler struct {
	client *gitlib.Service
}

func NewGitHandler(client *gitlib.Service) *GitHandler {
	return &GitHandler{client: client}
}

func (h *GitHandler) InitializeRepository(ctx context.Context, path string, repo interfaces.GitRepository) error {
	isBare := true

	initializedRepo, err := git.PlainInit(path, isBare)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOpt := model.NewPushOption(path, false, true, gpsconfig.AuthConfig{})
	if err := h.client.Push(ctx, repo, pushOpt); err != nil {
		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	if err := h.configureRepository(ctx, path, isBare, repo, initializedRepo); err != nil {
		return fmt.Errorf("failed to configure repository: %w", err)
	}

	return nil
}

func (h *GitHandler) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	return h.client.Push(ctx, repo, opt) //nolint
}

// configureRepository handles the internal repository configuration.
func (h *GitHandler) configureRepository(ctx context.Context, path string, isBare bool, repo interfaces.GitRepository, initializedRepo *git.Repository) error {
	if err := h.client.Ops.SetRemoteAndBranch(ctx, path, repo); err != nil {
		return fmt.Errorf("failed to set remote and branch: %w", err)
	}

	if isBare {
		if err := h.client.Ops.SetDefaultBranchBare(ctx, repo.ProjectInfo().DefaultBranch, initializedRepo); err != nil {
			return fmt.Errorf("failed to set default bare branch: %w", err)
		}
	} else {
		if err := h.client.Ops.SetDefaultBranch(ctx, repo.ProjectInfo().DefaultBranch, initializedRepo); err != nil {
			return fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	return nil
}
