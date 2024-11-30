// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package archive

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"

	"itiquette/git-provider-sync/internal/interfaces"
	"itiquette/git-provider-sync/internal/model"
	gpsconfig "itiquette/git-provider-sync/internal/model/configuration"
	"itiquette/git-provider-sync/internal/target/gitlib"
)

type GitHandler struct {
	client *gitlib.Service
}

func NewGitHandler(client *gitlib.Service) *GitHandler {
	return &GitHandler{client: client}
}

func (h *GitHandler) InitializeRepository(ctx context.Context, path string, repo interfaces.GitRepository) error {
	initializedRepo, err := git.PlainInit(path, false)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRepoInitialization, err)
	}

	pushOpt := model.NewPushOption(path, false, true, gpsconfig.HTTPClientOption{})
	if err := h.client.Push(ctx, repo, pushOpt, gpsconfig.GitOption{}); err != nil {
		return fmt.Errorf("%w: %w", ErrPushRepository, err)
	}

	if err := h.configureRepository(ctx, repo, initializedRepo, path); err != nil {
		return fmt.Errorf("failed to configure repository: %w", err)
	}

	return nil
}

func (h *GitHandler) Push(ctx context.Context, repo interfaces.GitRepository, opt model.PushOption) error {
	return h.client.Push(ctx, repo, opt, gpsconfig.GitOption{}) //nolint
}

// configureRepository handles the internal repository configuration.
func (h *GitHandler) configureRepository(ctx context.Context, repo interfaces.GitRepository, initializedRepo *git.Repository, path string) error {
	if err := h.client.Ops.SetRemoteAndBranch(ctx, repo, path); err != nil {
		return fmt.Errorf("failed to set remote and branch: %w", err)
	}

	if err := h.client.Ops.SetDefaultBranch(ctx, initializedRepo, repo.ProjectInfo().DefaultBranch); err != nil {
		return fmt.Errorf("failed to set default branch: %w", err)
	}

	return nil
}
