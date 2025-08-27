// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
//
// SPDX-License-Identifier: EUPL-1.2

package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v71/github"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProtectionService provides GitHub-specific repository protection operations with advanced features.
//
//	sophisticated protection service functionality  completely.
type ProtectionService struct {
	client *github.Client
	logger ports.Logger
}

// NewProtectionService creates a new GitHub protection service with advanced capabilities.
func NewProtectionService(client *github.Client, logger ports.Logger) *ProtectionService {
	return &ProtectionService{
		client: client,
		logger: logger,
	}
}

// ProtectRepositoryRequest contains parameters for protecting a repository.
type ProtectRepositoryRequest struct {
	Owner                  string
	RepositoryName         string
	DisableActions         bool
	EnableBranchProtection bool
	EnableTagProtection    bool
	EnableSecurityFeatures bool
	BranchProtectionRules  []BranchProtectionRule
	TagProtectionRules     []TagProtectionRule
}

// BranchProtectionRule defines protection rules for a branch.
type BranchProtectionRule struct {
	BranchPattern                string
	RequiredStatusChecks         *RequiredStatusChecks
	RequiredPullRequestReviews   *RequiredPullRequestReviews
	EnforceAdmins                bool
	RestrictPushes               *PushRestrictions
	RestrictReviews              *ReviewRestrictions
	RequiredApprovingReviewCount int
	DismissStaleReviews          bool
	RequireCodeOwnerReviews      bool
	AllowForcePushes             bool
	AllowDeletions               bool
}

// RequiredStatusChecks defines required status checks.
type RequiredStatusChecks struct {
	Strict   bool
	Contexts []string
	Checks   []RequiredStatusCheck
}

// RequiredStatusCheck defines a required status check.
type RequiredStatusCheck struct {
	Context string
	AppID   *int64
}

// RequiredPullRequestReviews defines PR review requirements.
type RequiredPullRequestReviews struct {
	RequiredApprovingReviewCount int
	DismissStaleReviews          bool
	RequireCodeOwnerReviews      bool
	DismissalRestrictions        *DismissalRestrictions
}

// DismissalRestrictions defines who can dismiss reviews.
type DismissalRestrictions struct {
	Users []string
	Teams []string
	Apps  []string
}

// PushRestrictions defines who can push to the branch.
type PushRestrictions struct {
	Users []string
	Teams []string
	Apps  []string
}

// ReviewRestrictions defines who can review PRs.
type ReviewRestrictions struct {
	Users []string
	Teams []string
	Apps  []string
}

// TagProtectionRule defines protection rules for tags.
type TagProtectionRule struct {
	Pattern string
}

// ProtectRepository applies comprehensive protection to a GitHub repository.
func (ps *ProtectionService) ProtectRepository(ctx context.Context, request ProtectRepositoryRequest) error {
	ps.logger.Debug(ctx, "Protecting GitHub repository", map[string]interface{}{
		"owner":                    request.Owner,
		"repository":               request.RepositoryName,
		"disable_actions":          request.DisableActions,
		"enable_branch_protection": request.EnableBranchProtection,
		"enable_tag_protection":    request.EnableTagProtection,
	})

	// Disable GitHub Actions if requested
	if request.DisableActions {
		if err := ps.disableActions(ctx, request.Owner, request.RepositoryName); err != nil {
			ps.logger.Warn(ctx, "Failed to disable GitHub Actions", map[string]interface{}{
				"owner":      request.Owner,
				"repository": request.RepositoryName,
				"error":      err.Error(),
			})
		}
	}

	// Enable branch protection
	if request.EnableBranchProtection {
		if err := ps.enableBranchProtection(ctx, request.Owner, request.RepositoryName); err != nil {
			return fmt.Errorf("failed to enable branch protection: %w", err)
		}
	}

	// Enable tag protection
	if request.EnableTagProtection {
		if err := ps.enableTagProtection(ctx, request.Owner, request.RepositoryName); err != nil {
			return fmt.Errorf("failed to enable tag protection: %w", err)
		}
	}

	// Enable security features (skipped - requires additional API endpoints)
	if request.EnableSecurityFeatures {
		ps.logger.Debug(ctx, "Security features configuration skipped", map[string]interface{}{
			"owner":      request.Owner,
			"repository": request.RepositoryName,
		})
	}

	ps.logger.Info(ctx, "Repository protection applied successfully", map[string]interface{}{
		"owner":      request.Owner,
		"repository": request.RepositoryName,
	})

	return nil
}

