// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2
package github

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strings"

	"github.com/google/go-github/v69/github"
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
		Enabled: github.Ptr(false),
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
func (p *ProtectionService) enableTagProtection(ctx context.Context, owner, projectName string) error {
	ruleset := github.RepositoryRuleset{
		Name:        "TagProtectionRule",
		Target:      github.Ptr(github.RulesetTargetTag),
		Enforcement: "active",
		Rules: &github.RepositoryRulesetRules{
			// github.NewTagNamePatternRule(&github.RulePatternParameters{
			// 	Operator: "starts_with", // Required operator field
			// 	Pattern:  *github.String("v")}),
			// Restrict tag creation
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

	_, _, err := p.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			return nil
		}

		return fmt.Errorf("failed to protect tags: %w", err)
	}

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:enableRulesetProtection")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Msg("enableRulesetProtection")

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

	_, _, err := p.client.Repositories.CreateRuleset(ctx, owner, projectName, ruleset)
	if err != nil {
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			return nil
		}

		return fmt.Errorf("failed to create ruleset protection. projectName: %s. err: %w", projectName, err)
	}

	return nil
}

func (p ProtectionService) unprotect(ctx context.Context, branch, owner, projectName string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitHub:unprotect")
	logger.Debug().Str("owner", owner).Str("projectName", projectName).Str("branch", branch).Msg("GitHub:unprotect")

	err := p.deleteAllRulesets(ctx, owner, projectName)
	if err != nil {
		return fmt.Errorf("failed to disable protected branches. projectName: %s. err: %w", projectName, err)
	}

	return nil
}
func (p *ProtectionService) deleteAllRulesets(ctx context.Context, owner, projectName string) error {
	// Get all rulesets
	rulesets, _, err := p.client.Repositories.GetAllRulesets(ctx, owner, projectName, false)
	if err != nil { // Check for upgrade requirement or 404 errors
		if strings.Contains(err.Error(), "403") && strings.Contains(err.Error(), "Upgrade to GitHub Pro") {
			// This is expected for non-Pro repositories, return nil to continue
			return nil
		}

		return fmt.Errorf("failed to list rulesets. projectName: %s, err: %w", projectName, err)
	}

	// Delete each ruleset
	for _, ruleset := range rulesets {
		if err := p.deleteRuleset(ctx, owner, projectName, *ruleset.ID); err != nil {
			return err
		}
	}

	return nil
}

func (p *ProtectionService) deleteRuleset(ctx context.Context, owner, projectName string, rulesetID int64) error {
	_, err := p.client.Repositories.DeleteRuleset(ctx, owner, projectName, rulesetID)
	if err != nil {
		return fmt.Errorf("failed to delete ruleset. projectName: %s, rulesetID: %d, err: %w",
			projectName, rulesetID, err)
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
