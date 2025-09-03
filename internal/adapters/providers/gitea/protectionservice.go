// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package gitea

import (
	"context"
	"fmt"

	"code.gitea.io/sdk/gitea"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProtectionService provides Gitea-specific repository protection operations.
type ProtectionService struct {
	client *gitea.Client
	logger ports.Logger
}

// NewProtectionService creates a new Gitea protection service.
func NewProtectionService(client *gitea.Client, logger ports.Logger) *ProtectionService {
	return &ProtectionService{
		client: client,
		logger: logger,
	}
}

// ProtectRepositoryRequest contains parameters for protecting a repository.
type ProtectRepositoryRequest struct {
	Owner                  string
	RepositoryName         string
	EnableBranchProtection bool
	BranchProtectionRules  []BranchProtectionRule
}

// BranchProtectionRule defines protection rules for a branch.
type BranchProtectionRule struct {
	BranchName              string
	EnablePush              bool
	EnablePushWhitelist     bool
	PushWhitelistUsernames  []string
	PushWhitelistTeams      []string
	EnableMergeWhitelist    bool
	MergeWhitelistUsernames []string
	MergeWhitelistTeams     []string
	ProtectedFilePatterns   string
	BlockOnRejectedReviews  bool
	DismissStaleApprovals   bool
	RequireSignedCommits    bool
	RequiredApprovals       int64
	EnableStatusCheck       bool
	StatusCheckContexts     []string
}

// ProtectRepository applies protection to a Gitea repository.
func (ps *ProtectionService) ProtectRepository(ctx context.Context, request ProtectRepositoryRequest) error {
	ps.logger.Debug(ctx, "Protecting Gitea repository", map[string]any{
		"owner":                    request.Owner,
		"repository":               request.RepositoryName,
		"enable_branch_protection": request.EnableBranchProtection,
	})

	if request.EnableBranchProtection {
		if err := ps.enableBranchProtection(ctx, request.Owner, request.RepositoryName, request.BranchProtectionRules); err != nil {
			return fmt.Errorf("failed to enable branch protection: %w", err)
		}
	}

	ps.logger.Info(ctx, "Repository protection applied successfully", map[string]any{
		"owner":      request.Owner,
		"repository": request.RepositoryName,
	})

	return nil
}

// UnprotectRepository removes protection from a Gitea repository.
func (ps *ProtectionService) UnprotectRepository(ctx context.Context, owner, repositoryName string) error {
	ps.logger.Debug(ctx, "Unprotecting Gitea repository", map[string]any{
		"owner":      owner,
		"repository": repositoryName,
	})

	// List and remove all branch protections
	protections, _, err := ps.client.ListBranchProtections(owner, repositoryName, gitea.ListBranchProtectionsOptions{})
	if err != nil {
		ps.logger.Warn(ctx, "Failed to list branch protections", map[string]any{
			"owner":      owner,
			"repository": repositoryName,
			"error":      err.Error(),
		})

		return fmt.Errorf("failed to list branch protections: %w", err)
	}

	for _, protection := range protections {
		_, err := ps.client.DeleteBranchProtection(owner, repositoryName, protection.BranchName)
		if err != nil {
			ps.logger.Warn(ctx, "Failed to remove branch protection", map[string]any{
				"owner":      owner,
				"repository": repositoryName,
				"branch":     protection.BranchName,
				"error":      err.Error(),
			})
		}
	}

	ps.logger.Info(ctx, "Repository protection removed", map[string]any{
		"owner":      owner,
		"repository": repositoryName,
	})

	return nil
}

// enableBranchProtection enables branch protection with specified rules.
func (ps *ProtectionService) enableBranchProtection(ctx context.Context, owner, repositoryName string, rules []BranchProtectionRule) error {
	// If no rules specified, apply default protection
	if len(rules) == 0 {
		rules = ps.getDefaultBranchProtectionRules()
	}

	for _, rule := range rules {
		if err := ps.applyBranchProtectionRule(ctx, owner, repositoryName, rule); err != nil {
			return fmt.Errorf("failed to apply branch protection rule for %s: %w", rule.BranchName, err)
		}
	}

	return nil
}

// getDefaultBranchProtectionRules returns default branch protection rules.
func (ps *ProtectionService) getDefaultBranchProtectionRules() []BranchProtectionRule {
	return []BranchProtectionRule{
		{
			BranchName:            "main",
			EnablePush:            false,
			RequiredApprovals:     1,
			DismissStaleApprovals: true,
			RequireSignedCommits:  false,
		},
		{
			BranchName:            "master",
			EnablePush:            false,
			RequiredApprovals:     1,
			DismissStaleApprovals: true,
			RequireSignedCommits:  false,
		},
	}
}

// applyBranchProtectionRule applies a single branch protection rule.
func (ps *ProtectionService) applyBranchProtectionRule(ctx context.Context, owner, repositoryName string, rule BranchProtectionRule) error {
	opt := gitea.CreateBranchProtectionOption{
		BranchName:              rule.BranchName,
		EnablePush:              rule.EnablePush,
		EnablePushWhitelist:     rule.EnablePushWhitelist,
		PushWhitelistUsernames:  rule.PushWhitelistUsernames,
		PushWhitelistTeams:      rule.PushWhitelistTeams,
		EnableMergeWhitelist:    rule.EnableMergeWhitelist,
		MergeWhitelistUsernames: rule.MergeWhitelistUsernames,
		MergeWhitelistTeams:     rule.MergeWhitelistTeams,
		ProtectedFilePatterns:   rule.ProtectedFilePatterns,
		BlockOnRejectedReviews:  rule.BlockOnRejectedReviews,
		DismissStaleApprovals:   rule.DismissStaleApprovals,
		RequireSignedCommits:    rule.RequireSignedCommits,
		RequiredApprovals:       rule.RequiredApprovals,
		EnableStatusCheck:       rule.EnableStatusCheck,
		StatusCheckContexts:     rule.StatusCheckContexts,
	}

	// First, try to delete existing protection (if any)
	_, err := ps.client.DeleteBranchProtection(owner, repositoryName, rule.BranchName)
	if err != nil {
		ps.logger.Debug(ctx, "Branch was not previously protected", map[string]any{
			"owner":       owner,
			"repository":  repositoryName,
			"branch_name": rule.BranchName,
		})
	}

	// Create new protection
	_, _, err = ps.client.CreateBranchProtection(owner, repositoryName, opt)
	if err != nil {
		return fmt.Errorf("failed to create branch protection: %w", err)
	}

	ps.logger.Debug(ctx, "Branch protection applied", map[string]any{
		"owner":      owner,
		"repository": repositoryName,
		"branch":     rule.BranchName,
	})

	return nil
}
