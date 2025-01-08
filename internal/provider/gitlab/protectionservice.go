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

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type ProtectionService struct {
	client *gitlab.Client
}

func NewProtectionService(client *gitlab.Client) ProtectionService {
	return ProtectionService{client: client}
}

func (p ProtectionService) Protect(ctx context.Context, _ string, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:Protect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("GitLab:Protect")

	projectID, _ := strconv.Atoi(projectIDStr)

	err := p.enableBranchProtection(ctx, branch, projectID)
	if err != nil {
		return fmt.Errorf("failed to enable branch protection. err: %w", err)
	}

	err = p.enableTagProtection(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to enable tag protection. err: %w", err)
	}

	return nil
}

func (p ProtectionService) enableTagProtection(ctx context.Context, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:enableTagProtection")
	logger.Debug().Int("projectID", projectID).Msg("GitLab:enableTagProtection")

	if _, _, err := p.client.ProtectedTags.ProtectRepositoryTags(projectID, &gitlab.ProtectRepositoryTagsOptions{
		Name:              gitlab.Ptr("*"),
		CreateAccessLevel: gitlab.Ptr(gitlab.NoPermissions),
	}); err != nil {
		return fmt.Errorf("failed to enable tag protection. err: %w", err)
	}

	return nil
}

func (p ProtectionService) enableBranchProtection(ctx context.Context, branch string, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:enableBranchProtection")
	logger.Debug().Int("projectID", projectID).Str("branch", branch).Msg("GitLab:enableBranchProtection")

	if _, _, err := p.client.ProtectedBranches.ProtectRepositoryBranches(projectID, &gitlab.ProtectRepositoryBranchesOptions{
		Name:                      gitlab.Ptr("*"),
		PushAccessLevel:           gitlab.Ptr(gitlab.NoPermissions),
		MergeAccessLevel:          gitlab.Ptr(gitlab.NoPermissions),
		AllowForcePush:            gitlab.Ptr(false),
		CodeOwnerApprovalRequired: gitlab.Ptr(true),
	}); err != nil {
		if strings.Contains(err.Error(), "409") {
			logger.Trace().Str("branch", "*").Str("err", err.Error()).Msg("failed to protect branch, normal upon project creation due to branches being protected upon project creation")

			return nil
		}

		return fmt.Errorf("failed to protect branches. err: %w", err)
	}

	if _, _, err := p.client.ProtectedBranches.ProtectRepositoryBranches(projectID, &gitlab.ProtectRepositoryBranchesOptions{
		Name:                      gitlab.Ptr(branch),
		PushAccessLevel:           gitlab.Ptr(gitlab.NoPermissions),
		MergeAccessLevel:          gitlab.Ptr(gitlab.NoPermissions),
		AllowForcePush:            gitlab.Ptr(false),
		CodeOwnerApprovalRequired: gitlab.Ptr(true),
	}); err != nil {
		if strings.Contains(err.Error(), "409") {
			logger.Trace().Str("branch", branch).Str("err", err.Error()).Msg("failed to protect branch, normal upon project creation due to branches being protected upon project creation")

			return nil
		}

		return fmt.Errorf("failed to protect branches. err: %w", err)
	}

	return nil
}

func (p ProtectionService) Unprotect(ctx context.Context, branch string, projectIDStr string) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:unprotect")
	logger.Debug().Str("projectIDStr", projectIDStr).Str("branch", branch).Msg("GitLab:unprotect")

	projectID, _ := strconv.Atoi(projectIDStr)

	err := p.disableBranchProtection(ctx, branch, projectID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected branches. err: %w", err)
		}
	}

	err = p.disableTagProtection(ctx, projectID)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return fmt.Errorf("failed to disable protected tags. projectIDStr: %s. err: %w", projectIDStr, err)
		}
	}

	return nil
}

func (p ProtectionService) disableBranchProtection(ctx context.Context, defaultBranch string, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:disableBranchProtection")
	logger.Debug().Int("projectID", projectID).Msg("GitLab:disableBranchProtection")

	_, err := p.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, "*")
	if err != nil {
		return fmt.Errorf("failed to remove existing protection: %w", err)
	}

	_, err = p.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to remove existing protection default branch. err: %w", err)
	}

	return nil
}

func (p ProtectionService) disableTagProtection(ctx context.Context, projectID int) error {
	logger := log.Logger(ctx)
	logger.Trace().Msg("Entering GitLab:removeAllTagProtection")
	logger.Debug().Int("projectID", projectID).Msg("GitLab:removeAllTagProtection")

	tags, _, err := p.client.ProtectedTags.ListProtectedTags(projectID, &gitlab.ListProtectedTagsOptions{})
	if err != nil {
		return fmt.Errorf("failed to list protected tags. err: %w", err)
	}

	for _, tag := range tags {
		_, err := p.client.ProtectedTags.UnprotectRepositoryTags(projectID, tag.Name)
		if err != nil {
			return fmt.Errorf("failed to unprotect tags. tag: %s, %w", tag.Name, err)
		}
	}

	return nil
}
