// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
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

// AdvancedProtectionService provides sophisticated GitHub protection using Rulesets.
// This restores the advanced protection service functionality from main branch.
type AdvancedProtectionService struct {
	client *github.Client
	logger ports.Logger
}

// NewAdvancedProtectionService creates a new advanced GitHub protection service.
func NewAdvancedProtectionService(client *github.Client, logger ports.Logger) *AdvancedProtectionService {
	return &AdvancedProtectionService{
		client: client,
		logger: logger,
	}
}

// ProtectRepositoryAdvanced applies comprehensive protection using GitHub Rulesets.
func (aps *AdvancedProtectionService) ProtectRepositoryAdvanced(ctx context.Context, owner, projectName string) error {
	aps.logger.Debug(ctx, "Applying advanced protection to GitHub repository", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
	})

	// Disable GitHub Actions
	if err := aps.disableActions(ctx, owner, projectName); err != nil {
		aps.logger.Warn(ctx, "Failed to disable GitHub Actions", map[string]interface{}{
			"owner":   owner,
			"project": projectName,
			"error":   err.Error(),
		})
	}

	// Enable branch protection using rulesets
	if err := aps.enableBranchProtection(ctx, owner, projectName); err != nil {
		if !aps.isUpgradeRequiredError(err) {
			return fmt.Errorf("failed to enable branch protection: %w", err)
		}

		aps.logger.Info(ctx, "Branch protection requires GitHub Pro - continuing without", map[string]interface{}{
			"owner":   owner,
			"project": projectName,
		})
	}

	// Enable tag protection using rulesets
	if err := aps.enableTagProtection(ctx, owner, projectName); err != nil {
		if !aps.isUpgradeRequiredError(err) {
			return fmt.Errorf("failed to enable tag protection: %w", err)
		}

		aps.logger.Info(ctx, "Tag protection requires GitHub Pro - continuing without", map[string]interface{}{
			"owner":   owner,
			"project": projectName,
		})
	}

	aps.logger.Info(ctx, "Advanced repository protection applied successfully", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
	})

	return nil
}

// UnprotectRepositoryAdvanced removes all protections using rulesets.
func (aps *AdvancedProtectionService) UnprotectRepositoryAdvanced(ctx context.Context, owner, projectName string) error {
	aps.logger.Debug(ctx, "Removing advanced protection from GitHub repository", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
	})

	if err := aps.deleteAllRulesets(ctx, owner, projectName); err != nil {
		if !aps.isUpgradeRequiredError(err) {
			return fmt.Errorf("failed to remove repository protection: %w", err)
		}
	}

	aps.logger.Info(ctx, "Advanced repository protection removed", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
	})

	return nil
}

// disableActions disables GitHub Actions for the repository.
func (aps *AdvancedProtectionService) disableActions(ctx context.Context, owner, projectName string) error {
	permissions := &github.ActionsPermissionsRepository{
		Enabled: github.Ptr(false),
	}

	_, _, err := aps.client.Repositories.EditActionsPermissions(ctx, owner, projectName, *permissions)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("failed to disable GitHub Actions: %w", err)
	}

	return nil
}

// enableBranchProtection enables sophisticated branch protection using rulesets.
func (aps *AdvancedProtectionService) enableBranchProtection(ctx context.Context, owner, projectName string) error {
	aps.logger.Debug(ctx, "Enabling branch protection with rulesets", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
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

	_, _, err := aps.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		return fmt.Errorf("failed to create branch protection ruleset: %w", err)
	}

	return nil
}

// enableTagProtection enables sophisticated tag protection using rulesets.
func (aps *AdvancedProtectionService) enableTagProtection(ctx context.Context, owner, projectName string) error {
	aps.logger.Debug(ctx, "Enabling tag protection with rulesets", map[string]interface{}{
		"owner":   owner,
		"project": projectName,
	})

	ruleset := github.RepositoryRuleset{
		Name:        "TagProtectionRule",
		Target:      github.Ptr(github.RulesetTargetTag),
		Enforcement: "active",
		Rules: &github.RepositoryRulesetRules{
			// Restrict tag creation, updates, and deletion
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

	_, _, err := aps.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		return fmt.Errorf("failed to create tag protection ruleset: %w", err)
	}

	return nil
}

// deleteAllRulesets removes all rulesets from a repository.
func (aps *AdvancedProtectionService) deleteAllRulesets(ctx context.Context, owner, projectName string) error {
	// Get all rulesets
	rulesets, _, err := aps.client.Repositories.GetAllRulesets(ctx, owner, projectName, false)
	if err != nil {
		return fmt.Errorf("failed to list rulesets: %w", err)
	}

	// Delete each ruleset
	for _, ruleset := range rulesets {
		if err := aps.deleteRuleset(ctx, owner, projectName, *ruleset.ID); err != nil {
			return err
		}
	}

	return nil
}

// deleteRuleset deletes a specific ruleset.
func (aps *AdvancedProtectionService) deleteRuleset(ctx context.Context, owner, projectName string, rulesetID int64) error {
	_, err := aps.client.Repositories.DeleteRuleset(ctx, owner, projectName, rulesetID)
	if err != nil {
		return fmt.Errorf("failed to delete ruleset %d: %w", rulesetID, err)
	}

	aps.logger.Debug(ctx, "Deleted ruleset", map[string]interface{}{
		"owner":      owner,
		"project":    projectName,
		"ruleset_id": rulesetID,
	})

	return nil
}

// isUpgradeRequiredError checks if an error indicates a GitHub Pro upgrade is required.
func (aps *AdvancedProtectionService) isUpgradeRequiredError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	return strings.Contains(errStr, "403") && strings.Contains(errStr, "Upgrade to GitHub Pro")
}

// GetRulesetInfo returns information about existing rulesets.
func (aps *AdvancedProtectionService) GetRulesetInfo(ctx context.Context, owner, projectName string) ([]*github.RepositoryRuleset, error) {
	rulesets, _, err := aps.client.Repositories.GetAllRulesets(ctx, owner, projectName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get ruleset information: %w", err)
	}

	return rulesets, nil
}

// CreateCustomBranchRuleset creates a custom branch protection ruleset.
func (aps *AdvancedProtectionService) CreateCustomBranchRuleset(ctx context.Context, owner, projectName string, config BranchRulesetConfig) error {
	ruleset := github.RepositoryRuleset{
		Name:        config.Name,
		Target:      github.Ptr(github.RulesetTargetBranch),
		Enforcement: github.RulesetEnforcement(config.Enforcement),
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: config.IncludeBranches,
				Exclude: config.ExcludeBranches,
			},
		},
		Rules: &github.RepositoryRulesetRules{},
	}

	// Configure rules based on config
	if config.PreventCreation {
		ruleset.Rules.Creation = &github.EmptyRuleParameters{}
	}

	if config.PreventDeletion {
		ruleset.Rules.Deletion = &github.EmptyRuleParameters{}
	}

	if config.RequirePullRequest {
		ruleset.Rules.PullRequest = &github.PullRequestRuleParameters{}
	}

	_, _, err := aps.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		return fmt.Errorf("failed to create custom branch ruleset: %w", err)
	}

	return nil
}

// BranchRulesetConfig contains configuration for custom branch rulesets.
type BranchRulesetConfig struct {
	Name               string
	Enforcement        string // "active", "evaluate", "disabled"
	IncludeBranches    []string
	ExcludeBranches    []string
	PreventCreation    bool
	PreventDeletion    bool
	RequirePullRequest bool
}