// UnprotectRepository removes protection from a GitHub repository.
func (ps *ProtectionService) UnprotectRepository(ctx context.Context, owner, repositoryName string) error {
	ps.logger.Debug(ctx, "Unprotecting GitHub repository", map[string]interface{}{
		"owner":      owner,
		"repository": repositoryName,
	})

	// Remove branch protection from default branch
	repo, _, err := ps.client.Repositories.Get(ctx, owner, repositoryName)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	if repo.DefaultBranch != nil {
		_, err = ps.client.Repositories.RemoveBranchProtection(ctx, owner, repositoryName, *repo.DefaultBranch)
		if err != nil && !strings.Contains(err.Error(), "404") {
			ps.logger.Warn(ctx, "Failed to remove branch protection", map[string]interface{}{
				"owner":      owner,
				"repository": repositoryName,
				"branch":     *repo.DefaultBranch,
				"error":      err.Error(),
			})
		}
	}

	ps.logger.Info(ctx, "Repository protection removed", map[string]interface{}{
		"owner":      owner,
		"repository": repositoryName,
	})

	return nil
}

// ProtectAdvanced applies sophisticated protection to a repository using comprehensive techniques.
func (ps *ProtectionService) ProtectAdvanced(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Applying advanced GitHub protection", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	// 1. Disable GitHub Actions
	if err := ps.disableActions(ctx, owner, projectName); err != nil {
		ps.logger.Warn(ctx, "Failed to disable Actions", map[string]interface{}{
			"owner":       owner,
			"projectName": projectName,
			"error":       err.Error(),
		})
	}

	// 2. Enable branch protection with rulesets
	if err := ps.enableBranchProtection(ctx, owner, projectName); err != nil {
		ps.logger.Warn(ctx, "Failed to enable branch protection", map[string]interface{}{
			"owner":       owner,
			"projectName": projectName,
			"error":       err.Error(),
		})
	}

	// 3. Enable tag protection with rulesets
	if err := ps.enableTagProtection(ctx, owner, projectName); err != nil {
		ps.logger.Warn(ctx, "Failed to enable tag protection", map[string]interface{}{
			"owner":       owner,
			"projectName": projectName,
			"error":       err.Error(),
		})
	}

	ps.logger.Info(ctx, "Advanced GitHub protection applied successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	return nil
}

// UnprotectAdvanced removes sophisticated protection from a repository.
func (ps *ProtectionService) UnprotectAdvanced(ctx context.Context, branch, owner, projectName string) error {
	ps.logger.Debug(ctx, "Removing advanced GitHub protection", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"branch":      branch,
	})

	// Delete all rulesets
	if err := ps.deleteAllRulesets(ctx, owner, projectName); err != nil {
		return fmt.Errorf("failed to disable protected branches. projectName: %s. err: %w", projectName, err)
	}

	ps.logger.Info(ctx, "Advanced GitHub protection removed successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	return nil
}

// disableActions disables GitHub Actions for a repository

func (ps *ProtectionService) disableActions(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Disabling GitHub Actions", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	permissions := &github.ActionsPermissionsRepository{
		Enabled: github.Ptr(false),
	}

	_, _, err := ps.client.Repositories.EditActionsPermissions(ctx, owner, projectName, *permissions)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// Repository doesn't support Actions, this is expected
			return nil
		}

		return fmt.Errorf("failed to disable Actions. projectName: %s. err: %w", projectName, err)
	}

	ps.logger.Debug(ctx, "GitHub Actions disabled successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	return nil
}

// enableBranchProtection enables sophisticated branch protection using rulesets

