// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"

	"github.com/google/go-github/v67/github"
)

type ProtectionService struct {
	client *github.Client
}

func NewProtectionService(client *github.Client) *ProtectionService {
	return &ProtectionService{client: client}
}

func (p ProtectionService) protect(ctx context.Context, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:protect")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Msg("GitHub:protect")

	permissions := &github.ActionsPermissionsRepository{
		Enabled: github.Bool(false),
	}

	_, _, err := p.client.Repositories.EditActionsPermissions(ctx, owner, projectName, *permissions)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable Actions. projectName: %s. err: %w", projectName, err)
		}
	}

	err = p.enableBranchProtection(ctx, owner, projectName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to enable branch protection. projectName: %s. err: %w", projectName, err)
		}
	}

	err = p.enableTagProtection(ctx, owner, projectName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to enable tag protection. projectName: %s. err: %w", projectName, err)
		}
	}

	return nil
}

//nolint
func (p ProtectionService) enableTagProtection(ctx context.Context, owner, projectName string) error {
	_, _, err := p.client.Repositories.CreateTagProtection(ctx, owner, projectName, "*") //lint:ignore SA1019 we will fix
	if err != nil {
		return fmt.Errorf("failed to protect tags: %w", err)
	}

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:enableBranchProtection")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Msg("enableBranchProtection")

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

	branches, _, err := p.client.Repositories.ListBranches(ctx, owner, projectName, &github.BranchListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list branches. projectName: %s. err: %w", projectName, err)
	}

	//TODO
	// Restrictions: &github.BranchRestrictionsRequest{
	// 	Users: []string{},
	// 	Teams: []string{},
	// 	Apps:  []string{},
	// },
	for _, branch := range branches {
		_, _, err := p.client.Repositories.UpdateBranchProtection(ctx, owner, projectName, *branch.Name, protectionReq)
		if err != nil {
			return fmt.Errorf("failed to protect branch. branch: %s. err: %w", *branch.Name, err)
		}
	}

	return nil
}

func (p ProtectionService) unprotect(ctx context.Context, branch, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:unprotect")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Str("branch", branch).Msg("GitHub:unprotect")

	err := p.disableBranchProtection(ctx, branch, owner, projectName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected branches. projectName: %s. err: %w", projectName, err)
		}
	}

	err = p.disableTagProtection(ctx, owner, projectName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected tags. projectName: %s. err: %w", projectName, err)
		}
	}

	return nil
}

func (p ProtectionService) disableBranchProtection(ctx context.Context, branch, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:disableBranchProtection")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Msg("GitHub:disableBranchProtection")

	branches, _, err := p.client.Repositories.ListBranches(ctx, owner, projectName, &github.BranchListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list branches. projectName: %s. err: %w", projectName, err)
	}

	for _, b := range branches {
		_, err := p.client.Repositories.RemoveBranchProtection(ctx, owner, projectName, b.GetName())
		if err != nil {
			return fmt.Errorf("failed to remove branch protection. projectName: %s. err: %w", projectName, err)
		}
	}

	_, err = p.client.Repositories.RemoveBranchProtection(ctx, owner, projectName, branch)
	if err != nil {
		return fmt.Errorf("failed to remove branch protection: projectName: %s. err: %w", projectName, err)
	}

	return nil
}

func (p ProtectionService) disableTagProtection(ctx context.Context, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:disableTagProtection")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Msg("GitHub:disableTagProtection")

	tags, _, err := p.client.Repositories.ListTags(ctx, owner, projectName, &github.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	for _, tag := range tags {
		// Remove protection if it exists
		_, err := p.client.Repositories.RemoveBranchProtection(ctx, owner, projectName, tag.GetName())
		if err != nil {
			return fmt.Errorf("failed to remove tag. tag: %s. err: %w", *tag.Name, err)
		}
	}

	return nil
}

func splitProjectPath(path string) (string, string, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("string was not in a/b format, failed to split: path: %s", path)
	}

	return parts[0], parts[1], nil
}
