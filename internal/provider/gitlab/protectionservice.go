// SPDX-FileCopyrightText: 2024 Josef Andersson
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"fmt"
	"itiquette/git-provider-sync/internal/log"
	"strconv"
	"strings"

	"github.com/xanzy/go-gitlab"
)

type ProtectionService struct {
	client *gitlab.Client
}

func NewProtectionService(client *gitlab.Client) *ProtectionService {
	return &ProtectionService{client: client}
}

func (p ProtectionService) protect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:protect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("protect")

	projectID, _ := strconv.Atoi(projectIDStr)

	err := p.enableBranchProtection(ctx, branch, projectID)
	if err != nil {
		return fmt.Errorf("failed to enable branch protection: %w", err)
	}

	err = p.enableTagProtection(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to enable tag protection: %w", err)
	}

	return nil
}

func (p ProtectionService) enableTagProtection(ctx context.Context, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering enableTagProtection")
	logger.Debug().Int("projectID", projectID).Msg("enableTagProtection")

	if _, _, err := p.client.ProtectedTags.ProtectRepositoryTags(projectID, &gitlab.ProtectRepositoryTagsOptions{
		Name:              gitlab.Ptr("*"),
		CreateAccessLevel: gitlab.Ptr(gitlab.NoPermissions),
	}); err != nil {
		return fmt.Errorf("failed to enable tagprotection: %w", err)
	}

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, branch string, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:enableBranchProtection")
	logger.Debug().Int("projectID", projectID).Str("branch", branch).Msg("enableBranchProtection")

	if _, _, err := p.client.ProtectedBranches.ProtectRepositoryBranches(projectID, &gitlab.ProtectRepositoryBranchesOptions{
		Name:                      gitlab.Ptr("*"),
		PushAccessLevel:           gitlab.Ptr(gitlab.NoPermissions),
		MergeAccessLevel:          gitlab.Ptr(gitlab.NoPermissions),
		AllowForcePush:            gitlab.Ptr(false),
		CodeOwnerApprovalRequired: gitlab.Ptr(true),
	}); err != nil {
		return fmt.Errorf("failed to protect branches: %w", err)
	}

	if _, _, err := p.client.ProtectedBranches.ProtectRepositoryBranches(projectID, &gitlab.ProtectRepositoryBranchesOptions{
		Name:                      gitlab.Ptr(branch),
		PushAccessLevel:           gitlab.Ptr(gitlab.NoPermissions),
		MergeAccessLevel:          gitlab.Ptr(gitlab.NoPermissions),
		AllowForcePush:            gitlab.Ptr(false),
		CodeOwnerApprovalRequired: gitlab.Ptr(true),
	}); err != nil {
		return fmt.Errorf("failed to protect default branch: %w", err)
	}

	return nil
}

func (p ProtectionService) unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering unprotect")
	logger.Debug().Str("projectID", projectIDStr).Str("branch", branch).Msg("unprotect")

	projectID, _ := strconv.Atoi(projectIDStr)

	err := p.disableBranchProtection(ctx, branch, projectID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected branches: %w", err)
		}
	}

	err = p.disableTagProtection(ctx, projectID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected tags: %w", err)
		}
	}

	return nil
}

func (p ProtectionService) disableBranchProtection(ctx context.Context, branch string, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:disableBranchProtection")
	logger.Debug().Int("projectID", projectID).Msg("GitLab:disableBranchProtection")

	_, err := p.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, "*")
	if err != nil {
		return fmt.Errorf("failed to remove existing protection: %w", err)
	}

	_, err = p.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, branch)
	if err != nil {
		return fmt.Errorf("failed to remove existing protection default branch: %w", err)
	}

	return nil
}

func (p ProtectionService) disableTagProtection(ctx context.Context, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:removeAllTagProtection")
	logger.Debug().Int("projectID", projectID).Msg("GitLab:removeAllTagProtection")

	tags, _, err := p.client.ProtectedTags.ListProtectedTags(projectID, &gitlab.ListProtectedTagsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list protected tags: %w", err)
	}

	for _, tag := range tags {
		_, err := p.client.ProtectedTags.UnprotectRepositoryTags(projectID, tag.Name)
		if err != nil {
			return fmt.Errorf("failed to unprotect tag %s: %w", tag.Name, err)
		}
	}

	return nil
}
