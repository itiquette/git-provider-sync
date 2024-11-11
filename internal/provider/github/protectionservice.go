// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"

	"github.com/google/go-github/v66/github"
)

type ProtectionService struct {
	client *github.Client
}

func NewProtectionService(client *github.Client) *ProtectionService {
	return &ProtectionService{client: client}
}

func (p ProtectionService) protect(ctx context.Context, owner, name string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:protect")
	logger.Debug().Str("owner", owner).Str("name", name).Msg("GitHub:protect")

	permissions := &github.ActionsPermissionsRepository{
		Enabled: github.Bool(false),
	}

	_, _, err := p.client.Repositories.EditActionsPermissions(ctx, owner, name, *permissions)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable Actions for repository: %w", err)
		}
	}

	err = p.enableBranchProtection(ctx, owner, name)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to enable branch protection: %w", err)
		}
	}

	err = p.enableTagProtection(ctx, owner, name)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to enable tag protection: %w", err)
		}
	}

	return nil
}

func (p ProtectionService) enableTagProtection(ctx context.Context, owner, repo string) error {
	_, _, err := p.client.Repositories.CreateTagProtection(ctx, owner, repo, "*")
	if err != nil {
		return fmt.Errorf("failed to protect tags: %w", err)
	}

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering enableBranchProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("enableBranchProtection")

	protectionReq := &github.ProtectionRequest{
		RequiredStatusChecks: &github.RequiredStatusChecks{
			Strict:   true,
			Contexts: &[]string{},
		},
		RequiredPullRequestReviews: &github.PullRequestReviewsEnforcementRequest{
			RequiredApprovingReviewCount: *github.Int(1),
			RequireCodeOwnerReviews:      *github.Bool(true),
			DismissStaleReviews:          *github.Bool(true),
		},
		EnforceAdmins:    *github.Bool(true),
		AllowForcePushes: github.Bool(false),
		AllowDeletions:   github.Bool(false),
	}

	branches, _, err := p.client.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	//TODO
	// Restrictions: &github.BranchRestrictionsRequest{
	// 	Users: []string{},
	// 	Teams: []string{},
	// 	Apps:  []string{},
	// },
	for _, branch := range branches {
		_, _, err := p.client.Repositories.UpdateBranchProtection(ctx, owner, repo, *branch.Name, protectionReq)
		if err != nil {
			return fmt.Errorf("failed to protect branch %s: %w", *branch.Name, err)
		}
	}

	return nil
}

func (p ProtectionService) unprotect(ctx context.Context, branch, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering unprotect")
	logger.Debug().Str("owner", owner).Str("repo", repo).Str("branch", branch).Msg("unprotect")

	err := p.disableBranchProtection(ctx, branch, owner, repo)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected branches: %w", err)
		}
	}

	err = p.disableTagProtection(ctx, owner, repo)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected tags: %w", err)
		}
	}

	return nil
}

func (p ProtectionService) disableBranchProtection(ctx context.Context, branch, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:disableBranchProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("GitHub:disableBranchProtection")

	branches, _, err := p.client.Repositories.ListBranches(ctx, owner, repo, &github.BranchListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	for _, b := range branches {
		_, err := p.client.Repositories.RemoveBranchProtection(ctx, owner, repo, b.GetName())
		if err != nil {
			return fmt.Errorf("failed to remove branch protection: %w", err)
		}
	}

	_, err = p.client.Repositories.RemoveBranchProtection(ctx, owner, repo, branch)
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: %w", err)
	}

	return nil
}

func (p ProtectionService) disableTagProtection(ctx context.Context, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:disableTagProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("GitHub:disableTagProtection")

	tags, _, err := p.client.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	for _, tag := range tags {
		// Remove protection if it exists
		_, err := p.client.Repositories.RemoveBranchProtection(ctx, owner, repo, tag.GetName())
		if err != nil {
			return fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	return nil
}
