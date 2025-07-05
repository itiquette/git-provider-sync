// SPDX-FileCopyrightText: 2024 itiquette/git-provider-sync
//
// SPDX-License-Identifier: EUPL-1.2

package gitlab

import (
	"context"
	"fmt"

	"gitlab.com/gitlab-org/api/client-go"

	"itiquette/git-provider-sync/internal/domain/ports"
)

// ProtectionService provides GitLab-specific repository protection operations.
// This restores the sophisticated protection service functionality from main branch.
type ProtectionService struct {
	client *gitlab.Client
	logger ports.Logger
}

// NewProtectionService creates a new GitLab protection service.
func NewProtectionService(client *gitlab.Client, logger ports.Logger) *ProtectionService {
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
	EnablePushRules        bool
	PushRules              *PushRules
}

// BranchProtectionRule defines protection rules for a branch in GitLab.
type BranchProtectionRule struct {
	BranchName                 string
	PushAccessLevel            string // "no one", "developer", "maintainer", "admin"
	MergeAccessLevel           string // "no one", "developer", "maintainer", "admin"
	UnprotectAccessLevel       string // "developer", "maintainer", "admin"
	AllowForcePush             bool
	CodeOwnerApprovalRequired  bool
	InheritedRequiredApprovals *int
	RequiredApprovals          *int
	UserID                     []int
	GroupID                    []int
	AllowedToPush              []ProtectionUser
	AllowedToMerge             []ProtectionUser
	AllowedToUnprotect         []ProtectionUser
}

// ProtectionUser represents a user or group with protection permissions.
type ProtectionUser struct {
	ID          int
	Username    string
	AccessLevel string
	UserType    string // "User", "Group"
}

// PushRules defines push rules for a GitLab project.
type PushRules struct {
	AuthorEmailRegex           string
	BranchNameRegex            string
	CommitMessageRegex         string
	CommitMessageNegativeRegex string
	FileNameRegex              string
	DenyDeleteTag              bool
	MemberCheck                bool
	PreventSecrets             bool
	CommitCommitterCheck       bool
	RejectUnsignedCommits      bool
	MaxFileSize                int
}

// ProtectRepository applies comprehensive protection to a GitLab repository.
func (ps *ProtectionService) ProtectRepository(ctx context.Context, request ProtectRepositoryRequest) error {
	ps.logger.Debug(ctx, "Protecting GitLab repository", map[string]interface{}{
		"owner":                    request.Owner,
		"repository":               request.RepositoryName,
		"enable_branch_protection": request.EnableBranchProtection,
		"enable_push_rules":        request.EnablePushRules,
	})

	// Get project ID
	projectPath := request.Owner + "/" + request.RepositoryName

	project, _, err := ps.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Enable branch protection
	if request.EnableBranchProtection {
		if err := ps.enableBranchProtection(ctx, project.ID, request.BranchProtectionRules); err != nil {
			return fmt.Errorf("failed to enable branch protection: %w", err)
		}
	}

	// Enable push rules
	if request.EnablePushRules && request.PushRules != nil {
		if err := ps.enablePushRules(ctx, project.ID, request.PushRules); err != nil {
			return fmt.Errorf("failed to enable push rules: %w", err)
		}
	}

	ps.logger.Info(ctx, "Repository protection applied successfully", map[string]interface{}{
		"owner":      request.Owner,
		"repository": request.RepositoryName,
		"project_id": project.ID,
	})

	return nil
}

// UnprotectRepository removes protection from a GitLab repository.
func (ps *ProtectionService) UnprotectRepository(ctx context.Context, owner, repositoryName string) error {
	ps.logger.Debug(ctx, "Unprotecting GitLab repository", map[string]interface{}{
		"owner":      owner,
		"repository": repositoryName,
	})

	// Get project ID
	projectPath := owner + "/" + repositoryName

	project, _, err := ps.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	// List and remove all branch protections
	protectedBranches, _, err := ps.client.ProtectedBranches.ListProtectedBranches(project.ID, nil)
	if err != nil {
		ps.logger.Warn(ctx, "Failed to list protected branches", map[string]interface{}{
			"owner":      owner,
			"repository": repositoryName,
			"project_id": project.ID,
			"error":      err.Error(),
		})

		return nil
	}

	for _, branch := range protectedBranches {
		_, err := ps.client.ProtectedBranches.UnprotectRepositoryBranches(project.ID, branch.Name)
		if err != nil {
			ps.logger.Warn(ctx, "Failed to unprotect branch", map[string]interface{}{
				"owner":      owner,
				"repository": repositoryName,
				"project_id": project.ID,
				"branch":     branch.Name,
				"error":      err.Error(),
			})
		}
	}

	// Remove push rules
	if err := ps.removePushRules(ctx, project.ID); err != nil {
		ps.logger.Warn(ctx, "Failed to remove push rules", map[string]interface{}{
			"owner":      owner,
			"repository": repositoryName,
			"project_id": project.ID,
			"error":      err.Error(),
		})
	}

	ps.logger.Info(ctx, "Repository protection removed", map[string]interface{}{
		"owner":      owner,
		"repository": repositoryName,
		"project_id": project.ID,
	})

	return nil
}

