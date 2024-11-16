// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package gitea

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"

	"code.gitea.io/sdk/gitea"
)

type ProtectionService struct {
	client *gitea.Client
}

func NewProtectionService(client *gitea.Client) *ProtectionService {
	return &ProtectionService{client: client}
}

func (p ProtectionService) protect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:protect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("protect")

	// projectIDStr is expected to be in format "owner/repo"
	owner, repoName := splitProjectPath(projectIDStr)
	if owner == "" || repoName == "" {
		return fmt.Errorf("invalid project path: %s", projectIDStr)
	}

	// Get repository to verify it exists
	_, _, err := p.client.GetRepo(owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get repository %s/%s: %w", owner, repoName, err)
	}

	err = p.enableBranchProtection(ctx, branch, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to enable branch protection: %w", err)
	}

	err = p.enableTagProtection(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to enable tag protection: %w", err)
	}

	return nil
}

func (p ProtectionService) unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:unprotect")
	logger.Debug().Str("projectID", projectIDStr).Str("branch", branch).Msg("unprotect")

	owner, repoName := splitProjectPath(projectIDStr)
	if owner == "" || repoName == "" {
		return fmt.Errorf("invalid project path: %s", projectIDStr)
	}

	// Get repository to verify it exists
	_, _, err := p.client.GetRepo(owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get repository %s/%s: %w", owner, repoName, err)
	}

	err = p.disableBranchProtection(ctx, branch, owner, repoName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected branches: %w", err)
		}
	}

	err = p.disableTagProtection(ctx, owner, repoName)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected tags: %w", err)
		}
	}

	return nil
}

func (p ProtectionService) enableTagProtection(ctx context.Context, owner, repo string) error { //nolint
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering enableTagProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("enableTagProtection")

	//Todo: check with gitea dev but does not seem to be supported in go-sdk for now
	//https://github.com/go-gitea/gitea/issues/15548

	// p.client
	// opts := gitea.CreateTagOption{
	// 	TagName: "*",
	// }

	// _, _, err := p.client.(owner, repo, opts)
	// if err != nil {
	// 	return fmt.Errorf("failed to enable tag protection: %w", err)
	// }

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, branch, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:enableBranchProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Str("branch", branch).Msg("enableBranchProtection")

	// Create protection for all branches
	allBranchesOpts := gitea.CreateBranchProtectionOption{
		BranchName:              "*",
		EnablePush:              false,
		EnablePushWhitelist:     true,
		PushWhitelistUsernames:  []string{},
		PushWhitelistTeams:      []string{},
		EnableMergeWhitelist:    true,
		MergeWhitelistUsernames: []string{},
		MergeWhitelistTeams:     []string{},
		RequireSignedCommits:    true,
		ProtectedFilePatterns:   "*",
		BlockOnRejectedReviews:  true,
		DismissStaleApprovals:   true,
		//RequireApprovals:        true,
		RequiredApprovals: 1,
	}

	_, _, err := p.client.CreateBranchProtection(owner, repo, allBranchesOpts)
	if err != nil {
		return fmt.Errorf("failed to protect all branches: %w", err)
	}

	// Create specific protection for the default branch
	defaultBranchOpts := gitea.CreateBranchProtectionOption{
		BranchName:              branch,
		EnablePush:              false,
		EnablePushWhitelist:     true,
		PushWhitelistUsernames:  []string{},
		PushWhitelistTeams:      []string{},
		EnableMergeWhitelist:    true,
		MergeWhitelistUsernames: []string{},
		MergeWhitelistTeams:     []string{},
		RequireSignedCommits:    true,
		ProtectedFilePatterns:   "*",
		BlockOnRejectedReviews:  true,
		DismissStaleApprovals:   true,
		//	RequireApprovals:        true,
		RequiredApprovals: 1,
	}

	_, _, err = p.client.CreateBranchProtection(owner, repo, defaultBranchOpts)
	if err != nil {
		return fmt.Errorf("failed to protect default branch: %w", err)
	}

	return nil
}

func (p ProtectionService) disableBranchProtection(ctx context.Context, _, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:disableBranchProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("Gitea:disableBranchProtection")

	// List all branch protections
	protections, _, err := p.client.ListBranchProtections(owner, repo, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list branch protections: %w", err)
	}

	// Remove each branch protection
	for _, protection := range protections {
		_, err := p.client.DeleteBranchProtection(owner, repo, protection.RuleName)
		if err != nil {
			return fmt.Errorf("failed to remove branch protection for %s: %w", protection.BranchName, err)
		}
	}

	return nil
}

func (p ProtectionService) disableTagProtection(ctx context.Context, owner, repo string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering Gitea:removeAllTagProtection")
	logger.Debug().Str("owner", owner).Str("repo", repo).Msg("Gitea:removeAllTagProtection")

	// List all protected tags
	protectedTags, _, err := p.client.ListRepoTags(owner, repo, gitea.ListRepoTagsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list protected tags: %w", err)
	}

	// Remove each tag protection
	for _, tag := range protectedTags {
		_, err := p.client.DeleteTag(owner, repo, tag.Name)
		if err != nil {
			return fmt.Errorf("failed to unprotect tag %s: %w", tag.Name, err)
		}
	}

	return nil
}

func splitProjectPath(path string) (string, string) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