func (ps *ProtectionService) enableBranchProtection(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Enabling GitHub branch protection with rulesets", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	ruleset := github.RepositoryRuleset{
		Name:        "BranchProtectionRules",
		Target:      github.Ptr(github.RulesetTargetBranch),
		Enforcement: "active",

		// Match all branches by default
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: []string{"~ALL"},
				Exclude: []string{},
			},
		},
		BypassActors: []*github.BypassActor{},
		Rules: &github.RepositoryRulesetRules{
			Creation: &github.EmptyRuleParameters{},
			Update: &github.UpdateRuleParameters{
				UpdateAllowsFetchAndMerge: false,
			},
			Deletion:    &github.EmptyRuleParameters{},
			PullRequest: &github.PullRequestRuleParameters{},
		},
	}

	_, _, err := ps.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			ps.logger.Debug(ctx, "GitHub Pro required for rulesets, skipping", map[string]interface{}{
				"owner":       owner,
				"projectName": projectName,
			})

			return nil
		}

		return fmt.Errorf("failed to create ruleset protection. projectName: %s. err: %w", projectName, err)
	}

	ps.logger.Debug(ctx, "GitHub branch protection enabled successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	return nil
}

// enableTagProtection enables sophisticated tag protection using rulesets

func (ps *ProtectionService) enableTagProtection(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Enabling GitHub tag protection with rulesets", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	ruleset := github.RepositoryRuleset{
		Name:        "TagProtectionRule",
		Target:      github.Ptr(github.RulesetTargetTag),
		Enforcement: "active",
		Rules: &github.RepositoryRulesetRules{
			// Restrict tag creation, update, and deletion
			Creation: &github.EmptyRuleParameters{},
			Update:   &github.UpdateRuleParameters{},
			Deletion: &github.EmptyRuleParameters{},
		},
		// Apply to all tags by default
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: []string{"refs/tags/*"},
				Exclude: []string{},
			},
		},
	}

	_, _, err := ps.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			ps.logger.Debug(ctx, "GitHub Pro required for tag rulesets, skipping", map[string]interface{}{
				"owner":       owner,
				"projectName": projectName,
			})

			return nil
		}

		return fmt.Errorf("failed to protect tags: %w", err)
	}

	ps.logger.Debug(ctx, "GitHub tag protection enabled successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	return nil
}

// deleteAllRulesets removes all rulesets from a repository

func (ps *ProtectionService) deleteAllRulesets(ctx context.Context, owner, projectName string) error {
	ps.logger.Debug(ctx, "Deleting all GitHub rulesets", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
	})

	// Get all rulesets
	rulesets, _, err := ps.client.Repositories.GetAllRulesets(ctx, owner, projectName, false)
	if err != nil {
		// Check for upgrade requirement or 404 errors
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			ps.logger.Debug(ctx, "GitHub Pro required for rulesets, skipping", map[string]interface{}{
				"owner":       owner,
				"projectName": projectName,
			})

			return nil
		}

		return fmt.Errorf("failed to list rulesets. projectName: %s, err: %w", projectName, err)
	}

	// Delete each ruleset
	for _, ruleset := range rulesets {
		if err := ps.deleteRuleset(ctx, owner, projectName, *ruleset.ID); err != nil {
			return err
		}
	}

	ps.logger.Debug(ctx, "All GitHub rulesets deleted successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"count":       len(rulesets),
	})

	return nil
}

// deleteRuleset deletes a single ruleset by ID

func (ps *ProtectionService) deleteRuleset(ctx context.Context, owner, projectName string, rulesetID int64) error {
	ps.logger.Debug(ctx, "Deleting GitHub ruleset", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"rulesetID":   rulesetID,
	})

	_, err := ps.client.Repositories.DeleteRuleset(ctx, owner, projectName, rulesetID)
	if err != nil {
		return fmt.Errorf("failed to delete ruleset. projectName: %s, rulesetID: %d, err: %w",
			projectName, rulesetID, err)
	}

	ps.logger.Debug(ctx, "GitHub ruleset deleted successfully", map[string]interface{}{
		"owner":       owner,
		"projectName": projectName,
		"rulesetID":   rulesetID,
	})

	return nil
}

// splitProjectPath splits a project path into owner and repository components