// enableBranchProtection enables branch protection with specified rules.
func (ps *ProtectionService) enableBranchProtection(ctx context.Context, projectID int, rules []BranchProtectionRule) error {
	// If no rules specified, apply default protection
	if len(rules) == 0 {
		rules = ps.getDefaultBranchProtectionRules()
	}

	for _, rule := range rules {
		if err := ps.applyBranchProtectionRule(ctx, projectID, rule); err != nil {
			return fmt.Errorf("failed to apply branch protection rule for %s: %w", rule.BranchName, err)
		}
	}

	return nil
}

// enablePushRules enables push rules for the project.
func (ps *ProtectionService) enablePushRules(ctx context.Context, projectID int, rules *PushRules) error {
	ps.logger.Debug(ctx, "Enabling push rules", map[string]interface{}{
		"project_id": projectID,
	})

	opts := &gitlab.AddProjectPushRuleOptions{
		AuthorEmailRegex:           &rules.AuthorEmailRegex,
		BranchNameRegex:            &rules.BranchNameRegex,
		CommitMessageRegex:         &rules.CommitMessageRegex,
		CommitMessageNegativeRegex: &rules.CommitMessageNegativeRegex,
		FileNameRegex:              &rules.FileNameRegex,
		DenyDeleteTag:              &rules.DenyDeleteTag,
		MemberCheck:                &rules.MemberCheck,
		PreventSecrets:             &rules.PreventSecrets,
		CommitCommitterCheck:       &rules.CommitCommitterCheck,
		RejectUnsignedCommits:      &rules.RejectUnsignedCommits,
	}

	if rules.MaxFileSize > 0 {
		opts.MaxFileSize = &rules.MaxFileSize
	}

	_, _, err := ps.client.Projects.AddProjectPushRule(projectID, opts)
	if err != nil {
		return fmt.Errorf("failed to add push rules: %w", err)
	}

	return nil
}

// removePushRules removes push rules from the project.
func (ps *ProtectionService) removePushRules(ctx context.Context, projectID int) error {
	_, err := ps.client.Projects.DeleteProjectPushRule(projectID)
	if err != nil && !isNotFoundError(err) {
		return fmt.Errorf("failed to remove push rules: %w", err)
	}

	return nil
}

// getDefaultBranchProtectionRules returns default branch protection rules.
func (ps *ProtectionService) getDefaultBranchProtectionRules() []BranchProtectionRule {
	return []BranchProtectionRule{
		{
			BranchName:                "main",
			PushAccessLevel:           "maintainer",
			MergeAccessLevel:          "developer",
			UnprotectAccessLevel:      "maintainer",
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		},
		{
			BranchName:                "master",
			PushAccessLevel:           "maintainer",
			MergeAccessLevel:          "developer",
			UnprotectAccessLevel:      "maintainer",
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		},
		{
			BranchName:                "develop",
			PushAccessLevel:           "developer",
			MergeAccessLevel:          "developer",
			UnprotectAccessLevel:      "maintainer",
			AllowForcePush:            false,
			CodeOwnerApprovalRequired: false,
		},
	}
}

// applyBranchProtectionRule applies a single branch protection rule.
func (ps *ProtectionService) applyBranchProtectionRule(ctx context.Context, projectID int, rule BranchProtectionRule) error {
	ps.logger.Debug(ctx, "Applying branch protection rule", map[string]interface{}{
		"project_id": projectID,
		"branch":     rule.BranchName,
	})

	// First, try to unprotect if already protected
	_, unprotectErr := ps.client.ProtectedBranches.UnprotectRepositoryBranches(projectID, rule.BranchName)
	if unprotectErr != nil {
		ps.logger.Debug(ctx, "Branch was not previously protected", map[string]interface{}{
			"project_id":  projectID,
			"branch_name": rule.BranchName,
		})
	}

	// Build protection options
	opts := &gitlab.ProtectRepositoryBranchesOptions{
		Name: &rule.BranchName,
	}

	// Set push access level
	pushAccessLevel := ps.convertAccessLevel(rule.PushAccessLevel)
	opts.PushAccessLevel = &pushAccessLevel

	// Set merge access level
	mergeAccessLevel := ps.convertAccessLevel(rule.MergeAccessLevel)
	opts.MergeAccessLevel = &mergeAccessLevel

	// Set unprotect access level
	unprotectAccessLevel := ps.convertAccessLevel(rule.UnprotectAccessLevel)
	opts.UnprotectAccessLevel = &unprotectAccessLevel

	// Set additional options
	opts.AllowForcePush = &rule.AllowForcePush
	opts.CodeOwnerApprovalRequired = &rule.CodeOwnerApprovalRequired

	// Apply protection
	_, _, err := ps.client.ProtectedBranches.ProtectRepositoryBranches(projectID, opts)
	if err != nil {
		return fmt.Errorf("failed to protect branch: %w", err)
	}

	ps.logger.Debug(ctx, "Branch protection applied", map[string]interface{}{
		"project_id": projectID,
		"branch":     rule.BranchName,
	})

	return nil
}

// convertAccessLevel converts string access level to GitLab access level.
func (ps *ProtectionService) convertAccessLevel(level string) gitlab.AccessLevelValue {
	switch level {
	case "no one", "none":
		return gitlab.NoPermissions
	case "developer":
		return gitlab.DeveloperPermissions
	case "maintainer", "master":
		return gitlab.MaintainerPermissions
	case "admin":
		return gitlab.OwnerPermissions
	default:
		return gitlab.DeveloperPermissions
	}
}

// isNotFoundError checks if an error is a 404 not found error.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	return fmt.Sprintf("%v", err) == "404 Not Found"
}
